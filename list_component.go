package main

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type listComponent struct {
	list         list.Model
	delegateKeys *delegateKeyMap
	width        int
	height       int
	focused      bool
	listName     string // For filtering items by list
}

func newListComponent(listName string, items []list.Item) listComponent {
	delegateKeys := newDelegateKeyMap()
	delegate := newItemDelegate(delegateKeys)

	l := list.New(items, delegate, 0, 0)
	l.Title = listName

	// Use default title style initially - will be updated when color is set
	l.Styles.Title = titleStyle

	// Customize list styles with theme colors
	l.Styles.PaginationStyle = lipgloss.NewStyle().
		Foreground(theme.BrightBlack())

	l.Styles.HelpStyle = lipgloss.NewStyle().
		Foreground(theme.BrightBlack())

	l.Styles.ActivePaginationDot = lipgloss.NewStyle().
		Foreground(theme.BrightCyan()).
		SetString("•")

	l.Styles.InactivePaginationDot = lipgloss.NewStyle().
		Foreground(theme.BrightBlack()).
		SetString("•")

	l.Styles.FilterPrompt = lipgloss.NewStyle().
		Foreground(theme.BrightCyan())

	l.Styles.FilterCursor = lipgloss.NewStyle().
		Foreground(theme.BrightCyan())

	l.SetShowPagination(true)
	l.SetShowHelp(false)
	l.SetShowTitle(true)
	l.SetShowFilter(false) // Disable individual filtering - use global filter
	l.SetFilteringEnabled(false)

	return listComponent{
		list:         l,
		delegateKeys: delegateKeys,
		focused:      false,
		listName:     listName,
	}
}

func (lc *listComponent) SetSize(width, height int) {
	lc.width = width
	lc.height = height
	lc.list.SetSize(width, height)
}

func (lc *listComponent) SetTitleColor(color lipgloss.TerminalColor) {
	lc.list.Styles.Title = lipgloss.NewStyle().
		Foreground(theme.Bg()).
		Background(color).
		Padding(0, 1)
}

func (lc *listComponent) SetItems(items []list.Item) tea.Cmd {
	return lc.list.SetItems(items)
}

func (lc *listComponent) Focus() {
	lc.focused = true
}

func (lc *listComponent) Blur() {
	lc.focused = false
}

func (lc listComponent) Update(msg tea.Msg) (listComponent, tea.Cmd) {
	// Update the underlying list
	newList, cmd := lc.list.Update(msg)
	lc.list = newList
	return lc, cmd
}

func (lc listComponent) View() string {
	return lc.list.View()
}

func (lc listComponent) SelectedItem() list.Item {
	if lc.list.Cursor() < len(lc.list.Items()) {
		return lc.list.Items()[lc.list.Cursor()]
	}
	return nil
}

func (lc listComponent) FilterState() list.FilterState {
	return lc.list.FilterState()
}
