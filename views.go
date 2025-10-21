package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// getContentHeight returns the available height for content (excluding footer)
func (m *model) getContentHeight() int {
	// Account for footer (1 line)
	return m.height - 1
}

// View renders the UI
func (m model) View() string {
	// Define base styles
	footerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Background(lipgloss.Color("235")).
		Padding(0, 1).
		Width(m.width)

	// Render footer
	helpText := "tab: view • s: focus sidebar • r: refresh • ↑/↓: items • q: quit"
	if m.sidebarFocused {
		if m.sidebarSection == SidebarColorConfig {
			helpText = "tab: section • 1-9/0: pick color • ↑/↓: lists • s: unfocus sidebar • q: quit"
		} else {
			helpText = "tab: section • space: toggle • ↑/↓: navigate • s: unfocus sidebar • q: quit"
		}
	} else if m.viewMode == ColumnView {
		helpText = "tab: view • s: focus sidebar • ctrl+←/→: reorder • ←/→: columns • ↑/↓: items • q: quit"
	}
	footer := footerStyle.Render(helpText)

	// Handle special states
	var content string
	if m.loading {
		content = renderLoading(m.width, m.height-1)
	} else if m.err != nil {
		content = renderError(m.err, m.width, m.height-1)
	} else {
		// Render main content and sidebar with responsive widths
		var mainContent string
		if m.viewMode == ColumnView {
			mainContent = m.renderColumnView()
		} else {
			mainContent = m.renderListView()
		}

		sidebar := m.renderSidebar()

		// Responsive sidebar width with clamping so total <= m.width
		minMain := 20
		desiredSidebar := m.width / 3
		if desiredSidebar < 24 {
			desiredSidebar = 24
		}
		if desiredSidebar > 48 {
			desiredSidebar = 48
		}

		// Ensure main + sidebar fits; shrink sidebar if necessary
		sidebarWidth := desiredSidebar
		if m.width-sidebarWidth < minMain {
			sidebarWidth = m.width - minMain
			if sidebarWidth < 0 {
				sidebarWidth = 0
			}
		}
		mainWidth := m.width - sidebarWidth
		contentHeight := m.getContentHeight()

		mainStyle := lipgloss.NewStyle().Width(mainWidth).Height(contentHeight)
		sidebarStyle := lipgloss.NewStyle().Width(sidebarWidth).Height(contentHeight)

		content = lipgloss.JoinHorizontal(lipgloss.Top,
			mainStyle.Render(mainContent),
			sidebarStyle.Render(sidebar))
	}

	// Combine all parts
	return lipgloss.JoinVertical(lipgloss.Left, content, footer)
}

func renderLoading(width, height int) string {
	style := lipgloss.NewStyle().
		Width(width).
		Height(height).
		Align(lipgloss.Center).
		AlignVertical(lipgloss.Center)

	return style.Render("Loading reminders...")
}

func renderError(err error, width, height int) string {
	style := lipgloss.NewStyle().
		Width(width).
		Height(height).
		Align(lipgloss.Center).
		AlignVertical(lipgloss.Center).
		Foreground(lipgloss.Color("196"))

	return style.Render(fmt.Sprintf("Error loading reminders:\n%v\n\nPress q to quit", err))
}

