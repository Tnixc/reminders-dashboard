package main

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
	}
}
