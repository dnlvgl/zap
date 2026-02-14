package ui

import "github.com/charmbracelet/lipgloss"

var (
	subtle = lipgloss.AdaptiveColor{Light: "#666666", Dark: "#999999"}
	accent = lipgloss.AdaptiveColor{Light: "#7B2FBE", Dark: "#BD93F9"}
	danger = lipgloss.AdaptiveColor{Light: "#CC0000", Dark: "#FF5555"}
	green  = lipgloss.AdaptiveColor{Light: "#007700", Dark: "#50FA7B"}
	yellow = lipgloss.AdaptiveColor{Light: "#886600", Dark: "#F1FA8C"}

	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(accent).
			MarginBottom(1)

	itemStyle = lipgloss.NewStyle().
			PaddingLeft(2)

	selectedStyle = lipgloss.NewStyle().
			PaddingLeft(1).
			Bold(true).
			Foreground(accent).
			SetString(">")

	unselectedStyle = lipgloss.NewStyle().
			PaddingLeft(2)

	pidStyle = lipgloss.NewStyle().
			Foreground(yellow).
			Bold(true)

	portStyle = lipgloss.NewStyle().
			Foreground(accent)

	commandStyle = lipgloss.NewStyle().
			Foreground(subtle)

	tagContainerStyle = lipgloss.NewStyle().
				Foreground(lipgloss.AdaptiveColor{Light: "#FFFFFF", Dark: "#282A36"}).
				Background(lipgloss.AdaptiveColor{Light: "#0077CC", Dark: "#8BE9FD"}).
				Padding(0, 1)

	tagSystemdStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#FFFFFF", Dark: "#282A36"}).
			Background(lipgloss.AdaptiveColor{Light: "#CC7700", Dark: "#FFB86C"}).
			Padding(0, 1)

	confirmStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(danger)

	successStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(green)

	errorStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(danger)

	helpStyle = lipgloss.NewStyle().
			Foreground(subtle).
			MarginTop(1)

	statusBarStyle = lipgloss.NewStyle().
			Foreground(subtle)
)
