package main

import (
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	tint "github.com/lrstanley/bubbletint"
	"github.com/sahilm/fuzzy"
	"sort"
	"time"
)

var (
	// Theme for the app
	theme = tint.TintSerendipityMidnight

	appStyle = lipgloss.NewStyle().Padding(1, 2)

	titleStyle = lipgloss.NewStyle().
			Foreground(theme.Bg()).
			Background(theme.Blue()).
			Padding(0, 1)
)

type item struct {
	title        string
	description  string
	listName     string
	color        string
	urgencyText  string
	urgencyColor string
	parsedDate   time.Time
}

func (i item) Title() string {
	// Return plain text with bullet, no colors (colors applied in delegate)
	if i.color != "" {
		return "● " + i.title
	}
	return i.title
}

func (i item) Description() string {
	// Return plain text (colors applied in delegate)
	if i.urgencyText != "" {
		return i.urgencyText + " • " + i.description
	}
	return i.description
}

func (i item) FilterValue() string {
	return i.title
}

// Helper to convert urgency color names to theme colors
func urgencyColorToTheme(colorName string) lipgloss.TerminalColor {
	switch colorName {
	case "red":
		return theme.Red()
	case "orange":
		return theme.Yellow()
	case "yellow":
		return theme.White()
	default:
		// Use subtle/dimmed color for non-urgent items
		return theme.BrightBlack()
	}
}

type listKeyMap struct{}

func newListKeyMap() *listKeyMap {
	return &listKeyMap{}
}

type listModel struct {
	list         list.Model
	delegateKeys *delegateKeyMap
	commonHelp   commonHelp
	width        int
	height       int

	// Custom filtering
	allItems    []list.Item
	filterValue string
	filtering   bool
	filterInput textinput.Model
}

func newListModel() listModel {
	var (
		delegateKeys = newDelegateKeyMap()
	)

	// Load reminders from command
	items, err := loadReminders()
	if err != nil {
		// Fallback to empty list if reminders command fails
		items = []list.Item{}
	}

	// Create text input for filtering
	ti := textinput.New()
	ti.Placeholder = "Filter..."
	ti.Prompt = "/"
	ti.CharLimit = 100
	ti.Width = 1000 // Prevent wrapping
	ti.PromptStyle = lipgloss.NewStyle().Foreground(theme.BrightCyan())
	ti.TextStyle = lipgloss.NewStyle().Foreground(theme.Fg())
	ti.PlaceholderStyle = lipgloss.NewStyle().Foreground(theme.BrightBlack())

	// Setup list
	delegate := newItemDelegate(delegateKeys)
	remindersList := list.New(items, delegate, 0, 0)
	remindersList.Title = ""
	remindersList.Styles.Title = titleStyle
	remindersList.SetShowTitle(false)

	// Customize list styles with theme colors
	remindersList.Styles.PaginationStyle = lipgloss.NewStyle().
		Foreground(theme.BrightBlack())

	remindersList.Styles.HelpStyle = lipgloss.NewStyle().
		Foreground(theme.BrightBlack())

	remindersList.Styles.ActivePaginationDot = lipgloss.NewStyle().
		Foreground(theme.BrightCyan()).
		SetString("•")

	remindersList.Styles.InactivePaginationDot = lipgloss.NewStyle().
		Foreground(theme.BrightBlack()).
		SetString("•")

	// Make filter UI use empty strings - we show it at bottom instead
	remindersList.Styles.FilterPrompt = lipgloss.NewStyle()
	remindersList.Styles.FilterCursor = lipgloss.NewStyle()

	// Remove "No items" text when list is empty
	remindersList.Styles.NoItems = lipgloss.NewStyle()

	// Keep filtering enabled but filter UI will be hidden
	remindersList.SetShowFilter(false)
	remindersList.SetFilteringEnabled(true)

	// Customize the help view styles
	remindersList.Help.Styles.ShortKey = lipgloss.NewStyle().
		Foreground(theme.BrightBlack())

	remindersList.Help.Styles.ShortDesc = lipgloss.NewStyle().
		Foreground(theme.Fg())

	remindersList.Help.Styles.ShortSeparator = lipgloss.NewStyle().
		Foreground(theme.BrightBlack())

	remindersList.Help.Styles.FullKey = lipgloss.NewStyle().
		Foreground(theme.BrightCyan())

	remindersList.Help.Styles.FullDesc = lipgloss.NewStyle().
		Foreground(theme.Fg())

	remindersList.Help.Styles.FullSeparator = lipgloss.NewStyle().
		Foreground(theme.BrightBlack())

	remindersList.Help.Styles.Ellipsis = lipgloss.NewStyle().
		Foreground(theme.BrightBlack())

	remindersList.SetShowPagination(true)
	remindersList.SetShowHelp(false) // Disable list's built-in help, we use commonHelp

	return listModel{
		list:         remindersList,
		delegateKeys: delegateKeys,
		commonHelp:   newCommonHelp(),
		allItems:     items,
		filterInput:  ti,
		filtering:    false,
		filterValue:  "",
	}
}

