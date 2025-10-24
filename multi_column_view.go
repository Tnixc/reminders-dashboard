package main

import (
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sahilm/fuzzy"
	"sort"
	"strings"
)

// getListColor returns a color for list titles, preferring config-defined colors
func getListColor(listName string, index int) lipgloss.TerminalColor {
	// First try to get color from config (listColorMap is loaded in reminders.go)
	if listColorMap != nil {
		if configColor, exists := listColorMap[strings.ToLower(listName)]; exists && configColor != "" {
			// Convert config color name to theme color
			switch strings.ToLower(configColor) {
			case "red":
				return theme.Red()
			case "blue":
				return theme.Blue()
			case "green":
				return theme.Green()
			case "yellow":
				return theme.Yellow()
			case "cyan":
				return theme.Cyan()
			case "magenta":
				return theme.BrightCyan() // closest available
			case "white":
				return theme.White()
			case "black":
				return theme.BrightBlack()
			default:
				// If unknown color, fall back to index-based
			}
		}
	}

	// Fall back to index-based colors
	colors := []lipgloss.TerminalColor{
		theme.Blue(),
		theme.Green(),
		theme.Yellow(),
		theme.Red(),
		theme.Cyan(),
		theme.BrightCyan(),
	}
	return colors[index%len(colors)]
}

type multiColumnView struct {
	listComponents []listComponent
	allItems       []item
	enabledLists   []string
	groupedItems   map[string][]item // Items grouped by list name

	// Filtering
	filterInput textinput.Model
	filtering   bool
	filterValue string

	// Help
	commonHelp commonHelp

	// Dimensions
	width  int
	height int

	// Focus
	focusedIndex int // Which list column is focused (-1 means none)
}

func newMultiColumnView(enabledLists []string) multiColumnView {
	// Create text input for filtering - styled consistently with list view
	ti := textinput.New()
	ti.Placeholder = "Filter..."
	ti.Prompt = "/"
	ti.CharLimit = 100
	ti.Width = 1000 // Prevent wrapping
	ti.PromptStyle = lipgloss.NewStyle().Foreground(theme.BrightCyan())
	ti.TextStyle = lipgloss.NewStyle().Foreground(theme.Fg())
	ti.PlaceholderStyle = lipgloss.NewStyle().Foreground(theme.BrightBlack())

	// Create list components for each enabled list
	var listComponents []listComponent
	for _, listName := range enabledLists {
		component := newListComponent(listName, []list.Item{})
		// Set title color based on list name (from config) or default
		color := getListColor(listName, 0) // Use 0 index - getListColor will use config color if available
		component.SetTitleColor(color)
		listComponents = append(listComponents, component)
	}

	m := multiColumnView{
		listComponents: listComponents,
		allItems:       []item{},
		enabledLists:   enabledLists,
		groupedItems:   make(map[string][]item),
		filterInput:    ti,
		filtering:      false,
		filterValue:    "",
		commonHelp:     newCommonHelp(),
		focusedIndex:   0, // Focus first list by default
	}

	return m
}

func (m *multiColumnView) loadItems() {
	// Load all items (no filter by list)
	items, err := loadRemindersFiltered(nil)
	if err != nil {
		m.allItems = []item{}
		return
	}

	// Convert from list.Item to item
	m.allItems = make([]item, len(items))
	for i, listItem := range items {
		if it, ok := listItem.(item); ok {
			m.allItems[i] = it
		}
	}

	m.applyFilter("")

	// Set initial focus
	if len(m.listComponents) > 0 && m.focusedIndex < len(m.listComponents) {
		m.listComponents[m.focusedIndex].Focus()
	}
}

func (m *multiColumnView) groupItemsByList(items []item) {
	m.groupedItems = make(map[string][]item)

	// Group items by list name
	for _, it := range items {
		// Only include items from enabled lists
		isEnabled := false
		for _, listName := range m.enabledLists {
			if it.listName == listName {
				isEnabled = true
				break
			}
		}

		if isEnabled {
			m.groupedItems[it.listName] = append(m.groupedItems[it.listName], it)
		}
	}
}

func (m *multiColumnView) updateListComponents() {
	// Update each list component with its filtered items
	for i, listName := range m.enabledLists {
		if i < len(m.listComponents) {
			items := m.groupedItems[listName]
			// Convert []item to []list.Item
			listItems := make([]list.Item, len(items))
			for j, item := range items {
				listItems[j] = item
			}
			m.listComponents[i].SetItems(listItems)

			// Update the title color based on first item's color in this list
			if len(items) > 0 && items[0].color != "" {
				m.listComponents[i].SetTitleColor(lipgloss.Color(items[0].color))
			}
		}
	}
}

func (m *multiColumnView) applyFilter(query string) {
	m.filterValue = query

	var filteredItems []item
	if query == "" {
		filteredItems = m.allItems
	} else {
		// Fuzzy search across all items
		var searchStrings []string
		for _, item := range m.allItems {
			searchStrings = append(searchStrings, item.title)
		}

		matches := fuzzy.Find(query, searchStrings)

		filteredItems = make([]item, len(matches))
		for i, match := range matches {
			filteredItems[i] = m.allItems[match.Index]
		}
	}

	// Sort filtered items by due date
	sort.Slice(filteredItems, func(i, j int) bool {
		// Items without due dates go to the end
		if filteredItems[i].parsedDate.IsZero() && !filteredItems[j].parsedDate.IsZero() {
			return false
		}
		if !filteredItems[i].parsedDate.IsZero() && filteredItems[j].parsedDate.IsZero() {
			return true
		}
		if filteredItems[i].parsedDate.IsZero() && filteredItems[j].parsedDate.IsZero() {
			return filteredItems[i].title < filteredItems[j].title
		}
		return filteredItems[i].parsedDate.Before(filteredItems[j].parsedDate)
	})

	// Regroup and update list components
	m.groupItemsByList(filteredItems)
	m.updateListComponents()
}

