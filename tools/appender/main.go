package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/erikgeiser/promptkit/selection"
	"github.com/muesli/termenv"
)

type Directory struct {
	Path string
}

type model struct {
	workDir   string
	selected  []string
	stage     int // 0: directory selection, 1: filename input
	textInput textinput.Model
	err       error
}

type TreeDir struct {
	Path     string
	Depth    int
	LastItem bool
	Parents  []bool // true means the parent level is a last item
}

func buildTreeDirs(paths []Directory) []TreeDir {
	var treeDirs []TreeDir
	for i, dir := range paths {
		parts := strings.Split(dir.Path, string(os.PathSeparator))
		depth := len(parts) - 1

		// Build parents slice
		parents := make([]bool, depth)
		currentPath := ""
		for j := 0; j < depth; j++ {
			currentPath = filepath.Join(currentPath, parts[j])
			isLast := true
			// Check if there are any more items at this level
			for k := i + 1; k < len(paths); k++ {
				otherParts := strings.Split(paths[k].Path, string(os.PathSeparator))
				if len(otherParts) > j+1 &&
					strings.HasPrefix(paths[k].Path, currentPath+string(os.PathSeparator)) {
					isLast = false
					break
				}
			}
			parents[j] = isLast
		}

		// Determine if this is the last item at its level
		isLast := true
		for j := i + 1; j < len(paths); j++ {
			otherParts := strings.Split(paths[j].Path, string(os.PathSeparator))
			if len(otherParts) == len(parts) {
				// Compare all parts except the last one
				isEqual := true
				for k := 0; k < len(parts)-1; k++ {
					if otherParts[k] != parts[k] {
						isEqual = false
						break
					}
				}
				if isEqual {
					isLast = false
					break
				}
			}
		}

		treeDirs = append(treeDirs, TreeDir{
			Path:     dir.Path,
			Depth:    depth,
			LastItem: isLast,
			Parents:  parents,
		})
	}
	return treeDirs
}

const customTemplate = `
{{- if .Prompt -}}
  {{ Bold .Prompt }}
{{ end -}}
{{ if .IsFiltered }}
  {{- print .FilterPrompt " " .FilterInput }}
{{ end }}

{{- range  $i, $choice := .Choices }}
  {{- $treeDir := (getTreeDir $choice) -}}
  {{- if IsScrollUpHintPosition $i }}
    {{- "⇡ " -}}
  {{- else if IsScrollDownHintPosition $i -}}
    {{- "⇣ " -}} 
  {{- else -}}
    {{- "  " -}}
  {{- end -}}

  {{- range $level := iterate $treeDir.Depth -}}
    {{- if lt $level (sub $treeDir.Depth 1) -}}
      {{- if index $treeDir.Parents $level -}}
        {{- "    " -}}
      {{- else -}}
        {{- "│   " -}}
      {{- end -}}
    {{- end -}}
  {{- end -}}

  {{- if gt $treeDir.Depth 0 -}}
    {{- if $treeDir.LastItem -}}
      {{- "└── " -}}
    {{- else -}}
      {{- "├── " -}}
    {{- end -}}
  {{- else -}}
    {{- "│ " -}}
  {{- end -}}

  {{- if eq $.SelectedIndex $i }}
   {{- print "[" (Foreground "32" (Bold "x")) "] " (Selected $choice) "\n" }}
  {{- else }}
    {{- print "[ ] " (Unselected $choice) "\n" }}
  {{- end }}
{{- end}}`

func initialModel(workDir string) model {
	ti := textinput.New()
	ti.Placeholder = "out.txt"
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 20

	return model{
		workDir:   workDir,
		selected:  []string{},
		stage:     0,
		textInput: ti,
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		m.selectDirectories,
		textinput.Blink,
	)
}

