package main

import (
	"log"
	"os"
	"path/filepath"
	"strings"
)

func visitNode(node *FileNode, prefix string, removeHidden bool) error {
	// Read directory contents
	entries, err := os.ReadDir(node.path)
	if err != nil {
		log.Printf("Error reading directory %s: %v", node.path, err)
		return err
	}

	// Process each entry in the directory
	for i, entry := range entries {
		if removeHidden && strings.HasPrefix(entry.Name(), ".") {
			continue
		}

		isLast := i == len(entries)-1

		// Create child node
		childPath := filepath.Join(node.path, entry.Name())
		childNode := &FileNode{
			name:     entry.Name(),
			path:     childPath,
			isDir:    entry.IsDir(),
			prefix:   buildPrefix(prefix, isLast),
			expanded: false,
		}

		// Add child to parent's children
		node.children = append(node.children, childNode)

		// If it's a directory, recursively visit it
		if childNode.isDir {
			// Calculate new prefix for children of this directory
			newPrefix := prefix
			if isLast {
				newPrefix += "    " // 4 spaces for alignment when last item
			} else {
				newPrefix += "│   " // vertical line + 3 spaces for non-last items
			}

			if err := visitNode(childNode, newPrefix, removeHidden); err != nil {
				log.Printf("Error visiting directory %s: %v", childPath, err)
				return err
			}
		}
	}

	return nil
}

// buildPrefix creates the proper prefix for tree visualization.
func buildPrefix(parentPrefix string, isLast bool) string {
	if isLast {
		return parentPrefix + "└── "
	}
	return parentPrefix + "├── "
}

// flattenNode returns a slice of nodes in display order.
func (node *FileNode) flatten() []*FileNode {
	result := []*FileNode{node} // Start with current node

	if node.isDir && node.expanded {
		for _, child := range node.children {
			// Recursively flatten each child and append to result
			result = append(result, child.flatten()...)
		}
	}

	return result
}
