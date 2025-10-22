package theme

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"

	"github.com/dlvhdr/reminders-dashboard/v4/internal/config"
)

type Theme struct {
	SelectedBackground lipgloss.AdaptiveColor // config.Theme.Colors.Background.Selected
	PrimaryBorder      lipgloss.AdaptiveColor // config.Theme.Colors.Border.Primary
	FaintBorder        lipgloss.AdaptiveColor // config.Theme.Colors.Border.Faint
	SecondaryBorder    lipgloss.AdaptiveColor // config.Theme.Colors.Border.Secondary
	FaintText          lipgloss.AdaptiveColor // config.Theme.Colors.Text.Faint
	PrimaryText        lipgloss.AdaptiveColor // config.Theme.Colors.Text.Primary
	SecondaryText      lipgloss.AdaptiveColor // config.Theme.Colors.Text.Secondary
	InvertedText       lipgloss.AdaptiveColor // config.Theme.Colors.Text.Inverted
	SuccessText        lipgloss.AdaptiveColor // config.Theme.Colors.Text.Success
	WarningText        lipgloss.AdaptiveColor // config.Theme.Colors.Text.Warning
	ErrorText          lipgloss.AdaptiveColor // config.Theme.Colors.Text.Error
}

var DefaultTheme = &Theme{
	// Background
	SelectedBackground: lipgloss.AdaptiveColor{Light: "#949cbb", Dark: "#414559"},
	
	// Borders - using surface colors
	PrimaryBorder:   lipgloss.AdaptiveColor{Light: "#626880", Dark: "#626880"},
	SecondaryBorder: lipgloss.AdaptiveColor{Light: "#51576d", Dark: "#51576d"},
	FaintBorder:     lipgloss.AdaptiveColor{Light: "#414559", Dark: "#414559"},
	
	// Text colors
	PrimaryText:   lipgloss.AdaptiveColor{Light: "#c6d0f5", Dark: "#c6d0f5"},
	SecondaryText: lipgloss.AdaptiveColor{Light: "#b5bfe2", Dark: "#b5bfe2"},
	FaintText:     lipgloss.AdaptiveColor{Light: "#a5adce", Dark: "#a5adce"},
	InvertedText:  lipgloss.AdaptiveColor{Light: "#303446", Dark: "#303446"},
	
	// Status colors - using standard terminal colors for urgency
	SuccessText: lipgloss.AdaptiveColor{Light: "#a6d189", Dark: "#a6d189"},  // Green for success
	WarningText: lipgloss.AdaptiveColor{Light: "#ef9f76", Dark: "#ef9f76"},  // Orange/peach for warning
	ErrorText:   lipgloss.AdaptiveColor{Light: "#e78284", Dark: "#e78284"},  // Red for error
}

func ParseTheme(cfg *config.Config) Theme {
	_shimHex := func(hex config.HexColor, fallback lipgloss.AdaptiveColor) lipgloss.AdaptiveColor {
		if hex == "" {
			return fallback
		}
		return lipgloss.AdaptiveColor{Light: string(hex), Dark: string(hex)}
	}

	if cfg.Theme.Colors != nil {
		DefaultTheme.SelectedBackground = _shimHex(
			cfg.Theme.Colors.Inline.Background.Selected,
			DefaultTheme.SelectedBackground,
		)
		DefaultTheme.PrimaryBorder = _shimHex(
			cfg.Theme.Colors.Inline.Border.Primary,
			DefaultTheme.PrimaryBorder,
		)
		DefaultTheme.FaintBorder = _shimHex(
			cfg.Theme.Colors.Inline.Border.Faint,
			DefaultTheme.FaintBorder,
		)
		DefaultTheme.SecondaryBorder = _shimHex(
			cfg.Theme.Colors.Inline.Border.Secondary,
			DefaultTheme.SecondaryBorder,
		)
		DefaultTheme.FaintText = _shimHex(
			cfg.Theme.Colors.Inline.Text.Faint,
			DefaultTheme.FaintText,
		)
		DefaultTheme.PrimaryText = _shimHex(
			cfg.Theme.Colors.Inline.Text.Primary,
			DefaultTheme.PrimaryText,
		)
		DefaultTheme.SecondaryText = _shimHex(
			cfg.Theme.Colors.Inline.Text.Secondary,
			DefaultTheme.SecondaryText,
		)
		DefaultTheme.InvertedText = _shimHex(
			cfg.Theme.Colors.Inline.Text.Inverted,
			DefaultTheme.InvertedText,
		)
		DefaultTheme.SuccessText = _shimHex(
			cfg.Theme.Colors.Inline.Text.Success,
			DefaultTheme.SuccessText,
		)
		DefaultTheme.WarningText = _shimHex(
			cfg.Theme.Colors.Inline.Text.Warning,
			DefaultTheme.WarningText,
		)
		DefaultTheme.ErrorText = _shimHex(
			cfg.Theme.Colors.Inline.Text.Error,
			DefaultTheme.ErrorText,
		)
	}

	log.Debug("Parsing theme", "config", cfg.Theme, "theme", DefaultTheme)

	return *DefaultTheme
}
