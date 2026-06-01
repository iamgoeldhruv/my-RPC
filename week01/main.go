package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

func printTree(path string, prefix string) error {
	entries, err := os.ReadDir(path)
	if err != nil {
		return err
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})

	for i, entry := range entries {
		isLast := i == len(entries)-1
		connector := "├── "
		if isLast {
			connector = "└── "
		}

		fmt.Println(prefix + connector + entry.Name())

		if entry.IsDir() {
			newPrefix := prefix + "│   "
			if isLast {
				newPrefix = prefix + "    "
			}

			fullPath := filepath.Join(path, entry.Name())
			if err := printTree(fullPath, newPrefix); err != nil {
				return err
			}
		}
	}

	return nil
}

func main() {
	targetDir := "."
	if len(os.Args) > 2 {
		fmt.Println("Usage: tree [directory]")
		os.Exit(1)
	}
	if len(os.Args) == 2 {
		targetDir = os.Args[1]
	}

	info, err := os.Stat(targetDir)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	if !info.IsDir() {
		fmt.Printf("Error: %s is not a directory\n", targetDir)
		os.Exit(1)
	}

	fmt.Println(filepath.Base(targetDir))
	if err := printTree(targetDir, ""); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}