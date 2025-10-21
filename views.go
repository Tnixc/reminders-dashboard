package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/76creates/stickers/flexbox"
	"github.com/charmbracelet/lipgloss"
)

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (m *model) getContentHeight() int {
	return m.height - 1 // minus footer
}

// View renders the entire application UI using flexbox
func (m model) View() string {
	// Handle special states
	if m.loading {
		return renderLoading(m.width, m.height)
	}
	if m.err != nil {
		return renderError(m.err, m.width, m.height)
	}

	// If settings panel is open, render overlay layout
	if m.sidebarFocused {
		return m.renderSettingsOverlay()
	}

	// Normal layout: render using flexbox
	return m.renderNormalLayout()
}

// renderNormalLayout renders the default layout with flexbox
func (m model) renderNormalLayout() string {
	// Create new flexbox for this render
	styleBackground := lipgloss.NewStyle().Align(lipgloss.Center)
	flex := flexbox.New(m.width, m.height).SetStyle(styleBackground)

	// Create three rows: top padding, content, footer
	topPadding := 0
	footerHeight := 1
	contentHeight := m.height - topPadding - footerHeight

	// Row 1: Top padding
	paddingRow := flex.NewRow().AddCells(
		flexbox.NewCell(1, topPadding).SetContent(""),
	)

	// Row 2: Main content (main view + sidebar)
	sidebarWidth := 48
	if m.width < 120 {
		sidebarWidth = 40
	}
	if sidebarWidth > m.width/3 {
		sidebarWidth = m.width / 3
	}
	if sidebarWidth < 36 {
		sidebarWidth = 36
	}

	mainWidth := m.width - sidebarWidth

	// Create main content
	var mainContent string
	if m.viewMode == ColumnView {
		mainContent = m.renderColumnView(mainWidth, contentHeight)
	} else {
		mainContent = m.renderListView(mainWidth, contentHeight)
	}

	// Create sidebar content
	sidebarContent := m.renderSidebar(sidebarWidth, contentHeight)

	// Main content row with two cells
	contentRow := flex.NewRow().AddCells(
		flexbox.NewCell(mainWidth, contentHeight).SetContent(mainContent),
		flexbox.NewCell(sidebarWidth, contentHeight).SetContent(sidebarContent),
	)

	// Row 3: Footer
	footerContent := m.renderFooter()
	footerRow := flex.NewRow().AddCells(
		flexbox.NewCell(1, footerHeight).SetContent(footerContent),
	)

	// Add all rows to flexbox
	flex.AddRows([]*flexbox.Row{paddingRow, contentRow, footerRow})

	// Force recalculate and render
	flex.ForceRecalculate()
	return flex.Render()
}

