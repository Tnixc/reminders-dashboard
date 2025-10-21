package main

import (
	"fmt"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// formatDueDate formats a due date nicely
func formatDueDate(dueDateStr string) string {
	if dueDateStr == "" {
		return "No due date"
	}

	dueDate, err := time.Parse(time.RFC3339, dueDateStr)
	if err != nil {
		return dueDateStr
	}

	return dueDate.Format("Jan 02, 3:04 PM")
}

// getCountdown calculates time until due and returns a formatted string
func getCountdown(dueDateStr string) (string, int) {
	if dueDateStr == "" {
		return "No deadline", 999
	}

	dueDate, err := time.Parse(time.RFC3339, dueDateStr)
	if err != nil {
		return "", 999
	}

	now := time.Now()
	diff := dueDate.Sub(now)

	days := int(diff.Hours() / 24)
	hours := int(diff.Hours()) % 24

	if days < 0 {
		return fmt.Sprintf("Overdue by %d days", -days), -1
	} else if days == 0 {
		if hours < 0 {
			return "Overdue", -1
		} else if hours == 0 {
			return "Due now", 0
		}
		return fmt.Sprintf("Due in %d hours", hours), 0
	} else if days == 1 {
		return "Due tomorrow", 1
	} else if days < 7 {
		return fmt.Sprintf("Due in %d days", days), days
	} else {
		weeks := days / 7
		if weeks == 1 {
			return "Due in 1 week", days
		}
		return fmt.Sprintf("Due in %d weeks", weeks), days
	}
}

// getUrgencyColor returns a color based on urgency
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
