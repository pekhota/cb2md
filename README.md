# cb2md (codebase to markdown) - ASCII Tree + Full File Contents Generator

A command-line tool that:

1. Prints an ASCII tree of a directory (skipping hidden and ignored files).
2. Optionally writes a Markdown file showing both the ASCII tree and the contents of every included file.

## Features

- Recursively scans a folder structure.
- Skips hidden files/directories and those matched by an `.ignore` file.
- Outputs an ASCII tree at the top.
- Optionally shows file contents in Markdown code blocks if the output file ends with `.md`.

## Installation

1. **Clone** the repository:

```bash
git clone git@github.com:pekhota/cb2md.git
cd cb2md
```

2. **Build or Install**:

```bash
# Build a local binary named `ascii-tree`
go build -o ascii-tree

# OR install into your GOPATH/bin
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

- `-ignore=.ignore` — A file containing glob patterns to skip (default: `.ignore`).
- `-o=tree.md` — Output file. If it ends with `.md`, will also include file contents in Markdown.

### Examples

```bash
# Print ASCII tree to stdout
./cb2md ./my-project

# Print ASCII tree + file contents in Markdown to tree.md
./cb2md ./my-project -ignore=.ignore -o=tree.md
```

## .ignore File

- Each line is a glob pattern (e.g., `*.log`, `secret.txt`, or `build/`).
- Lines starting with `#` are treated as comments.
- Empty lines are skipped.

Example `.ignore`:

```
*.log
secret.txt
build/
```

## Contributing

Feel free to open issues or pull requests if you find any bugs, or have suggestions for new features.

## License

This project is available under the [MIT License](./LICENSE).
