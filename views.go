package main

import (
	"fmt"
	"time"

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

// View renders the entire application UI
func (m model) View() string {
	// Handle special states
	if m.loading {
		return renderLoading(m.width, m.height)
	}
	if m.err != nil {
		return renderError(m.err, m.width, m.height)
	}

	// Calculate dimensions - reserve space for padding, footer
	topPadding := 3
	contentHeight := m.height - topPadding - 1 // Reserve 5 lines for top padding, 1 for footer
	if contentHeight < 1 {
		contentHeight = 1
	}

	// Render footer
	footer := m.renderFooter()

	// Add top padding
	var paddingLines []string
	for i := 0; i < topPadding; i++ {
		paddingLines = append(paddingLines, "")
	}
	topPaddingStr := lipgloss.JoinVertical(lipgloss.Left, paddingLines...)

	// If settings panel is open, render overlay layout
	if m.sidebarFocused {
		content := m.renderSettingsOverlay(contentHeight, footer)
		return lipgloss.JoinVertical(lipgloss.Left, topPaddingStr, content)
	}

	// Normal layout: main content + sidebar
	content := m.renderNormalLayout(contentHeight, footer)
	return lipgloss.JoinVertical(lipgloss.Left, topPaddingStr, content)
}

// renderNormalLayout renders the default layout with sidebar
func (m model) renderNormalLayout(contentHeight int, footer string) string {
	// Calculate sidebar width - make it twice as wide
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

	// Render main content (column view or list view)
	var mainContent string
	if m.viewMode == ColumnView {
		mainContent = m.renderColumnView(mainWidth, contentHeight)
	} else {
		mainContent = m.renderListView(mainWidth, contentHeight)
	}

	// Render sidebar
	sidebar := m.renderSidebar(sidebarWidth, contentHeight)

	// Right-align sidebar by placing it after main content
	content := lipgloss.JoinHorizontal(
		lipgloss.Top,
		mainContent,
		sidebar,
	)

	return lipgloss.JoinVertical(lipgloss.Left, content, footer)
}

// renderSettingsOverlay renders settings as an overlay
func (m model) renderSettingsOverlay(contentHeight int, footer string) string {
	// Render settings panel
	settingsPanel := m.renderSettingsPanel()

	// Overlay settings in center
	content := lipgloss.Place(
		m.width,
		contentHeight,
		lipgloss.Center,
		lipgloss.Center,
		settingsPanel,
		lipgloss.WithWhitespaceChars(" "),
		lipgloss.WithWhitespaceForeground(lipgloss.Color("235")),
	)

	return lipgloss.JoinVertical(lipgloss.Left, content, footer)
}

// renderColumnView renders the column/card view
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

	// Header inside the box at the top
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(listColor)).
		Width(width - 2).
		Align(lipgloss.Left)

	header := headerStyle.Render(listName)

	// Calculate visible cards
	// Each card: border(2) + title(1) + due(1) + countdown(1) = 5 lines
	linesPerCard := 5
	// Column overhead: header(1) + blank after header(1) + footer(1) = 3 lines
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
	cardWidth := width - 2

	for i := startIdx; i < endIdx; i++ {
		reminder := reminders[i]
		isFocused := focused && (i == m.cursor)
		card := m.renderCard(reminder, cardWidth, isFocused, listColor)
		cards = append(cards, card)
	}

	// Footer with count in the right
	footerStyle := lipgloss.NewStyle().
		Width(width - 2).
		Foreground(lipgloss.Color("245")).
		Align(lipgloss.Right)

	columnFooter := footerStyle.Render(fmt.Sprintf("%d of %d", startIdx+1, len(reminders)))

	// Build column content with header inside
	var lines []string
	lines = append(lines, header)
	lines = append(lines, "")

	if len(cards) > 0 {
		lines = append(lines, cards...)
	} else {
		emptyMsg := lipgloss.NewStyle().
			Foreground(lipgloss.Color("245")).
			Italic(true).
			Render("No items")
		lines = append(lines, emptyMsg)
	}

	lines = append(lines, columnFooter)

	content := lipgloss.JoinVertical(lipgloss.Left, lines...)

	// Wrap in border with full height
	borderColor := "240"
	if focused {
		borderColor = listColor
	}

	columnStyle := lipgloss.NewStyle().
		Width(width).
		Height(height).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color(borderColor))

	return columnStyle.Render(content)
}

