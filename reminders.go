package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/charmbracelet/bubbles/list"
)

type Config struct {
	ListColors map[string]string `toml:"listColors"`
}

var listColorMap map[string]string

func loadConfig() error {
	configPath := filepath.Join(os.ExpandEnv("$HOME"), ".config", "reminders-dashboard", "config.toml")

	data, err := os.ReadFile(configPath)
	if err != nil {
		// Config file is optional
		listColorMap = make(map[string]string)
		return nil
	}

	var config Config
	if err := toml.Unmarshal(data, &config); err != nil {
		// Invalid config, ignore and continue
		listColorMap = make(map[string]string)
		return nil
	}

	// Store colors with lowercase keys for case-insensitive lookup
	listColorMap = make(map[string]string)
	for name, color := range config.ListColors {
		listColorMap[strings.ToLower(name)] = color
	}

	return nil
}

type Reminder struct {
	Title       string    `json:"title"`
	DueDate     string    `json:"dueDate,omitempty"`
	StartDate   string    `json:"startDate,omitempty"`
	List        string    `json:"list"`
	Priority    int       `json:"priority"`
	IsCompleted bool      `json:"isCompleted"`
	ExternalID  string    `json:"externalId"`
	Notes       string    `json:"notes,omitempty"`
	parsedDate  time.Time // for sorting
	Color       string    // color from config
	TimeColor   string    // color for urgency display
}

func loadReminders() ([]list.Item, error) {
	return loadRemindersFiltered(nil)
}

func loadRemindersFiltered(enabledLists []string) ([]list.Item, error) {
	// Load config on first call
	if listColorMap == nil {
		loadConfig()
	}

	// Execute the reminders command
	cmd := exec.Command("reminders", "show-all", "-f", "json")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	// Parse JSON
	var reminders []Reminder
	if err := json.Unmarshal(output, &reminders); err != nil {
		return nil, err
	}

	// Parse dates and filter out completed reminders
	var activeReminders []Reminder
	for _, r := range reminders {
		if r.IsCompleted {
			continue
		}

		// Lookup color for this list (case-insensitive)
		r.Color = listColorMap[strings.ToLower(r.List)]

		// Filter by enabled lists if specified
		if enabledLists != nil && len(enabledLists) > 0 {
			listEnabled := false
			for _, enabledList := range enabledLists {
				if r.List == enabledList {
					listEnabled = true
					break
				}
			}
			if !listEnabled {
				continue
			}
		}

		// Parse due date for sorting
		if r.DueDate != "" {
			if t, err := time.Parse(time.RFC3339, r.DueDate); err == nil {
				r.parsedDate = t
			}
		}
		activeReminders = append(activeReminders, r)
	}

	// Sort by due date
	sort.Slice(activeReminders, func(i, j int) bool {
		// Items without due dates go to the end
		if activeReminders[i].parsedDate.IsZero() && !activeReminders[j].parsedDate.IsZero() {
			return false
		}
		if !activeReminders[i].parsedDate.IsZero() && activeReminders[j].parsedDate.IsZero() {
			return true
		}
		if activeReminders[i].parsedDate.IsZero() && activeReminders[j].parsedDate.IsZero() {
			return activeReminders[i].Title < activeReminders[j].Title
		}
		return activeReminders[i].parsedDate.Before(activeReminders[j].parsedDate)
	})

	// Convert to list items
	items := make([]list.Item, len(activeReminders))
	for i, r := range activeReminders {
		items[i] = reminderToItem(r)
	}

	return items, nil
}

func getUniqueLists() ([]string, error) {
	// Execute the reminders command
	cmd := exec.Command("reminders", "show-all", "-f", "json")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	// Parse JSON
	var reminders []Reminder
	if err := json.Unmarshal(output, &reminders); err != nil {
		return nil, err
	}

	// Collect unique list names
	listSet := make(map[string]bool)
	for _, r := range reminders {
		if !r.IsCompleted && r.List != "" {
			listSet[r.List] = true
		}
	}

	// Convert to slice and sort
	lists := make([]string, 0, len(listSet))
	for listName := range listSet {
		lists = append(lists, listName)
	}
	sort.Strings(lists)

	return lists, nil
}

