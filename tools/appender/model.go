//nolint:forbidigo
package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
)

type model struct {
	workDir            string
	rootNode           *FileNode
	cursor             int
	flatNodes          []*FileNode
	nodeLookup         map[string]*FileNode
	windowSize         windowSize // Number of items to show at once
	offset             int        // Starting index for the window
	renderer           *glamour.TermRenderer
	removeHidden       bool
	leftViewport       viewport.Model
	rightViewport      viewport.Model
	showClipboardModal bool
	clipboardError     error
	showSaveModal      bool
	outputPath         textarea.Model
	keys               keyMap
	help               help.Model
	// Find mode related fields
	inFindMode      bool
	findPattern     textarea.Model
	matchedNodes    []*FileNode
	currentMatchIdx int
}

type windowSize struct {
	height int
	width  int
}

func (m *model) buildFileTree() error {
	info, err := os.Stat(m.workDir)
	if err != nil {
		return err
	}

	if m.nodeLookup == nil {
		m.nodeLookup = make(map[string]*FileNode)
	}

	m.rootNode = &FileNode{
		name:     info.Name(),
		path:     m.workDir,
		isDir:    info.IsDir(),
		isRoot:   true,
		expanded: true,
		selected: false,
	}

	err = visitNode(m.rootNode, "", m.removeHidden, m.nodeLookup)
	return err
}

func (m *model) flattenTree() {
	filters := make([]FilterFunc, 0)
	if m.removeHidden {
		filters = append(filters, FilterHidden)
	}
	filters = append(filters, FilterBinary)
	m.flatNodes = m.rootNode.flatten(m.nodeLookup, filters...)
}

func (m *model) toggleDirSelection(node *FileNode) {
	node.selected = !node.selected
	m.nodeLookup[node.path] = node
	for _, child := range node.children {
		if child.isDir {
			m.toggleDirSelection(child)
		} else {
			child.selected = node.selected
			m.nodeLookup[child.path] = child
		}
	}
}

func (m *model) generateOutput(w io.Writer) {
	var output strings.Builder
	m.collectSelectedFiles(m.rootNode, &output)

	_, err := w.Write([]byte(output.String()))
	if err != nil {
		fmt.Printf("Error writing output file: %v\n", err)
		os.Exit(1)
	}
}

func (m *model) collectSelectedFiles(node *FileNode, output *strings.Builder) {
	if node.selected && !node.isDir {
		relPath, _ := filepath.Rel(m.workDir, node.path)
		content, err := os.ReadFile(node.path)
		if err == nil {
			fmt.Fprintf(output, "# %s\n%s\n", relPath, string(content))
		}
	}

	for _, child := range node.children {
		m.collectSelectedFiles(child, output)
	}
}

// Add a new function to model.go that handles node display with highlighting.
func (m *model) getNodeDisplay(node *FileNode) string {
	// Start with the standard string representation
	display := node.String()

	// Check if this node is a match
	isMatch := false
	isCurrentMatch := false

	for matchIdx, matchNode := range m.matchedNodes {
		if node.path == matchNode.path {
			isMatch = true
			if matchIdx == m.currentMatchIdx {
				isCurrentMatch = true
			}
			break
		}
	}

	// Apply highlighting if it's a match
	if isMatch {
		// Extract the name part from the display string
		parts := strings.Split(display, node.name)
		prefix := parts[0]
		suffix := ""
		if len(parts) > 1 {
			suffix = parts[1]
		}

		// Apply different highlighting based on match type
		var highlightedName string
		if isCurrentMatch {
			// Current match gets a different highlight
			highlightedName = lipgloss.NewStyle().
				Foreground(lipgloss.Color("205")).
				Bold(true).
				Render(node.name)
		} else {
			// Other matches get a lighter highlight
			highlightedName = lipgloss.NewStyle().
				Foreground(lipgloss.Color("39")).
				Render(node.name)
		}

		// Reconstruct the display string with highlighted name
		display = prefix + highlightedName + suffix
	}

	return display
}

// Add this method to update content.
func (m *model) updateContent() tea.Cmd {
	buf := bytes.NewBuffer([]byte{})

	// Generate and render markdown content
	m.generateOutput(buf)
	renderedContent, err := m.renderer.Render(buf.String())
	if err != nil {
		renderedContent = fmt.Sprintf("Error rendering content: %v", err)
	}

	// Reset viewport
	m.rightViewport = viewport.New(
		2*m.windowSize.width/3-4, // Width
		m.windowSize.height-2,    // Height
	)

	// Set content and explicitly set viewport to top
	m.rightViewport.SetContent(renderedContent)
	m.rightViewport.YOffset = 0
	return nil
}

const helpMsg = "\nPress space to select, l/h to expand/collapse directories, enter to generate output, q to quit\n"