func (m model) renderSidebar() string {
	contentHeight := m.getContentHeight()

	// Minimum heights for each section
	minDaysHeight := 5
	minListHeight := 7
	minColorHeight := 7
	totalMinHeight := minDaysHeight + minListHeight + minColorHeight

	var daysHeight, listHeight, colorHeight int

	if contentHeight >= totalMinHeight {
		// Allocate height proportionally
		// Days Filter: ~20% of height
		// List Filter: ~45% of height
		// Color Config: ~35% of height
		daysHeight = contentHeight * 20 / 100
		listHeight = contentHeight * 45 / 100
		colorHeight = contentHeight - daysHeight - listHeight // remainder

		// Ensure minimum heights
		if daysHeight < minDaysHeight {
			daysHeight = minDaysHeight
		}
		if listHeight < minListHeight {
			listHeight = minListHeight
		}
		if colorHeight < minColorHeight {
			colorHeight = minColorHeight
		}

		// If total exceeds contentHeight, reduce proportionally
		totalAllocated := daysHeight + listHeight + colorHeight
		if totalAllocated > contentHeight {
			excess := totalAllocated - contentHeight
			// Reduce from color section first, then list, then days
			if colorHeight > minColorHeight {
				reduce := min(excess, colorHeight - minColorHeight)
				colorHeight -= reduce
				excess -= reduce
			}
			if excess > 0 && listHeight > minListHeight {
				reduce := min(excess, listHeight - minListHeight)
				listHeight -= reduce
				excess -= reduce
			}
			if excess > 0 && daysHeight > minDaysHeight {
				reduce := min(excess, daysHeight - minDaysHeight)
				daysHeight -= reduce
			}
		}
	} else {
		// Terminal too short, use minimum heights proportionally
		daysHeight = contentHeight * minDaysHeight / totalMinHeight
		listHeight = contentHeight * minListHeight / totalMinHeight
		colorHeight = contentHeight - daysHeight - listHeight
		if daysHeight < 3 {
			daysHeight = 3
		}
		if listHeight < 4 {
			listHeight = 4
		}
		if colorHeight < 4 {
			colorHeight = 4
		}
	}

	var sections []string

	// Days Filter Section
	sections = append(sections, m.renderDaysFilterSection(daysHeight))

	// List Filter Section
	sections = append(sections, m.renderListFilterSection(listHeight))

	// Color Config Section
	sections = append(sections, m.renderColorConfigSection(colorHeight))

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

func (m model) renderDaysFilterSection(height int) string {
	var content strings.Builder

	// Section title
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("248"))
	
	if m.sidebarFocused && m.sidebarSection == SidebarDaysFilter {
		titleStyle = titleStyle.Foreground(lipgloss.Color("205"))
	}
	
	content.WriteString(titleStyle.Render("📅 Days Filter") + "\n\n")

	// Section content
	if m.sidebarFocused && m.sidebarSection == SidebarDaysFilter {
		// Compute visible rows based on available height (minus borders and title)
		innerHeight := height - 2 // top+bottom border
		overhead := 2             // title + blank line
		visibleRows := innerHeight - overhead
		if visibleRows < 1 {
			visibleRows = 1
		}

		startIdx := 0
		endIdx := len(m.daysFilterOptions)

		// Scroll if needed
		if len(m.daysFilterOptions) > visibleRows {
			if m.sidebarCursor >= visibleRows {
				startIdx = m.sidebarCursor - visibleRows + 1
			}
			endIdx = startIdx + visibleRows
			if endIdx > len(m.daysFilterOptions) {
				endIdx = len(m.daysFilterOptions)
			}
		}

		for i := startIdx; i < endIdx; i++ {
			days := m.daysFilterOptions[i]
			cursor := " "
			if m.sidebarCursor == i {
				cursor = "▸"
			}

			checked := "○"
			label := "All reminders"
			if days > 0 {
				label = fmt.Sprintf("Next %d days", days)
			}

			if m.daysFilter == days {
				checked = "●"
			}

			style := lipgloss.NewStyle()
			if m.sidebarCursor == i {
				style = style.Bold(true).Foreground(lipgloss.Color("205"))
			}

			line := fmt.Sprintf("%s %s %s", cursor, checked, label)
			content.WriteString(style.Render(line) + "\n")
		}
	} else {
		// Show current selection when not focused
		dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("243"))
		if m.daysFilter == 0 {
			content.WriteString(dimStyle.Render(" ● All reminders\n"))
		} else {
			content.WriteString(dimStyle.Render(fmt.Sprintf(" ● Next %d days\n", m.daysFilter)))
		}
	}

	// Wrap in border with fixed height
	borderStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(0, 1).
		Height(height)

	if m.sidebarFocused && m.sidebarSection == SidebarDaysFilter {
		borderStyle = borderStyle.BorderForeground(lipgloss.Color("205"))
	}

	return borderStyle.Render(content.String())
}