// renderCard renders a single reminder card
func (m model) renderCard(reminder Reminder, width int, focused bool, listColor string) string {
	titleStyle := lipgloss.NewStyle().Bold(true)
	metaStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("245"))

	countdown, urgency := getCountdown(reminder.DueDate)
	urgencyColor := getUrgencyColor(urgency)
	countdownStyle := lipgloss.NewStyle().Foreground(urgencyColor)

	// Title
	title := reminder.Title
	if focused {
		title = "▸ " + title
	}

	// Due date
	dueText := formatDueDate(reminder.DueDate)

	// Build card content
	var lines []string
	lines = append(lines, titleStyle.Render(title))
	lines = append(lines, metaStyle.Render(dueText))
	lines = append(lines, countdownStyle.Render(countdown))
	
	content := lipgloss.JoinVertical(lipgloss.Left, lines...)

	// Card styling
	borderColor := "240"
	if focused {
		borderColor = "205"
	}

	cardStyle := lipgloss.NewStyle().
		Width(width).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color(borderColor))

	return cardStyle.Render(content)
}

// renderListView renders the compact list view
func (m model) renderListView(width, height int) string {
	// Debug: show info about what we have
	if len(m.reminders) == 0 {
		return renderEmpty(width, height, "No reminders loaded (press 'r' to refresh)")
	}
	
	filtered := getFilteredReminders(&m)

	if len(filtered) == 0 {
		return renderEmpty(width, height, fmt.Sprintf("No reminders match filter (have %d total, check settings 's')", len(m.reminders)))
	}

	metaStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	
	var lines []string

	// Calculate visible range
	visibleStart := m.scrollOffset
	visibleEnd := m.scrollOffset + height

	for i, reminder := range filtered {
		if i < visibleStart {
			continue
		}
		if i >= visibleEnd {
			break
		}

		// Get info
		countdown, urgency := getCountdown(reminder.DueDate)
		urgencyStyle := lipgloss.NewStyle().Foreground(getUrgencyColor(urgency))

		listColor := "248"
		if color, exists := m.listColors[reminder.List]; exists {
			listColor = color
		}
		listStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(listColor))

		dueText := formatDueDate(reminder.DueDate)

		// Format: "Item 1  list 1  in 1 day  due ..."
		line := fmt.Sprintf("%s  %s  %s  %s",
			reminder.Title,
			listStyle.Render(reminder.List),
			urgencyStyle.Render(countdown),
			metaStyle.Render("due "+dueText))

		// Highlight if focused
		if i == m.cursor {
			line = lipgloss.NewStyle().Bold(true).Render("▸ " + line)
		} else {
			line = "  " + line
		}

		lines = append(lines, line)
	}

	// Pad to fill height
	for len(lines) < height {
		lines = append(lines, "")
	}

	// Join and style
	content := lipgloss.JoinVertical(lipgloss.Left, lines...)
	
	return content
}

// renderSidebar renders the sidebar with settings button and calendar
func (m model) renderSidebar(width, height int) string {
	// Calculate heights for sections
	settingsButtonHeight := 9
	// Account for blank line between sections
	calendarHeight := height - settingsButtonHeight - 2

	if calendarHeight < 8 {
		calendarHeight = 8
		settingsButtonHeight = height - calendarHeight - 2
	}

	settingsButton := m.renderSettingsButton(width, settingsButtonHeight)
	calendar := m.renderCalendar(width, calendarHeight)

	// Join sections vertically
	content := lipgloss.JoinVertical(lipgloss.Left,
		settingsButton,
		"",
		calendar,
	)

	return content
}

// renderSettingsButton renders the settings button
func (m model) renderSettingsButton(width, height int) string {
	text := "View settings\n\n(e.g. next n days,\nwhich lists)\n\nPress 's' to open"

	buttonStyle := lipgloss.NewStyle().
		Width(width).
		Height(height).
		Padding(1).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("141")).
		Foreground(lipgloss.Color("248")).
		Align(lipgloss.Center, lipgloss.Center)

	return buttonStyle.Render(text)
}

