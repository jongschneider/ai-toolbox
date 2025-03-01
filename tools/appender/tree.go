package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type FileNode struct {
	name     string // name represents the name of the file or directory
	path     string // path represents the full path of the file or directory
	isDir    bool   // isDir is used to identify directories
	isRoot   bool   // isRoot is only used to identify the root node.
	expanded bool   // expanded is used to show/hide the children of a directory
	selected bool
	prefix   string      // prefix is used in the View method to draw the tree structure
	children []*FileNode // includes directories and files
}

func (node *FileNode) String() string {
	dirIndicator := ""
	if node.isDir && !node.isRoot {
		if node.expanded {
			dirIndicator = " "
		} else {
			dirIndicator = " "
		}
	}

	selected := ""
	if node.selected {
		selected = "  "
	}

	return fmt.Sprintf("%s%s%s%s", node.prefix, dirIndicator, node.name, selected)
}

func visitNode(
	node *FileNode,
	prefix string,
	removeHidden bool,
	nodeMap map[string]*FileNode,
) error {
	// Read directory contents
	entries, err := os.ReadDir(node.path)
	if err != nil {
		log.Printf("Error reading directory %s: %v", node.path, err)
		return err
	}

	// Process each entry in the directory
	for i, entry := range entries {
		// if removeHidden && strings.HasPrefix(entry.Name(), ".") {
		// 	continue
		// }

		isLast := i == len(entries)-1

		// Create child node
		childPath := filepath.Join(node.path, entry.Name())
		childNode, found := nodeMap[childPath]
		if !found {
			childNode = &FileNode{
				name:     entry.Name(),
				path:     childPath,
				isDir:    entry.IsDir(),
				prefix:   buildPrefix(prefix, isLast),
				expanded: false,
			}
		}

		// Add child to parent's children
		node.children = append(node.children, childNode)
		nodeMap[childPath] = childNode
		// If it's a directory, recursively visit it
		if childNode.isDir {
			// Calculate new prefix for children of this directory
			newPrefix := prefix
			if isLast {
				newPrefix += "    " // 4 spaces for alignment when last item
			} else {
				newPrefix += "│   " // vertical line + 3 spaces for non-last items
			}

			if err := visitNode(childNode, newPrefix, removeHidden, nodeMap); err != nil {
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

type FilterFunc func(node *FileNode) bool

func FilterHidden(node *FileNode) bool {
	isHidden := strings.HasPrefix(node.name, ".") && node.name != "."
	return isHidden
}

func include(node *FileNode, filters ...FilterFunc) bool {
	for _, filter := range filters {
		if filter(node) {
			return false
		}
	}
	return true
}

// flattenNode returns a slice of nodes in display order.
func (node *FileNode) flatten(nodeMap map[string]*FileNode, filters ...FilterFunc) []*FileNode {
	result := []*FileNode{}
	stateNode, ok := nodeMap[node.path]
	if !ok {
		nodeMap[node.path] = node
		stateNode = node
	}
	if !include(stateNode, filters...) {
		return result
	}

	result = append(result, stateNode)
	if stateNode.isDir && stateNode.expanded {
		for _, child := range stateNode.children {
			cn, ok := nodeMap[child.path]
			if !ok {
				nodeMap[child.path] = node
				cn = child
			}

			if !include(cn, filters...) {
				continue
			}

			// Recursively flatten each child and append to result
			result = append(result, cn.flatten(nodeMap, filters...)...)
		}
	}

	return result
}
