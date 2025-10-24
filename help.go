package main

import (
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/lipgloss"
)

// Common key bindings shared across views
type commonKeyMap struct {
	filter     key.Binding
	navigate   key.Binding
	switchTabs key.Binding
	settings   key.Binding
	quit       key.Binding
}

func newCommonKeyMap() commonKeyMap {
	return commonKeyMap{
		filter: key.NewBinding(
			key.WithKeys("/"),
			key.WithHelp("/", "filter"),
		),
		navigate: key.NewBinding(
			key.WithKeys("h", "j", "k", "l"),
			key.WithHelp("hjkl", "navigate"),
		),
		switchTabs: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "switch tabs"),
		),
		settings: key.NewBinding(
			key.WithKeys("s"),
			key.WithHelp("s", "settings"),
		),
		quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
	}
}

func (k commonKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.filter, k.navigate, k.switchTabs, k.settings, k.quit}
}

func (k commonKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.filter, k.navigate, k.switchTabs},
		{k.settings, k.quit},
	}
}

// Common help model that can be shared across views
type commonHelp struct {
	help help.Model
	keys commonKeyMap
}

func newCommonHelp() commonHelp {
	h := help.New()

	// Customize the help view styles with theme
	h.Styles.ShortKey = lipgloss.NewStyle().
		Foreground(theme.BrightBlack())

	h.Styles.ShortDesc = lipgloss.NewStyle().
		Foreground(theme.Fg())

	h.Styles.ShortSeparator = lipgloss.NewStyle().
		Foreground(theme.BrightBlack())

	h.Styles.FullKey = lipgloss.NewStyle().
		Foreground(theme.BrightCyan())

	h.Styles.FullDesc = lipgloss.NewStyle().
		Foreground(theme.Fg())

	h.Styles.FullSeparator = lipgloss.NewStyle().
		Foreground(theme.BrightBlack())

	h.Styles.Ellipsis = lipgloss.NewStyle().
		Foreground(theme.BrightBlack())

	return commonHelp{
		help: h,
		keys: newCommonKeyMap(),
	}
}

func (h commonHelp) View(width int) string {
	h.help.Width = width
	return h.help.View(h.keys)
}
