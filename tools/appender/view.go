package main

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

func (m *model) View() string {
	if m.showSaveModal {
		modalStyle := lipgloss.NewStyle().
			Width(60).
			Align(lipgloss.Center).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62")).
			Padding(1)

		modal := modalStyle.Render(
			"Save output to file\n\n" +
				m.outputPath.View() + "\n\n" +
				"[enter to save, esc to cancel]",
		)

		return lipgloss.Place(
			m.windowSize.width,
			m.windowSize.height,
			lipgloss.Center,
			lipgloss.Center,
			modal,
		)
	}

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

	// // Render both panes
	mainView := lipgloss.JoinHorizontal(lipgloss.Top,
		treeStyle.Render(m.leftViewport.View()),
		contentStyle.Render(m.rightViewport.View()),
	)

	// Add help view at the bottom
	return fmt.Sprintf("%s\n%s", mainView, m.help.View(m.keys))
}
