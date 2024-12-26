# cb2md (Codebase to Markdown)

A command-line tool that:

1. Prints an ASCII tree of a directory (skipping hidden files, directories, and anything matched by `.ignore`).
2. Optionally writes a **Markdown file** containing both the ASCII tree **and** file contents for all included files.

## Features

- Recursively scans a folder structure, producing an ASCII tree at the top.
- Skips:
    - **Hidden files/directories** (those starting with `.`).
    - Anything matched by your `.ignore` file (glob patterns).
    - A set of **built-in "skip content" patterns** (e.g. common images like `.png`, `.jpg`, `.gif`, lock files, etc.), which appear in the tree but **do not** appear in the “Full File List” section.
- If an **output file** is specified (e.g. `-o=tree.md`), the tool:
    - Overwrites that file if it exists.
    - **Skips** re-including the generated file in its own output (no recursion).
    - Wraps the ASCII tree in triple backticks (````` ``` `````), and then prints a “Full File List” of included files below, each in its own code block.

## Installation

1. **Clone** the repository:

```bash
git clone git@github.com:pekhota/cb2md.git
cd cb2md
```

2. **Build or Install**:

```bash
# Build a local binary named `cb2md`
go build -o cb2md

# OR install into your GOPATH/bin (adjust if you keep your modules elsewhere)
go install
```

3. **Run**:

```bash
./cb2md /path/to/directory
```

## Usage

```bash
./cb2md [OPTIONS] /path/to/directory
```

### Flags

- **`-ignore=.ignore`**  
  Path to a file containing **glob patterns** to skip entirely. Default is `.ignore`.
    - For instance, if `.ignore` has `*.log`, then any `.log` file won’t appear in **either** the tree or the file list.

- **`-o=tree.md`**  
  Output file.
    - If the file name ends with `.md`, the ASCII tree is wrapped in triple backticks.
    - The tool also prints a “Full File List” section for files that **aren’t** matched by skip-content patterns (like `.jpg`, `.png`, etc.).
    - This file is **skipped** from the scan to prevent recursion, and is **overwritten** if it exists.

### Examples

```bash
# Print ASCII tree to stdout (plain text only)
./cb2md ./my-project

# Print ASCII tree + file contents in Markdown to tree.md
./cb2md ./my-project -ignore=.ignore -o=tree.md
```

In the second example:

- The ASCII tree is generated and written to `tree.md` (wrapped in triple backticks).
- A “Full File List” follows, showing each included file path plus its contents in a code block.

## Built-in Skip-Content Patterns

By default, the tool has a **built-in set** of **“skip content”** patterns for common **image files** (`*.png`, `*.jpg`, `*.gif`, etc.) and lock files (`package-lock.json`, `composer.lock`). These files:

1. **Appear in the ASCII tree** (so you know they exist),
2. **But are omitted from the “Full File List”** to avoid dumping large/binary data.

If you want to include these files in the “Full File List,” remove or adjust this logic in the `skipContentPatterns` section of `main.go`.

## .ignore File

Each line in your `.ignore` file is a glob pattern. Anything matching these patterns is **excluded entirely** (not even in the ASCII tree). For example:

```
*.log
secret.txt
build/
```

- `*.log`  — skip all `.log` files.
- `secret.txt` — skip exactly that file.
- `build/` — skip everything under the `build` folder.

Lines starting with `#` are comments; empty lines are ignored.

## Contributing

Feel free to open issues or pull requests if you find any bugs or have suggestions for new features. This tool is designed to be easily customizable for your own patterns or filtering needs.

## License

This project is available under the [MIT License](./LICENSE). If you need a more permissive or restrictive license, you’re free to change it accordingly.
