package config

import (
	"strings"
)



func (cfg RemindersSectionConfig) ToSectionConfig() SectionConfig {
	return SectionConfig{
		Title:   cfg.Title,
		Filters: cfg.Filters,
		Limit:   cfg.Limit,
		Type:    cfg.Type,
	}
}

func MergeColumnConfigs(defaultCfg, sectionCfg ColumnConfig) ColumnConfig {
	colCfg := defaultCfg
	if sectionCfg.Width != nil {
		colCfg.Width = sectionCfg.Width
	}
	if sectionCfg.Hidden != nil {
		colCfg.Hidden = sectionCfg.Hidden
	}
	return colCfg
}

func TruncateCommand(cmd string) string {
	cmd = strings.ReplaceAll(cmd, "\n", "")
	if len(cmd) > 30 {
		return cmd[:30] + "..."
	}
	return cmd
}
