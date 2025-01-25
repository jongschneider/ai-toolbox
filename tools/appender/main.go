//nolint:forbidigo
package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/jongschneider/ai-toolbox/tools/appender/config"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"golang.org/x/term"
)

func main() {
	if err := config.InitConfig(); err != nil {
		fmt.Printf("Error initializing config: %v\n", err)
		os.Exit(1)
	}

	flags := pflag.NewFlagSet("appender", pflag.ExitOnError)
	flags.IntP("logging", "l", 0, "Logging level (1=DEBUG, 2=INFO, 3=WARN, 4=ERROR)")
	if err := flags.Parse(os.Args[1:]); err != nil {
		fmt.Printf("Error parsing flags: %v\n", err)
		os.Exit(1)
	}
	if err := viper.BindPFlags(flags); err != nil {
		fmt.Printf("Error binding flags: %v\n", err)
		os.Exit(1)
	}

	workDir := "."
	if flags.NArg() > 0 {
		workDir = flags.Arg(0)
	}

	if err := setupLogging(); err != nil {
		fmt.Printf("Error setting up logging: %v\n", err)
		os.Exit(1)
	}
	slog.Info("starting application")

	// Get terminal height and set window size to leave room for help text
	w, h, _ := term.GetSize(int(os.Stdout.Fd())) //nolint:varnamelen
	// Initialize glamour renderer
	renderer, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(80),
	)
	if err != nil {
		fmt.Printf("Error creating renderer: %v\n", err)
		os.Exit(1)
	}
	initialModel := &model{
		workDir: workDir,
		windowSize: windowSize{
			width:  w,
			height: h - 2, // Leave space for help text,
		},
		renderer:     renderer,
		removeHidden: true,
		leftViewport: viewport.New(
			w/3-4, // Width (adjusted for borders and padding)
			h-4,   // Height (adjusted for borders and padding)
		),
		rightViewport: viewport.New(
			2*w/3-4, // Width (adjusted for borders and padding)
			h-4,     // Height (adjusted for borders and padding)
		),
	}

	if err := initialModel.buildFileTree(); err != nil {
		fmt.Printf("Error building file tree: %v\n", err)
		os.Exit(1)
	}

	initialModel.flattenTree()

	p := tea.NewProgram(initialModel, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.windowSize.height = msg.Height - 4
		m.windowSize.width = msg.Width

		// Update viewport sizes
		m.leftViewport.Width = m.windowSize.width/3 - 4
		m.leftViewport.Height = m.windowSize.height - 2
		m.rightViewport.Width = 2*m.windowSize.width/3 - 4
		m.rightViewport.Height = m.windowSize.height - 2

		return m, tea.Batch(
			m.updateTree(),
			m.updateContent(),
		)

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		// Navigation keys are handled by handleKeyPress
		case "up", "k", "down", "j", "pgup", "pgdown", "home", "end":
			return m.handleKeyPress(msg)

		// Right viewport scrolling
		case "K":
			m.rightViewport.HalfViewUp()

		case "J":
			m.rightViewport.HalfViewDown()

		case "g":
			m.rightViewport.GotoTop()

		case "G":
			m.rightViewport.GotoBottom()

		case " ":
			currentNode := m.flatNodes[m.cursor]
			if currentNode.isDir {
				m.toggleDirSelection(currentNode)
			} else {
				currentNode.selected = !currentNode.selected
				m.nodeLookup[currentNode.path] = currentNode
			}
			// Update both tree and content after selection changes
			return m, tea.Batch(
				m.updateTree(),
				m.updateContent(),
			)

		case "l", "h":
			currentNode := m.flatNodes[m.cursor]
			if currentNode.isDir {
				currentNode.expanded = !currentNode.expanded
				m.nodeLookup[currentNode.path] = currentNode
				m.flattenTree()
				// Adjust cursor if necessary after tree changes
				if m.cursor >= len(m.flatNodes) {
					m.cursor = len(m.flatNodes) - 1
				}
			}
			return m, m.updateTree()

		case ".":
			m.removeHidden = !m.removeHidden
			m.flattenTree()
			return m, m.updateTree()

		case "enter":
			f, err := os.Create("output.txt")
			if err != nil {
				slog.With("err", err).Error("Error creating output file")
				return m, tea.Quit
			}
			defer f.Close()

			m.generateOutput(f)
			return m, tea.Quit
		}
	}

	// Handle viewport updates
	var leftViewportCmd, rightViewportCmd tea.Cmd
	m.leftViewport, leftViewportCmd = m.leftViewport.Update(msg)
	m.rightViewport, rightViewportCmd = m.rightViewport.Update(msg)

	return m, tea.Batch(cmd, leftViewportCmd, rightViewportCmd)
}