// renderSettingsOverlay renders the settings panel overlay using flexbox
func (m model) renderSettingsOverlay() string {
	// Create new flexbox for this render
	styleBackground := lipgloss.NewStyle().Align(lipgloss.Center)
	flex := flexbox.New(m.width, m.height).SetStyle(styleBackground)

	topPadding := 3
	footerHeight := 1
	contentHeight := m.height - topPadding - footerHeight

	// Row 1: Top padding
	paddingRow := flex.NewRow().AddCells(
		flexbox.NewCell(1, topPadding).SetContent(""),
	)

	// Row 2: Main content with overlay
	// Background content (main view + sidebar)
	sidebarWidth := 48
	if m.width < 120 {
		sidebarWidth = 40
	}
	mainWidth := m.width - sidebarWidth

	var mainContent string
	if m.viewMode == ColumnView {
		mainContent = m.renderColumnView(mainWidth, contentHeight)
	} else {
		mainContent = m.renderListView(mainWidth, contentHeight)
	}

	sidebarContent := m.renderSidebar(sidebarWidth, contentHeight)
	
	// Combine background
	background := lipgloss.JoinHorizontal(lipgloss.Top, mainContent, sidebarContent)

	// Settings panel (centered overlay)
	settingsPanel := m.renderSettingsPanel()
	
	// Calculate overlay position
	panelWidth := lipgloss.Width(settingsPanel)
	panelHeight := lipgloss.Height(settingsPanel)
	
	leftPadding := (m.width - panelWidth) / 2
	topOverlayPadding := (contentHeight - panelHeight) / 2
	
	if leftPadding < 0 {
		leftPadding = 0
	}
	if topOverlayPadding < 0 {
		topOverlayPadding = 0
	}

	// Create overlay by positioning the settings panel
	var overlayLines []string
	backgroundLines := strings.Split(background, "\n")
	settingsPanelLines := strings.Split(settingsPanel, "\n")

	for i := 0; i < contentHeight; i++ {
		if i >= topOverlayPadding && i < topOverlayPadding+panelHeight {
			// Overlay the settings panel
			settingsLineIdx := i - topOverlayPadding
			if settingsLineIdx < len(settingsPanelLines) {
				line := ""
				if i < len(backgroundLines) {
					bgLine := backgroundLines[i]
					if len(bgLine) > leftPadding {
						line = bgLine[:leftPadding]
					} else {
						line = bgLine + strings.Repeat(" ", leftPadding-len(bgLine))
					}
				} else {
					line = strings.Repeat(" ", leftPadding)
				}
				line += settingsPanelLines[settingsLineIdx]
				// Fill the rest of the line
				currentLen := lipgloss.Width(line)
				if currentLen < m.width {
					line += strings.Repeat(" ", m.width-currentLen)
				}
				overlayLines = append(overlayLines, line)
			} else if i < len(backgroundLines) {
				overlayLines = append(overlayLines, backgroundLines[i])
			}
		} else if i < len(backgroundLines) {
			overlayLines = append(overlayLines, backgroundLines[i])
		}
	}

	overlayContent := strings.Join(overlayLines, "\n")

	contentRow := flex.NewRow().AddCells(
		flexbox.NewCell(1, contentHeight).SetContent(overlayContent),
	)

	// Row 3: Footer
	footerContent := m.renderFooter()
	footerRow := flex.NewRow().AddCells(
		flexbox.NewCell(1, footerHeight).SetContent(footerContent),
	)

	flex.AddRows([]*flexbox.Row{paddingRow, contentRow, footerRow})
	flex.ForceRecalculate()
	return flex.Render()
}

// renderColumnView renders the column view
func (m model) renderColumnView(width, height int) string {
	columns := getColumns(&m)

	// Debug: show info about what we have
	if len(m.reminders) == 0 {
		return renderEmpty(width, height, "No reminders loaded (press 'r' to refresh)")
	}

	filtered := getFilteredReminders(&m)
	if len(filtered) == 0 {
		return renderEmpty(width, height, fmt.Sprintf("No reminders match filter (have %d total, check settings)", len(m.reminders)))
	}

	if len(columns) == 0 {
		return renderEmpty(width, height, fmt.Sprintf("No columns to display (%d filtered reminders)", len(filtered)))
	}

	// Calculate column dimensions
	minCardWidth := 20
	maxCardWidth := 36
	cardPadding := 2

	numColumns := len(columns)
	if numColumns > 6 {
		numColumns = 6 // Max 6 visible columns
	}

	totalPadding := (numColumns - 1) * cardPadding
	columnWidth := (width - totalPadding) / numColumns

	if columnWidth < minCardWidth {
		numColumns = width / (minCardWidth + cardPadding)
		if numColumns < 1 {
			numColumns = 1
		}
		columnWidth = (width - (numColumns-1)*cardPadding) / numColumns
	}
	if columnWidth > maxCardWidth {
		columnWidth = maxCardWidth
	}

	// Render columns
	var renderedColumns []string
	visibleCount := min(numColumns, len(columns))

	for i := 0; i < visibleCount; i++ {
		column := columns[i]
		scrollOffset := 0
		if i < len(m.columnScrolls) {
			scrollOffset = m.columnScrolls[i]
		}

		listName := ""
		if len(column) > 0 {
			listName = column[0].List
		}

		columnStr := m.renderColumn(column, listName, columnWidth, height, i == m.columnCursor, scrollOffset)
		renderedColumns = append(renderedColumns, columnStr)
	}

	// Show indicator if more columns exist
	if visibleCount < len(columns) {
		remaining := len(columns) - visibleCount
		indicator := renderEmpty(columnWidth, height, fmt.Sprintf("+ %d more", remaining))
		renderedColumns = append(renderedColumns, indicator)
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, renderedColumns...)
}

