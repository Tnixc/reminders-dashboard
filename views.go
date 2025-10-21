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

	// Modern header with badge-style design
	badgeColor := listColor
	if focused {
		badgeColor = listColor // Keep same color, just background highlight later
	}
	
	headerBadgeStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("235")).
		Background(lipgloss.Color(badgeColor)).
		Padding(0, 1).
		Align(lipgloss.Left)

	countBadgeStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("245")).
		Italic(true)

	header := headerBadgeStyle.Render(listName) + " " + countBadgeStyle.Render(fmt.Sprintf("(%d)", len(reminders)))

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

	// Column styling with consistent border - no change on focus
	columnStyle := lipgloss.NewStyle().
		Width(width).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("237")).
		Padding(1)

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

	// Countdown with urgency-based badge
	countdownBadgeStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("235")).
		Background(urgencyColor).
		Padding(0, 1).
		Bold(true)

	// Build card content
	var lines []string
	lines = append(lines, titleStyle.Render(title))
	lines = append(lines, metaStyle.Render("  📅 "+dueText))
	lines = append(lines, "  "+countdownBadgeStyle.Render(countdown))

	content := lipgloss.JoinVertical(lipgloss.Left, lines...)

	// Card styling - background changes on focus, no border changes
	backgroundColor := lipgloss.Color("236")
	if focused {
		backgroundColor = lipgloss.Color("237")
	}

	cardStyle := lipgloss.NewStyle().
		Width(width).
		Background(backgroundColor).
		Padding(1).
		MarginBottom(1)

	return cardStyle.Render(content)
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

	// Add a modern header
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("183")).
		Background(lipgloss.Color("53")).
		Padding(0, 2).
		Width(width - 8)
	
	lines = append(lines, headerStyle.Render(fmt.Sprintf("📋 All Reminders (%d)", len(filtered))))
	lines = append(lines, "")

	// Define fixed column widths for table layout
	contentWidth := width - 8
	cursorWidth := 2
	titleWidth := (contentWidth - cursorWidth - 30) // Remaining space for title
	if titleWidth < 20 {
		titleWidth = 20
	}
	listWidth := 15
	countdownWidth := 12

	// Add table header
	headerRowStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("245")).
		Bold(true)
	
	headerRow := fmt.Sprintf("  %-*s  %-*s  %-*s  %s",
		titleWidth, "Title",
		listWidth, "List",
		countdownWidth, "Countdown",
		"Due Date")
	lines = append(lines, headerRowStyle.Render(headerRow))
	
	// Add separator
	separatorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("237"))
	lines = append(lines, separatorStyle.Render("  "+strings.Repeat("─", contentWidth)))

	// Calculate visible range (adjust for header)
	adjustedHeight := height - 7 // Account for header, table header, separator, footer
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

		// Create modern badges
		listBadgeStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("235")).
			Background(lipgloss.Color(listColor)).
			Padding(0, 1).
			Bold(true)
		
		countdownBadgeStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("235")).
			Background(urgencyColor).
			Padding(0, 1).
			Bold(true)

		listBadge := listBadgeStyle.Render(reminder.List)
		countdownBadge := countdownBadgeStyle.Render(countdown)

		dueText := formatDueDate(reminder.DueDate)
		dueStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("245")).
			Italic(true)

		// Truncate title if too long
		title := reminder.Title
		if len(title) > titleWidth-3 {
			title = title[:titleWidth-3] + "..."
		}

		// Build table row with fixed widths
		cursor := "  "
		titleColor := lipgloss.Color("255")
		bgColor := lipgloss.Color("")
		
		if i == m.cursor {
			cursor = "▸ "
			titleColor = lipgloss.Color(listColor)
			bgColor = lipgloss.Color("237")
		}

		titleStyle := lipgloss.NewStyle().
			Foreground(titleColor).
			Bold(true)

		// Format as fixed-width table row
		rowContent := fmt.Sprintf("%-*s  %-*s  %-*s  %s",
			titleWidth, titleStyle.Render(title),
			listWidth+4, listBadge, // +4 for badge padding/styling
			countdownWidth+4, countdownBadge, // +4 for badge padding/styling
			dueStyle.Render("📅 "+dueText))

		// Apply background only, no border changes
		if i == m.cursor {
			rowStyle := lipgloss.NewStyle().
				Background(bgColor).
				Width(contentWidth)
			lines = append(lines, cursor+rowStyle.Render(rowContent))
		} else {
			lines = append(lines, cursor+rowContent)
		}
	}

	// Add scroll indicator at bottom if needed
	if len(filtered) > adjustedHeight {
		lines = append(lines, "")
		scrollStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("245")).
			Italic(true)
		lines = append(lines, scrollStyle.Render(fmt.Sprintf("  ↕ Showing %d-%d of %d", visibleStart+1, min(visibleEnd, len(filtered)), len(filtered))))
	}

	// Join and style
	content := lipgloss.JoinVertical(lipgloss.Left, lines...)

	// Add overall border - never changes
	containerStyle := lipgloss.NewStyle().
		Width(width - 4).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("237")).
		Padding(1)

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
	text := "⚙ View settings\n\n(next n days,\nfilter lists)\n\nPress 's' to open"

	buttonStyle := lipgloss.NewStyle().
		Width(width).
		Height(height).
		Padding(1).
		Foreground(lipgloss.Color("183")).
		Align(lipgloss.Center, lipgloss.Center).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("237"))

	return buttonStyle.Render(text)
}

