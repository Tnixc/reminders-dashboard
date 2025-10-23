package main

import (
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/lipgloss"
)

// Common key bindings shared across views
type commonKeyMap struct {
	filter    key.Binding
	clearFilter key.Binding
	toggleView key.Binding
	togglePane key.Binding
	switchFocus key.Binding
	quit      key.Binding
}

func newCommonKeyMap() commonKeyMap {
	return commonKeyMap{
		filter: key.NewBinding(
			key.WithKeys("/"),
			key.WithHelp("/", "filter"),
		),
		clearFilter: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "clear filter"),
		),
		toggleView: key.NewBinding(
			key.WithKeys("v"),
			key.WithHelp("v", "toggle view"),
		),
		togglePane: key.NewBinding(
			key.WithKeys("s"),
			key.WithHelp("s", "toggle sidebar"),
		),
		switchFocus: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "switch focus"),
		),
		quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
	}
}

func (k commonKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.filter, k.toggleView, k.togglePane, k.switchFocus, k.quit}
}

func (k commonKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.filter, k.clearFilter, k.toggleView, k.togglePane},
		{k.switchFocus, k.quit},
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
