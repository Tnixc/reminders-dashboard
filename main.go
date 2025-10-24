package main

import (
	"fmt"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"go.dalton.dog/bubbleup"
	"os"
	"os/exec"
	"strings"
	"time"
)

// Root tabs hosting the existing views; Settings opens as a modal overlay.
type rootModel struct {
	tabs      []string
	activeTab int
	width     int
	height    int

	// content models
	single listModel
	multi  multiColumnView

	// settings overlay
	settingsOpen bool
	picker       listPicker

	// edit overlay
	editOpen     bool
	editFocus    int // 0=list, 1=title, 2=notes, 3=complete, 4=delete
	editList     textinput.Model
	editTitle    textinput.Model
	editNotes    textinput.Model
	editComplete bool
	editDelete   bool
	editItem     *item

	// shared filter state
	sharedFilter string

	// alerts
	alert bubbleup.AlertModel
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

	// Edit inputs
	editList := textinput.New()
	editList.Placeholder = "List name..."
	editList.CharLimit = 50
	editList.Width = 50
	editList.CursorStyle = lipgloss.NewStyle().Foreground(theme.BrightCyan())

	editTitle := textinput.New()
	editTitle.Placeholder = "Reminder title..."
	editTitle.CharLimit = 200
	editTitle.Width = 50
	editTitle.CursorStyle = lipgloss.NewStyle().Foreground(theme.BrightCyan())

	editNotes := textinput.New()
	editNotes.Placeholder = "Notes (optional)..."
	editNotes.CharLimit = 500
	editNotes.Width = 50
	editNotes.CursorStyle = lipgloss.NewStyle().Foreground(theme.BrightCyan())

	// Alerts
	alert := *bubbleup.NewAlertModel(80, true)

	return rootModel{
		tabs:         []string{"List", "Columns"},
		activeTab:    0,
		single:       single,
		multi:        multi,
		picker:       picker,
		editOpen:     false,
		editFocus:    1, // start with title
		editList:     editList,
		editTitle:    editTitle,
		editNotes:    editNotes,
		editComplete: false,
		editDelete:   false,
		editItem:     nil,
		sharedFilter: "",
		alert:        alert,
	}
}

func (m rootModel) Init() tea.Cmd { return m.alert.Init() }