// renderColumn renders a single column with cards
func (m model) renderColumn(reminders []Reminder, listName string, width, height int, focused bool, scrollOffset int) string {
	// Get list color
	listColor := "248"
	if color, exists := m.listColors[listName]; exists {
		listColor = color
	}

	// Header with list name and count
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(listColor))

	countStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("245"))

	header := headerStyle.Render(listName) + " " + countStyle.Render(fmt.Sprintf("(%d)", len(reminders)))

	// Calculate visible cards
	linesPerCard := 5
	columnOverhead := 3

	availableHeight := height - columnOverhead
	if availableHeight < linesPerCard {
		availableHeight = linesPerCard
	}
	visibleCards := availableHeight / linesPerCard
	if visibleCards < 1 {
		visibleCards = 1
	}

	startIdx := scrollOffset / linesPerCard
	if startIdx < 0 {
		startIdx = 0
	}
	endIdx := startIdx + visibleCards
	if endIdx > len(reminders) {
		endIdx = len(reminders)
	}

	// Keep focused item visible
	if focused && m.cursor < len(reminders) {
		cursorIdx := m.cursor
		if cursorIdx < startIdx {
			startIdx = cursorIdx
			endIdx = startIdx + visibleCards
			if endIdx > len(reminders) {
				endIdx = len(reminders)
			}
		}
		if cursorIdx >= endIdx {
			endIdx = cursorIdx + 1
			startIdx = endIdx - visibleCards
			if startIdx < 0 {
				startIdx = 0
			}
		}
	}

	// Render cards
	var cards []string
	cardWidth := width - 4 // Account for padding

	for i := startIdx; i < endIdx; i++ {
		reminder := reminders[i]
		isFocused := focused && (i == m.cursor)
		card := m.renderCard(reminder, cardWidth, isFocused, listColor)
		cards = append(cards, card)
	}

	// Build column content with header inside
	var lines []string
	lines = append(lines, header)
	lines = append(lines, "")

	if len(cards) > 0 {
		lines = append(lines, cards...)
	} else {
		emptyMsg := lipgloss.NewStyle().
			Foreground(lipgloss.Color("183")).
			Italic(true).
			Render("No items")
		lines = append(lines, emptyMsg)
	}

	// Footer with scroll indicator
	if len(reminders) > visibleCards {
		footerStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("245")).
			Italic(true)
		lines = append(lines, "")
		lines = append(lines, footerStyle.Render(fmt.Sprintf("  ↕ %d-%d of %d", startIdx+1, endIdx, len(reminders))))
	}

	content := lipgloss.JoinVertical(lipgloss.Left, lines...)

	// Column styling with square border
	columnStyle := lipgloss.NewStyle().
		Width(width).
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("237"))

	return columnStyle.Render(content)
}

// renderCard renders a single reminder card with modern styling
func (m model) renderCard(reminder Reminder, width int, focused bool, listColor string) string {
	countdown, urgency := getCountdown(reminder.DueDate)
	urgencyColor := getUrgencyColor(urgency)

	// Title with fixed width to prevent layout shifts
	cursor := "  "
	titleColor := lipgloss.Color("255")
	if focused {
		cursor = "▸ "
		titleColor = lipgloss.Color(listColor)
	}

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(titleColor)

	title := cursor + reminder.Title

	// Due date with subtle styling
	dueText := formatDueDate(reminder.DueDate)
	metaStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("245")).
		Italic(true)

	// Countdown with urgency color
	countdownStyle := lipgloss.NewStyle().
		Foreground(urgencyColor).
		Bold(true)

	// Build card content
	var lines []string
	lines = append(lines, titleStyle.Render(title))
	lines = append(lines, metaStyle.Render("  "+dueText))
	lines = append(lines, "  "+countdownStyle.Render(countdown))

	content := lipgloss.JoinVertical(lipgloss.Left, lines...)

	return content
}

