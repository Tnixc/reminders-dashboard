package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// ColumnViewConfig holds configuration for the column view
type ColumnViewConfig struct {
	Width          int
	Height         int
	Columns        [][]Reminder
	ColumnNames    []string
	FocusedColumn  int
	FocusedItem    int
	ListColors     map[string]string
	ScrollOffsets  []int
	GetCountdown   func(string) (string, int)
}

// RenderColumnView renders the multi-column card layout
func RenderColumnView(config ColumnViewConfig) string {
	if len(config.Columns) == 0 {
		return RenderEmptyCard(config.Width, config.Height, "No reminders to display")
	}

	// Calculate how many columns we can fit
	minCardWidth := 20
	maxCardWidth := 36
	cardPadding := 2 // Space between columns

	availableWidth := config.Width
	numColumns := len(config.Columns)

	// Calculate column width
	totalPadding := (numColumns - 1) * cardPadding
	columnWidth := (availableWidth - totalPadding) / numColumns

	// Ensure reasonable bounds
	if columnWidth < minCardWidth {
		// Too narrow, reduce number of visible columns
		numColumns = availableWidth / (minCardWidth + cardPadding)
		if numColumns < 1 {
			numColumns = 1
		}
		columnWidth = (availableWidth - (numColumns-1)*cardPadding) / numColumns
	}
	if columnWidth > maxCardWidth {
		columnWidth = maxCardWidth
	}

	// Render visible columns
	var renderedColumns []string
	visibleCount := numColumns
	if visibleCount > len(config.Columns) {
		visibleCount = len(config.Columns)
	}

	for i := 0; i < visibleCount; i++ {
		column := config.Columns[i]
		columnName := ""
		if i < len(config.ColumnNames) {
			columnName = config.ColumnNames[i]
		}

		scrollOffset := 0
		if i < len(config.ScrollOffsets) {
			scrollOffset = config.ScrollOffsets[i]
		}

		listColor := "248"
		if columnName != "" && config.ListColors != nil {
			if color, exists := config.ListColors[columnName]; exists {
				listColor = color
			}
		}

		columnStr := renderSingleColumn(
			column,
			columnName,
			columnWidth,
			config.Height,
			i == config.FocusedColumn,
			config.FocusedItem,
			listColor,
			scrollOffset,
			config.GetCountdown,
		)

		renderedColumns = append(renderedColumns, columnStr)
	}

	// If we can't fit all columns, show indicator
	if visibleCount < len(config.Columns) {
		remaining := len(config.Columns) - visibleCount
		indicator := lipgloss.NewStyle().
			Width(columnWidth).
			Height(config.Height).
			Padding(1).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240")).
			Foreground(lipgloss.Color("245")).
			Italic(true).
			Align(lipgloss.Center, lipgloss.Center).
			Render(fmt.Sprintf("+ %d more", remaining))
		renderedColumns = append(renderedColumns, indicator)
	}

	// Join columns horizontally
	return lipgloss.JoinHorizontal(lipgloss.Top, renderedColumns...)
}

// renderSingleColumn renders a single column with its cards
func renderSingleColumn(
	reminders []Reminder,
	columnName string,
	width int,
	height int,
	focused bool,
	focusedItem int,
	listColor string,
	scrollOffset int,
	getCountdown func(string) (string, int),
) string {
	// Header with list name
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(listColor)).
		Width(width - 4).
		Padding(0, 1)

	header := headerStyle.Render(columnName)

	// Render cards
	var cards []string
	cardHeight := 0 // Auto height for cards

	// Calculate visible range based on scroll
	// Assume each card takes ~6 lines (title + due + countdown + notes + spacing)
	linesPerCard := 6
	visibleCards := (height - 3) / linesPerCard // Subtract header and footer
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

	for i := startIdx; i < endIdx; i++ {
		reminder := reminders[i]
		countdown := ""
		urgency := 999

		if getCountdown != nil {
			countdown, urgency = getCountdown(reminder.DueDate)
		}

		cardStyle := DefaultCardStyle()
		cardStyle.Width = width - 4 // Account for column borders
		cardStyle.Height = cardHeight
		cardStyle.Focused = focused && (i == focusedItem)
		cardStyle.ListColor = listColor

		card := RenderCard(reminder, cardStyle, countdown, urgency)
		cards = append(cards, card)
	}

	// Footer with item count
	footerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("245")).
		Width(width - 4).
		Padding(0, 1)

	footer := footerStyle.Render(fmt.Sprintf("%d items", len(reminders)))

	// Combine header, cards, and footer
	content := header + "\n"
	if len(cards) > 0 {
		content += strings.Join(cards, "\n") + "\n"
	} else {
		emptyMsg := lipgloss.NewStyle().
			Foreground(lipgloss.Color("245")).
			Italic(true).
			Padding(2, 1).
			Render("No items")
		content += emptyMsg + "\n"
	}
	content += footer

	// Wrap in column container
	borderColor := "240"
	if focused {
		borderColor = "205"
	}

	columnStyle := lipgloss.NewStyle().
		Width(width).
		Height(height).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(borderColor)).
		Padding(0, 1)

	return columnStyle.Render(content)
}