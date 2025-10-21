package main

import (
	"fmt"
	"strings"
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

	// Insert flexible spacer so the sidebar hugs the right edge
	mainActualWidth := lipgloss.Width(mainContent)
	sidebarActualWidth := lipgloss.Width(sidebar)
	spacerWidth := m.width - mainActualWidth - sidebarActualWidth
	if spacerWidth < 0 {
		spacerWidth = 0
	}
	spacer := strings.Repeat(" ", spacerWidth)

	content := lipgloss.JoinHorizontal(
		lipgloss.Top,
		mainContent,
		spacer,
		sidebar,
	)

	return lipgloss.JoinVertical(lipgloss.Left, content, footer)
}

// renderSettingsOverlay renders settings as an overlay
func (m model) renderSettingsOverlay(contentHeight int, footer string) string {
	// Render settings panel
	settingsPanel := m.renderSettingsPanel()

	// Overlay settings in center without background fill; allow up to 90% width
	content := lipgloss.Place(
		m.width,
		contentHeight,
		lipgloss.Center,
		lipgloss.Center,
		lipgloss.NewStyle().MaxWidth((m.width*9)/10).Render(settingsPanel),
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

	// Header with list name and count
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(listColor))

	countStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("245"))

	header := headerStyle.Render(listName) + " " + countStyle.Render(fmt.Sprintf("(%d)", len(reminders)))

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

	// Column styling with square border - no change on focus
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

// renderListView renders the compact list view with table-style alignment
func (m model) renderListView(width, height int) string {
	// Debug: show info about what we have
	if len(m.reminders) == 0 {
		return renderEmpty(width, height, "No reminders loaded (press 'r' to refresh)")
	}

	filtered := getFilteredReminders(&m)

	if len(filtered) == 0 {
		return renderEmpty(width, height, fmt.Sprintf("No reminders match filter (have %d total, check settings 's')", len(m.reminders)))
	}

	var lines []string

	// Add header
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("183"))
	
	lines = append(lines, headerStyle.Render(fmt.Sprintf("All Reminders (%d)", len(filtered))))

	// Define fixed column widths for table layout
	contentWidth := width - 4
	cursorWidth := 2
	titleWidth := 30 // Fixed reasonable width for title
	if titleWidth > contentWidth-40 {
		titleWidth = contentWidth - 40
	}
	if titleWidth < 15 {
		titleWidth = 15
	}
	listWidth := 15 // Increased width for badges
	countdownWidth := 10

	// Add table header
	headerRowStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("245")).
		Bold(true)
	
	headerRow := fmt.Sprintf("%-*s  %-*s  %-*s  %s",
		titleWidth+cursorWidth, "Title",
		listWidth, "List",
		countdownWidth, "Countdown",
		"Due Date")
	lines = append(lines, headerRowStyle.Render(headerRow))
	
	// Add separator
	separatorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("237"))
	lines = append(lines, separatorStyle.Render(strings.Repeat("─", contentWidth)))

	// Calculate visible range (adjust for header)
	adjustedHeight := height - 4 // Account for header, table header, separator
	visibleStart := m.scrollOffset
	visibleEnd := m.scrollOffset + adjustedHeight

	for i, reminder := range filtered {
		if i < visibleStart {
			continue
		}
		if i >= visibleEnd {
			break
		}

		// Get info
		countdown, urgency := getCountdown(reminder.DueDate)
		urgencyColor := getUrgencyColor(urgency)

		listColor := "248"
		if color, exists := m.listColors[reminder.List]; exists {
			listColor = color
		}

		// Create badge-style list display with colored background
		badgeStyle := lipgloss.NewStyle().
			Background(lipgloss.Color(listColor)).
			Foreground(lipgloss.Color("0")).
			Bold(true).
			Padding(0, 1)
		
		// Truncate list name if too long for badge
		listName := reminder.List
		if len(listName) > listWidth-3 {
			listName = listName[:listWidth-4] + "…"
		}
		listBadge := badgeStyle.Render(listName)
	
		countdownStyle := lipgloss.NewStyle().
			Foreground(urgencyColor).
			Bold(true)

		countdownText := countdownStyle.Render(countdown)

		dueText := formatDueDate(reminder.DueDate)
		dueStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("245")).
			Italic(true)

		// Truncate title if too long
		title := reminder.Title
		if len(title) > titleWidth {
			title = title[:titleWidth-3] + "..."
		}

		// Build table row with fixed widths
		cursor := "  "
		titleColor := lipgloss.Color("255")
		
		if i == m.cursor {
			cursor = "> "
			titleColor = lipgloss.Color("183")
		}

		titleStyle := lipgloss.NewStyle().
			Foreground(titleColor).
			Bold(i == m.cursor)

		// Format as fixed-width table row
		// Note: we need to account for the badge rendering width separately
		row := fmt.Sprintf("%s%-*s  %s%s  %-*s  %s",
			cursor,
			titleWidth, titleStyle.Render(title),
			listBadge,
			strings.Repeat(" ", listWidth-len(listName)-2), // Adjust spacing after badge
			countdownWidth, countdownText,
			dueStyle.Render(dueText))

		lines = append(lines, row)
	}

	// Join and style
	content := lipgloss.JoinVertical(lipgloss.Left, lines...)

	// Add overall border - never changes
	containerStyle := lipgloss.NewStyle().
		Width(width - 2).
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("237"))

	return containerStyle.Render(content)
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

	settingsButton := m.renderSettingsButton(width-4, settingsButtonHeight)
	calendar := m.renderCalendar(width-4, calendarHeight)

	// Join sections vertically
	content := lipgloss.JoinVertical(lipgloss.Right,
		settingsButton,
		"",
		calendar,
	)

	return content
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

	// Panel styling with square border - never changes
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
		lines = append(lines, focusedStyle.Render("> " + sectionHeader))
	} else {
		lines = append(lines, titleStyle.Render("  " + sectionHeader))
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
		lines = append(lines, focusedStyle.Render("> " + sectionHeader))
	} else {
		lines = append(lines, titleStyle.Render("  " + sectionHeader))
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
		lines = append(lines, focusedStyle.Render("> " + sectionHeader))
		if m.colorPickerActive {
			lines = append(lines, instructionStyle.Render("  1-9/0: color, esc: cancel"))
		} else {
			lines = append(lines, instructionStyle.Render("  enter: pick, 1-9/0: color"))
		}
	} else {
		lines = append(lines, titleStyle.Render("  " + sectionHeader))
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