// updateTableData updates the table with current filtered reminders
func (m *model) updateTableData() {
	filtered := getFilteredReminders(m)

	// Clear existing rows
	m.table.ClearRows()

	rows := make([][]any, 0, len(filtered))
	for _, reminder := range filtered {
		countdown, urgency := getCountdown(reminder.DueDate)
		dueText := formatDueDate(reminder.DueDate)

		// Get list color
		listColor := "248"
		if color, exists := m.listColors[reminder.List]; exists {
			listColor = color
		}

		// Create badge for list
		listName := reminder.List
		if len(listName) > 11 {
			listName = listName[:10] + "…"
		}

		badgeStyle := lipgloss.NewStyle().
			Background(lipgloss.Color(listColor)).
			Foreground(lipgloss.Color("0")).
			Bold(true).
			Padding(0, 1)

		listBadge := badgeStyle.Render(listName)

		// Style countdown with urgency color
		urgencyColor := getUrgencyColor(urgency)
		countdownStyle := lipgloss.NewStyle().
			Foreground(urgencyColor).
			Bold(true)
		countdownText := countdownStyle.Render(countdown)

		rows = append(rows, []any{
			reminder.Title,
			listBadge,
			countdownText,
			dueText,
		})
	}

	m.table.AddRows(rows)
}

// renderListView renders the list view using stickers table component
func (m model) renderListView(width, height int) string {
	// Debug: show info about what we have
	if len(m.reminders) == 0 {
		return renderEmpty(width, height, "No reminders loaded (press 'r' to refresh)")
	}

	filtered := getFilteredReminders(&m)

	if len(filtered) == 0 {
		return renderEmpty(width, height, fmt.Sprintf("No reminders match filter (have %d total, check settings 's')", len(m.reminders)))
	}

	// Update table data
	m.updateTableData()

	// Set table dimensions
	tableWidth := width - 6
	tableHeight := height - 6

	if tableWidth < 40 {
		tableWidth = 40
	}
	if tableHeight < 5 {
		tableHeight = 5
	}

	m.table.SetWidth(tableWidth)
	m.table.SetHeight(tableHeight)

	// Add header
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("255"))

	header := headerStyle.Render(fmt.Sprintf("# Reminders (%d)", len(filtered)))

	// Render table
	tableView := m.table.Render()

	// Combine header and table
	content := lipgloss.JoinVertical(lipgloss.Left,
		header,
		"",
		tableView,
	)

	// Add overall border
	containerStyle := lipgloss.NewStyle().
		Width(width - 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(0, 1)

	return containerStyle.Render(content)
}

// renderSidebar renders the sidebar with settings button and calendar using flexbox
func (m model) renderSidebar(width, height int) string {
	// Create a temporary flexbox for sidebar layout
	sidebarFlex := flexbox.New(width, height)

	// Calculate heights for sections
	settingsButtonHeight := 9
	blankHeight := 1
	calendarHeight := height - settingsButtonHeight - blankHeight

	if calendarHeight < 8 {
		calendarHeight = 8
		settingsButtonHeight = height - calendarHeight - blankHeight
	}

	// Create settings button content
	settingsContent := m.renderSettingsButton(width, settingsButtonHeight)

	// Create calendar content
	calendarContent := m.renderCalendar(width, calendarHeight)

	// Create rows for sidebar
	settingsRow := sidebarFlex.NewRow().AddCells(
		flexbox.NewCell(1, settingsButtonHeight).SetContent(settingsContent),
	)

	blankRow := sidebarFlex.NewRow().AddCells(
		flexbox.NewCell(1, blankHeight).SetContent(""),
	)

	calendarRow := sidebarFlex.NewRow().AddCells(
		flexbox.NewCell(1, calendarHeight).SetContent(calendarContent),
	)

	sidebarFlex.AddRows([]*flexbox.Row{settingsRow, blankRow, calendarRow})
	sidebarFlex.ForceRecalculate()

	return sidebarFlex.Render()
}

// renderSettingsButton renders the settings button
func (m model) renderSettingsButton(width, height int) string {
	text := "Settings\n\n(next n days,\nfilter lists)\n\nPress 's'"

	buttonStyle := lipgloss.NewStyle().
		Width(width).
		Height(height).
		Foreground(lipgloss.Color("183")).
		Align(lipgloss.Center, lipgloss.Center).
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("237"))

	return buttonStyle.Render(text)
}

