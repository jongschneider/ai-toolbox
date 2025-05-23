//nolint:forbidigo
package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textarea"
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
	txtArea := textarea.New()
	txtArea.SetValue("output.txt")
	txtArea.ShowLineNumbers = false
	txtArea.Placeholder = "Enter filename..."
	txtArea.Focus()
	txtArea.SetHeight(1)
	txtArea.CharLimit = 255
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
		outputPath:      txtArea,
		keys:            keys,
		help:            help.New(),
		findPattern:     initFindInput(),
		inFindMode:      false,
		matchedNodes:    []*FileNode{},
		currentMatchIdx: -1,
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

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) { //nolint:gocyclo
	var cmd tea.Cmd

	switch msg := msg.(type) {
	// Handle the pattern changed message in the Update function
	case findPatternChangedMsg:
		if m.inFindMode {
			m.performFind()
		}
		return m, m.updateTree()

	case tea.WindowSizeMsg:
		m.windowSize.height = msg.Height - 4
		m.windowSize.width = msg.Width

		// Update viewport sizes
		m.leftViewport.Width = m.windowSize.width/3 - 4
		m.leftViewport.Height = m.windowSize.height - 2
		m.rightViewport.Width = 2*m.windowSize.width/3 - 4
		m.rightViewport.Height = m.windowSize.height - 2

		m.help.Width = msg.Width
		return m, tea.Batch(
			m.updateTree(),
			m.updateContent(),
		)

	case tea.KeyMsg:
		// Handle keys in find mode
		if m.inFindMode {
			switch msg.String() {
			case tea.KeyEsc.String():
				// ESC completely exits find mode and clears highlighting
				m.inFindMode = false
				m.findPattern.Reset()
				m.findPattern.Blur()
				m.matchedNodes = nil
				m.currentMatchIdx = -1
				return m, m.updateTree()

			case tea.KeyEnter.String():
				// When Enter is pressed while the input is focused:
				// 1. Perform the final search with current pattern
				m.performFind()

				// 2. Keep the pattern but unfocus the input field
				m.findPattern.Blur()

				// 3. Keep inFindMode false so we return to normal mode
				m.inFindMode = false

				// 3. Keep inFindMode true to preserve highlighting and n/N navigation

				return m, m.updateTree()
			}

			// Only pass keypresses to the textarea if it's focused
			if m.findPattern.Focused() {
				var cmd tea.Cmd
				m.findPattern, cmd = m.findPattern.Update(msg)

				// Perform search on each keypress for real-time results
				return m, tea.Batch(cmd, func() tea.Msg {
					m.performFind()
					return findPatternChangedMsg{}
				})
			}

			return m, nil
		}
		if m.showSaveModal {
			switch msg.String() {
			case tea.KeyEsc.String():
				m.showSaveModal = false
				m.outputPath.Reset()
				m.outputPath.SetValue("output.txt")
				return m, nil
			case tea.KeyEnter.String():
				f, err := os.Create(m.outputPath.Value())
				if err != nil {
					slog.Error("Failed to create file", "error", err)
					return m, nil
				}
				defer f.Close()
				m.generateOutput(f)
				return m, tea.Quit
			}
			var cmd tea.Cmd
			m.outputPath, cmd = m.outputPath.Update(msg)
			return m, cmd
		}

		if m.showClipboardModal {
			switch msg.String() {
			case "y":
				err := m.copyToClipboard()
				if err != nil {
					m.clipboardError = err
					return m, nil
				}
				m.showClipboardModal = false
				return m, tea.Quit
			case "n", tea.KeyEsc.String():
				m.showClipboardModal = false
				m.clipboardError = nil
				return m, nil
			default:
				if m.clipboardError != nil {
					m.showClipboardModal = false
					m.clipboardError = nil
				}
				return m, nil
			}
		}

		if key.Matches(msg, m.keys.Help) {
			m.help.ShowAll = !m.help.ShowAll
			return m, nil
		}

		switch msg.String() {
		case "/":
			if !m.showSaveModal && !m.showClipboardModal {
				m.inFindMode = true
				m.findPattern.Reset()
				m.findPattern.Focus()
				m.matchedNodes = nil
				m.currentMatchIdx = -1

				// Immediately update the tree to show the search bar
				return m, m.updateTree()
			}
			fallthrough

		case "ctrl+c", "q":
			return m, tea.Quit

		case "c":
			slog.Info("pressing c")
			m.showClipboardModal = true
			return m, nil

		// Navigation keys are handled by handleKeyPress
		case "up", "k", "down", "j", "pgup", "pgdown", "home", "end", "n", "N", tea.KeyEsc.String():
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
		case tea.KeyEnter.String():
			if !m.showSaveModal {
				m.showSaveModal = true
				m.outputPath.Focus()
				return m, nil
			}
		}
	}

	// Handle viewport updates
	var leftViewportCmd, rightViewportCmd tea.Cmd
	m.leftViewport, leftViewportCmd = m.leftViewport.Update(msg)
	m.rightViewport, rightViewportCmd = m.rightViewport.Update(msg)

	return m, tea.Batch(cmd, leftViewportCmd, rightViewportCmd)
}

// Add a custom message type for pattern changes.
type findPatternChangedMsg struct{}