func (m model) renderListFilterSection(height int) string {
	var content strings.Builder

	// Section title
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("248"))
	
	if m.sidebarFocused && m.sidebarSection == SidebarListFilter {
		titleStyle = titleStyle.Foreground(lipgloss.Color("205"))
	}
	
	content.WriteString(titleStyle.Render("📋 List Filter") + "\n\n")

	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("243"))

	if m.sidebarFocused && m.sidebarSection == SidebarListFilter {
		// Compute visible rows based on section height (minus borders and title)
		innerHeight := height - 2 // top+bottom border
		overhead := 2             // title + blank line
		visibleRows := innerHeight - overhead
		if visibleRows < 1 {
			visibleRows = 1
		}

		startIdx := 0
		endIdx := len(m.availableLists)

		// Scroll if needed
		if len(m.availableLists) > visibleRows {
			if m.sidebarCursor >= visibleRows {
				startIdx = m.sidebarCursor - visibleRows + 1
			}
			endIdx = startIdx + visibleRows
			if endIdx > len(m.availableLists) {
				endIdx = len(m.availableLists)
			}
		}

		for i := startIdx; i < endIdx; i++ {
			listName := m.availableLists[i]
			cursor := " "
			if m.sidebarCursor == i {
				cursor = "▸"
			}

			checked := "☐"
			if m.selectedLists[listName] {
				checked = "☑"
			}

			style := lipgloss.NewStyle()
			if m.sidebarCursor == i {
				style = style.Bold(true).Foreground(lipgloss.Color("205"))
			}

			// Show list color indicator
			color := m.listColors[listName]
			colorIndicator := lipgloss.NewStyle().Foreground(lipgloss.Color(color)).Render("●")

			line := fmt.Sprintf("%s %s %s %s", cursor, checked, colorIndicator, listName)
			content.WriteString(style.Render(line) + "\n")
		}
	} else {
		// Show count of selected lists when not focused
		selectedCount := 0
		for _, selected := range m.selectedLists {
			if selected {
				selectedCount++
			}
		}
		content.WriteString(dimStyle.Render(fmt.Sprintf(" %d/%d lists selected\n", selectedCount, len(m.availableLists))))
	}

	// Wrap in border with fixed height
	borderStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(0, 1).
		Height(height)

	if m.sidebarFocused && m.sidebarSection == SidebarListFilter {
		borderStyle = borderStyle.BorderForeground(lipgloss.Color("205"))
	}

	return borderStyle.Render(content.String())
}