// renderCalendar renders the small calendar widget
func (m model) renderCalendar(width, height int) string {
	now := time.Now()

	// Create calendar title with modern styling
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("183")).
		Italic(true)

	title := titleStyle.Render(now.Format("January 2006"))

	// Create day number with vibrant color
	dayStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("235")).
		Background(lipgloss.Color("205")).
		Padding(0, 2).
		Width(width - 2).
		Align(lipgloss.Center)

	day := dayStyle.Render(fmt.Sprintf("%d", now.Day()))

	// Create weekday
	weekdayStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("248")).
		Italic(true)

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
		Align(lipgloss.Center, lipgloss.Center).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("237"))

	return calendarStyle.Render(content)
}

// renderSettingsPanel renders the settings overlay panel
func (m model) renderSettingsPanel() string {
	panelWidth := (m.width * 9) / 10

	panelHeight := 35
	if panelHeight > m.height*3/4 {
		panelHeight = m.height * 3 / 4
	}

	// Header with modern gradient-like styling
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205")).
		Background(lipgloss.Color("53")).
		Padding(0, 2).
		MarginBottom(1).
		Width(panelWidth - 4)

	header := headerStyle.Render("⚙ Settings")

	// Build content sections
	content := header + "\n\n"
	content += m.renderDaysFilterSection() + "\n\n"
	content += m.renderListFilterSection() + "\n\n"
	content += m.renderColorConfigSection()

	// Panel styling with rounded border - never changes
	panelStyle := lipgloss.NewStyle().
		Width(panelWidth).
		Height(panelHeight).
		Padding(1, 2).
		Background(lipgloss.Color("235")).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("237"))

	return panelStyle.Render(content)
}

// renderDaysFilterSection renders the days filter settings
func (m model) renderDaysFilterSection() string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("141")).
		Italic(true)

	focusedStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205"))

	normalStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("248"))

	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("141")).
		Background(lipgloss.Color("53")).
		Padding(0, 2).
		MarginBottom(1)

	var lines []string

	sectionHeader := "  Days Filter"
	if m.sidebarSection == SidebarDaysFilter {
		sectionHeader = "▸ Days Filter"
		lines = append(lines, headerStyle.Render(sectionHeader))
	} else {
		lines = append(lines, titleStyle.Render(sectionHeader))
	}

	for i, days := range m.daysFilterOptions {
		label := "All"
		if days > 0 {
			label = fmt.Sprintf("Next %d days", days)
		}

		// Fixed width cursor to prevent layout shifts
		cursor := "  "
		lineColor := normalStyle
		if m.sidebarSection == SidebarDaysFilter && m.sidebarCursor == i {
			cursor = "▸ "
			lineColor = focusedStyle
		}

		// Modern radio button style
		check := "○"
		checkColor := lipgloss.Color("248")
		if days == m.daysFilter {
			check = "●"
			checkColor = lipgloss.Color("141")
		}
		
		checkStyle := lipgloss.NewStyle().
			Foreground(checkColor).
			Bold(days == m.daysFilter)

		line := fmt.Sprintf("%s%s %s", cursor, checkStyle.Render(check), label)
		line = lineColor.Render(line)

		lines = append(lines, line)
	}

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