// getDirectories returns a list of non-hidden directories
func getDirectories(root string) ([]Directory, error) {
	var dirs []Directory

	err := filepath.Walk(root, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Get the base name of the current path
		name := info.Name()

		// Skip hidden files and directories (starting with .)
		if strings.HasPrefix(name, ".") {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Check if any parent directory is hidden
		parts := strings.Split(path, string(os.PathSeparator))
		for _, part := range parts {
			if strings.HasPrefix(part, ".") {
				if info.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
		}

		// Skip if not a directory
		if !info.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}

		// Skip root directory
		if relPath == "." {
			return nil
		}

		dirs = append(dirs, Directory{Path: relPath})
		return nil
	})

	return dirs, err
}

func (m model) selectDirectories() tea.Msg {
	dirs, err := getDirectories(m.workDir)
	if err != nil {
		return errMsg{err}
	}

	blue := termenv.String().Foreground(termenv.ANSI256Color(32))
	selected := []string{}

	// Keep selecting directories until user is done
	for {
		treeDirs := buildTreeDirs(dirs)

		sp := selection.New("Select directories (press 'q' when done):", dirs)
		sp.Template = customTemplate
		sp.PageSize = 15 // Increased to show more of the tree structure
		sp.FilterPrompt = "Filter directories:"
		sp.FilterPlaceholder = "Type to filter"
		sp.SelectedChoiceStyle = func(c *selection.Choice[Directory]) string {
			return blue.Bold().Styled(filepath.Base(c.Value.Path))
		}
		sp.UnselectedChoiceStyle = func(c *selection.Choice[Directory]) string {
			return filepath.Base(c.Value.Path)
		}
		sp.Filter = func(filter string, choice *selection.Choice[Directory]) bool {
			return strings.Contains(strings.ToLower(choice.Value.Path), strings.ToLower(filter))
		}
		sp.ExtendedTemplateFuncs = map[string]interface{}{
			"getTreeDir": func(c *selection.Choice[Directory]) TreeDir {
				for _, td := range treeDirs {
					if td.Path == c.Value.Path {
						return td
					}
				}
				return TreeDir{}
			},
			"iterate": func(n int) []int {
				var result []int
				for i := 0; i < n; i++ {
					result = append(result, i)
				}
				return result
			},
			"sub": func(a, b int) int {
				return a - b
			},
		}

		choice, err := sp.RunPrompt()
		if err != nil {
			if err.Error() == "interrupted" {
				break
			}
			return errMsg{err}
		}

		// Add to selected if not already selected
		alreadySelected := false
		for _, s := range selected {
			if s == choice.Path {
				alreadySelected = true
				break
			}
		}
		if !alreadySelected {
			selected = append(selected, choice.Path)
		}
	}

	return dirsSelectedMsg(selected)
}

type (
	dirsSelectedMsg []string
	errMsg          struct{ error }
)

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "enter":
			if m.stage == 1 {
				return m, m.processFiles
			}
		}

	case dirsSelectedMsg:
		m.selected = []string(msg)
		m.stage = 1
		return m, nil

	case errMsg:
		m.err = msg.error
		return m, tea.Quit
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

	if m.stage == 1 {
		s := "Selected directories:\n"
		for _, dir := range m.selected {
			s += fmt.Sprintf("  %s\n", dir)
		}
		s += "\nEnter output filename (default: out.txt):\n"
		s += m.textInput.View()
		return s
	}

	return "\n"
}

func (m model) processFiles() tea.Msg {
	outputFile := m.textInput.Value()
	if outputFile == "" {
		outputFile = "out.txt"
	}

	f, err := os.Create(outputFile)
	if err != nil {
		return errMsg{err}
	}
	defer f.Close()

	for _, dir := range m.selected {
		err := filepath.Walk(
			filepath.Join(m.workDir, dir),
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
			return errMsg{err}
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
		fmt.Printf("Error resolving path: %v\n", err)
		os.Exit(1)
	}

	p := tea.NewProgram(initialModel(absPath))
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}
