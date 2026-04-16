package tui

import "github.com/charmbracelet/lipgloss"

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("39")). // bright cyan
			MarginBottom(1)

	subtitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("245")) // dim grey

	selectedStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("15")).  // white
			Background(lipgloss.Color("236")). // dark grey bg
			PaddingLeft(1).PaddingRight(1)

	normalRowStyle = lipgloss.NewStyle().
			PaddingLeft(1).PaddingRight(1)

	criticalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).Bold(true) // bright red

	highStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("214")).Bold(true) // orange

	mediumStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("226")) // yellow

	lowStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("250")) // light grey

	infoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("39")) // cyan

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).Bold(true)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			MarginTop(1)

	detailKeyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("39")).Bold(true)

	detailValStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252"))

	spinnerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("205"))

	borderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("236"))

	bannerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("39")).Bold(true) // bright cyan

	versionBadgeStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("245")) // dim grey
)