func (m *multiColumnView) updateEnabledLists(enabledLists []string) {
	m.enabledLists = enabledLists

	// Recreate list components for new enabled lists, preserving colors by name
	var listComponents []listComponent
	for _, listName := range enabledLists {
		component := newListComponent(listName, []list.Item{})
		// Set title color based on list name (from config) or default
		color := getListColor(listName, 0) // Use 0 index - getListColor will use config color if available
		component.SetTitleColor(color)
		listComponents = append(listComponents, component)
	}
	m.listComponents = listComponents

	// Adjust focus if needed
	if m.focusedIndex >= len(m.listComponents) {
		m.focusedIndex = len(m.listComponents) - 1
	}
	if m.focusedIndex < 0 && len(m.listComponents) > 0 {
		m.focusedIndex = 0
	}

	// Regroup items and update list components
	var itemsToUse []item
	if m.filterValue != "" {
		// If filtering, use the filtered items logic
		var searchStrings []string
		for _, item := range m.allItems {
			searchStrings = append(searchStrings, item.title)
		}

		matches := fuzzy.Find(m.filterValue, searchStrings)
		itemsToUse = make([]item, len(matches))
		for i, match := range matches {
			itemsToUse[i] = m.allItems[match.Index]
		}
	} else {
		itemsToUse = m.allItems
	}

	m.groupItemsByList(itemsToUse)
	m.updateListComponents()
}

func (m multiColumnView) Init() tea.Cmd {
	return nil
}

func (m multiColumnView) Update(msg tea.Msg) (multiColumnView, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		// List components will be resized in View
		return m, nil

	case tea.KeyMsg:
		// If filtering, handle filter input
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

		// Handle focus switching between lists using h/l or left/right arrows
		switch msg.String() {
		case "right", "l":
			if len(m.listComponents) > 0 {
				if m.focusedIndex >= 0 && m.focusedIndex < len(m.listComponents) {
					m.listComponents[m.focusedIndex].Blur()
				}
				m.focusedIndex = (m.focusedIndex + 1) % len(m.listComponents)
				m.listComponents[m.focusedIndex].Focus()
			}
			return m, nil
		case "left", "h":
			if len(m.listComponents) > 0 {
				if m.focusedIndex >= 0 && m.focusedIndex < len(m.listComponents) {
					m.listComponents[m.focusedIndex].Blur()
				}
				m.focusedIndex--
				if m.focusedIndex < 0 {
					m.focusedIndex = len(m.listComponents) - 1
				}
				m.listComponents[m.focusedIndex].Focus()
			}
			return m, nil
		}

		// Handle list navigation keys - map jk to arrow keys for up/down
		var mappedMsg tea.Msg = msg
		switch msg.String() {
		case "j":
			mappedMsg = tea.KeyMsg{Type: tea.KeyDown}
		case "k":
			mappedMsg = tea.KeyMsg{Type: tea.KeyUp}
		}

		switch msg.String() {
		case "/":
			// Start filtering
			m.filtering = true
			m.filterInput.Focus()
			return m, nil

		case "esc":
			// Clear filter
			if m.filterValue != "" {
				m.filterInput.SetValue("")
				m.applyFilter("")
				return m, nil
			}
		}

		// Pass navigation keys to the focused list
		if m.focusedIndex >= 0 && m.focusedIndex < len(m.listComponents) {
			newComponent, cmd := m.listComponents[m.focusedIndex].Update(mappedMsg)
			m.listComponents[m.focusedIndex] = newComponent
			cmds = append(cmds, cmd)
		}
	}

	// Don't update other list components - only the focused one should update

	return m, tea.Batch(cmds...)
}

func (m multiColumnView) View() string {
	if m.width == 0 || m.height == 0 {
		return ""
	}

	// Render help first to get its actual height
	helpMaxWidth := m.width
	if helpMaxWidth > 120 {
		helpMaxWidth = 120
	}
	helpView := m.commonHelp.View(helpMaxWidth)

	// Account for padding when calculating available space
	// We have 1 line top padding for the whole view + 1 line padding above columns
	const topPadding = 1
	const columnTopPadding = 1
	helpHeight := lipgloss.Height(helpView)

	// Available height for lists = total height - help height - top padding - column top padding
	listHeight := m.height - helpHeight - topPadding - columnTopPadding
	if listHeight < 0 {
		listHeight = 0
	}

	// Set size for all list components
	numLists := len(m.listComponents)
	if numLists == 0 {
		return ""
	}

	// Always render columns in column view - each column shows its own "No items" if empty
	// Fixed width for each column
	const fixedColumnWidth = 45
	listWidth := fixedColumnWidth

	for i := range m.listComponents {
		m.listComponents[i].SetSize(listWidth, listHeight)
	}

	// Render all list components horizontally with 1 line padding above
	var listViews []string
	for _, component := range m.listComponents {
		columnView := component.View()
		// Add 1 line padding above each column
		columnWithPadding := "\n" + columnView
		listViews = append(listViews, columnWithPadding)
	}

	listsView := lipgloss.JoinHorizontal(lipgloss.Top, listViews...)

	// Join vertically - lipgloss handles the layout
	content := lipgloss.JoinVertical(lipgloss.Left, listsView, helpView)

	// Add 2ch left padding and 1 line top padding only
	paddingStyle := lipgloss.NewStyle().PaddingLeft(2).PaddingTop(1)
	return paddingStyle.Render(content)
}
