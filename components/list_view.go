package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// ListViewConfig holds configuration for the list view
type ListViewConfig struct {
	Width         int
	Height        int
	Reminders     []Reminder
	FocusedIndex  int
	ScrollOffset  int
	ListColors    map[string]string
	GetCountdown  func(string) (string, int)
	FormatDueDate func(string) string
}

// RenderListView renders reminders in a compact list format
func RenderListView(config ListViewConfig) string {
	if len(config.Reminders) == 0 {
		emptyStyle := lipgloss.NewStyle().
			Width(config.Width).
			Height(config.Height).
			Foreground(lipgloss.Color("245")).
			Italic(true).
			Padding(2, 4)
		return emptyStyle.Render("No reminders to display")
	}

	titleStyle := lipgloss.NewStyle().Bold(true)
	metaStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("245"))

	var lines []string
	currentLine := 0

	// Calculate which lines are visible
	visibleStart := config.ScrollOffset
	visibleEnd := config.ScrollOffset + config.Height

	for i, reminder := range config.Reminders {
		// Check if this reminder is in the visible range
		reminderStartLine := currentLine

		// Build reminder display
		cursor := "  "
		if i == config.FocusedIndex {
			cursor = "▸ "
		}

		// Get countdown info
		countdown := ""
		urgency := 999
		if config.GetCountdown != nil {
			countdown, urgency = config.GetCountdown(reminder.DueDate)
		}

		// Get list color
		listColor := "248"
		if config.ListColors != nil {
			if color, exists := config.ListColors[reminder.List]; exists {
				listColor = color
			}
		}
		listStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(listColor))
		urgencyStyle := lipgloss.NewStyle().Foreground(getUrgencyColor(urgency))

		// Format title line
		titleText := reminder.Title
		if i == config.FocusedIndex {
			titleText = titleStyle.Render(titleText)
		}
		titleLine := fmt.Sprintf("%s%s", cursor, titleText)

		// Format meta line (list, countdown, due date)
		dueText := ""
		if config.FormatDueDate != nil {
			dueText = config.FormatDueDate(reminder.DueDate)
		}

		metaLine := fmt.Sprintf("   %s • %s • %s",
			listStyle.Render(reminder.List),
			urgencyStyle.Render(countdown),
			metaStyle.Render(dueText))

		// Add lines if in visible range
		if currentLine >= visibleStart && currentLine < visibleEnd {
			lines = append(lines, titleLine)
		}
		currentLine++

		if currentLine >= visibleStart && currentLine < visibleEnd {
			lines = append(lines, metaLine)
		}
		currentLine++

		// Add notes if present
		if reminder.Notes != "" {
			noteLines := strings.Split(reminder.Notes, "\n")
			for _, noteLine := range noteLines {
				trimmed := strings.TrimSpace(noteLine)
				if trimmed != "" {
					noteText := "   " + metaStyle.Render(trimmed)
					if currentLine >= visibleStart && currentLine < visibleEnd {
						lines = append(lines, noteText)
					}
					currentLine++
				}
			}
		}

		// Add spacing between items
		if currentLine >= visibleStart && currentLine < visibleEnd {
			lines = append(lines, "")
		}
		currentLine++

		// If we're past the visible area and haven't started yet, skip
		if reminderStartLine > visibleEnd {
			break
		}
	}

	// Join lines and pad to fill height
	result := strings.Join(lines, "\n")

	// Pad with empty lines to fill height
	currentLines := len(lines)
	for currentLines < config.Height {
		result += "\n"
		currentLines++
	}

	// Apply container styling
	containerStyle := lipgloss.NewStyle().
		Width(config.Width).
		Height(config.Height).
		Padding(0, 2)

	return containerStyle.Render(result)
}

// RenderCompactList renders a more compact list view (as shown in the mockup)
func RenderCompactList(config ListViewConfig) string {
	if len(config.Reminders) == 0 {
		emptyStyle := lipgloss.NewStyle().
			Width(config.Width).
			Height(config.Height).
			Foreground(lipgloss.Color("245")).
			Italic(true).
			Padding(2, 4)
		return emptyStyle.Render("No reminders to display")
	}

	var lines []string
	metaStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("245"))

	// Calculate which lines are visible
	visibleStart := config.ScrollOffset
	visibleEnd := config.ScrollOffset + config.Height

	for i, reminder := range config.Reminders {
		if i < visibleStart {
			continue
		}
		if i >= visibleEnd {
			break
		}

		// Get countdown info
		countdown := ""
		if config.GetCountdown != nil {
			countdown, _ = config.GetCountdown(reminder.DueDate)
		}

		// Get list color
		listColor := "248"
		if config.ListColors != nil {
			if color, exists := config.ListColors[reminder.List]; exists {
				listColor = color
			}
		}
		listStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(listColor))

		// Format due date
		dueText := ""
		if config.FormatDueDate != nil {
			dueText = config.FormatDueDate(reminder.DueDate)
		}

		// Compact format: "Item 1  list 1  in 1 day  due ..."
		line := fmt.Sprintf("%s  %s  %s  %s",
			reminder.Title,
			listStyle.Render(reminder.List),
			countdown,
			metaStyle.Render("due "+dueText))

		// Highlight if focused
		if i == config.FocusedIndex {
			line = lipgloss.NewStyle().Bold(true).Render(line)
		}

		lines = append(lines, line)
	}

	// Join lines and pad to fill height
	result := strings.Join(lines, "\n")

	// Pad with empty lines to fill height
	currentLines := len(lines)
	for currentLines < config.Height {
		result += "\n"
		currentLines++
	}

	// Apply container styling
	containerStyle := lipgloss.NewStyle().
		Width(config.Width).
		Height(config.Height).
		Padding(0, 2)

	return containerStyle.Render(result)
}
