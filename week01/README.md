# Week 01 Tree CLI

This folder contains a small Go CLI that prints a tree view of files and directories.

## What it does

- Accepts an optional directory argument.
- Prints the directory structure using `├──` and `└──` connectors.
- Recursively walks subdirectories with `os` and `path/filepath`.

## Requirements

- Go 1.26.3 or newer

## Build

From inside `week01/`:

```bash
go build -o tree-cli
```

## Run

Print the current directory:

```bash
go run .
```

Print a specific directory:

```bash
go run . /path/to/project
```

If you built the binary:

```bash
./tree-cli /path/to/project
```

## Notes

- If no directory is provided, the CLI defaults to `.`.
- The output is sorted alphabetically for stable results.