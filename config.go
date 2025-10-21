package main

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// getConfigPath returns the path to the config file
func getConfigPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "config.json"
	}
	return filepath.Join(homeDir, ".config", "reminders-dashboard", "config.json")
}

// loadConfig loads the configuration from disk
func loadConfig(path string) Config {
	config := Config{
		ListColors:    make(map[string]string),
		SelectedLists: make(map[string]bool),
		DaysFilter:    0,
		ColumnOrder:   []string{},
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return config // Return default config if file doesn't exist
	}

	json.Unmarshal(data, &config)
	return config
}

// saveConfig saves the configuration to disk
func (m *model) saveConfig() error {
	config := Config{
		ListColors:    m.listColors,
		SelectedLists: m.selectedLists,
		DaysFilter:    m.daysFilter,
		ColumnOrder:   m.columnOrder,
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	// Create config directory if it doesn't exist
	configDir := filepath.Dir(m.configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	return os.WriteFile(m.configPath, data, 0644)
}
