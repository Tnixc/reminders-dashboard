package main

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// Init initializes the model
func (m model) Init() tea.Cmd {
	return loadRemindersCmd
}

// Update handles messages and updates the model
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		// Keep cursor visible after resize
		if !m.sidebarFocused {
			if m.viewMode == ListView {
				filtered := getFilteredReminders(&m)
				visibleHeight := m.getContentHeight()
				bounds := make([]reminderBounds, 0, len(filtered))
				current := 0
				for _, r := range filtered {
					start := current
					current += 2 // title + info
					if r.Notes != "" {
						noteLines := strings.Split(r.Notes, "\n")
						for _, line := range noteLines {
							if strings.TrimSpace(line) != "" {
								current++
							}
						}
					}
					current++ // blank
					bounds = append(bounds, reminderBounds{startLine: start, endLine: current})
				}
				if m.cursor < len(bounds) {
					cursorStart := bounds[m.cursor].startLine
					cursorEnd := bounds[m.cursor].endLine
					if cursorStart < m.scrollOffset {
						m.scrollOffset = cursorStart
					}
					if cursorEnd > m.scrollOffset+visibleHeight {
						m.scrollOffset = cursorEnd - visibleHeight
						if m.scrollOffset < 0 {
							m.scrollOffset = 0
						}
					}
				} else {
					m.scrollOffset = 0
				}
			} else {
				columns := getColumns(&m)
				if len(m.columnScrolls) != len(columns) {
					m.columnScrolls = make([]int, len(columns))
				}
				if m.columnCursor < len(columns) {

					visibleHeight := m.getContentHeight()
					cursorLine := m.cursor * 4
					scroll := m.columnScrolls[m.columnCursor]
					if cursorLine < scroll {
						m.columnScrolls[m.columnCursor] = cursorLine
					}
					if cursorLine >= scroll+visibleHeight {
						newScroll := cursorLine - visibleHeight + 4
						if newScroll < 0 {
							newScroll = 0
						}
						m.columnScrolls[m.columnCursor] = newScroll
					}
				}
			}
		}
		return m, nil

	case remindersLoadedMsg:
		m.loading = false
		m.err = msg.err
		if msg.err == nil {
			m.reminders = msg.reminders
			sortReminders(&m)
			updateAvailableLists(&m)
			
			// Update table data for list view
			m.updateTableData()

			// Initialize selectedLists if empty (first run)
			if len(m.selectedLists) == 0 {
				for _, list := range m.availableLists {
					m.selectedLists[list] = true
				}
			} else {
				// Add any new lists that weren't in the config
				for _, list := range m.availableLists {
					if _, exists := m.selectedLists[list]; !exists {
						m.selectedLists[list] = true
					}
				}
			}

			// Assign default colors to lists that don't have one
			for i, list := range m.availableLists {
				if _, exists := m.listColors[list]; !exists {
					m.listColors[list] = m.availableColors[i%len(m.availableColors)]
				}
			}

			// Save initial config
			m.saveConfig()

			// Initialize column scrolls
			columns := getColumns(&m)
			m.columnScrolls = make([]int, len(columns))

			// Initialize column order if not set
			if len(m.columnOrder) == 0 {
				m.columnOrder = getColumnListNames(&m)
				m.saveConfig()
			}
		}
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "esc":
			// Close color picker or settings panel
			if m.sidebarFocused {
				if m.colorPickerActive {
					m.colorPickerActive = false
					return m, nil
				}
				m.sidebarFocused = false
				return m, nil
			}

		case "tab":
			if !m.sidebarFocused {
				// Switch view mode
				if m.viewMode == ColumnView {
					m.viewMode = ListView
					m.updateTableData()
				} else {
					m.viewMode = ColumnView
					// Initialize column scrolls for new view
					cols := getColumns(&m)
					m.columnScrolls = make([]int, len(cols))
				}
				m.cursor = 0
				m.columnCursor = 0
				m.scrollOffset = 0
			} else {
				// Tab switches between sidebar sections
				if m.sidebarSection == SidebarDaysFilter {
					m.sidebarSection = SidebarListFilter
				} else if m.sidebarSection == SidebarListFilter {
					m.sidebarSection = SidebarColorConfig
				} else {
					m.sidebarSection = SidebarDaysFilter
				}
				m.sidebarCursor = 0
				m.colorPickerActive = false
			}

		case "s":
			m.sidebarFocused = !m.sidebarFocused
			m.sidebarCursor = 0
			m.colorPickerActive = false

		case "r":
			m.loading = true
			return m, loadRemindersCmd

		case "up", "k":
			if m.sidebarFocused {
				if m.sidebarCursor > 0 {
					m.sidebarCursor--
				}
			} else if m.viewMode == ListView {
				// Move cursor up in table
				m.table.CursorUp()
			} else {
				// Column view - move within column
				if m.cursor > 0 {
					m.cursor--
				}
				// Keep cursor visible in column view
				{
					columns := getColumns(&m)
					if len(m.columnScrolls) != len(columns) {
						m.columnScrolls = make([]int, len(columns))
					}
					if m.columnCursor < len(columns) {

						visibleHeight := m.getContentHeight()
						cursorLine := m.cursor * 4
						scroll := m.columnScrolls[m.columnCursor]
						if cursorLine < scroll {
							m.columnScrolls[m.columnCursor] = cursorLine
						}
						if cursorLine >= scroll+visibleHeight {
							newScroll := cursorLine - visibleHeight + 4
							if newScroll < 0 {
								newScroll = 0
							}
							m.columnScrolls[m.columnCursor] = newScroll
						}
					}
				}
			}

		case "down", "j":
			if m.sidebarFocused {
				maxCursor := 0
				if m.sidebarSection == SidebarDaysFilter {
					maxCursor = len(m.daysFilterOptions) - 1
				} else if m.sidebarSection == SidebarListFilter {
					maxCursor = len(m.availableLists) - 1
				} else if m.sidebarSection == SidebarColorConfig {
					maxCursor = len(m.availableLists) - 1
				}
				if m.sidebarCursor < maxCursor {
					m.sidebarCursor++
				}
			} else if m.viewMode == ListView {
				// Move cursor down in table
				m.table.CursorDown()
			} else {
				// Column view - move within column
				columns := getColumns(&m)
				if m.columnCursor < len(columns) {
					col := columns[m.columnCursor]
					if m.cursor < len(col)-1 {
						m.cursor++
					}
					// Keep cursor visible in column view
					visibleHeight := m.getContentHeight()
					cursorLine := m.cursor * 4
					if len(m.columnScrolls) != len(columns) {
						m.columnScrolls = make([]int, len(columns))
					}
					scroll := m.columnScrolls[m.columnCursor]
					if cursorLine < scroll {
						m.columnScrolls[m.columnCursor] = cursorLine
					}
					if cursorLine >= scroll+visibleHeight {
						newScroll := cursorLine - visibleHeight + 4
						if newScroll < 0 {
							newScroll = 0
						}
						m.columnScrolls[m.columnCursor] = newScroll
					}
				}
			}

		case "left", "h":
			if !m.sidebarFocused {
				if m.viewMode == ColumnView {
					if m.columnCursor > 0 {
						m.columnCursor--
						m.cursor = 0
						// Reset scroll for the newly focused column
						columns := getColumns(&m)
						if len(m.columnScrolls) != len(columns) {
							m.columnScrolls = make([]int, len(columns))
						}
						if m.columnCursor < len(m.columnScrolls) {
							m.columnScrolls[m.columnCursor] = 0
						}
					}
				} else if m.viewMode == ListView {
					// Move cursor left in table
					m.table.CursorLeft()
				}
			}

		case "right", "l":
			if !m.sidebarFocused {
				if m.viewMode == ColumnView {
					columns := getColumns(&m)
					if m.columnCursor < len(columns)-1 {
						m.columnCursor++
						m.cursor = 0
						// Reset scroll for the newly focused column
						if len(m.columnScrolls) != len(columns) {
							m.columnScrolls = make([]int, len(columns))
						}
						if m.columnCursor < len(m.columnScrolls) {
							m.columnScrolls[m.columnCursor] = 0
						}
					}
				} else if m.viewMode == ListView {
					// Move cursor right in table
					m.table.CursorRight()
				}
			}

		case "ctrl+left", "ctrl+h":
			// Reorder columns - move current column left
			if !m.sidebarFocused && m.viewMode == ColumnView {
				if m.columnCursor > 0 {
					listNames := getColumnListNames(&m)
					if len(m.columnOrder) == 0 {
						m.columnOrder = listNames
					}
					// Swap with previous
					m.columnOrder[m.columnCursor], m.columnOrder[m.columnCursor-1] =
						m.columnOrder[m.columnCursor-1], m.columnOrder[m.columnCursor]
					m.columnCursor--
					// Rebuild column scrolls to match new order
					columns := getColumns(&m)
					m.columnScrolls = make([]int, len(columns))
					m.saveConfig()
				}
			}

		case "ctrl+right", "ctrl+l":
			// Reorder columns - move current column right
			if !m.sidebarFocused && m.viewMode == ColumnView {
				listNames := getColumnListNames(&m)
				if len(m.columnOrder) == 0 {
					m.columnOrder = listNames
				}
				if m.columnCursor < len(m.columnOrder)-1 {
					// Swap with next
					m.columnOrder[m.columnCursor], m.columnOrder[m.columnCursor+1] =
						m.columnOrder[m.columnCursor+1], m.columnOrder[m.columnCursor]
					m.columnCursor++
					// Rebuild column scrolls to match new order
					columns := getColumns(&m)
					m.columnScrolls = make([]int, len(columns))
					m.saveConfig()
				}
			}

		case "1", "2", "3", "4", "5", "6", "7", "8", "9", "0":
			// Numeric color picker (requires selecting a list first)
			if m.sidebarFocused && m.sidebarSection == SidebarColorConfig && m.colorPickerActive {
				// Convert key to color index
				colorIdx := -1
				switch msg.String() {
				case "1":
					colorIdx = 0
				case "2":
					colorIdx = 1
				case "3":
					colorIdx = 2
				case "4":
					colorIdx = 3
				case "5":
					colorIdx = 4
				case "6":
					colorIdx = 5
				case "7":
					colorIdx = 6
				case "8":
					colorIdx = 7
				case "9":
					colorIdx = 8
				case "0":
					colorIdx = 9
				}
				if colorIdx >= 0 && colorIdx < len(m.availableColors) && m.colorPickerList < len(m.availableLists) {
					listName := m.availableLists[m.colorPickerList]
					m.colorPickerCursor = colorIdx
					m.listColors[listName] = m.availableColors[colorIdx]
					m.updateTableData()
					m.saveConfig()
					// Move to next list for quick editing and exit picker
					if m.sidebarCursor < len(m.availableLists)-1 {
						m.sidebarCursor++
					}
					m.colorPickerActive = false
				}
			}

		case " ", "enter":
			if m.sidebarFocused {
				if m.sidebarSection == SidebarDaysFilter {
					if m.sidebarCursor < len(m.daysFilterOptions) {
						m.daysFilter = m.daysFilterOptions[m.sidebarCursor]
						m.cursor = 0
						m.scrollOffset = 0
						// Reset column scrolls
						columns := getColumns(&m)
						m.columnScrolls = make([]int, len(columns))
						m.updateTableData()
						m.saveConfig()
					}
				} else if m.sidebarSection == SidebarListFilter {
					if m.sidebarCursor < len(m.availableLists) {
						listName := m.availableLists[m.sidebarCursor]
						m.selectedLists[listName] = !m.selectedLists[listName]
						m.cursor = 0
						m.scrollOffset = 0
						// Reset column scrolls
						columns := getColumns(&m)
						m.columnScrolls = make([]int, len(columns))
						m.updateTableData()
						m.saveConfig()
					}
				} else if m.sidebarSection == SidebarColorConfig {
					// Step 1: choose a list to recolor
					if !m.colorPickerActive {
						if m.sidebarCursor < len(m.availableLists) {
							m.colorPickerActive = true
							m.colorPickerList = m.sidebarCursor
							// initialize cursor to current color index if possible
							listName := m.availableLists[m.colorPickerList]
							current := m.listColors[listName]
							idx := -1
							for i, c := range m.availableColors {
								if c == current {
									idx = i
									break
								}
							}
							if idx >= 0 {
								m.colorPickerCursor = idx
							} else {
								m.colorPickerCursor = 0
							}
						}
					} else {
						// If already active, pressing enter cancels selection
						m.colorPickerActive = false
					}
				}
			}
		}
	}

	return m, nil
}