func (m listModel) Init() tea.Cmd {
	return nil
}

func (m *listModel) reloadWithFilter(enabledLists []string) tea.Cmd {
	items, err := loadRemindersFiltered(enabledLists)
	if err != nil {
		// On error, just keep current items
		return nil
	}

	// SetItems returns a command that needs to be executed
	// This command handles refiltering if a filter is active
	return m.list.SetItems(items)
}

func (m listModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.list.SetSize(msg.Width, msg.Height)

	case tea.KeyMsg:
		// Handle custom filtering
		if m.filtering {
			switch msg.String() {
			case "esc":
				m.filtering = false
				m.filterInput.Blur()
				return m, nil
			case "enter":
				m.filtering = false
				m.filterInput.Blur()
				m.applyFilter(m.filterInput.Value())
				return m, nil
			default:
				newFilterInput, cmd := m.filterInput.Update(msg)
				m.filterInput = newFilterInput
				cmds = append(cmds, cmd)
				return m, tea.Batch(cmds...)
			}
		}

		// Handle "/" to start filtering
		if msg.String() == "/" {
			m.filtering = true
			m.filterInput.Focus()
			return m, nil
		}

		// Handle "esc" to clear filter
		if msg.String() == "esc" && m.filterValue != "" {
			m.filterInput.SetValue("")
			m.filterValue = ""
			m.applyFilter("")
			return m, nil
		}

		// Don't match any of the keys below if we're actively filtering.
		if m.list.FilterState() == list.Filtering {
			break
		}

	}

	// This will also call our delegate's update function.
	newListModel, cmd := m.list.Update(msg)
	m.list = newListModel
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m *listModel) applyFilter(query string) {
	m.filterValue = query

	var filteredItems []list.Item
	if query == "" {
		filteredItems = m.allItems
	} else {
		// Fuzzy search across all items
		var searchStrings []string
		for _, listItem := range m.allItems {
			if it, ok := listItem.(item); ok {
				searchStrings = append(searchStrings, it.title)
			}
		}

		matches := fuzzy.Find(query, searchStrings)

		filteredItems = make([]list.Item, len(matches))
		for i, match := range matches {
			filteredItems[i] = m.allItems[match.Index]
		}
	}

	// Sort filtered items by due date
	sort.Slice(filteredItems, func(i, j int) bool {
		it1, ok1 := filteredItems[i].(item)
		it2, ok2 := filteredItems[j].(item)
		if !ok1 || !ok2 {
			return false
		}
		// Items without due dates go to the end
		if it1.parsedDate.IsZero() && !it2.parsedDate.IsZero() {
			return false
		}
		if !it1.parsedDate.IsZero() && it2.parsedDate.IsZero() {
			return true
		}
		if it1.parsedDate.IsZero() && it2.parsedDate.IsZero() {
			return it1.title < it2.title
		}
		return it1.parsedDate.Before(it2.parsedDate)
	})

	m.list.SetItems(filteredItems)
}

func (m listModel) View() string {
	// Render help first to get its actual height
	helpMaxWidth := m.width
	if helpMaxWidth > 120 {
		helpMaxWidth = 120
	}
	helpView := m.commonHelp.View(helpMaxWidth)

	// Account for padding when calculating available space
	// We have 1 line top padding
	const topPadding = 1
	helpHeight := lipgloss.Height(helpView)

	// Available height for list = total height - help height - top padding
	listHeight := m.height - helpHeight - topPadding
	if listHeight < 0 {
		listHeight = 0
	}
	m.list.SetSize(m.width, listHeight)

	listView := m.list.View()

	// Join vertically - lipgloss handles the layout
	content := lipgloss.JoinVertical(lipgloss.Left, listView, helpView)

	// Add 2ch left padding and 1 line top padding only
	paddingStyle := lipgloss.NewStyle().PaddingLeft(2).PaddingTop(1)
	return paddingStyle.Render(content)
}