func (m model) renderColorConfigSection(height int) string {
	var content strings.Builder

	// Section title
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("248"))
	
	if m.sidebarFocused && m.sidebarSection == SidebarColorConfig {
		titleStyle = titleStyle.Foreground(lipgloss.Color("205"))
	}
	
	content.WriteString(titleStyle.Render("🎨 List Colors") + "\n\n")

	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("243"))

	if m.sidebarFocused && m.sidebarSection == SidebarColorConfig {
		// Compute layout based on available height
		innerHeight := height - 2 // top+bottom border
		titleOverhead := 2        // title + blank line

		// Reserve at least one line for the palette header
		minPaletteLines := 1

		// Visible list rows allocated dynamically
		visibleLists := innerHeight - titleOverhead - minPaletteLines - 1 // -1 to show at least one palette row
		if visibleLists < 1 {
			visibleLists = 1
		}

		startIdx := 0
		endIdx := len(m.availableLists)

		// Scroll lists if needed
		if len(m.availableLists) > visibleLists {
			if m.sidebarCursor >= visibleLists {
				startIdx = m.sidebarCursor - visibleLists + 1
			}
			endIdx = startIdx + visibleLists
			if endIdx > len(m.availableLists) {
				endIdx = len(m.availableLists)
			}
		}

		// Show which list we're editing
		for i := startIdx; i < endIdx; i++ {
			listName := m.availableLists[i]
			cursor := " "
			if m.sidebarCursor == i {
				cursor = "▸"
			}

			color := m.listColors[listName]
			colorIndicator := lipgloss.NewStyle().Foreground(lipgloss.Color(color)).Bold(true).Render("●")

			style := lipgloss.NewStyle()
			if m.sidebarCursor == i {
				style = style.Bold(true).Foreground(lipgloss.Color("205"))
			}

			line := fmt.Sprintf("%s %s %s", cursor, colorIndicator, listName)
			content.WriteString(style.Render(line) + "\n")
		}

		if m.sidebarCursor < len(m.availableLists) {
			content.WriteString("\n" + dimStyle.Render("Pick a color:") + "\n")
			
			listName := m.availableLists[m.sidebarCursor]
			currentColor := m.listColors[listName]

			// Determine how many palette rows we can show
			paletteRows := innerHeight - titleOverhead - (endIdx-startIdx) - 1 // minus "Pick a color:" line
			if paletteRows < 1 {
				paletteRows = 1
			}
			if paletteRows > len(m.availableColors) {
				paletteRows = len(m.availableColors)
			}

			for i := 0; i < paletteRows; i++ {
				color := m.availableColors[i]
				colorName := m.colorNames[i]
				colorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(color)).Bold(true)

				// Show number key
				keyNum := (i + 1) % 10 // 1-9, then 0 for 10th
				
				indicator := " "
				if color == currentColor {
					indicator = "▸"
				}

				// Create color preview
				preview := colorStyle.Render("██")
				line := fmt.Sprintf("%s %d %s %s", indicator, keyNum, preview, colorName)
				
				if color == currentColor {
					line = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205")).Render(line)
				}
				
				content.WriteString(line + "\n")
			}

			// Indicate there are more palette colors not shown
			if paletteRows < len(m.availableColors) {
				content.WriteString(dimStyle.Render(" ...more") + "\n")
			}
		}
	} else {
		// Show instruction when not focused
		content.WriteString(dimStyle.Render(" Press 's' to configure\n"))
	}

	// Wrap in border with fixed height
	borderStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(0, 1).
		Height(height)

	if m.sidebarFocused && m.sidebarSection == SidebarColorConfig {
		borderStyle = borderStyle.BorderForeground(lipgloss.Color("205"))
	}

	return borderStyle.Render(content.String())
}

