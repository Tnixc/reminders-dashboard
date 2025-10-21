package main

import (
	"github.com/76creates/stickers/flexbox"
	"github.com/76creates/stickers/table"
)

// Reminder represents a single reminder from the JSON
type Reminder struct {
	DueDate     string `json:"dueDate"`
	ExternalID  string `json:"externalId"`
	IsCompleted bool   `json:"isCompleted"`
	List        string `json:"list"`
	Priority    int    `json:"priority"`
	StartDate   string `json:"startDate"`
	Title       string `json:"title"`
	Notes       string `json:"notes"`
}

// Config represents persistent user configuration
type Config struct {
	ListColors    map[string]string `json:"listColors"`
	SelectedLists map[string]bool   `json:"selectedLists"`
	DaysFilter    int               `json:"daysFilter"`
	ColumnOrder   []string          `json:"columnOrder"`
}

// ViewMode represents the current view type
type ViewMode int

const (
	ColumnView ViewMode = iota
	ListView
)

// SidebarSection represents which part of the sidebar is focused
type SidebarSection int

const (
	SidebarDaysFilter SidebarSection = iota
	SidebarListFilter
	SidebarColorConfig
)

// Model represents the application state
type model struct {
	reminders         []Reminder
	viewMode          ViewMode
	sidebarSection    SidebarSection
	availableLists    []string
	selectedLists     map[string]bool
	listColors        map[string]string // list name -> color code
	availableColors   []string          // palette of colors to choose from
	colorNames        []string          // human-readable color names
	daysFilter        int               // 0 = all, 7 = next 7 days, etc.
	daysFilterOptions []int
	columnOrder       []string // custom order for columns
	cursor            int
	columnCursor      int   // which column is focused in column view
	scrollOffset      int   // scroll position for list view
	columnScrolls     []int // scroll positions for each column in column view
	sidebarCursor     int
	sidebarFocused    bool // whether sidebar or main content is focused
	width             int
	height            int
	err               error
	// Color picker (two-step) state
	colorPickerActive bool // true when awaiting color selection for the chosen list
	colorPickerList   int  // index of the list currently targeted for recolor
	colorPickerCursor int  // index within availableColors palette for preview/selection
	loading           bool
	configPath        string
	// Flexbox and table components
	flexBox           *flexbox.FlexBox
	table             *table.Table
}

// remindersLoadedMsg is sent when reminders are loaded
type remindersLoadedMsg struct {
	reminders []Reminder
	err       error
}

// reminderBounds is used for calculating line boundaries in list view
type reminderBounds struct {
	startLine int
	endLine   int
}
