package main

import "github.com/charmbracelet/bubbles/key"

type keyMap struct {
	Up         key.Binding
	Down       key.Binding
	ToggleDir  key.Binding
	Select     key.Binding
	ToggleHide key.Binding
	Save       key.Binding
	Copy       key.Binding
	Help       key.Binding
	Quit       key.Binding
	Find       key.Binding
	NextMatch  key.Binding
	PrevMatch  key.Binding
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Find, k.Help, k.Quit}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.ToggleDir},
		{k.Select, k.ToggleHide, k.Save},
		{k.Find, k.NextMatch, k.PrevMatch},
		{k.Copy, k.Help, k.Quit},
	}
}

var keys = keyMap{
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "move up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "move down"),
	),
	ToggleDir: key.NewBinding(
		key.WithKeys("l", "h"),
		key.WithHelp("l/h", "expand/collapse"),
	),
	Select: key.NewBinding(
		key.WithKeys("space"),
		key.WithHelp("space", "select"),
	),
	ToggleHide: key.NewBinding(
		key.WithKeys("."),
		key.WithHelp(".", "toggle hidden"),
	),
	Save: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "save"),
	),
	Copy: key.NewBinding(
		key.WithKeys("c"),
		key.WithHelp("c", "copy"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "toggle help"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
	Find: key.NewBinding(
		key.WithKeys("/"),
		key.WithHelp("/", "find"),
	),
	NextMatch: key.NewBinding(
		key.WithKeys("n"),
		key.WithHelp("n", "next match"),
	),
	PrevMatch: key.NewBinding(
		key.WithKeys("N"),
		key.WithHelp("N", "prev match"),
	),
}