func (m model) renderListView() string {
	contentHeight := m.height - 1 // minus footer
	var lines []string

	lines = append(lines, "")

	filtered := getFilteredReminders(&m)

	if len(filtered) == 0 {
		content := "  No reminders to display"
		// Pad with empty lines to fill height
		result := strings.Join(lines, "\n") + "\n" + content
		currentLines := strings.Count(result, "\n") + 1
		for currentLines < contentHeight {
			result += "\n"
			currentLines++
		}
		return result
	}

	titleStyle := lipgloss.NewStyle().Bold(true)
	dateStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))

	visibleHeight := m.getContentHeight()

	// Calculate line boundaries for each reminder
	bounds := []reminderBounds{}
	currentLine := 0

	for _, reminder := range filtered {
		start := currentLine
		currentLine++ // title line
		currentLine++ // date/countdown line

		if reminder.Notes != "" {
			noteLines := strings.Split(reminder.Notes, "\n")
			for _, line := range noteLines {
				if strings.TrimSpace(line) != "" {
					currentLine++
				}
			}
		}
		currentLine++ // blank line

		bounds = append(bounds, reminderBounds{start, currentLine})
	}

	// Compute local scroll to keep current reminder visible (do not mutate model here)
	scroll := m.scrollOffset
	if m.cursor < len(bounds) {
		cursorStart := bounds[m.cursor].startLine
		cursorEnd := bounds[m.cursor].endLine

		// Scroll up if cursor is above visible area
		if cursorStart < scroll {
			scroll = cursorStart
		}

		// Scroll down if cursor is below visible area
		if cursorEnd > scroll+visibleHeight {
			scroll = cursorEnd - visibleHeight
			if scroll < 0 {
				scroll = 0
			}
		}
	}

	// Now render with scrolling applied
	currentLine = 0
	for i, reminder := range filtered {
		cursor := "  "
		if m.cursor == i {
			cursor = "▸ "
		}

		countdown, urgency := getCountdown(reminder.DueDate)
		urgencyColor := getUrgencyColor(urgency)
		countdownStyle := lipgloss.NewStyle().Foreground(urgencyColor)

		// Get list color
		listColor := m.listColors[reminder.List]
		listColorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(listColor))

		title := reminder.Title
		if m.cursor == i {
			title = titleStyle.Render(title)
		}

		titleLine := fmt.Sprintf("%s %s", cursor, title)
		infoLine := fmt.Sprintf("   %s • %s • %s",
			listColorStyle.Render(reminder.List),
			formatDueDate(reminder.DueDate),
			countdownStyle.Render(countdown))

		// Check if lines are in visible range
		if currentLine >= scroll && currentLine < scroll+visibleHeight {
			lines = append(lines, titleLine)
		}
		currentLine++

		if currentLine >= scroll && currentLine < scroll+visibleHeight {
			lines = append(lines, infoLine)
		}
		currentLine++

		if reminder.Notes != "" {
			noteLines := strings.Split(reminder.Notes, "\n")
			for _, line := range noteLines {
				if strings.TrimSpace(line) != "" {
					noteLine := fmt.Sprintf("   %s", dateStyle.Render(line))
					if currentLine >= scroll && currentLine < scroll+visibleHeight {
						lines = append(lines, noteLine)
					}
					currentLine++
				}
			}
		}

		if currentLine >= scroll && currentLine < scroll+visibleHeight {
			lines = append(lines, "")
		}
		currentLine++
	}

	result := strings.Join(lines, "\n")
	// Pad with empty lines to fill height
	currentLines := strings.Count(result, "\n") + 1
	for currentLines < contentHeight {
		result += "\n"
		currentLines++
	}
	return result
}

