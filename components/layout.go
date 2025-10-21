package components

import (
	"github.com/charmbracelet/lipgloss"
)

// LayoutConfig holds configuration for the main layout
type LayoutConfig struct {
	Width            int
	Height           int
	ViewMode         string // "column" or "list"
	SettingsVisible  bool
	CalendarVisible  bool
	MainContent      string
	SettingsPanel    string
	Calendar         string
	SettingsButton   string
	Footer           string
}

// RenderLayout renders the complete application layout
func RenderLayout(config LayoutConfig) string {
	// Calculate dimensions
	footerHeight := 1
	contentHeight := config.Height - footerHeight

	// Two main layouts:
	// 1. Main content + sidebar (settings button + calendar)
	// 2. If settings visible, overlay settings panel

	if config.SettingsVisible {
		// Settings overlay mode
		return renderLayoutWithSettingsOverlay(config, contentHeight)
	}

	// Normal mode with sidebar
	return renderLayoutWithSidebar(config, contentHeight)
}

// renderLayoutWithSidebar renders the layout with sidebar
func renderLayoutWithSidebar(config LayoutConfig, contentHeight int) string {
	// Calculate sidebar width
	sidebarWidth := 24
	if config.Width < 80 {
		sidebarWidth = 20
	}
	if sidebarWidth > config.Width/4 {
		sidebarWidth = config.Width / 4
	}

	mainWidth := config.Width - sidebarWidth

	// Build sidebar
	sidebar := buildSidebar(config, sidebarWidth, contentHeight)

	// Style main content area
	mainStyle := lipgloss.NewStyle().
		Width(mainWidth).
		Height(contentHeight)

	sidebarStyle := lipgloss.NewStyle().
		Width(sidebarWidth).
		Height(contentHeight)

	// Combine horizontally
	content := lipgloss.JoinHorizontal(
		lipgloss.Top,
		mainStyle.Render(config.MainContent),
		sidebarStyle.Render(sidebar),
	)

	// Add footer
	return lipgloss.JoinVertical(
		lipgloss.Left,
		content,
		config.Footer,
	)
}

// renderLayoutWithSettingsOverlay renders layout with settings as overlay
func renderLayoutWithSettingsOverlay(config LayoutConfig, contentHeight int) string {
	// Settings takes up center portion of screen
	settingsWidth := config.Width * 2 / 3
	if settingsWidth > 60 {
		settingsWidth = 60
	}
	if settingsWidth < 40 {
		settingsWidth = 40
	}

	settingsHeight := contentHeight * 3 / 4
	if settingsHeight > 30 {
		settingsHeight = 30
	}

	// Position settings in center
	leftPadding := (config.Width - settingsWidth) / 2
	topPadding := (contentHeight - settingsHeight) / 2

	// Create background (dimmed main content)
	backgroundStyle := lipgloss.NewStyle().
		Width(config.Width).
		Height(contentHeight).
		Foreground(lipgloss.Color("240"))

	background := backgroundStyle.Render(config.MainContent)

	// Overlay settings panel
	// For simplicity, we'll place it using padding
	settingsContainer := lipgloss.NewStyle().
		Width(config.Width).
		Height(contentHeight).
		PaddingLeft(leftPadding).
		PaddingTop(topPadding)

	// This is a simplified overlay - in a real implementation,
	// you'd use lipgloss.Place to properly overlay
	content := lipgloss.Place(
		config.Width,
		contentHeight,
		lipgloss.Center,
		lipgloss.Center,
		config.SettingsPanel,
		lipgloss.WithWhitespaceChars(" "),
		lipgloss.WithWhitespaceForeground(lipgloss.Color("235")),
	)

	// Add footer
	return lipgloss.JoinVertical(
		lipgloss.Left,
		content,
		config.Footer,
	)
}

// buildSidebar creates the sidebar content
func buildSidebar(config LayoutConfig, width, height int) string {
	// Calculate heights for sidebar elements
	settingsButtonHeight := 8
	calendarHeight := height - settingsButtonHeight - 2 // 2 for spacing

	if calendarHeight < 6 {
		calendarHeight = 6
		settingsButtonHeight = height - calendarHeight - 2
	}

	var elements []string

	// Settings button
	if config.SettingsButton != "" {
		elements = append(elements, config.SettingsButton)
	} else {
		elements = append(elements, RenderSettingsButton(width-2, settingsButtonHeight, false))
	}

	// Spacing
	elements = append(elements, "")

	// Calendar
	if config.Calendar != "" {
		elements = append(elements, config.Calendar)
	} else {
		elements = append(elements, RenderMiniCalendar(width-2, calendarHeight, false))
	}

	// Join vertically and pad to full height
	sidebarContent := lipgloss.JoinVertical(lipgloss.Left, elements...)

	sidebarStyle := lipgloss.NewStyle().
		Width(width).
		Height(height).
		Padding(0, 1)

	return sidebarStyle.Render(sidebarContent)
}

// RenderMainArea renders just the main content area (list or column view)
func RenderMainArea(config LayoutConfig) string {
	if config.ViewMode == "column" {
		return config.MainContent
	}
	return config.MainContent
}