// renderCalendar renders the small calendar widget
func (m model) renderCalendar(width, height int) string {
	now := time.Now()

	// Create calendar title
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("183"))

	title := titleStyle.Render(now.Format("January 2006"))

	// Create day number
	dayStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205")).
		Width(width - 2).
		Align(lipgloss.Center)

	day := dayStyle.Render(fmt.Sprintf("%d", now.Day()))

	// Create weekday
	weekdayStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("248"))

	weekday := weekdayStyle.Render(now.Format("Monday"))

	content := lipgloss.JoinVertical(lipgloss.Center,
		title,
		day,
		weekday,
	)

	calendarStyle := lipgloss.NewStyle().
		Width(width).
		Height(height).
		Align(lipgloss.Center, lipgloss.Center).
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("237"))

	return calendarStyle.Render(content)
}

// renderSettingsPanel renders the settings overlay panel
func (m model) renderSettingsPanel() string {
	panelWidth := (m.width * 9) / 10

	panelHeight := 30
	if panelHeight > m.height*3/4 {
		panelHeight = m.height * 3 / 4
	}

	// Header
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205"))

	header := headerStyle.Render("Settings")

	// Build content sections
	content := header + "\n"
	content += m.renderDaysFilterSection() + "\n"
	content += m.renderListFilterSection() + "\n"
	content += m.renderColorConfigSection()

	// Panel styling with square border
	panelStyle := lipgloss.NewStyle().
		Width(panelWidth).
		Height(panelHeight).
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("237"))

	return panelStyle.Render(content)
}

// renderDaysFilterSection renders the days filter settings
func (m model) renderDaysFilterSection() string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("141"))

	focusedStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205"))

	normalStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("248"))

	var lines []string

	sectionHeader := "Days Filter"
	if m.sidebarSection == SidebarDaysFilter {
		lines = append(lines, focusedStyle.Render("> "+sectionHeader))
	} else {
		lines = append(lines, titleStyle.Render("  "+sectionHeader))
	}

	for i, days := range m.daysFilterOptions {
		label := "All"
		if days > 0 {
			label = fmt.Sprintf("Next %d days", days)
		}

		// Fixed width cursor to prevent layout shifts
		cursor := "  "
		if m.sidebarSection == SidebarDaysFilter && m.sidebarCursor == i {
			cursor = "> "
		}

		// Radio button style
		check := "( )"
		checkColor := lipgloss.Color("248")
		if days == m.daysFilter {
			check = "(x)"
			checkColor = lipgloss.Color("141")
		}

		checkStyle := lipgloss.NewStyle().
			Foreground(checkColor)

		line := fmt.Sprintf("%s%s %s", cursor, checkStyle.Render(check), label)

		if m.sidebarSection == SidebarDaysFilter && m.sidebarCursor == i {
			line = focusedStyle.Render(line)
		} else {
			line = normalStyle.Render(line)
		}

		lines = append(lines, line)
	}

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

// renderListFilterSection renders the list filter settings
func (m model) renderListFilterSection() string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("141"))

	focusedStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205"))

	normalStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("248"))

	var lines []string

	sectionHeader := "List Filter"
	if m.sidebarSection == SidebarListFilter {
		lines = append(lines, focusedStyle.Render("> "+sectionHeader))
	} else {
		lines = append(lines, titleStyle.Render("  "+sectionHeader))
	}

	for i, listName := range m.availableLists {
		// Fixed width cursor to prevent layout shifts
		cursor := "  "
		if m.sidebarSection == SidebarListFilter && m.sidebarCursor == i {
			cursor = "> "
		}

		// Checkbox style
		check := "[ ]"
		checkColor := lipgloss.Color("248")
		if selected, exists := m.selectedLists[listName]; exists && selected {
			check = "[x]"
			checkColor = lipgloss.Color("48")
		}

		checkStyle := lipgloss.NewStyle().
			Foreground(checkColor)

		line := fmt.Sprintf("%s%s %s", cursor, checkStyle.Render(check), listName)

		if m.sidebarSection == SidebarListFilter && m.sidebarCursor == i {
			line = focusedStyle.Render(line)
		} else {
			line = normalStyle.Render(line)
		}

		lines = append(lines, line)
	}

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

