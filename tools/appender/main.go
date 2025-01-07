//nolint:forbidigo
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
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

type model struct {
	workDir   string
	rootNode  *FileNode
	cursor    int
	flatNodes []*FileNode
}

func buildFileTree(path string, removeHidden bool) (*FileNode, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	node := &FileNode{
		name:     info.Name(),
		path:     path,
		isDir:    info.IsDir(),
		isRoot:   true,
		expanded: true,
		selected: false,
	}

	err = visitNode(node, "", removeHidden)
	return node, err
}

func (m *model) flattenTree() {
	m.flatNodes = m.rootNode.flatten()
}

func (m *model) Init() tea.Cmd {
	return nil
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) { //nolint:gocritic
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.flatNodes)-1 {
				m.cursor++
			}
		case " ":
			currentNode := m.flatNodes[m.cursor]
			if currentNode.isDir {
				m.toggleDirSelection(currentNode)
			} else {
				currentNode.selected = !currentNode.selected
			}
		case "l", "h":
			currentNode := m.flatNodes[m.cursor]
			if currentNode.isDir {
				currentNode.expanded = !currentNode.expanded
				m.flattenTree()
			}
		case "enter":
			m.generateOutput()
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m *model) toggleDirSelection(node *FileNode) {
	node.selected = !node.selected
	for _, child := range node.children {
		if child.isDir {
			m.toggleDirSelection(child)
		} else {
			child.selected = node.selected
		}
	}
}

func (m *model) View() string {
	var s strings.Builder

	for i, node := range m.flatNodes {
		line := node.String()
		if i == m.cursor {
			line = "> " + line
		} else {
			line = "  " + line
		}
		s.WriteString(line + "\n")
	}

	s.WriteString(
		"\nPress space to select, l/h to expand/collapse directories, enter to generate output, q to quit\n",
	)
	return s.String()
}

func (m *model) generateOutput() {
	var output strings.Builder
	m.collectSelectedFiles(m.rootNode, &output)

	err := os.WriteFile("output.txt", []byte(output.String()), 0o644)
	if err != nil {
		fmt.Printf("Error writing output file: %v\n", err)
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

func main() {
	workDir := "."
	if len(os.Args) > 1 {
		workDir = os.Args[1]
	}

	rootNode, err := buildFileTree(workDir, true)
	if err != nil {
		fmt.Printf("Error building file tree: %v\n", err)
		os.Exit(1)
	}

	initialModel := &model{
		workDir:  workDir,
		rootNode: rootNode,
		cursor:   0,
	}

	initialModel.flattenTree()

	p := tea.NewProgram(initialModel)
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}
