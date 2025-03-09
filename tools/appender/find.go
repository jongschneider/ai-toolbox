package main

import (
	"log/slog"
	"path/filepath"
	"strings"

	doublestar "github.com/bmatcuk/doublestar/v4"
	"github.com/charmbracelet/bubbles/textarea"
)

func initFindInput() textarea.Model {
	ti := textarea.New()
	ti.Placeholder = "Enter glob pattern..."
	ti.ShowLineNumbers = false
	ti.SetHeight(1)
	ti.CharLimit = 255
	// Set initial value to empty string explicitly
	ti.SetValue("")
	return ti
}

// performFind executes a glob pattern search and stores matches.
func (m *model) performFind() {
	pattern := m.findPattern.Value()
	if pattern == "" {
		m.matchedNodes = nil
		m.currentMatchIdx = -1
		return
	}

	// Create filters for matching
	filters := make([]FilterFunc, 0)
	if m.removeHidden {
		filters = append(filters, FilterHidden)
	}
	filters = append(filters, FilterBinary)

	// Instead of using doublestar.FilepathGlob directly,
	// we'll walk the tree structure we already have and match against the pattern
	searchPattern := filepath.Join(m.workDir, pattern)

	slog.Info("Searching for pattern", "pattern", searchPattern, "workDir", m.workDir)

	// Reset matches
	m.matchedNodes = make([]*FileNode, 0)

	// Use our existing tree structure to find matches
	m.findMatchesInNode(m.rootNode, searchPattern, filters)

	// Reset current match index
	if len(m.matchedNodes) > 0 {
		m.currentMatchIdx = 0
		// Navigate to first match
		m.navigateToMatch(0)
	} else {
		m.currentMatchIdx = -1
	}

	// Refresh the tree view
	m.flattenTree()
}

// findMatchesInNode recursively finds nodes that match the pattern
func (m *model) findMatchesInNode(node *FileNode, pattern string, filters []FilterFunc) {
	// Skip this node if it doesn't pass the filters
	if !include(node, filters...) {
		return
	}

	// Check if this node matches the pattern
	matched, err := doublestar.PathMatch(pattern, node.path)
	if err == nil && matched {
		m.matchedNodes = append(m.matchedNodes, node)
		// Ensure parent directories are expanded to show match
		m.ensureNodeVisible(node)
	}

	// Skip recursion if this is a directory that's filtered out (like .git)
	if node.isDir {
		// Only recurse into children if we're at the root (always check root)
		// or if this directory passes all filters
		for _, child := range node.children {
			m.findMatchesInNode(child, pattern, filters)
		}
	}
}

// ensureNodeVisible expands all parent directories of a node.
func (m *model) ensureNodeVisible(node *FileNode) {
	// Get the parent directory path
	parentPath := filepath.Dir(node.path)

	// If we're already at the root, no need to continue
	if parentPath == m.workDir || parentPath == "." || parentPath == node.path {
		return
	}

	// Find and expand the parent node
	if parentNode, ok := m.nodeLookup[parentPath]; ok {
		parentNode.expanded = true

		// Recursively ensure parent's parent is expanded
		m.ensureNodeVisible(parentNode)
	}
}

// navigateToMatch moves the cursor to the specified match index.
func (m *model) navigateToMatch(idx int) {
	if idx < 0 || idx >= len(m.matchedNodes) {
		return
	}

	// Update current match index
	m.currentMatchIdx = idx

	// Find the flat index of the matched node
	matchedNode := m.matchedNodes[idx]
	for i, node := range m.flatNodes {
		if node.path == matchedNode.path {
			// Set cursor to this node
			m.cursor = i

			// Ensure the node is visible in viewport
			m.ensureNodeInViewport()
			break
		}
	}
}

// ensureNodeInViewport adjusts the offset to make sure the current cursor is visible.
func (m *model) ensureNodeInViewport() {
	helpLines := len(strings.Split(helpMsg, "\n"))
	maxVisibleNodes := m.windowSize.height - helpLines - 2

	// If cursor is below the viewport, adjust offset
	if m.cursor >= m.offset+maxVisibleNodes {
		m.offset = m.cursor - maxVisibleNodes + 1
	}

	// If cursor is above the viewport, adjust offset
	if m.cursor < m.offset {
		m.offset = m.cursor
	}
}

// nextMatch navigates to the next match.
func (m *model) nextMatch() {
	if len(m.matchedNodes) == 0 {
		return
	}

	nextIdx := (m.currentMatchIdx + 1) % len(m.matchedNodes)
	m.navigateToMatch(nextIdx)
}

// prevMatch navigates to the previous match.
func (m *model) prevMatch() {
	if len(m.matchedNodes) == 0 {
		return
	}

	prevIdx := m.currentMatchIdx - 1
	if prevIdx < 0 {
		prevIdx = len(m.matchedNodes) - 1
	}
	m.navigateToMatch(prevIdx)
}
