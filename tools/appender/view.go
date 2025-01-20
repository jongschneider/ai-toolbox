package main

import "github.com/charmbracelet/lipgloss"

func (m *model) View() string {
	// Style definitions
	treeStyle := lipgloss.NewStyle().
		Width(m.windowSize.width/3 - 2).
		Height(m.windowSize.height).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Padding(1)

	contentStyle := lipgloss.NewStyle().
		Width(2*m.windowSize.width/3 - 2).
		Height(m.windowSize.height).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Padding(0)

	// Render both panes
	leftPane := treeStyle.Render(m.leftViewport.View())
	rightPane := contentStyle.Render(m.rightViewport.View())

	// Join panes horizontally
	return lipgloss.JoinHorizontal(lipgloss.Top, leftPane, rightPane)
}
