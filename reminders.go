package main

import (
	"encoding/json"
	"os/exec"
	"sort"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// loadRemindersCmd fetches reminders from the CLI
func loadRemindersCmd() tea.Msg {
	cmd := exec.Command("reminders", "show-all", "-f", "json")
	output, err := cmd.Output()
	if err != nil {
		return remindersLoadedMsg{err: err}
	}

	var reminders []Reminder
	if err := json.Unmarshal(output, &reminders); err != nil {
		return remindersLoadedMsg{err: err}
	}

	// Filter out completed reminders
	var active []Reminder
	for _, r := range reminders {
		if !r.IsCompleted {
			active = append(active, r)
		}
	}

	return remindersLoadedMsg{reminders: active}
}

// sortReminders sorts reminders by due date
func sortReminders(m *model) {
	sort.Slice(m.reminders, func(i, j int) bool {
		// Reminders without due dates go last
		if m.reminders[i].DueDate == "" {
			return false
		}
		if m.reminders[j].DueDate == "" {
			return true
		}

		ti, _ := time.Parse(time.RFC3339, m.reminders[i].DueDate)
		tj, _ := time.Parse(time.RFC3339, m.reminders[j].DueDate)
		return ti.Before(tj)
	})
}

// updateAvailableLists extracts unique list names
func updateAvailableLists(m *model) {
	listMap := make(map[string]bool)
	for _, r := range m.reminders {
		listMap[r.List] = true
	}

	m.availableLists = make([]string, 0, len(listMap))
	for list := range listMap {
		m.availableLists = append(m.availableLists, list)
	}
	sort.Strings(m.availableLists)
}

// getFilteredReminders returns reminders filtered by selected lists and days
func getFilteredReminders(m *model) []Reminder {
	var filtered []Reminder
	now := time.Now()

	for _, r := range m.reminders {
		// Filter by selected lists
		if !m.selectedLists[r.List] {
			continue
		}

		// Filter by days
		if m.daysFilter > 0 && r.DueDate != "" {
			dueDate, err := time.Parse(time.RFC3339, r.DueDate)
			if err == nil {
				daysDiff := int(dueDate.Sub(now).Hours() / 24)
				if daysDiff > m.daysFilter {
					continue
				}
				// Also skip if overdue by more than daysFilter
				if daysDiff < -m.daysFilter {
					continue
				}
			}
		}

		filtered = append(filtered, r)
	}
	return filtered
}

// getColumns returns reminders grouped by list for column view
func getColumns(m *model) [][]Reminder {
	filtered := getFilteredReminders(m)
	
	// Group reminders by list
	listMap := make(map[string][]Reminder)
	for _, r := range filtered {
		listMap[r.List] = append(listMap[r.List], r)
	}
	
	// Get list names that have reminders
	var availableLists []string
	for listName := range listMap {
		availableLists = append(availableLists, listName)
	}
	
	// Determine column order
	var columnOrder []string
	if len(m.columnOrder) > 0 {
		// Use custom order, but only for lists that exist
		for _, listName := range m.columnOrder {
			if _, exists := listMap[listName]; exists {
				columnOrder = append(columnOrder, listName)
			}
		}
		// Add any new lists that aren't in the custom order
		for _, listName := range availableLists {
			found := false
			for _, ordered := range columnOrder {
				if ordered == listName {
					found = true
					break
				}
			}
			if !found {
				columnOrder = append(columnOrder, listName)
			}
		}
	} else {
		// No custom order - use alphabetical
		columnOrder = availableLists
		sort.Strings(columnOrder)
	}
	
	// Build columns in order
	columns := make([][]Reminder, 0)
	for _, listName := range columnOrder {
		if reminders, exists := listMap[listName]; exists && len(reminders) > 0 {
			columns = append(columns, reminders)
		}
	}
	
	return columns
}

// getColumnListNames returns the list names for each column in order
func getColumnListNames(m *model) []string {
	columns := getColumns(m)
	names := make([]string, len(columns))
	for i, col := range columns {
		if len(col) > 0 {
			names[i] = col[0].List
		}
	}
	return names
}
