package components

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// CalendarConfig holds configuration for the calendar widget
type CalendarConfig struct {
	Width       int
	Height      int
	CurrentDate time.Time
	Focused     bool
}

// RenderCalendar renders a small calendar widget
func RenderCalendar(config CalendarConfig) string {
	borderColor := "240"
	if config.Focused {
		borderColor = "205"
	}

	calendarStyle := lipgloss.NewStyle().
		Width(config.Width).
		Height(config.Height).
		Padding(1).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(borderColor))

	content := renderCalendarContent(config)

	return calendarStyle.Render(content)
}

// renderCalendarContent generates the calendar content
func renderCalendarContent(config CalendarConfig) string {
	now := config.CurrentDate
	if now.IsZero() {
		now = time.Now()
	}

	// Header with month and year
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("141")).
		Align(lipgloss.Center)

	header := headerStyle.Render(now.Format("Jan 2006"))

	// Day headers (Mon-Sun)
	dayHeaderStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("245"))

	dayHeaders := "Mo Tu We Th Fr Sa Su"

	// Get the first day of the month
	firstDay := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	
	// Get the last day of the month
	lastDay := firstDay.AddDate(0, 1, -1)

	// Calculate starting position (weekday offset)
	// Monday = 0, Sunday = 6
	offset := int(firstDay.Weekday()) - 1
	if offset < 0 {
		offset = 6 // Sunday
	}

	var rows []string
	var week []string

	// Add empty cells for offset
	for i := 0; i < offset; i++ {
		week = append(week, "  ")
	}

	// Add days of the month
	currentDayStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205"))

	dayStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("248"))

	for day := 1; day <= lastDay.Day(); day++ {
		dayStr := fmt.Sprintf("%2d", day)

		// Highlight current day
		if day == now.Day() {
			dayStr = currentDayStyle.Render(dayStr)
		} else {
			dayStr = dayStyle.Render(dayStr)
		}

		week = append(week, dayStr)

		// If week is complete (7 days), add to rows
		if len(week) == 7 {
			rows = append(rows, strings.Join(week, " "))
			week = []string{}
		}
	}

	// Add remaining week if not complete
	if len(week) > 0 {
		// Pad with empty cells
		for len(week) < 7 {
			week = append(week, "  ")
		}
		rows = append(rows, strings.Join(week, " "))
	}

	// Combine all parts
	var content strings.Builder
	content.WriteString(header)
	content.WriteString("\n")
	content.WriteString(dayHeaderStyle.Render(dayHeaders))
	content.WriteString("\n")
	content.WriteString(strings.Join(rows, "\n"))

	return content.String()
}

// RenderMiniCalendar renders a minimal calendar placeholder
func RenderMiniCalendar(width, height int, focused bool) string {
	borderColor := "240"
	if focused {
		borderColor = "205"
	}

	now := time.Now()
	
	// Simple mini calendar with just month and current day
	contentStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("248")).
		Align(lipgloss.Center, lipgloss.Center)

	content := fmt.Sprintf("%s\n\n%d",
		now.Format("Jan 2006"),
		now.Day())

	calendarStyle := lipgloss.NewStyle().
		Width(width).
		Height(height).
		Padding(1).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(borderColor)).
		Align(lipgloss.Center, lipgloss.Center)

	return calendarStyle.Render(contentStyle.Render(content))
}