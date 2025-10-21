package context

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/dlvhdr/gh-dash/v4/internal/config"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/theme"
)

type State = int

const (
	TaskStart State = iota
	TaskFinished
	TaskError
)

type Task struct {
	Id           string
	StartText    string
	FinishedText string
	State        State
	Error        error
	StartTime    time.Time
	FinishedTime *time.Time
}

type ProgramContext struct {
	RepoPath          string
	RepoUrl           string
	User              string
	ScreenHeight      int
	ScreenWidth       int
	MainContentWidth  int
	MainContentHeight int
	Config            *config.Config
	ConfigFlag        string
	Version           string
	View              config.ViewType
	Error             error
	StartTask         func(task Task) tea.Cmd
	Theme             theme.Theme
	Styles            Styles
}

func (ctx *ProgramContext) GetViewSectionsConfig() []config.SectionConfig {
	var configs []config.SectionConfig
	switch ctx.View {
	case config.RemindersView:
		for _, cfg := range ctx.Config.RemindersSections {
			configs = append(configs, cfg.ToSectionConfig())
		}
	}

	return append([]config.SectionConfig{{Title: ""}}, configs...)
}
