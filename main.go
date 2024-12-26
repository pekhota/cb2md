package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Node represents a file or directory in our tree structure.
type Node struct {
	Name     string
	IsDir    bool
	Children []*Node
}

// includedFiles holds only those files we want to show in the “Full File List” section (i.e., we print their contents).
var includedFiles []string

// skipContentPatterns: these file types appear in the ASCII tree
// but will NOT be added to includedFiles, so we do not print their contents.
// Feel free to expand this list with other binary file types.
var skipContentPatterns = []string{
	"*.png", "*.jpg", "*.jpeg", "*.gif", "*.svg", "*.webp",
	"package-lock.json", "composer.lock",
}

func main() {
	var ignoreFile string
	var outFile string

	flag.StringVar(&ignoreFile, "ignore", ".ignore", "Path to ignore file (glob patterns). Default is .ignore")
	flag.StringVar(&outFile, "o", "", "Output file path (if empty, prints to stdout). If ends with .md, we also print file contents.")
	flag.Parse()

	if flag.NArg() < 1 {
		log.Fatalf("Usage: go run main.go [-ignore=.ignore] [-o=tree.md] /path/to/directory")
	}

	// Root directory to scan
	rootDir := flag.Arg(0)

	// Convert to absolute path
	absRoot, err := filepath.Abs(rootDir)
	if err != nil {
		log.Fatalf("Error getting absolute path: %v\n", err)
	}

	// Load ignore patterns (if any) from the .ignore file
	ignorePatterns := loadIgnorePatterns(filepath.Join(absRoot, ignoreFile))

	// Maintain a map of visited directories (real paths) to prevent loops
	visited := make(map[string]bool)

	// Build in-memory tree
	rootNode, err := buildTree(absRoot, absRoot, ignorePatterns, visited)
	if err != nil {
		log.Fatalf("Error building tree: %v\n", err)
	}

	// Sort the list of included files so we have a predictable order
	sort.Strings(includedFiles)

	// Determine output destination (stdout or file)
	var w io.Writer = os.Stdout
	if outFile != "" {
		f, err := os.Create(outFile)
		if err != nil {
			log.Fatalf("Error creating output file '%s': %v", outFile, err)
		}
		defer f.Close()
		w = f
	}

	// Check if we need Markdown fences for the ASCII tree
	outIsMarkdown := strings.HasSuffix(strings.ToLower(outFile), ".md") || outFile == ""

	// Print the ASCII tree
	if outIsMarkdown && outFile != "" {
		// If user specifically gave a .md outFile, wrap the tree in triple backticks
		fmt.Fprintln(w, "```")
		printTree(rootNode, "", true, w)
		fmt.Fprintln(w, "```")
	} else {
		// Otherwise just print the ASCII tree as plain text
		printTree(rootNode, "", true, w)
	}

	// If it's Markdown, we also print each file’s path + contents
	if outIsMarkdown {
		// A heading for file list
		fmt.Fprintln(w)
		fmt.Fprintln(w, "## Full File List")
		fmt.Fprintln(w)

		for _, fpath := range includedFiles {
			// Print the file’s path
			fmt.Fprintf(w, "### %s\n", fpath)

			// Determine language for code block
			language := guessLanguage(fpath)
			fmt.Fprintf(w, "```%s\n", language)

			// Print file contents
			err := printFileContents(filepath.Join(absRoot, fpath), w)
			if err != nil {
				fmt.Fprintf(w, "Error reading file: %v\n", err)
			}
			fmt.Fprintln(w, "```")
			fmt.Fprintln(w)
		}
	}
}

