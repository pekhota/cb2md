// main_test.go
package main

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// TestLoadIgnorePatterns ensures patterns are loaded from .ignore correctly.
func TestLoadIgnorePatterns(t *testing.T) {
	tmpDir := t.TempDir()
	ignoreFile := filepath.Join(tmpDir, ".ignore")

	content := []byte(`
# This is a comment
*.log
secret.txt

build/
`)
	if err := os.WriteFile(ignoreFile, content, 0o644); err != nil {
		t.Fatalf("Failed to write .ignore: %v", err)
	}

	patterns := loadIgnorePatterns(ignoreFile)
	want := []string{"*.log", "secret.txt", "build/"}
	if !reflect.DeepEqual(patterns, want) {
		t.Errorf("loadIgnorePatterns got %v, want %v", patterns, want)
	}
}

// TestMatchesAnySkipContent checks we do case-insensitive filename-only match.
func TestMatchesAnySkipContent(t *testing.T) {
	patterns := []string{
		"*.png", "*.jpg", "*.jpeg", "*.gif", "*.svg", "*.webp",
		"package-lock.json", "composer.lock",
	}
	tests := []struct {
		relPath string
		want    bool
	}{
		{"photo.JPG", true}, // baseName=photo.jpg, match *.jpg
		{"photo.jpeg", true},
		{"photo.JPEg", true},
		{"image.png", true},
		{"image.PNG", true},
		{"composer.lock", true},
		{"COMPOSER.LOCK", true}, // baseName=composer.lock
		{"package-lock.JSON", true},
		{"photo.txt", false},
		{"photo.jpg.bak", false}, // baseName=photo.jpg.bak; doesn't match *.jpg
	}
	for _, tt := range tests {
		got := matchesAnySkipContent(tt.relPath, patterns)
		if got != tt.want {
			t.Errorf("matchesAnySkipContent(%q) = %v; want %v", tt.relPath, got, tt.want)
		}
	}
}

// TestGuessLanguage verifies extension-to-language mapping.
func TestGuessLanguage(t *testing.T) {
	tests := []struct {
		filename string
		want     string
	}{
		{"main.go", "go"},
		{"script.js", "javascript"},
		{"script.JS", "javascript"}, // checks case-insensitivity of extension
		{"styles.css", "css"},
		{"readme.md", "markdown"},
		{"data.json", "json"},
		{"unknownfile.xyz", ""},
	}
	for _, tt := range tests {
		got := guessLanguage(tt.filename)
		if got != tt.want {
			t.Errorf("guessLanguage(%q) = %q; want %q", tt.filename, got, tt.want)
		}
	}
}

// TestPrintFileContents ensures file contents are printed as expected.
func TestPrintFileContents(t *testing.T) {
	tmp := t.TempDir()
	filePath := filepath.Join(tmp, "hello.txt")
	content := "Hello\nWorld"
	if err := os.WriteFile(filePath, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	var buf strings.Builder
	if err := printFileContents(filePath, &buf); err != nil {
		t.Errorf("printFileContents error: %v", err)
	}
	got := buf.String()
	want := "Hello\nWorld\n"
	if got != want {
		t.Errorf("printFileContents got:\n%q\nwant:\n%q", got, want)
	}
}
