package main

import (
	"encoding/json"
	"os/exec"
	"sort"
	"time"

	"github.com/charmbracelet/bubbles/list"
)

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
}

func loadReminders() ([]list.Item, error) {
	return loadRemindersFiltered(nil)
}

func loadRemindersFiltered(enabledLists []string) ([]list.Item, error) {
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
			desc = "Overdue"
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

	return item{
		title:       r.Title,
		description: desc,
	}
}