// renderCalendar renders the small calendar widget
func (m model) renderCalendar(width, height int) string {
	now := time.Now()

	// Create calendar title
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("141"))

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
		"",
		day,
		weekday,
	)

	calendarStyle := lipgloss.NewStyle().
		Width(width).
		Height(height).
		Padding(1).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("141")).
		Align(lipgloss.Center, lipgloss.Center)

	return calendarStyle.Render(content)
}

// renderSettingsPanel renders the settings overlay panel
func (m model) renderSettingsPanel() string {
	panelWidth := 50
	if panelWidth > m.width*2/3 {
		panelWidth = m.width * 2 / 3
	}

	panelHeight := 30
	if panelHeight > m.height*3/4 {
		panelHeight = m.height * 3 / 4
	}

	// Header
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205")).
		MarginBottom(1)

	header := headerStyle.Render("settings menu")

	// Build content sections
	content := header + "\n\n"
	content += m.renderDaysFilterSection() + "\n\n"
	content += m.renderListFilterSection() + "\n\n"
	content += m.renderColorConfigSection()

	// Panel styling
	panelStyle := lipgloss.NewStyle().
		Width(panelWidth).
		Height(panelHeight).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("205")).
		Padding(1, 2).
		Background(lipgloss.Color("235"))

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

	if m.sidebarSection == SidebarDaysFilter {
		lines = append(lines, focusedStyle.Render("▸ Days Filter"))
	} else {
		lines = append(lines, titleStyle.Render("  Days Filter"))
	}

	for i, days := range m.daysFilterOptions {
		label := "All"
		if days > 0 {
			label = fmt.Sprintf("Next %d days", days)
		}

		cursor := "  "
		if m.sidebarSection == SidebarDaysFilter && m.sidebarCursor == i {
			cursor = "▸ "
		}

		check := "○"
		if days == m.daysFilter {
			check = "●"
		}

		line := fmt.Sprintf("  %s%s %s", cursor, check, label)

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

	if m.sidebarSection == SidebarListFilter {
		lines = append(lines, focusedStyle.Render("▸ Visible Lists"))
	} else {
		lines = append(lines, titleStyle.Render("  Visible Lists"))
	}

	for i, listName := range m.availableLists {
		cursor := "  "
		if m.sidebarSection == SidebarListFilter && m.sidebarCursor == i {
			cursor = "▸ "
		}

		check := "☐"
		if m.selectedLists[listName] {
			check = "☑"
		}

		listColor := "248"
		if color, exists := m.listColors[listName]; exists {
			listColor = color
		}
		nameStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(listColor))

		line := fmt.Sprintf("  %s%s %s", cursor, check, nameStyle.Render(listName))

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

	var lines []string

	if m.sidebarSection == SidebarColorConfig {
		lines = append(lines, focusedStyle.Render("▸ List Colors"))
		infoStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("245")).
			Italic(true)
		lines = append(lines, infoStyle.Render("  Press 1-9, 0 to set color"))
	} else {
		lines = append(lines, titleStyle.Render("  List Colors"))
	}

	for i, listName := range m.availableLists {
		cursor := "  "
		if m.sidebarSection == SidebarColorConfig && m.sidebarCursor == i {
			cursor = "▸ "
		}

		currentColor := "248"
		if color, exists := m.listColors[listName]; exists {
			currentColor = color
		}

		colorBlock := lipgloss.NewStyle().
			Foreground(lipgloss.Color(currentColor)).
			Render("●")

		line := fmt.Sprintf("  %s%s %s", cursor, colorBlock, listName)

		if m.sidebarSection == SidebarColorConfig && m.sidebarCursor == i {
			line = focusedStyle.Render(line)
		} else {
			line = normalStyle.Render(line)
		}

		lines = append(lines, line)
	}

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

// renderFooter renders the footer with help text
func (m model) renderFooter() string {
	helpText := "tab: view • s: settings • ←/→: columns • ↑/↓: items • r: refresh • q: quit"
	
	if m.sidebarFocused {
		helpText = "esc: close • tab: section • space: toggle • 1-9/0: color • ↑/↓: navigate • q: quit"
	} else if m.viewMode == ListView {
		helpText = "tab: view • s: settings • ↑/↓: items • r: refresh • q: quit"
	}

	footerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Background(lipgloss.Color("235")).
		Padding(0, 1).
		Width(m.width)

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
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Foreground(lipgloss.Color("245")).
		Italic(true).
		Align(lipgloss.Center, lipgloss.Center)

	return style.Render(message)
}
