package main

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// TestLoadIgnorePatterns verifies we correctly load patterns from a file.
func TestLoadIgnorePatterns(t *testing.T) {
	// Create a temporary directory for this test
	tmpDir := t.TempDir()
	ignoreFilePath := filepath.Join(tmpDir, ".ignore")

	// Write some patterns to .ignore
	content := []byte(`
# This is a comment
*.log
secret.txt

build/
`)
	if err := os.WriteFile(ignoreFilePath, content, 0o644); err != nil {
		t.Fatalf("Failed to write .ignore file: %v", err)
	}

	// Call loadIgnorePatterns
	patterns := loadIgnorePatterns(ignoreFilePath)
	expected := []string{"*.log", "secret.txt", "build/"}

	if !reflect.DeepEqual(patterns, expected) {
		t.Errorf("Got patterns %v, want %v", patterns, expected)
	}
}

// TestMatchesAnyPattern ensures that file paths match expected patterns.
func TestMatchesAnyPattern(t *testing.T) {
	patterns := []string{
		"*.log",
		"build/",
		"secret.txt",
	}

	tests := []struct {
		path   string
		expect bool
	}{
		{"error.log", true},
		{"info.LOG", false}, // case sensitive, won't match .LOG
		{"build/index.js", true},
		{"secret.txt", true},
		{"random.txt", false},
	}

	for _, tc := range tests {
		got := matchesAnyPattern(tc.path, patterns)
		if got != tc.expect {
			t.Errorf("matchesAnyPattern(%q) = %v, want %v", tc.path, got, tc.expect)
		}
	}
}

// TestGuessLanguage checks that we map certain file extensions to the right language.
func TestGuessLanguage(t *testing.T) {
	tests := []struct {
		filename string
		want     string
	}{
		{"main.go", "go"},
		{"script.js", "javascript"},
		{"styles.css", "css"},
		{"unknownfile.xyz", ""},
		{"Dockerfile", ""}, // won't match
		{"README.md", "markdown"},
	}

	for _, tc := range tests {
		got := guessLanguage(tc.filename)
		if got != tc.want {
			t.Errorf("guessLanguage(%q) = %q; want %q", tc.filename, got, tc.want)
		}
	}
}

// TestBuildTree checks we correctly build a Node tree, skipping hidden files and ignoring .ignore patterns.
func TestBuildTree(t *testing.T) {
	// Create a temp directory with some test files/folders
	root := t.TempDir()

	// Set up directory structure:
	//   root/
	//     .hiddenDir/
	//       hiddenFile.txt
	//     visibleDir/
	//       file.go
	//       file.log
	//     .hiddenFile
	//     included.txt
	//     .ignore (patterns: *.log)
	hiddenDir := filepath.Join(root, ".hiddenDir")
	os.Mkdir(hiddenDir, 0o755)
	if err := os.WriteFile(filepath.Join(hiddenDir, "hiddenFile.txt"), []byte("secret"), 0o644); err != nil {
		t.Fatal(err)
	}

	visibleDir := filepath.Join(root, "visibleDir")
	os.Mkdir(visibleDir, 0o755)
	if err := os.WriteFile(filepath.Join(visibleDir, "file.go"), []byte("package main"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(visibleDir, "file.log"), []byte("log content"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(root, ".hiddenFile"), []byte("cannot see me"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "included.txt"), []byte("Hello"), 0o644); err != nil {
		t.Fatal(err)
	}

	ignoreContent := []byte("*.log\n")
	if err := os.WriteFile(filepath.Join(root, ".ignore"), ignoreContent, 0o644); err != nil {
		t.Fatal(err)
	}

	// Now call buildTree
	ignorePatterns := loadIgnorePatterns(filepath.Join(root, ".ignore"))
	visited := make(map[string]bool)
	node, err := buildTree(root, root, ignorePatterns, visited)
	if err != nil {
		t.Fatalf("buildTree failed: %v", err)
	}

	// The ASCII tree root node should be the top-level directory name
	if node.Name != filepath.Base(root) {
		t.Errorf("Root node name = %q; want %q", node.Name, filepath.Base(root))
	}

	// Check children
	// We expect visibleDir/ and included.txt to appear.
	// .hiddenDir and .hiddenFile are skipped because they start with '.'
	// file.log is skipped by .ignore pattern
	foundVisibleDir := false
	foundIncludedTxt := false

	// Because node is a directory, its Children correspond to those that weren't ignored
	for _, child := range node.Children {
		switch child.Name {
		case "visibleDir":
			foundVisibleDir = true
			// Within visibleDir, we should only see file.go, not file.log
			foundGo := false
			for _, subChild := range child.Children {
				if subChild.Name == "file.go" {
					foundGo = true
				}
				if subChild.Name == "file.log" {
					t.Error("Unexpectedly found file.log; it should be ignored by pattern *.log")
				}
			}
			if !foundGo {
				t.Error("Expected to find file.go inside visibleDir")
			}

		case "included.txt":
			foundIncludedTxt = true
		}
	}

	if !foundVisibleDir {
		t.Error("Expected to find visibleDir in root node children")
	}
	if !foundIncludedTxt {
		t.Error("Expected to find included.txt in root node children")
	}
}

// TestPrintFileContents (Optional):
// If you'd like to test file printing, you can capture the output by writing
// to a bytes.Buffer and verifying the contents. For example:
func TestPrintFileContents(t *testing.T) {
	tmpFile := t.TempDir() + "/hello.txt"
	if err := os.WriteFile(tmpFile, []byte("Hello\nWorld"), 0o644); err != nil {
		t.Fatal(err)
	}
	var buf strings.Builder
	if err := printFileContents(tmpFile, &buf); err != nil {
		t.Errorf("printFileContents failed: %v", err)
	}
	want := "Hello\nWorld\n"
	got := buf.String()
	if got != want {
		t.Errorf("printFileContents got:\n%q\nwant:\n%q", got, want)
	}
}
