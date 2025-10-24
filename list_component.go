package main

import (
	"strings"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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
	// Subtract 2 for left and right borders
	lc.list.SetSize(width-2, height)

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
	return lc.list.SetItems(items)
}

func (lc *listComponent) Focus() {
	lc.focused = true
	lc.updateCursorStyle()
}

func (lc *listComponent) Blur() {
	lc.focused = false
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
	view := lc.list.View()

	// If not focused, dim the entire list but keep title readable
	if !lc.focused {
		lines := strings.Split(view, "\n")
		for i := range lines {
			if i == 0 { // keep title as-is
				continue
			}
			lines[i] = lipgloss.NewStyle().Foreground(theme.BrightBlack()).Render(lines[i])
		}
		view = strings.Join(lines, "\n")
	}

	// Add left and right borders with fixed height
	borderStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderLeft(true).
		BorderRight(true).
		BorderForeground(lipgloss.Color("236")).  // Dark gray/black
		Width(lc.width - 2).
		Height(lc.height)

	return borderStyle.Render(view)
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
