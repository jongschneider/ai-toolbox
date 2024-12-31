package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// Directory represents a directory in the file system.
type Directory struct {
	path     string
	selected bool
	depth    int
}

// Model represents the application state.
type model struct {
	workDir   string
	dirs      []Directory
	cursor    int
	stage     int // 0: directory selection, 1: filename input
	textInput textinput.Model
	err       error
}

func initialModel(workDir string) model {
	ti := textinput.New()
	ti.Placeholder = "out.txt"
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 20

	return model{
		workDir:   workDir,
		dirs:      []Directory{},
		stage:     0,
		textInput: ti,
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		m.loadDirectories,
		textinput.Blink,
	)
}

// loadDirectories walks through the directory tree and populates the dirs slice.
func (m model) loadDirectories() tea.Msg {
	var dirs []Directory
	err := filepath.Walk(m.workDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			relPath, err := filepath.Rel(m.workDir, path)
			if err != nil {
				return err
			}
			depth := len(strings.Split(relPath, string(os.PathSeparator))) - 1
			dirs = append(dirs, Directory{
				path:     relPath,
				selected: false,
				depth:    depth,
			})
		}
		return nil
	})
	if err != nil {
		return msgError{err}
	}

	return dirsLoadedMsg(dirs)
}

type (
	dirsLoadedMsg []Directory
	msgError      struct{ error }
)

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "up", "k":
			if m.stage == 0 && m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.stage == 0 && m.cursor < len(m.dirs)-1 {
				m.cursor++
			}
		case " ":
			if m.stage == 0 {
				m.dirs[m.cursor].selected = !m.dirs[m.cursor].selected
			}
		case "enter":
			if m.stage == 0 {
				m.stage = 1
				return m, nil
			} else {
				return m, m.processFiles
			}
		}

	case dirsLoadedMsg:
		m.dirs = []Directory(msg)

	case msgError:
		m.err = msg.error
		return m, nil
	}

	if m.stage == 1 {
		var cmd tea.Cmd
		m.textInput, cmd = m.textInput.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m model) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error: %v\n", m.err)
	}

	if m.stage == 0 {
		s := "Select directories to process (space to select, enter to confirm):\n\n"

		for i, dir := range m.dirs {
			cursor := " "
			if i == m.cursor {
				cursor = ">"
			}

			checked := " "
			if dir.selected {
				checked = "x"
			}

			indent := strings.Repeat("  ", dir.depth)
			s += fmt.Sprintf("%s %s [%s] %s\n", cursor, indent, checked, dir.path)
		}

		s += "\nPress q to quit.\n"
		return s
	} else {
		s := "Selected directories:\n"
		for _, dir := range m.dirs {
			if dir.selected {
				s += fmt.Sprintf("  %s\n", dir.path)
			}
		}
		s += "\nEnter output filename (default: out.txt):\n"
		s += m.textInput.View()
		return s
	}
}

func (m model) processFiles() tea.Msg {
	outputFile := m.textInput.Value()
	if outputFile == "" {
		outputFile = "out.txt"
	}

	f, err := os.Create(outputFile) //nolint:varnamelen
	if err != nil {
		return msgError{err}
	}
	defer f.Close()

	for _, dir := range m.dirs {
		if !dir.selected {
			continue
		}

		err := filepath.Walk(
			filepath.Join(m.workDir, dir.path),
			func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				if !info.IsDir() {
					relPath, err := filepath.Rel(m.workDir, path)
					if err != nil {
						return err
					}

					content, err := os.ReadFile(path)
					if err != nil {
						return err
					}

					_, err = fmt.Fprintf(f, "# %s\n%s\n\n", relPath, string(content))
					if err != nil {
						return err
					}
				}
				return nil
			},
		)
		if err != nil {
			return msgError{err}
		}
	}

	return tea.Quit
}

func main() {
	workDir := "."
	if len(os.Args) > 1 {
		workDir = os.Args[1]
	}

	absPath, err := filepath.Abs(workDir)
	if err != nil {
		fmt.Printf("Error resolving path: %v\n", err) //nolint:forbidigo
		os.Exit(1)
	}

	p := tea.NewProgram(initialModel(absPath))
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err) //nolint:forbidigo
		os.Exit(1)
	}
}
