//nolint:forbidigo
package main

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"
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
	workDir      string
	rootNode     *FileNode
	cursor       int
	flatNodes    []*FileNode
	nodeLookup   map[string]*FileNode
	windowSize   windowSize // Number of items to show at once
	offset       int        // Starting index for the window
	removeHidden bool
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

	// handle janky "." right now
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
	slog.Info("before flatten")
	m.flatNodes = m.rootNode.flatten(m.nodeLookup, filters...)
}

func (m *model) Init() tea.Cmd {
	return nil
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
	if err := setupLogging(); err != nil {
		fmt.Printf("Error setting up logging: %v\n", err)
		os.Exit(1)
	}
	slog.Info("starting application")

	workDir := "."
	if len(os.Args) > 1 {
		workDir = os.Args[1]
	}

	// Get terminal height and set window size to leave room for help text
	w, h, _ := term.GetSize(int(os.Stdout.Fd()))
	// windowSize := h - 2 // Leave space for help text

	initialModel := &model{
		workDir: workDir,
		windowSize: windowSize{
			width:  w,
			height: h - 2, // Leave space for help text,
		},
		removeHidden: true,
	}

	if err := initialModel.buildFileTree(); err != nil {
		fmt.Printf("Error building file tree: %v\n", err)
		os.Exit(1)
	}

	slog.Info("before flattenTree")
	initialModel.flattenTree()

	slog.Info("after flattenTree")
	p := tea.NewProgram(initialModel, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// Update window size when terminal is resized
		m.windowSize.height = msg.Height - 4

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
				// Adjust offset if cursor moves above window
				if m.cursor < m.offset {
					m.offset = m.cursor
				}
			}

		case "down", "j":
			if m.cursor < len(m.flatNodes)-1 {
				m.cursor++
				// Adjust offset if cursor moves below window
				if m.cursor >= m.offset+m.windowSize.height {
					m.offset = m.cursor - m.windowSize.height + 1
				}
			}

		case " ":
			currentNode := m.flatNodes[m.cursor]
			if currentNode.isDir {
				m.toggleDirSelection(currentNode)
			} else {
				currentNode.selected = !currentNode.selected
				m.nodeLookup[currentNode.path] = currentNode
			}

		case "l", "h":
			currentNode := m.flatNodes[m.cursor]
			if currentNode.isDir {
				currentNode.expanded = !currentNode.expanded
				m.nodeLookup[currentNode.path] = currentNode
				m.flattenTree()
				// Adjust offset if necessary after tree changes
				if m.cursor >= len(m.flatNodes) {
					m.cursor = len(m.flatNodes) - 1
				}
				if m.offset > m.cursor {
					m.offset = m.cursor
				}
			}

		case ".":
			m.removeHidden = !m.removeHidden
			// if err := m.buildFileTree(); err != nil {
			// 	fmt.Printf("Error building file tree: %v\n", err)
			// 	os.Exit(1)
			// }
			slog.With("removehidden", m.removeHidden).Info("before hidden flatten:")
			m.flattenTree()
			slog.With("removehidden", m.removeHidden).Info("after hidden flatten")

		case "enter":
			m.generateOutput()
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m *model) View() string {
	slog.Info("rendering view...")
	var builder strings.Builder

	// Calculate the visible range
	end := m.offset + m.windowSize.height
	if end > len(m.flatNodes) {
		end = len(m.flatNodes)
	}

	if m.cursor >= len(m.flatNodes) {
		m.cursor = 0
	}

	// Only render the visible portion of the tree
	for i := m.offset; i < end; i++ {
		node := m.flatNodes[i]
		line := node.String()
		if i == m.cursor {
			line = "> " + line
		} else {
			line = "  " + line
		}
		builder.WriteString(line + "\n")
	}

	// Add scrolling indicators if necessary
	if m.offset > 0 {
		builder.WriteString("(↑ more above)\n")
	}
	if end < len(m.flatNodes) {
		builder.WriteString("(↓ more below)\n")
	}

	builder.WriteString(
		"\nPress space to select, l/h to expand/collapse directories, enter to generate output, q to quit\n",
	)

	// Style definitions
	treeStyle := lipgloss.NewStyle().
		Width(m.windowSize.width / 3).
		Height(m.windowSize.height).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Padding(1)

	contentStyle := lipgloss.NewStyle().
		Width(2 * m.windowSize.width / 3).
		Height(m.windowSize.height).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Padding(1)

	treeContent := builder.String()
	leftPane := treeStyle.Render(treeContent)
	rightPane := contentStyle.Render("")
	return lipgloss.JoinHorizontal(lipgloss.Top, leftPane, rightPane)
}

func setupLogging() error {
	// Create logs directory if it doesn't exist
	if err := os.MkdirAll("logs", 0o755); err != nil {
		return fmt.Errorf("could not create logs directory: %w", err)
	}

	// Open log file
	logFile, err := tea.LogToFile("logs/debug.log", "debug")
	if err != nil {
		return fmt.Errorf("could not open log file: %w", err)
	}

	// Configure slog
	logger := slog.New(slog.NewJSONHandler(logFile, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	slog.SetDefault(logger)

	return nil
}
