package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// SettingsConfig holds configuration for the settings panel
type SettingsConfig struct {
	Width           int
	Height          int
	Visible         bool
	DaysFilter      int
	DaysOptions     []int
	SelectedLists   map[string]bool
	AvailableLists  []string
	ListColors      map[string]string
	AvailableColors []string
	ColorNames      []string
	FocusedSection  int // 0 = days filter, 1 = list filter, 2 = color config
	CursorPosition  int
}

// SettingsSection represents different sections in settings
const (
	SectionDaysFilter = iota
	SectionListFilter
	SectionColorConfig
)

// RenderSettingsPanel renders the settings overlay/panel
func RenderSettingsPanel(config SettingsConfig) string {
	if !config.Visible {
		return ""
	}

	// Panel styling
	panelStyle := lipgloss.NewStyle().
		Width(config.Width).
		Height(config.Height).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("205")).
		Padding(1, 2).
		Background(lipgloss.Color("235"))

	// Header
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205")).
		MarginBottom(1)

	var content strings.Builder
	content.WriteString(headerStyle.Render("⚙ Settings"))
	content.WriteString("\n\n")

	// Days filter section
	content.WriteString(renderDaysFilterSettings(config))
	content.WriteString("\n\n")

	// List filter section
	content.WriteString(renderListFilterSettings(config))
	content.WriteString("\n\n")

	// Color configuration section
	content.WriteString(renderColorSettings(config))

	return panelStyle.Render(content.String())
}

// renderDaysFilterSettings renders the days filter section
func renderDaysFilterSettings(config SettingsConfig) string {
	sectionStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("248"))

	focusedStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205"))

	var lines []string

	// Section title
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("141"))

	if config.FocusedSection == SectionDaysFilter {
		lines = append(lines, focusedStyle.Render("▸ Days Filter"))
	} else {
		lines = append(lines, titleStyle.Render("  Days Filter"))
	}

	// Options
	for i, days := range config.DaysOptions {
		var label string
		if days == 0 {
			label = "All"
		} else {
			label = fmt.Sprintf("Next %d days", days)
		}

		cursor := "  "
		if config.FocusedSection == SectionDaysFilter && config.CursorPosition == i {
			cursor = "▸ "
		}

		check := "○"
		if days == config.DaysFilter {
			check = "●"
		}

		line := fmt.Sprintf("  %s%s %s", cursor, check, label)

		if config.FocusedSection == SectionDaysFilter && config.CursorPosition == i {
			line = focusedStyle.Render(line)
		} else {
			line = sectionStyle.Render(line)
		}

		lines = append(lines, line)
	}

	return strings.Join(lines, "\n")
}

// renderListFilterSettings renders the list filter section
func renderListFilterSettings(config SettingsConfig) string {
	sectionStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("248"))

	focusedStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205"))

	var lines []string

	// Section title
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("141"))

	if config.FocusedSection == SectionListFilter {
		lines = append(lines, focusedStyle.Render("▸ Visible Lists"))
	} else {
		lines = append(lines, titleStyle.Render("  Visible Lists"))
	}

	// List options
	for i, listName := range config.AvailableLists {
		cursor := "  "
		if config.FocusedSection == SectionListFilter && config.CursorPosition == i {
			cursor = "▸ "
		}

		check := "☐"
		if config.SelectedLists != nil && config.SelectedLists[listName] {
			check = "☑"
		}

		// Get list color
		listColor := "248"
		if config.ListColors != nil {
			if color, exists := config.ListColors[listName]; exists {
				listColor = color
			}
		}
		nameStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(listColor))

		line := fmt.Sprintf("  %s%s %s", cursor, check, nameStyle.Render(listName))

		if config.FocusedSection == SectionListFilter && config.CursorPosition == i {
			line = focusedStyle.Render(line)
		} else {
			line = sectionStyle.Render(line)
		}

		lines = append(lines, line)
	}

	return strings.Join(lines, "\n")
}

// renderColorSettings renders the color configuration section
func renderColorSettings(config SettingsConfig) string {
	sectionStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("248"))

	focusedStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205"))

	var lines []string

	// Section title
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("141"))

	if config.FocusedSection == SectionColorConfig {
		lines = append(lines, focusedStyle.Render("▸ List Colors"))
	} else {
		lines = append(lines, titleStyle.Render("  List Colors"))
	}

	// Color picker info
	if config.FocusedSection == SectionColorConfig {
		infoStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("245")).
			Italic(true)
		lines = append(lines, infoStyle.Render("  Press 1-9, 0 to set color"))
	}

	// List color assignments
	for i, listName := range config.AvailableLists {
		cursor := "  "
		if config.FocusedSection == SectionColorConfig && config.CursorPosition == i {
			cursor = "▸ "
		}

		// Get current color
		currentColor := "248"
		if config.ListColors != nil {
			if color, exists := config.ListColors[listName]; exists {
				currentColor = color
			}
		}

		colorBlock := lipgloss.NewStyle().
			Foreground(lipgloss.Color(currentColor)).
			Render("●")

		line := fmt.Sprintf("  %s%s %s", cursor, colorBlock, listName)

		if config.FocusedSection == SectionColorConfig && config.CursorPosition == i {
			line = focusedStyle.Render(line)
		} else {
			line = sectionStyle.Render(line)
		}

		lines = append(lines, line)
	}

	// Color palette
	if config.FocusedSection == SectionColorConfig {
		lines = append(lines, "")
		paletteStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("245"))
		lines = append(lines, paletteStyle.Render("  Color Palette:"))

		var palette strings.Builder
		palette.WriteString("  ")
		for i := 0; i < len(config.AvailableColors) && i < len(config.ColorNames); i++ {
			colorCode := config.AvailableColors[i]
			colorName := config.ColorNames[i]

			key := ""
			if i < 9 {
				key = fmt.Sprintf("%d", i+1)
			} else if i == 9 {
				key = "0"
			}

			colorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(colorCode))
			palette.WriteString(fmt.Sprintf("%s:%s ", key, colorStyle.Render(colorName)))

			// Line break every 5 colors
			if (i+1)%5 == 0 && i < len(config.AvailableColors)-1 {
				palette.WriteString("\n  ")
			}
		}

		lines = append(lines, paletteStyle.Render(palette.String()))
	}

	return strings.Join(lines, "\n")
}

// RenderSettingsButton renders a button to open settings
func RenderSettingsButton(width, height int, focused bool) string {
	text := "View settings\n(e.g. next n days,\nwhich lists)"

	borderColor := "240"
	if focused {
		borderColor = "205"
	}

	buttonStyle := lipgloss.NewStyle().
		Width(width).
		Height(height).
		Padding(1).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(borderColor)).
		Foreground(lipgloss.Color("248")).
		Align(lipgloss.Center, lipgloss.Center)

	return buttonStyle.Render(text)
}
