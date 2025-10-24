package main

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"strings"
)

type listComponent struct {
	list         list.Model
	delegateKeys *delegateKeyMap
	delegate     customItemDelegate
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

	// Keep left padding consistent with list view title padding (unchanged)

	l.SetShowPagination(true)
	l.SetShowHelp(false)
	l.SetShowTitle(true)
	l.SetShowFilter(false) // Disable individual filtering - use global filter
	l.SetFilteringEnabled(false)

	return listComponent{
		list:         l,
		delegateKeys: delegateKeys,
		delegate:     delegate,
		focused:      false,
		listName:     listName,
	}
}

func (lc *listComponent) SetSize(width, height int) {
	lc.width = width
	lc.height = height
	// No need to subtract for borders since we only have left border
	lc.list.SetSize(width, height)

	// Update cursor visibility based on focus
	lc.updateCursorStyle()
}

func (lc *listComponent) SetTitleColor(color lipgloss.TerminalColor) {
	lc.list.Styles.Title = lipgloss.NewStyle().
		Foreground(theme.Bg()).
		Background(color).
		Padding(0, 1)
}

func (lc *listComponent) SetItems(items []list.Item) tea.Cmd {
	// For column view, remove the redundant list name from descriptions
	modifiedItems := make([]list.Item, len(items))
	for i, it := range items {
		if item, ok := it.(item); ok {
			// Remove " • listName" from description if present
			suffix := " • " + lc.listName
			if strings.HasSuffix(item.description, suffix) {
				item.description = strings.TrimSuffix(item.description, suffix)
			}
			modifiedItems[i] = item
		} else {
			modifiedItems[i] = it
		}
	}
	return lc.list.SetItems(modifiedItems)
}

func (lc *listComponent) Focus() {
	lc.focused = true
	// Title indicator will be added in View() outside the badge
	lc.list.Title = lc.listName
	lc.updateCursorStyle()
}

func (lc *listComponent) Blur() {
	lc.focused = false
	lc.list.Title = lc.listName
	lc.updateCursorStyle()
}

func (lc *listComponent) updateCursorStyle() {
	if lc.focused {
		// Show selection indicator with normal border
		lc.delegate.Styles.SelectedTitle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), false, false, false, true).
			BorderForeground(theme.BrightCyan()).
			Padding(0, 0, 0, 1)

		lc.delegate.Styles.SelectedDesc = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), false, false, false, true).
			BorderForeground(theme.BrightCyan()).
			Foreground(theme.BrightBlack()).
			Padding(0, 0, 0, 1)
	} else {
		// Completely hide selection indicator - make selected styles identical to normal styles
		// Copy the exact same styles as NormalTitle and NormalDesc so there's no visual difference
		lc.delegate.Styles.SelectedTitle = lc.delegate.Styles.NormalTitle.Copy()
		lc.delegate.Styles.SelectedDesc = lc.delegate.Styles.NormalDesc.Copy()
	}

	// Update the delegate on the list
	lc.list.SetDelegate(lc.delegate)
}

func (lc listComponent) Update(msg tea.Msg) (listComponent, tea.Cmd) {
	// Only process navigation and selection when focused; allow size and non-key msgs
	if !lc.focused {
		if _, ok := msg.(tea.WindowSizeMsg); ok {
			// still update size changes
			newList, cmd := lc.list.Update(msg)
			lc.list = newList
			return lc, cmd
		}
		if _, ok := msg.(tea.KeyMsg); ok {
			return lc, nil
		}
		newList, cmd := lc.list.Update(msg)
		lc.list = newList
		return lc, cmd
	}

	// Update the underlying list
	newList, cmd := lc.list.Update(msg)
	lc.list = newList
	return lc, cmd
}

func (lc listComponent) View() string {
	var view string
	if len(lc.list.Items()) == 0 {
		// For empty lists, show only the title
		view = lc.list.Styles.Title.Render(lc.list.Title) + "\n"
	} else {
		view = lc.list.View()
	}
	lines := strings.Split(view, "\n")

	// Add focus indicator icon on the left if focused
	if lc.focused && len(lines) > 0 {
		// Indicator for current list
		focusIcon := lipgloss.NewStyle().
			Foreground(theme.BrightCyan()).
			Render("󰻿 ")
		lines[0] = focusIcon + lines[0]
	}

	// If not focused, dim the entire list but keep title readable
	if !lc.focused {
		for i := range lines {
			if i == 0 { // keep title as-is
				continue
			}
			lines[i] = lipgloss.NewStyle().Foreground(theme.BrightBlack()).Render(lines[i])
		}
	}

	view = strings.Join(lines, "\n")

	// Use highlighted border color when focused
	var borderColor lipgloss.TerminalColor
	if lc.focused {
		borderColor = theme.BrightCyan() // Highlighted for focused
	} else {
		borderColor = lipgloss.Color("236") // Dark gray/black for unfocused
	}

	// Create vertical bar: space + repeating │ + space
	barHeight := strings.Count(view, "\n") + 1
	bar := strings.Repeat("│\n", barHeight-1) + "│"
	barStyled := lipgloss.NewStyle().Foreground(borderColor).Render(bar)

	// Combine: space + bar + space + view
	combined := lipgloss.JoinHorizontal(lipgloss.Top, " ", barStyled, " ", view)

	// Ensure fixed width
	return lipgloss.NewStyle().Width(53).Render(combined)
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