// buildTree recursively walks directories to build a tree of Nodes.
// Also populates "includedFiles" for any files we keep.
func buildTree(basePath, currentPath string, ignorePatterns []string, visited map[string]bool) (*Node, error) {
	// Resolve symbolic links to prevent infinite loops
	realPath, err := filepath.EvalSymlinks(currentPath)
	if err != nil {
		return nil, err
	}

	// If we've already seen this real path, skip
	if visited[realPath] {
		return nil, nil
	}
	visited[realPath] = true

	info, err := os.Stat(currentPath)
	if err != nil {
		return nil, err
	}

	node := &Node{
		Name:  info.Name(),
		IsDir: info.IsDir(),
	}

	if info.IsDir() {
		// If it's a directory, read its contents
		entries, err := os.ReadDir(currentPath)
		if err != nil {
			return nil, err
		}

		for _, e := range entries {
			name := e.Name()

			// Skip hidden (files/folders starting with ".")
			if strings.HasPrefix(name, ".") {
				continue
			}

			childPath := filepath.Join(currentPath, name)
			relPath, err := filepath.Rel(basePath, childPath)
			if err != nil {
				return nil, err
			}

			// Check .ignore patterns to see if we skip it entirely
			if matchesAnyPattern(relPath, ignorePatterns) {
				continue
			}

			childNode, err := buildTree(basePath, childPath, ignorePatterns, visited)
			if err != nil {
				return nil, err
			}
			// childNode could be nil if recursion returned early (e.g. repeated real path).
			if childNode != nil {
				node.Children = append(node.Children, childNode)
			}
		}

		// Sort children so the output is predictable
		sort.Slice(node.Children, func(i, j int) bool {
			return node.Children[i].Name < node.Children[j].Name
		})

	} else {
		// It's a file, so let's check if we skip content
		relPath, err := filepath.Rel(basePath, currentPath)
		if err != nil {
			return nil, err
		}

		// If this file doesn't match skipContentPatterns, we'll include it in the Full File List
		if !matchesAnyPattern(relPath, skipContentPatterns) {
			includedFiles = append(includedFiles, relPath)
		}
	}

	return node, nil
}

// printTree prints a Node (directory or file) in ASCII tree format.
func printTree(node *Node, prefix string, isLast bool, w io.Writer) {
	connector := "├── "
	if isLast {
		connector = "└── "
	}

	// Print this node
	fmt.Fprintln(w, prefix+connector+node.Name)

	if node.IsDir {
		// Prepare prefix for children
		var childPrefix string
		if isLast {
			childPrefix = prefix + "    "
		} else {
			childPrefix = prefix + "│   "
		}

		for i, child := range node.Children {
			last := (i == len(node.Children)-1)
			printTree(child, childPrefix, last, w)
		}
	}
}

// loadIgnorePatterns reads lines from the ignore file and returns them as patterns.
func loadIgnorePatterns(ignorePath string) []string {
	var patterns []string
	f, err := os.Open(ignorePath)
	if err != nil {
		// If not found, no patterns
		return patterns
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		patterns = append(patterns, line)
	}
	return patterns
}

// matchesAnyPattern checks if relPath matches any pattern in patterns.
func matchesAnyPattern(relPath string, patterns []string) bool {
	for _, p := range patterns {
		match, err := filepath.Match(p, relPath)
		if err == nil && match {
			return true
		}
	}
	return false
}

// guessLanguage attempts to guess a code block language from the file extension.
func guessLanguage(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	languageMap := map[string]string{
		".go":   "go",
		".py":   "python",
		".js":   "javascript",
		".jsx":  "jsx",
		".ts":   "typescript",
		".tsx":  "tsx",
		".html": "html",
		".css":  "css",
		".scss": "scss",
		".java": "java",
		".rs":   "rust",
		".sh":   "bash",
		".rb":   "ruby",
		".php":  "php",
		".yaml": "yaml",
		".yml":  "yaml",
		".json": "json",
		".md":   "markdown",
	}
	if lang, ok := languageMap[ext]; ok {
		return lang
	}
	return "" // unknown
}

// printFileContents prints the contents of a file to the given writer.
func printFileContents(path string, w io.Writer) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		fmt.Fprintln(w, scanner.Text())
	}
	return scanner.Err()
}