// renderListFilterSection renders the list filter settings
func (m model) renderListFilterSection() string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("141")).
		Italic(true)

	focusedStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205"))

	normalStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("248"))

	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("81")).
		Background(lipgloss.Color("53")).
		Padding(0, 2).
		MarginBottom(1)

	var lines []string

	sectionHeader := "  List Filter"
	if m.sidebarSection == SidebarListFilter {
		sectionHeader = "▸ List Filter"
		lines = append(lines, headerStyle.Render(sectionHeader))
	} else {
		lines = append(lines, titleStyle.Render(sectionHeader))
	}

	for i, listName := range m.availableLists {
		// Fixed width cursor to prevent layout shifts
		cursor := "  "
		lineColor := normalStyle
		if m.sidebarSection == SidebarListFilter && m.sidebarCursor == i {
			cursor = "▸ "
			lineColor = focusedStyle
		}

		// Modern checkbox style
		check := "☐"
		checkColor := lipgloss.Color("248")
		isSelected := false
		if selected, exists := m.selectedLists[listName]; exists && selected {
			check = "☑"
			checkColor = lipgloss.Color("48")
			isSelected = true
		}
		
		checkStyle := lipgloss.NewStyle().
			Foreground(checkColor).
			Bold(isSelected)

		line := fmt.Sprintf("%s%s %s", cursor, checkStyle.Render(check), listName)
		line = lineColor.Render(line)

		lines = append(lines, line)
	}

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

// renderColorConfigSection renders the color configuration settings
func (m model) renderColorConfigSection() string {
	// Modern gradient styles
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("141")).
		Italic(true)

	focusedStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205"))

	normalStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("248"))

	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205")).
		Background(lipgloss.Color("53")).
		Padding(0, 2).
		MarginBottom(1)

	instructionStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("183")).
		Italic(true).
		MarginBottom(1)

	var lines []string

	// Header + instructions with modern styling
	sectionHeader := "  List Colors"
	if m.sidebarSection == SidebarColorConfig {
		sectionHeader = "▸ List Colors"
		lines = append(lines, headerStyle.Render(sectionHeader))
		if m.colorPickerActive {
			lines = append(lines, instructionStyle.Render("  Choose a color (1-9, 0) • esc: cancel"))
		} else {
			lines = append(lines, instructionStyle.Render("  Enter: pick list • 1-9/0: set color"))
		}
	} else {
		lines = append(lines, titleStyle.Render(sectionHeader))
	}

	// Lists with modern color badge design
	for i, listName := range m.availableLists {
		currentColor := "248"
		if color, exists := m.listColors[listName]; exists {
			currentColor = color
		}

		// Create a colorful badge for the list
		listBadgeStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("235")).
			Background(lipgloss.Color(currentColor)).
			Padding(0, 1).
			Bold(true)

		badge := listBadgeStyle.Render("●")

		// Fixed width cursor to prevent layout shifts
		cursor := "  "
		nameColor := normalStyle
		nameText := listName
		
		if m.sidebarSection == SidebarColorConfig && m.sidebarCursor == i {
			cursor = "▸ "
			nameColor = focusedStyle
		}
		
		// Highlight the list targeted for recolor when active
		if m.colorPickerActive && m.colorPickerList == i {
			nameText = listName + " ✓"
			nameStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("205")).
				Bold(true)
			nameText = nameStyle.Render(nameText)
		} else {
			nameText = nameColor.Render(nameText)
		}

		line := fmt.Sprintf("%s%s %s", cursor, badge, nameText)

		lines = append(lines, line)
	}

	// Modern color palette with cards
	if len(m.availableColors) > 0 {
		lines = append(lines, "")
		
		paletteHeaderStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("183")).
			Bold(true).
			MarginTop(1).
			MarginBottom(1)
		lines = append(lines, paletteHeaderStyle.Render("  Color Palette"))

		// Create color swatches in a grid with rounded borders
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

			// Create a card-style color swatch
			swatchStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("235")).
				Background(lipgloss.Color(code)).
				Padding(0, 1).
				Bold(true)

			keyStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("245")).
				Bold(true)

			nameStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("248"))

			swatch := swatchStyle.Render("████")
			
			// Add selection indicator with fixed width
			nameText := name
			if m.colorPickerActive && m.colorPickerCursor == i {
				// Highlighted with background, no layout shift
				highlightStyle := lipgloss.NewStyle().
					Foreground(lipgloss.Color(code)).
					Bold(true)
				nameText = highlightStyle.Render(name + " ◀")
			} else {
				nameText = nameStyle.Render(name + "   ") // Padding to match "◀" width
			}
			
			item := fmt.Sprintf("  %s %s %s", keyStyle.Render(key), swatch, nameText)

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
