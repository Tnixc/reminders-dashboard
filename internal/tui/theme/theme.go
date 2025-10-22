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

func getDefaultTheme() Theme {
	return Theme{
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
		SuccessText: lipgloss.AdaptiveColor{Light: "#a6d189", Dark: "#a6d189"}, // Green for success
		WarningText: lipgloss.AdaptiveColor{Light: "#ef9f76", Dark: "#ef9f76"}, // Orange/peach for warning
		ErrorText:   lipgloss.AdaptiveColor{Light: "#e78284", Dark: "#e78284"}, // Red for error
	}
}

func ParseTheme(cfg *config.Config) Theme {
	theme := getDefaultTheme()

	// Helper function to convert hex color to AdaptiveColor
	_shimHex := func(hex config.HexColor, fallback lipgloss.AdaptiveColor) lipgloss.AdaptiveColor {
		if hex == "" {
			return fallback
		}
		return lipgloss.AdaptiveColor{Light: string(hex), Dark: string(hex)}
	}

	// Only parse colors if theme config exists
	if cfg.Theme == nil {
		log.Debug("No theme config found, using defaults")
		return theme
	}

	log.Debug("Theme config exists", "hasColors", cfg.Theme.Colors != nil)

	if cfg.Theme.Colors != nil {
		log.Debug("Parsing theme colors from config",
			"primaryText", cfg.Theme.Colors.Text.Primary,
			"secondaryText", cfg.Theme.Colors.Text.Secondary,
			"faintText", cfg.Theme.Colors.Text.Faint,
			"selectedBg", cfg.Theme.Colors.Background.Selected,
			"errorText", cfg.Theme.Colors.Text.Error,
			"warningText", cfg.Theme.Colors.Text.Warning,
			"successText", cfg.Theme.Colors.Text.Success,
		)

		theme.SelectedBackground = _shimHex(
			cfg.Theme.Colors.Background.Selected,
			theme.SelectedBackground,
		)
		theme.PrimaryBorder = _shimHex(
			cfg.Theme.Colors.Border.Primary,
			theme.PrimaryBorder,
		)
		theme.FaintBorder = _shimHex(
			cfg.Theme.Colors.Border.Faint,
			theme.FaintBorder,
		)
		theme.SecondaryBorder = _shimHex(
			cfg.Theme.Colors.Border.Secondary,
			theme.SecondaryBorder,
		)
		theme.FaintText = _shimHex(
			cfg.Theme.Colors.Text.Faint,
			theme.FaintText,
		)
		theme.PrimaryText = _shimHex(
			cfg.Theme.Colors.Text.Primary,
			theme.PrimaryText,
		)
		theme.SecondaryText = _shimHex(
			cfg.Theme.Colors.Text.Secondary,
			theme.SecondaryText,
		)
		theme.InvertedText = _shimHex(
			cfg.Theme.Colors.Text.Inverted,
			theme.InvertedText,
		)
		theme.SuccessText = _shimHex(
			cfg.Theme.Colors.Text.Success,
			theme.SuccessText,
		)
		theme.WarningText = _shimHex(
			cfg.Theme.Colors.Text.Warning,
			theme.WarningText,
		)
		theme.ErrorText = _shimHex(
			cfg.Theme.Colors.Text.Error,
			theme.ErrorText,
		)

		log.Debug("Parsed theme from config",
			"primaryText", theme.PrimaryText,
			"secondaryText", theme.SecondaryText,
			"faintText", theme.FaintText,
			"selectedBg", theme.SelectedBackground,
			"primaryBorder", theme.PrimaryBorder,
		)
	} else {
		log.Debug("No theme colors config found, using defaults")
	}

	return theme
}