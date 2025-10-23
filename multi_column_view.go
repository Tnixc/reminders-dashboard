package main

import (
	"strings"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sahilm/fuzzy"
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
	for i, listName := range enabledLists {
		component := newListComponent(listName, []list.Item{})
		// Set title color based on list name (from config) or index fallback
		color := getListColor(listName, i)
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

	// Regroup and update list components
	m.groupItemsByList(filteredItems)
	m.updateListComponents()
}

func (m *multiColumnView) updateEnabledLists(enabledLists []string) {
	m.enabledLists = enabledLists

	// Recreate list components for new enabled lists
	var listComponents []listComponent
	for i, listName := range enabledLists {
		component := newListComponent(listName, []list.Item{})
		// Set title color based on list name (from config) or index fallback
		color := getListColor(listName, i)
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

		// Handle focus switching between lists
		switch msg.String() {
		case "tab":
			// Move focus to next list
			if len(m.listComponents) > 0 {
				m.listComponents[m.focusedIndex].Blur()
				m.focusedIndex = (m.focusedIndex + 1) % len(m.listComponents)
				m.listComponents[m.focusedIndex].Focus()
			}
			return m, nil
		case "shift+tab":
			// Move focus to previous list
			if len(m.listComponents) > 0 {
				m.listComponents[m.focusedIndex].Blur()
				m.focusedIndex--
				if m.focusedIndex < 0 {
					m.focusedIndex = len(m.listComponents) - 1
				}
				m.listComponents[m.focusedIndex].Focus()
			}
			return m, nil
		}

		// Handle list navigation keys - map hjkl to arrow keys
		var mappedMsg tea.Msg = msg
		switch msg.String() {
		case "h":
			mappedMsg = tea.KeyMsg{Type: tea.KeyLeft}
		case "j":
			mappedMsg = tea.KeyMsg{Type: tea.KeyDown}
		case "k":
			mappedMsg = tea.KeyMsg{Type: tea.KeyUp}
		case "l":
			mappedMsg = tea.KeyMsg{Type: tea.KeyRight}
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

	// Update all list components with other messages
	for i := range m.listComponents {
		newComponent, cmd := m.listComponents[i].Update(msg)
		m.listComponents[i] = newComponent
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m multiColumnView) View() string {
	if m.width == 0 || m.height == 0 {
		return ""
	}

	// Title or filter input
	var topView string
	if m.filtering {
		topView = m.filterInput.View()
	} else {
		// Show title, or filter indicator if filtered
		if m.filterValue != "" {
			displayValue := m.filterValue
			if len(displayValue) > 20 {
				displayValue = displayValue[:17] + "..."
			}
			topView = "🔍 " + displayValue
		} else {
			topView = titleStyle.Render("Reminders")
		}
	}

	// Render help with large width to avoid wrapping
	helpView := m.commonHelp.View(1000)

	// Compute dynamic list height
	topHeight := lipgloss.Height(topView)
	helpHeight := lipgloss.Height(helpView)
	totalPadding := 2 // vertical padding from appStyle
	listHeight := m.height - totalPadding - topHeight - helpHeight
	if listHeight < 0 {
		listHeight = 0
	}

	// Cap list height to prevent exceeding bubbleboxer line limit (52 lines total)
	maxListHeight := 45 // Leave buffer for variations in help height
	if listHeight > maxListHeight {
		listHeight = maxListHeight
	}

	// Set size for all list components
	// Be very conservative with width to avoid bubbleboxer limits (186 chars max)
	numLists := len(m.listComponents)
	if numLists == 0 {
		return ""
	}

	// Allocate width per list with conservative limits
	maxListWidth := 35 // Very conservative per-list width
	listWidth := m.width / numLists
	if listWidth > maxListWidth {
		listWidth = maxListWidth
	}
	if listWidth < 15 { // Minimum reasonable width for a list
		listWidth = 15
	}

	for i := range m.listComponents {
		m.listComponents[i].SetSize(listWidth, listHeight)
	}

	// Render all list components horizontally with padding above
	var listViews []string
	for _, component := range m.listComponents {
		listViews = append(listViews, component.View())
	}

	listsView := lipgloss.JoinHorizontal(lipgloss.Top, listViews...)

	// Add one line of padding above the columns
	paddedListsView := "\n" + listsView

	// Build the layout: top (title/filter) + padded lists + help
	parts := []string{topView, paddedListsView, helpView}

	return appStyle.Render(lipgloss.JoinVertical(lipgloss.Left, parts...))
}
