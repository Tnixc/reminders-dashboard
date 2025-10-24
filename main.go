package main

import (
"fmt"
"os"
"strings"
tea "github.com/charmbracelet/bubbletea"
"github.com/charmbracelet/lipgloss"
)

// Root tabs hosting the existing views; Settings opens as a modal overlay.
type rootModel struct {
tabs       []string
activeTab  int
width      int
height     int

// content models
single listModel
multi  multiColumnView

// settings overlay
settingsOpen bool
picker       listPicker

// shared filter state
sharedFilter string
}

func initialModel() rootModel {
// Build list picker from existing reminders
lists, err := getUniqueLists()
if err != nil {
lists = []string{}
}
picker := newListPicker(lists)
enabled := picker.getEnabledLists()

// Child models
single := newListModel()
multi := newMultiColumnView(enabled)
multi.loadItems()

return rootModel{
tabs:         []string{"List", "Columns"},
activeTab:    0,
single:       single,
multi:        multi,
picker:       picker,
sharedFilter: "",
}
}

func (m rootModel) Init() tea.Cmd { return nil }

func (m rootModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
var cmds []tea.Cmd

switch t := msg.(type) {
case tea.WindowSizeMsg:
m.width, m.height = t.Width, t.Height

// Calculate available height for content (tabs at bottom, single line)
tabHeight := 1           // simple text tabs are 1 line tall
availableHeight := m.height - tabHeight - 1 // -1 for newline between content and tabs

// Create adjusted size message for children
adjustedMsg := tea.WindowSizeMsg{Width: t.Width, Height: availableHeight}

// forward to children with adjusted height
var cmd tea.Cmd
v, c := m.single.Update(adjustedMsg); m.single = v.(listModel); cmd = c
cmds = append(cmds, cmd)
m.multi, cmd = m.multi.Update(adjustedMsg)
cmds = append(cmds, cmd)
// picker size
m.picker.width, m.picker.height = t.Width, t.Height

case tea.KeyMsg:
if m.settingsOpen {
// Esc closes settings
if t.String() == "esc" || t.String() == "q" {
m.settingsOpen = false
return m, nil
}
}

// Check if any child view is filtering - if so, skip global hotkeys (except ctrl+c)
isFiltering := false
if m.activeTab == 0 {
// Check if single list view is filtering
isFiltering = m.single.filtering
} else {
// Check if multi-column view is filtering
isFiltering = m.multi.filtering
}

switch t.String() {
case "ctrl+c":
return m, tea.Quit
case "q":
if isFiltering {
break // Let child handle it
}
if m.settingsOpen {
m.settingsOpen = false
return m, nil
}
return m, tea.Quit
case "s":
if isFiltering {
break // Let child handle it
}
// toggle settings overlay
m.settingsOpen = !m.settingsOpen
return m, nil
case "tab":
if isFiltering {
break // Let child handle it
}
if !m.settingsOpen {
// Save current filter before switching
if m.activeTab == 0 {
m.sharedFilter = m.single.filterValue
} else {
m.sharedFilter = m.multi.filterValue
}

// Cycle to next tab (wrap around)
m.activeTab = (m.activeTab + 1) % len(m.tabs)

// Apply shared filter to new tab
if m.activeTab == 0 {
// Switching to single list view - apply shared filter
m.single.filterInput.SetValue(m.sharedFilter)
m.single.filterValue = m.sharedFilter
m.single.applyFilter(m.sharedFilter)
} else if m.activeTab == 1 {
// Switching to multi-column view - apply shared filter
m.multi.filterInput.SetValue(m.sharedFilter)
m.multi.filterValue = m.sharedFilter
m.multi.applyFilter(m.sharedFilter)

// ensure focus on first column
m.multi.focusedIndex = 0
if len(m.multi.listComponents) > 0 {
for i := range m.multi.listComponents {
m.multi.listComponents[i].Blur()
}
m.multi.listComponents[0].Focus()
}
}
}
return m, tea.Batch(cmds...)
case "shift+tab":
if isFiltering {
break // Let child handle it
}
if !m.settingsOpen {
// Cycle to previous tab (wrap around)
m.activeTab--
if m.activeTab < 0 {
m.activeTab = len(m.tabs) - 1
}
}
return m, nil
}

// route keys to active view when not in settings
if !m.settingsOpen {
var cmd tea.Cmd
if m.activeTab == 0 {
v, c := m.single.Update(t)
m.single = v.(listModel)
cmd = c
} else {
m.multi, cmd = m.multi.Update(t)
}
cmds = append(cmds, cmd)
}

case filterChangeMsg:
// Update filters for both views
cmd := m.single.reloadWithFilter(t.enabledLists)
cmds = append(cmds, cmd)
m.multi.updateEnabledLists(t.enabledLists)
}

// Update picker if settings open
if m.settingsOpen {
v, cmd := m.picker.Update(msg)
m.picker = v.(listPicker)
cmds = append(cmds, cmd)
}

return m, tea.Batch(cmds...)
}

