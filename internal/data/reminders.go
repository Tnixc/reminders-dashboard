package data

import (
	"encoding/json"
	"os/exec"
	"sort"
	"time"

	"github.com/dlvhdr/reminders-dashboard/v4/internal/tui/theme"
)

type Reminder struct {
	Id          int
	Title       string
	Notes       string
	DueDate     time.Time
	List        string
	Priority    int
	IsCompleted bool
}

type tempReminder struct {
	DueDate     string `json:"dueDate"`
	ExternalID  string `json:"externalId"`
	IsCompleted bool   `json:"isCompleted"`
	List        string `json:"list"`
	Priority    int    `json:"priority"`
	StartDate   string `json:"startDate"`
	Title       string `json:"title"`
	Notes       string `json:"notes"`
}

func (data Reminder) GetTitle() string {
	return data.Title
}

func (data Reminder) GetNumber() int {
	return data.Id
}

func (data Reminder) GetUrl() string {
	return "" // No URL for reminders
}

func (data Reminder) GetUpdatedAt() time.Time {
	return data.DueDate
}

func (data Reminder) GetCreatedAt() time.Time {
	return data.DueDate
}

func (data Reminder) GetAuthor(theme theme.Theme, showAuthorIcon bool) string {
	return "" // No author
}

func (data Reminder) GetRepoNameWithOwner() string {
	return data.List
}

type RemindersResponse struct {
	Reminders  []Reminder
	TotalCount int
	PageInfo   PageInfo
}

func FetchReminders(query string, limit int, pageInfo *PageInfo) (RemindersResponse, error) {
	cmd := exec.Command("reminders", "show-all", "-f", "json")
	output, err := cmd.Output()
	if err != nil {
		return RemindersResponse{}, err
	}

	var tempReminders []tempReminder
	if err := json.Unmarshal(output, &tempReminders); err != nil {
		return RemindersResponse{}, err
	}

	// Filter out completed and convert
	var active []Reminder
	for i, r := range tempReminders {
		if !r.IsCompleted {
			due, _ := time.Parse(time.RFC3339, r.DueDate)
			active = append(active, Reminder{
				Id:          i + 1,
				Title:       r.Title,
				Notes:       r.Notes,
				DueDate:     due,
				List:        r.List,
				Priority:    r.Priority,
				IsCompleted: r.IsCompleted,
			})
		}
	}

	// Sort by due date
	sort.Slice(active, func(i, j int) bool {
		if active[i].DueDate.IsZero() {
			return false
		}
		if active[j].DueDate.IsZero() {
			return true
		}
		return active[i].DueDate.Before(active[j].DueDate)
	})

	return RemindersResponse{
		Reminders:  active,
		TotalCount: len(active),
		PageInfo: PageInfo{
			HasNextPage: false,
		},
	}, nil
}