func calculateRelativeTime(dueDate time.Time) (string, string) {
	now := time.Now()

	if dueDate.Before(now) {
		return "Overdue", "red"
	}

	diff := dueDate.Sub(now)

	// Determine urgency color
	urgencyColor := "" // default, no special color
	if diff <= 24*time.Hour {
		urgencyColor = "red"
	} else if diff <= 3*24*time.Hour {
		urgencyColor = "orange"
	} else if diff <= 7*24*time.Hour {
		urgencyColor = "yellow"
	}

	// Calculate days and hours
	days := int(diff.Hours() / 24)
	hours := int(diff.Hours()) % 24

	// Format the time string
	var timeStr string
	if days == 0 {
		// Less than a day
		if hours == 1 {
			timeStr = "1 hour"
		} else {
			timeStr = fmt.Sprintf("%d hours", hours)
		}
	} else if days == 1 {
		if hours == 0 {
			timeStr = "1 day"
		} else if hours == 1 {
			timeStr = "1 day 1 hour"
		} else {
			timeStr = fmt.Sprintf("1 day %d hours", hours)
		}
	} else if days < 7 {
		if hours == 0 {
			timeStr = fmt.Sprintf("%d days", days)
		} else if hours == 1 {
			timeStr = fmt.Sprintf("%d days 1 hour", days)
		} else {
			timeStr = fmt.Sprintf("%d days %d hours", days, hours)
		}
	} else {
		// Format as weeks and days
		weeks := days / 7
		remainingDays := days % 7

		if weeks == 1 {
			if remainingDays == 0 {
				timeStr = "1 week"
			} else if remainingDays == 1 {
				timeStr = "1 week 1 day"
			} else {
				timeStr = fmt.Sprintf("1 week %d days", remainingDays)
			}
		} else {
			if remainingDays == 0 {
				timeStr = fmt.Sprintf("%d weeks", weeks)
			} else if remainingDays == 1 {
				timeStr = fmt.Sprintf("%d weeks 1 day", weeks)
			} else {
				timeStr = fmt.Sprintf("%d weeks %d days", weeks, remainingDays)
			}
		}
	}

	return "Due in " + timeStr, urgencyColor
}

func reminderToItem(r Reminder) item {
	var desc string

	// Format the description with due date and list
	if !r.parsedDate.IsZero() {
		// Format date nicely
		now := time.Now()
		dueDate := r.parsedDate

		// Calculate relative time
		if dueDate.Year() == now.Year() && dueDate.Month() == now.Month() && dueDate.Day() == now.Day() {
			desc = "Today"
		} else if dueDate.Year() == now.Year() && dueDate.Month() == now.Month() && dueDate.Day() == now.Day()+1 {
			desc = "Tomorrow"
		} else if dueDate.Before(now) {
			// For overdue items, show the actual date instead of "Overdue"
			// The urgency will be shown in the colored text
			if dueDate.Year() == now.Year() {
				desc = dueDate.Format("Jan 2")
			} else {
				desc = dueDate.Format("Jan 2, 2006")
			}
		} else {
			// Show date in format like "Nov 1" or "Nov 1, 2026" if not current year
			if dueDate.Year() == now.Year() {
				desc = dueDate.Format("Jan 2")
			} else {
				desc = dueDate.Format("Jan 2, 2006")
			}
		}

		// Add list name
		desc += " • " + r.List
	} else {
		desc = r.List
	}

	// Calculate urgency text and color
	urgencyText := ""
	urgencyColor := ""
	if !r.parsedDate.IsZero() {
		urgencyText, urgencyColor = calculateRelativeTime(r.parsedDate)
	}

	return item{
		title:        r.Title,
		description:  desc,
		listName:     r.List,
		color:        r.Color,
		urgencyText:  urgencyText,
		urgencyColor: urgencyColor,
		parsedDate:   r.parsedDate,
		externalID:   r.ExternalID,
		completed:    r.IsCompleted,
	}
}