func (m *model) updateTree() tea.Cmd {
	var builder strings.Builder
	// Add search input at top if in find mode
	if m.inFindMode && m.findPattern.Focused() {
		searchLabel := "Find: "

		searchInfo := fmt.Sprintf("%s%s", searchLabel, m.findPattern.Value())
		if len(m.matchedNodes) > 0 {
			searchInfo += fmt.Sprintf(" (%d/%d matches)", m.currentMatchIdx+1, len(m.matchedNodes))
		} else if m.findPattern.Value() != "" {
			searchInfo += " (no matches)"
		}

		builder.WriteString(lipgloss.NewStyle().
			Foreground(lipgloss.Color("205")).
			Render(searchInfo) + "\n\n")
	}

	// Calculate the actual visible height
	// Subtract help message height and borders/padding
	helpLines := len(strings.Split(helpMsg, "\n"))
	maxVisibleNodes := m.windowSize.height - helpLines - 2 // -2 for top/bottom borders
	if m.inFindMode {
		maxVisibleNodes -= 2 // Account for search line
	}

	// Ensure cursor stays within bounds
	if m.cursor >= len(m.flatNodes) {
		m.cursor = len(m.flatNodes) - 1
	}
	if m.cursor < 0 {
		m.cursor = 0
	}

	// Adjust offset to keep cursor in view
	if m.cursor < m.offset {
		// Cursor moved above viewport
		m.offset = m.cursor
	} else if m.cursor >= m.offset+maxVisibleNodes {
		// Cursor moved below viewport
		m.offset = m.cursor - maxVisibleNodes + 1
	}

	// Ensure offset stays within bounds
	if m.offset < 0 {
		m.offset = 0
	}
	maxOffset := len(m.flatNodes) - maxVisibleNodes
	if maxOffset < 0 {
		maxOffset = 0
	}
	if m.offset > maxOffset {
		m.offset = maxOffset
	}

	// Calculate the visible range
	end := m.offset + maxVisibleNodes
	if end > len(m.flatNodes) {
		end = len(m.flatNodes)
	}

	// Render visible nodes
	for i := m.offset; i < end; i++ {
		node := m.flatNodes[i]
		// Get the node display with potential highlighting
		line := m.getNodeDisplay(node)
		if i == m.cursor {
			line = "> " + line
		} else {
			line = "  " + line
		}
		builder.WriteString(line + "\n")
	}

	// Add up arrow if there's content above
	if m.offset > 0 {
		builder.WriteString("↑ more above\n")
		// maxVisibleNodes-- // Reduce visible nodes to account for indicator
	}
	// Add down arrow if there's content below
	if end < len(m.flatNodes) {
		builder.WriteString("↓ more below\n")
	}

	builder.WriteString(helpMsg)

	// Update viewport with new content
	m.leftViewport = viewport.New(
		m.windowSize.width/3-4, // Width
		m.windowSize.height-2,  // Height
	)
	m.leftViewport.SetContent(builder.String())

	// Important: Track viewport position
	m.leftViewport.YOffset = 0 // Reset to top since we're managing scroll position via offset

	return nil
}

// Update the key handling in the Update method.
func (m *model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg.String() {
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
			cmd = m.updateTree()
		}

	case "down", "j":
		if m.cursor < len(m.flatNodes)-1 {
			m.cursor++
			cmd = m.updateTree()
		}

	case "pgup":
		// Move cursor up by viewport height
		visibleNodes := m.windowSize.height - len(strings.Split(helpMsg, "\n")) - 2
		m.cursor -= visibleNodes
		if m.cursor < 0 {
			m.cursor = 0
		}
		cmd = m.updateTree()

	case "pgdown":
		// Move cursor down by viewport height
		visibleNodes := m.windowSize.height - len(strings.Split(helpMsg, "\n")) - 2
		m.cursor += visibleNodes
		if m.cursor >= len(m.flatNodes) {
			m.cursor = len(m.flatNodes) - 1
		}
		cmd = m.updateTree()

	case "home":
		m.cursor = 0
		m.offset = 0
		cmd = m.updateTree()

	case "end":
		m.cursor = len(m.flatNodes) - 1
		cmd = m.updateTree()

	case "n":
		// Allow n to work in normal mode (after find has been used)
		if len(m.matchedNodes) > 0 {
			m.nextMatch()
			cmd = m.updateTree()
		}

	case "N":
		// Allow N to work in normal mode (after find has been used)
		if len(m.matchedNodes) > 0 {
			m.prevMatch()
			cmd = m.updateTree()
		}

	case "esc":
		// ESC completely exits find mode and clears highlighting
		m.inFindMode = false
		m.findPattern.Reset()
		m.findPattern.Blur()
		m.matchedNodes = nil
		m.currentMatchIdx = -1
		return m, m.updateTree()
	}

	return m, cmd
}

func (m *model) copyToClipboard() error {
	var output strings.Builder
	m.collectSelectedFiles(m.rootNode, &output)
	return clipboard.WriteAll(output.String())
}