// renderColorConfigSection renders the color configuration settings
func (m model) renderColorConfigSection() string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("141"))

	focusedStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205"))

	normalStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("248"))

	instructionStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("245"))

	var lines []string

	// Header + instructions
	sectionHeader := "List Colors"
	if m.sidebarSection == SidebarColorConfig {
		lines = append(lines, focusedStyle.Render("> "+sectionHeader))
		if m.colorPickerActive {
			lines = append(lines, instructionStyle.Render("  1-9/0: color, esc: cancel"))
		} else {
			lines = append(lines, instructionStyle.Render("  enter: pick, 1-9/0: color"))
		}
	} else {
		lines = append(lines, titleStyle.Render("  "+sectionHeader))
	}

	// Lists with color indicators
	for i, listName := range m.availableLists {
		currentColor := "248"
		if color, exists := m.listColors[listName]; exists {
			currentColor = color
		}

		// Color indicator
		colorStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color(currentColor))

		indicator := colorStyle.Render("*")

		// Fixed width cursor to prevent layout shifts
		cursor := "  "

		if m.sidebarSection == SidebarColorConfig && m.sidebarCursor == i {
			cursor = "> "
		}

		// Show selection indicator when active
		suffix := "  "
		if m.colorPickerActive && m.colorPickerList == i {
			suffix = " <"
		}

		line := fmt.Sprintf("%s%s %s%s", cursor, indicator, listName, suffix)

		if m.sidebarSection == SidebarColorConfig && m.sidebarCursor == i {
			line = focusedStyle.Render(line)
		} else {
			line = normalStyle.Render(line)
		}

		lines = append(lines, line)
	}

	// Color palette
	if len(m.availableColors) > 0 {
		lines = append(lines, "")

		paletteHeaderStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("183"))
		lines = append(lines, paletteHeaderStyle.Render("  Colors"))

		// Create color list
		for i := 0; i < len(m.availableColors); i++ {
			code := m.availableColors[i]
			name := ""
			if i < len(m.colorNames) {
				name = m.colorNames[i]
			}

			key := ""
			if i < 9 {
				key = fmt.Sprintf("%d", i+1)
			} else if i == 9 {
				key = "0"
			}

			// Color swatch
			swatchStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color(code))

			keyStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("245"))

			nameStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("248"))

			swatch := swatchStyle.Render("██")

			// Add selection indicator with fixed width
			suffix := "  "
			if m.colorPickerActive && m.colorPickerCursor == i {
				suffix = " <"
			}

			item := fmt.Sprintf("  %s %s %s%s", keyStyle.Render(key), swatch, nameStyle.Render(name), suffix)

			lines = append(lines, item)
		}
	}

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

// renderFooter renders the footer with help text
func (m model) renderFooter() string {
	helpText := "tab: view • s: settings • ←/→: columns • ↑/↓: items • r: refresh • q: quit"

	if m.sidebarFocused {
		helpText = "esc: close • tab: section • enter: pick list • 1-9/0: color • space: toggle • ↑/↓: navigate • q: quit"
	} else if m.viewMode == ListView {
		helpText = "tab: view • s: settings • ↑/↓: items • r: refresh • q: quit"
	}

	footerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("183")).
		Background(lipgloss.Color("235")).
		Padding(0, 1).
		Width(m.width).
		Bold(true)

	return footerStyle.Render(helpText)
}

// Helper rendering functions

func renderLoading(width, height int) string {
	content := "Loading reminders..."

	style := lipgloss.NewStyle().
		Width(width).
		Height(height).
		Align(lipgloss.Center, lipgloss.Center).
		Foreground(lipgloss.Color("248"))

	return style.Render(content)
}

func renderError(err error, width, height int) string {
	content := fmt.Sprintf("Error: %v", err)

	style := lipgloss.NewStyle().
		Width(width).
		Height(height).
		Align(lipgloss.Center, lipgloss.Center).
		Foreground(lipgloss.Color("196"))

	return style.Render(content)
}

func renderEmpty(width, height int, message string) string {
	style := lipgloss.NewStyle().
		Width(width).
		Height(height).
		Padding(1).
		Foreground(lipgloss.Color("245")).
		Italic(true).
		Align(lipgloss.Center, lipgloss.Center)

	return style.Render(message)
}