func (m model) renderColumnView() string {
	contentHeight := m.height - 1 // minus footer
	columns := getColumns(&m)

	if len(columns) == 0 {
		content := "\n  No reminders to display"
		// Pad with empty lines to fill height
		currentLines := strings.Count(content, "\n") + 1
		for currentLines < contentHeight {
			content += "\n"
			currentLines++
		}
		return content
	}

	// Use local scrolls to keep rendering pure
	columnScrolls := make([]int, len(columns))
	for i := range columns {
		if i < len(m.columnScrolls) {
			columnScrolls[i] = m.columnScrolls[i]
		} else {
			columnScrolls[i] = 0
		}
	}

	// Render each column with scrolling
	var columnStrs []string

	for colIdx, col := range columns {
		scroll := columnScrolls[colIdx]
		isFocused := colIdx == m.columnCursor

		// Compute scroll position for focused column (do not mutate model)
		if isFocused && m.cursor < len(col) {
			visibleHeight := m.getContentHeight()
			// Each reminder takes approximately 4 lines
			cursorLine := m.cursor * 4

			// Scroll up if cursor is above visible area
			if cursorLine < scroll {
				scroll = cursorLine
			}

			// Scroll down if cursor is below visible area
			if cursorLine >= scroll+visibleHeight {
				scroll = cursorLine - visibleHeight + 4
				if scroll < 0 {
					scroll = 0
				}
			}
		}

		columnStr := m.renderColumn(col, isFocused, scroll)
		columnStrs = append(columnStrs, columnStr)
	}

	// Calculate column width based on responsive layout and include borders/padding
	// Keep this in sync with View() sidebar calculation
	minMain := 20
	desiredSidebar := m.width / 3
	if desiredSidebar < 24 {
		desiredSidebar = 24
	}
	if desiredSidebar > 48 {
		desiredSidebar = 48
	}
	sidebarWidth := desiredSidebar
	if m.width-sidebarWidth < minMain {
		sidebarWidth = m.width - minMain
		if sidebarWidth < 0 {
			sidebarWidth = 0
		}
	}

	availableWidth := m.width - sidebarWidth
	if availableWidth < 0 {
		availableWidth = 0
	}

	n := len(columns)
	extraPerCol := 4 // 2 border + 2 padding
	contentWidth := availableWidth - n*extraPerCol
	// Ensure a reasonable minimum per column
	if contentWidth < n*10 {
		contentWidth = n * 10
	}
	columnWidth := contentWidth / n
	if columnWidth > 40 {
		columnWidth = 40
	}
	if columnWidth < 10 {
		columnWidth = 10
	}

	// Style each column with borders
	styledColumns := make([]string, len(columnStrs))
	for i, colStr := range columnStrs {
		style := lipgloss.NewStyle().
			Width(columnWidth).
			Padding(0, 1).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240"))

		// Highlight border on focus
		if i == m.columnCursor {
			style = style.BorderForeground(lipgloss.Color("205"))
		}

		styledColumns[i] = style.Render(colStr)
	}

	result := "\n" + lipgloss.JoinHorizontal(lipgloss.Top, styledColumns...)
	// Pad with empty lines to fill height
	currentLines := strings.Count(result, "\n") + 1
	for currentLines < contentHeight {
		result += "\n"
		currentLines++
	}
	return result
}

func (m model) renderColumn(reminders []Reminder, isFocused bool, scrollOffset int) string {
	var lines []string

	if len(reminders) == 0 {
		return ""
	}

	// Column header (list name)
	listName := reminders[0].List
	listColor := m.listColors[listName]
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(listColor)).
		Underline(true)

	lines = append(lines, headerStyle.Render(listName))
	lines = append(lines, "")

	titleStyle := lipgloss.NewStyle().Bold(true)
	dateStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))

	visibleHeight := m.getContentHeight()
	currentLine := 0

	for i, reminder := range reminders {
		cursor := " "
		if isFocused && m.cursor == i {
			cursor = "▸"
		}

		countdown, urgency := getCountdown(reminder.DueDate)
		urgencyColor := getUrgencyColor(urgency)
		countdownStyle := lipgloss.NewStyle().Foreground(urgencyColor)

		title := reminder.Title
		if isFocused && m.cursor == i {
			title = titleStyle.Render(title)
		}

		titleLine := fmt.Sprintf("%s %s", cursor, title)
		dateLine := fmt.Sprintf("  %s", dateStyle.Render(formatDueDate(reminder.DueDate)))
		countdownLine := fmt.Sprintf("  %s", countdownStyle.Render(countdown))

		// Check if lines are in visible range
		if currentLine >= scrollOffset && currentLine < scrollOffset+visibleHeight {
			lines = append(lines, titleLine)
		}
		currentLine++

		if currentLine >= scrollOffset && currentLine < scrollOffset+visibleHeight {
			lines = append(lines, dateLine)
		}
		currentLine++

		if currentLine >= scrollOffset && currentLine < scrollOffset+visibleHeight {
			lines = append(lines, countdownLine)
		}
		currentLine++

		if i < len(reminders)-1 {
			if currentLine >= scrollOffset && currentLine < scrollOffset+visibleHeight {
				lines = append(lines, "")
			}
			currentLine++
		}
	}

	return strings.Join(lines, "\n")
}