func (m rootModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch t := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = t.Width, t.Height

		// Reserve space for footer (tabs/filter) + bottom padding
		// Footer is typically 1 line for tabs + 1 line for bottom padding
		const footerReservedHeight = 2

		// Create adjusted size message for children
		adjustedHeight := t.Height - footerReservedHeight
		if adjustedHeight < 0 {
			adjustedHeight = 0
		}
		adjustedMsg := tea.WindowSizeMsg{Width: t.Width, Height: adjustedHeight}

		// Forward adjusted size to children
		var cmd tea.Cmd
		v, c := m.single.Update(adjustedMsg)
		m.single = v.(listModel)
		cmd = c
		cmds = append(cmds, cmd)
		m.multi, cmd = m.multi.Update(adjustedMsg)
		cmds = append(cmds, cmd)
		// picker size
		m.picker.width, m.picker.height = t.Width, t.Height

	case tea.KeyMsg:
		if m.editOpen {
			switch t.String() {
			case "enter":
				// Save the edit
				newList := strings.TrimSpace(m.editList.Value())
				newTitle := strings.TrimSpace(m.editTitle.Value())
				newNotes := strings.TrimSpace(m.editNotes.Value())
				if newTitle != "" && m.editItem != nil {
					// Run reminders edit command
					args := []string{"edit", newList, m.editItem.externalID}
					if newNotes != "" {
						args = append(args, "--notes", newNotes)
					}
					args = append(args, newTitle)
					cmd := exec.Command("reminders", args...)
					err := cmd.Run()
					if err == nil {
						alertCmd := m.alert.NewAlertCmd(bubbleup.InfoKey, "Reminder updated successfully")
						cmds = append(cmds, alertCmd)

						// Handle complete toggle
						if m.editComplete != m.editItem.completed {
							var completeCmd *exec.Cmd
							if m.editComplete {
								completeCmd = exec.Command("reminders", "complete", newList, m.editItem.externalID)
							} else {
								completeCmd = exec.Command("reminders", "uncomplete", newList, m.editItem.externalID)
							}
							if completeCmd.Run() == nil {
								alertCmd := m.alert.NewAlertCmd(bubbleup.InfoKey, "Completion status updated")
								cmds = append(cmds, alertCmd)
							}
						}
						// Handle delete toggle
						if m.editDelete {
							deleteCmd := exec.Command("reminders", "delete", newList, m.editItem.externalID)
							if deleteCmd.Run() == nil {
								alertCmd := m.alert.NewAlertCmd(bubbleup.InfoKey, "Reminder deleted")
								cmds = append(cmds, alertCmd)
							}
						}
						// Refresh the data
						m.single.reloadWithFilter(nil)
						m.multi.loadItems()
					} else {
						alertCmd := m.alert.NewAlertCmd(bubbleup.InfoKey, "Failed to update reminder")
						cmds = append(cmds, alertCmd)
					}
				}
				m.editOpen = false
				m.editItem = nil
				return m, nil
			case "esc":
				// Cancel edit
				m.editOpen = false
				m.editItem = nil
				return m, nil
			case "tab":
				// Cycle focus forward
				m.editFocus = (m.editFocus + 1) % 5
				m.editList.Blur()
				m.editTitle.Blur()
				m.editNotes.Blur()
				switch m.editFocus {
				case 0:
					m.editList.Focus()
				case 1:
					m.editTitle.Focus()
				case 2:
					m.editNotes.Focus()
				case 3, 4:
					// checkboxes, no focus
				}
				return m, nil
			case "shift+tab":
				// Cycle focus backward
				m.editFocus = (m.editFocus + 4) % 5 // +4 is -1 mod 5
				m.editList.Blur()
				m.editTitle.Blur()
				m.editNotes.Blur()
				switch m.editFocus {
				case 0:
					m.editList.Focus()
				case 1:
					m.editTitle.Focus()
				case 2:
					m.editNotes.Focus()
				case 3, 4:
					// checkboxes, no focus
				}
				return m, nil
			default:
				// Handle input for focused field
				var cmd tea.Cmd
				if m.editFocus < 3 {
					switch m.editFocus {
					case 0:
						m.editList, cmd = m.editList.Update(msg)
					case 1:
						m.editTitle, cmd = m.editTitle.Update(msg)
					case 2:
						m.editNotes, cmd = m.editNotes.Update(msg)
					}
				} else {
					// checkboxes: only handle space to toggle
					if t.String() == " " {
						if m.editFocus == 3 {
							m.editComplete = !m.editComplete
						} else if m.editFocus == 4 {
							m.editDelete = !m.editDelete
						}
					}
					cmd = nil
				}
				return m, cmd
			}
		}

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
		case "enter":
			if isFiltering {
				break // Let child handle it
			}
			if !m.settingsOpen {
				// Open edit overlay for selected reminder
				var selectedItem item
				var ok bool
				if m.activeTab == 0 {
					// Single list view
					if cursor := m.single.list.Cursor(); cursor < len(m.single.list.Items()) {
						selectedItem, ok = m.single.list.Items()[cursor].(item)
					}
				} else {
					// Multi-column view
					if selected := m.multi.listComponents[m.multi.focusedIndex].SelectedItem(); selected != nil {
						selectedItem, ok = selected.(item)
					}
				}
				if ok {
					m.editOpen = true
					m.editFocus = 1 // start with title
					m.editList.SetValue(selectedItem.listName)
					m.editTitle.SetValue(selectedItem.title)
					m.editNotes.SetValue("") // notes not stored in item, could add if needed
					m.editComplete = selectedItem.completed
					m.editDelete = false // default to not delete
					m.editItem = &selectedItem
					m.editList.Blur()
					m.editTitle.Focus()
					m.editNotes.Blur()
					return m, nil
				}
			}

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

	// Update alerts
	alertOut, alertCmd := m.alert.Update(msg)
	m.alert = alertOut.(bubbleup.AlertModel)
	cmds = append(cmds, alertCmd)

	return m, tea.Batch(cmds...)
}

var (
	// Use tint theme colors
	docStyle = lipgloss.NewStyle().Padding(1, 2)
)

