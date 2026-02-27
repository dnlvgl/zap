package ui

import "github.com/charmbracelet/lipgloss"

// Color palette
var (
	colorSubtle = lipgloss.AdaptiveColor{Light: "#666666", Dark: "#999999"}
	colorAccent = lipgloss.AdaptiveColor{Light: "#7B2FBE", Dark: "#BD93F9"}
	colorDanger = lipgloss.AdaptiveColor{Light: "#CC0000", Dark: "#FF5555"}
	colorGreen  = lipgloss.AdaptiveColor{Light: "#007700", Dark: "#50FA7B"}
	colorYellow = lipgloss.AdaptiveColor{Light: "#886600", Dark: "#F1FA8C"}
	colorCyan   = lipgloss.AdaptiveColor{Light: "#0077CC", Dark: "#8BE9FD"}
	colorOrange = lipgloss.AdaptiveColor{Light: "#CC7700", Dark: "#FFB86C"}
	colorMuted  = lipgloss.AdaptiveColor{Light: "#888888", Dark: "#6272A4"}
)

// Table styles (used inside StyleFunc callback)
var (
	tableHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(colorSubtle)

	tableCellStyle = lipgloss.NewStyle()

	tableSelectedStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.AdaptiveColor{Light: "#FFFFFF", Dark: "#FFFFFF"}).
				Background(colorAccent)
)

// Tag styles (used in detail panel only)
var (
	tagContainerStyle = lipgloss.NewStyle().
				Foreground(lipgloss.AdaptiveColor{Light: "#FFFFFF", Dark: "#282A36"}).
				Background(colorCyan).
				Padding(0, 1)

	tagSystemdStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#FFFFFF", Dark: "#282A36"}).
			Background(colorOrange).
			Padding(0, 1)

	tagSudoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#FFFFFF", Dark: "#282A36"}).
			Background(colorDanger).
			Padding(0, 1)
)

// Detail panel styles
var (
	detailPanelStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(colorMuted).
				Padding(0, 1)

	detailLabelStyle = lipgloss.NewStyle().
				Width(9).
				Foreground(colorMuted)

	detailValueStyle = lipgloss.NewStyle().
				Foreground(colorSubtle)

	strategyStyle = lipgloss.NewStyle().
			Foreground(colorAccent).
			Bold(true)

	warningStyle = lipgloss.NewStyle().
			Foreground(colorDanger)
)

// Confirm styles
var (
	confirmPanelStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(colorDanger).
				Padding(0, 1)

	confirmPromptStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(colorDanger)

	confirmDescStyle = lipgloss.NewStyle().
				Foreground(colorSubtle)
)

// General styles
var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorAccent).
			MarginBottom(1)

	successStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorGreen)

	errorStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorDanger)

	helpStyle = lipgloss.NewStyle().
			Foreground(colorSubtle).
			MarginTop(1)

	searchStyle = lipgloss.NewStyle().
			Foreground(colorAccent).
			Bold(true)

	searchPromptStyle = lipgloss.NewStyle().
				Foreground(colorMuted)

	searchPlaceholderStyle = lipgloss.NewStyle().
				Foreground(colorMuted).
				Italic(true)
)
