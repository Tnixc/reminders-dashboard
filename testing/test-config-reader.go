package main

import (
	"fmt"
	"os"

	"github.com/dlvhdr/reminders-dashboard/v4/internal/config"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run test-config-reader.go <config-file-path>")
		os.Exit(1)
	}

	configPath := os.Args[1]
	
	fmt.Println("=================================================================")
	fmt.Println("Config File Reading Test")
	fmt.Println("=================================================================")
	fmt.Printf("Reading config from: %s\n\n", configPath)

	// Parse the config
	cfg, err := config.ParseConfig(config.Location{
		RepoPath:   "",
		ConfigFlag: configPath,
	})

	if err != nil {
		fmt.Printf("❌ Error reading config: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✅ Config loaded successfully!\n")

	// Print out the loaded configuration
	fmt.Println("=================================================================")
	fmt.Println("Configuration Values")
	fmt.Println("=================================================================")
	
	fmt.Printf("Confirm Quit:              %v\n", cfg.ConfirmQuit)
	fmt.Printf("Smart Filtering at Launch: %v\n", cfg.SmartFilteringAtLaunch)
	fmt.Println()

	fmt.Println("--- Defaults ---")
	fmt.Printf("Reminders Limit:           %d\n", cfg.Defaults.RemindersLimit)
	fmt.Printf("View:                      %s\n", cfg.Defaults.View)
	fmt.Printf("Refetch Interval (min):    %d\n", cfg.Defaults.RefetchIntervalMinutes)
	fmt.Printf("Date Format:               %s\n", cfg.Defaults.DateFormat)
	fmt.Println()

	fmt.Println("--- Preview ---")
	fmt.Printf("Preview Open:              %v\n", cfg.Defaults.Preview.Open)
	fmt.Printf("Preview Width:             %d\n", cfg.Defaults.Preview.Width)
	fmt.Println()

	fmt.Println("--- Layout (Reminders) ---")
	if cfg.Defaults.Layout.Reminders.Title.Width != nil {
		fmt.Printf("Title Width:               %d\n", *cfg.Defaults.Layout.Reminders.Title.Width)
	}
	if cfg.Defaults.Layout.Reminders.List.Width != nil {
		fmt.Printf("List Width:                %d\n", *cfg.Defaults.Layout.Reminders.List.Width)
	}
	if cfg.Defaults.Layout.Reminders.DueIn.Width != nil {
		fmt.Printf("DueIn Width:               %d\n", *cfg.Defaults.Layout.Reminders.DueIn.Width)
	}
	if cfg.Defaults.Layout.Reminders.Date.Hidden != nil {
		fmt.Printf("Date Hidden:               %v\n", *cfg.Defaults.Layout.Reminders.Date.Hidden)
	}
	if cfg.Defaults.Layout.Reminders.Priority.Width != nil {
		fmt.Printf("Priority Width:            %d\n", *cfg.Defaults.Layout.Reminders.Priority.Width)
	}
	if cfg.Defaults.Layout.Reminders.Completed.Width != nil {
		fmt.Printf("Completed Width:           %d\n", *cfg.Defaults.Layout.Reminders.Completed.Width)
	}
	fmt.Println()

	fmt.Println("--- Reminders Sections ---")
	fmt.Printf("Number of sections:        %d\n", len(cfg.RemindersSections))
	for i, section := range cfg.RemindersSections {
		fmt.Printf("  [%d] Title:   %s\n", i, section.Title)
		fmt.Printf("      Filters: %s\n", section.Filters)
		if section.Limit != nil {
			fmt.Printf("      Limit:   %d\n", *section.Limit)
		}
	}
	fmt.Println()

	fmt.Println("--- List Colors ---")
	if len(cfg.ListColors) > 0 {
		for list, color := range cfg.ListColors {
			fmt.Printf("  %-12s → %s\n", list, color)
		}
	} else {
		fmt.Println("  (none)")
	}
	fmt.Println()

	if cfg.Theme != nil {
		fmt.Println("--- Theme ---")
		fmt.Printf("Sections Show Count:       %v\n", cfg.Theme.Ui.SectionsShowCount)
		fmt.Printf("Table Show Separator:      %v\n", cfg.Theme.Ui.Table.ShowSeparator)
		fmt.Printf("Table Compact:             %v\n", cfg.Theme.Ui.Table.Compact)
		
		if cfg.Theme.Colors != nil {
			fmt.Println()
			fmt.Println("--- Theme Colors ---")
			fmt.Printf("Text Primary:              %s\n", cfg.Theme.Colors.Text.Primary)
			fmt.Printf("Text Secondary:            %s\n", cfg.Theme.Colors.Text.Secondary)
			fmt.Printf("Text Faint:                %s\n", cfg.Theme.Colors.Text.Faint)
			fmt.Printf("Text Inverted:             %s\n", cfg.Theme.Colors.Text.Inverted)
			fmt.Printf("Text Success:              %s\n", cfg.Theme.Colors.Text.Success)
			fmt.Printf("Text Warning:              %s\n", cfg.Theme.Colors.Text.Warning)
			fmt.Printf("Text Error:                %s\n", cfg.Theme.Colors.Text.Error)
			fmt.Printf("Border Primary:            %s\n", cfg.Theme.Colors.Border.Primary)
			fmt.Printf("Border Secondary:          %s\n", cfg.Theme.Colors.Border.Secondary)
			fmt.Printf("Border Faint:              %s\n", cfg.Theme.Colors.Border.Faint)
			fmt.Printf("Background Selected:       %s\n", cfg.Theme.Colors.Background.Selected)
		}
		fmt.Println()
	}

	fmt.Println("--- Keybindings ---")
	fmt.Printf("Universal keybindings:     %d\n", len(cfg.Keybindings.Universal))
	for i, kb := range cfg.Keybindings.Universal {
		name := kb.Name
		if name == "" {
			name = kb.Builtin
		}
		fmt.Printf("  [%d] %-10s → %s\n", i, kb.Key, name)
	}
	fmt.Printf("Reminders keybindings:     %d\n", len(cfg.Keybindings.Reminders))
	for i, kb := range cfg.Keybindings.Reminders {
		name := kb.Name
		if name == "" {
			name = kb.Builtin
		}
		fmt.Printf("  [%d] %-10s → %s\n", i, kb.Key, name)
	}
	fmt.Println()

	fmt.Println("=================================================================")
	fmt.Println("✅ All configuration values loaded successfully!")
	fmt.Println("=================================================================")
}