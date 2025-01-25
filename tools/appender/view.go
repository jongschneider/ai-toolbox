package main

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

func (m *model) View() string {
	if m.showClipboardModal {
		modalStyle := lipgloss.NewStyle().
			Width(40).
			Align(lipgloss.Center).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62")).
			Padding(1)

		dialog := "Copy selected files to clipboard?\n\n" +
			"y - copy\n" +
			"n - cancel"

		if m.clipboardError != nil {
			dialog = fmt.Sprintf(
				"Error copying to clipboard:\n%v\n\n[press any key to continue]",
				m.clipboardError,
			)
		}

		modal := modalStyle.Render(dialog)
		return lipgloss.Place(
			m.windowSize.width,
			m.windowSize.height,
			lipgloss.Center,
			lipgloss.Center,
			modal,
			lipgloss.WithWhitespaceChars(""),
			lipgloss.WithWhitespaceForeground(lipgloss.Color("0")),
		)
	}

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