var (
// Use tint theme colors
docStyle = lipgloss.NewStyle().Padding(1, 2)
)

func (m rootModel) renderTabs(filterText string, isFiltering bool, filterInput string) string {
	var parts []string
	for i, t := range m.tabs {
		var tabText string
		if i == m.activeTab {
			// Active tab: highlighted
			tabText = lipgloss.NewStyle().
				Foreground(theme.BrightCyan()).
				Bold(true).
				Render("[" + t + "]")
		} else {
			// Inactive tab: dimmed
			tabText = lipgloss.NewStyle().
				Foreground(theme.BrightBlack()).
				Render("[" + t + "]")
		}
		parts = append(parts, tabText)
	}

	tabsRow := strings.Join(parts, " ")

	// Always show filter (prevents newline when toggling filter)
	if isFiltering {
		// Show filter input box with cursor
		cursor := "_"
		filterBox := lipgloss.NewStyle().
			Foreground(theme.BrightCyan()).
			Render(" / " + filterInput + cursor)
		tabsRow = tabsRow + filterBox
	} else if filterText != "" {
		// Show filter indicator
		displayValue := filterText
		if len(displayValue) > 30 {
			displayValue = displayValue[:27] + "..."
		}
		filterIndicator := lipgloss.NewStyle().
			Foreground(theme.Yellow()).
			Render(" " + displayValue)
		tabsRow = tabsRow + filterIndicator
	} else {
		// Show empty filter placeholder to prevent layout shift
		filterPlaceholder := lipgloss.NewStyle().
			Foreground(theme.BrightBlack()).
			Render(" /")
		tabsRow = tabsRow + filterPlaceholder
	}

	return tabsRow
}

func (m rootModel) View() string {
// Get filter text and filtering status from active view
var filterText string
var isFiltering bool
var filterInput string

if m.activeTab == 0 {
// Single list view
isFiltering = m.single.filtering
filterText = m.single.filterValue
if isFiltering {
filterInput = m.single.filterInput.Value()
}
} else {
// Multi-column view
isFiltering = m.multi.filtering
filterText = m.multi.filterValue
if isFiltering {
filterInput = m.multi.filterInput.Value()
}
}

// Background view from active tab
var body string
if m.activeTab == 0 {
body = m.single.View()
} else {
body = m.multi.View()
}

// Render tabs at the bottom with filter
footer := m.renderTabs(filterText, isFiltering, filterInput)

// Calculate available height for content (tabs at bottom, single line)
footerHeight := lipgloss.Height(footer)
availableHeight := m.height - footerHeight - 1 // -1 for newline

// Place content to fill available width and height
content := lipgloss.Place(m.width, availableHeight, lipgloss.Left, lipgloss.Top, body)

if !m.settingsOpen {
return content + "\n" + footer
}

// Settings modal (use existing picker styled by theme)
modal := lipgloss.NewStyle().
Border(lipgloss.NormalBorder()).
BorderForeground(theme.BrightCyan()).
Padding(0, 0).
Render(m.picker.View())

// Simple centered modal substitute
_ = lipgloss.Width(modal)
boxed := lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, modal)
dimmed := lipgloss.NewStyle().Foreground(theme.BrightBlack()).Render(content + "\n" + footer)
return dimmed + "\n" + boxed
}

func main() {
p := tea.NewProgram(initialModel(), tea.WithAltScreen())
if _, err := p.Run(); err != nil {
fmt.Println("Error running program:", err)
os.Exit(1)
}
}