func (m rootModel) renderTabs(filterText string, isFiltering bool, filterInput string, width int) string {
	const paddingLeft = 2
	const paddingRight = 2
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
		cursor := "│"
		filterBox := lipgloss.NewStyle().
			Foreground(theme.BrightRed()).
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

	// Add current time and date on the right
	currentTime := time.Now().Format("Monday, January 2, 2006 15:04:05")
	timeStr := lipgloss.NewStyle().
		Foreground(theme.BrightWhite()).
		Render(currentTime)

	// Right-align the time: calculate spaces needed
	effectiveWidth := width - paddingLeft - paddingRight // account for left padding
	tabsWidth := lipgloss.Width(tabsRow)
	timeWidth := lipgloss.Width(timeStr)
	spaces := ""
	if effectiveWidth > tabsWidth+timeWidth {
		spaces = strings.Repeat(" ", effectiveWidth-tabsWidth-timeWidth)
	}

	// Join tabs/filter with spaces and time
	footerLine := tabsRow + spaces + timeStr

	// Add left padding to align with help text
	return lipgloss.NewStyle().PaddingLeft(paddingLeft).Render(footerLine)
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

	// Render tabs at the bottom with filter
	footer := m.renderTabs(filterText, isFiltering, filterInput, m.width)

	// Add bottom padding under the footer
	footerWithPadding := footer + "\n"

	// Background view from active tab
	var body string
	if m.activeTab == 0 {
		body = m.single.View()
	} else {
		body = m.multi.View()
	}

	if !m.settingsOpen && !m.editOpen {
		// Use lipgloss.Height() to properly calculate - no manual arithmetic
		content := lipgloss.JoinVertical(lipgloss.Left, body, footerWithPadding)
		return m.alert.Render(content)
	}

	if m.editOpen {
		// Edit modal
		labelStyle := lipgloss.NewStyle().Foreground(theme.BrightCyan())
		keyStyle := lipgloss.NewStyle().Foreground(theme.BrightBlack())
		descStyle := lipgloss.NewStyle().Foreground(theme.Fg())

		completeCheck := "[ ]"
		if m.editComplete {
			completeCheck = "[x]"
		}
		deleteCheck := "[ ]"
		if m.editDelete {
			deleteCheck = "[x]"
		}

		// Add cursor for focused checkboxes
		completeLine := labelStyle.Render("Complete: ") + completeCheck
		deleteLine := labelStyle.Render("Delete: ") + deleteCheck
		if m.editFocus == 3 {
			completeLine = "> " + completeLine
		} else if m.editFocus == 4 {
			deleteLine = "> " + deleteLine
		}

		editContent := labelStyle.Render("List: ") + m.editList.View() + "\n\n" +
			labelStyle.Render("Title: ") + m.editTitle.View() + "\n\n" +
			labelStyle.Render("Notes: ") + m.editNotes.View() + "\n\n" +
			completeLine + "\n\n" +
			deleteLine + "\n\n" +
			keyStyle.Render("Tab") + descStyle.Render(" / ") + keyStyle.Render("Shift+Tab") + descStyle.Render(" to navigate, ") +
			keyStyle.Render("Space") + descStyle.Render(" to toggle, ") +
			keyStyle.Render("Enter") + descStyle.Render(" to save, ") +
			keyStyle.Render("Esc") + descStyle.Render(" to cancel")
		modal := lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(theme.BrightCyan()).
			Padding(1, 2).
			Render(editContent)

		// Simple centered modal substitute
		_ = lipgloss.Width(modal)
		fullView := lipgloss.JoinVertical(lipgloss.Left, body, footerWithPadding)
		boxed := lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, modal)
		dimmed := lipgloss.NewStyle().Foreground(theme.BrightBlack()).Render(fullView)
		content := dimmed + "\n" + boxed
		return m.alert.Render(content)
	}

	// Settings modal (use existing picker styled by theme)
	modal := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(theme.BrightCyan()).
		Padding(0, 0).
		Render(m.picker.View())

	// Simple centered modal substitute
	_ = lipgloss.Width(modal)
	fullView := lipgloss.JoinVertical(lipgloss.Left, body, footerWithPadding)
	boxed := lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, modal)
	dimmed := lipgloss.NewStyle().Foreground(theme.BrightBlack()).Render(fullView)
	content := dimmed + "\n" + boxed
	return m.alert.Render(content)
}

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
