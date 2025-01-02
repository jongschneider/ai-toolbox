package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type FileNode struct {
	name     string
	path     string
	isDir    bool
	expanded bool
	selected bool
	children []*FileNode
}

type model struct {
	workDir    string
	rootNode   *FileNode
	cursor     int
	flatNodes  []*FileNode
	linePrefix map[*FileNode]string
}

func buildFileTree(path string, isRoot bool) (*FileNode, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	node := &FileNode{
		name:     info.Name(),
		path:     path,
		isDir:    info.IsDir(),
		expanded: true,
		selected: false,
	}

	if isRoot {
		node.name = "."
	} else if strings.HasPrefix(node.name, ".") {
		return nil, nil
	}

	if node.isDir {
		entries, err := os.ReadDir(path)
		if err != nil {
			return nil, err
		}

		for _, entry := range entries {
			childPath := filepath.Join(path, entry.Name())
			childNode, err := buildFileTree(childPath, false)
			if err != nil || childNode == nil {
				continue
			}
			node.children = append(node.children, childNode)
		}
	}

	return node, nil
}

func (m *model) flattenTree() {
	m.flatNodes = make([]*FileNode, 0)
	m.linePrefix = make(map[*FileNode]string)
	m.flattenNode(m.rootNode, "", true)
}

func (m *model) flattenNode(node *FileNode, prefix string, isLast bool) {
	m.flatNodes = append(m.flatNodes, node)
	m.linePrefix[node] = prefix

	if !node.isDir || !node.expanded {
		return
	}

	for i, child := range node.children {
		newPrefix := prefix
		if i == len(node.children)-1 {
			m.flattenNode(child, newPrefix+"└──", true)
		} else {
			m.flattenNode(child, newPrefix+"├──", false)
		}
	}
}

func (m *model) getNodeLine(node *FileNode) string {
	prefix := m.linePrefix[node]
	dirIndicator := ""
	if node.isDir {
		if node.expanded {
			dirIndicator = "󱞣 "
		} else {
			dirIndicator = " "
		}
	}

	selected := ""
	if node.selected {
		selected = " "
	}

	return fmt.Sprintf("%s%s%s%s", prefix, dirIndicator, node.name, selected)
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
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

func (m model) View() string {
	var s strings.Builder

	for i, node := range m.flatNodes {
		line := m.getNodeLine(node)
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
			output.WriteString(fmt.Sprintf("# %s\n%s\n", relPath, string(content)))
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
