package components

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Reminder represents a single reminder (copied from main types)
type Reminder struct {
	DueDate     string
	ExternalID  string
	IsCompleted bool
	List        string
	Priority    int
	StartDate   string
	Title       string
	Notes       string
}

// CardStyle defines styling for reminder cards
type CardStyle struct {
	Width         int
	Height        int
	Focused       bool
	ListColor     string
	BorderColor   string
	PaddingTop    int
	PaddingBottom int
	PaddingLeft   int
	PaddingRight  int
}

// DefaultCardStyle returns the default card styling
func DefaultCardStyle() CardStyle {
	return CardStyle{
		Width:         30,
		Height:        0, // auto
		Focused:       false,
		ListColor:     "248",
		BorderColor:   "240",
		PaddingTop:    0,
		PaddingBottom: 0,
		PaddingLeft:   1,
		PaddingRight:  1,
	}
}

// RenderCard renders a single reminder as a card
func RenderCard(reminder Reminder, style CardStyle, countdown string, urgency int) string {
	// Styles
	titleStyle := lipgloss.NewStyle().Bold(true)
	metaStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	urgencyColor := getUrgencyColor(urgency)
	countdownStyle := lipgloss.NewStyle().Foreground(urgencyColor)

	// Build card content
	var lines []string

	// Title
	title := reminder.Title
	if len(title) > style.Width-4 {
		title = title[:style.Width-7] + "..."
	}
	lines = append(lines, titleStyle.Render(title))

	// Due date
	if reminder.DueDate != "" {
		dueText := formatDueDate(reminder.DueDate)
		if len(dueText) > style.Width-4 {
			dueText = dueText[:style.Width-7] + "..."
		}
		lines = append(lines, metaStyle.Render(dueText))
	}

	// Countdown
	lines = append(lines, countdownStyle.Render(countdown))

	// Notes (if any, truncated)
	if reminder.Notes != "" {
		noteLines := strings.Split(reminder.Notes, "\n")
		notesAdded := 0
		for _, line := range noteLines {
			if strings.TrimSpace(line) != "" && notesAdded < 2 {
				if len(line) > style.Width-4 {
					line = line[:style.Width-7] + "..."
				}
				lines = append(lines, metaStyle.Render(line))
				notesAdded++
			}
		}
	}

	content := strings.Join(lines, "\n")



	cardStyle := lipgloss.NewStyle().
		Width(style.Width).
		Padding(style.PaddingTop, style.PaddingRight, style.PaddingBottom, style.PaddingLeft)

	if style.Height > 0 {
		cardStyle = cardStyle.Height(style.Height)
	}

	return cardStyle.Render(content)
}

// RenderEmptyCard renders an empty card placeholder
func RenderEmptyCard(width, height int, message string) string {
	style := lipgloss.NewStyle().
		Width(width).
		Height(height).
		Padding(1).

		Foreground(lipgloss.Color("245")).
		Italic(true).
		Align(lipgloss.Center, lipgloss.Center)

	return style.Render(message)
}

// Helper functions (these match the main package utils)

func formatDueDate(dueDateStr string) string {
	if dueDateStr == "" {
		return "No due date"
	}
	// Simple format - in real implementation this would parse RFC3339
	return dueDateStr
}

func getUrgencyColor(urgency int) lipgloss.Color {
	if urgency < 0 {
		return lipgloss.Color("196") // red - overdue
	} else if urgency == 0 {
		return lipgloss.Color("208") // orange - today
	} else if urgency <= 3 {
		return lipgloss.Color("226") // yellow - very soon
	} else if urgency <= 7 {
		return lipgloss.Color("220") // light yellow - this week
	}
	return lipgloss.Color("248") // gray - future
}
