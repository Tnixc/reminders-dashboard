package components

import (
	"github.com/charmbracelet/lipgloss"
)

// FooterConfig holds configuration for the footer
type FooterConfig struct {
	Width   int
	HelpText string
}

// RenderFooter renders the footer with help text
func RenderFooter(config FooterConfig) string {
	footerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Background(lipgloss.Color("235")).
		Padding(0, 1).
		Width(config.Width)

	return footerStyle.Render(config.HelpText)
}

// GetHelpText returns appropriate help text based on current mode
func GetHelpText(mode string, settingsVisible bool) string {
	switch mode {
	case "column":
		if settingsVisible {
			return "esc: close settings • tab: section • space: toggle • ↑/↓: navigate • q: quit"
		}
		return "tab: view • s: settings • ←/→: columns • ↑/↓: items • c: calendar • r: refresh • q: quit"
	
	case "list":
		if settingsVisible {
			return "esc: close settings • tab: section • space: toggle • ↑/↓: navigate • q: quit"
		}
		return "tab: view • s: settings • ↑/↓: items • c: calendar • r: refresh • q: quit"
	
	case "settings":
		return "esc: close • tab: section • space: toggle • 1-9/0: pick color • ↑/↓: navigate • q: quit"
	
	default:
		return "tab: view • s: settings • r: refresh • q: quit"
	}
}