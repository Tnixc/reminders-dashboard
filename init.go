package main

import (
	"github.com/76creates/stickers/flexbox"
	"github.com/76creates/stickers/table"
	"github.com/charmbracelet/lipgloss"
)

// initialModel creates and returns the initial application model
func initialModel() model {
	// Define color palette
	availableColors := []string{
		"205", // pink/magenta
		"141", // purple
		"81",  // blue
		"51",  // cyan
		"48",  // teal
		"118", // green
		"226", // yellow
		"208", // orange
		"196", // red
		"248", // gray
	}
	colorNames := []string{
		"Pink",
		"Purple",
		"Blue",
		"Cyan",
		"Teal",
		"Green",
		"Yellow",
		"Orange",
		"Red",
		"Gray",
	}

	configPath := getConfigPath()
	config := loadConfig(configPath)

	// Initialize flexbox for layout management
	styleBackground := lipgloss.NewStyle().Align(lipgloss.Center)
	flexBox := flexbox.New(80, 24).SetStyle(styleBackground)

	// Initialize stickers table component
	headers := []string{"Title", "List", "Countdown", "Due Date"}
	t := table.NewTable(0, 0, headers)
	
	// Set table ratios and minimum widths
	ratio := []int{6, 3, 2, 2}
	minSize := []int{20, 10, 10, 10}
	t.SetRatio(ratio).SetMinWidth(minSize)
	t.SetStylePassing(true)

	return model{
		viewMode:          ColumnView,
		sidebarSection:    SidebarDaysFilter,
		sidebarFocused:    false,
		selectedLists:     config.SelectedLists,
		listColors:        config.ListColors,
		availableColors:   availableColors,
		colorNames:        colorNames,
		daysFilter:        config.DaysFilter,
		daysFilterOptions: []int{0, 1, 3, 7, 14, 30},
		columnOrder:       config.ColumnOrder,
		columnScrolls:     make([]int, 0),
		width:             80, // Default terminal width
		height:            24, // Default terminal height
		loading:           true,
		configPath:        configPath,
		flexBox:           flexBox,
		table:             t,
	}
}
