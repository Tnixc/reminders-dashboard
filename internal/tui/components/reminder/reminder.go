package reminder

import (
	"strconv"

	"github.com/charmbracelet/lipgloss"

	"github.com/dlvhdr/gh-dash/v4/internal/data"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/table"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/context"
	"github.com/dlvhdr/gh-dash/v4/internal/utils"
)

type Reminder struct {
	Ctx     *context.ProgramContext
	Data    *data.Reminder
	Columns []table.Column
}

func (r *Reminder) getTextStyle() lipgloss.Style {
	return components.GetIssueTextStyle(r.Ctx)
}

func (r *Reminder) renderTitle() string {
	if r.Data == nil {
		return ""
	}
	return r.getTextStyle().Render(r.Data.Title)
}

func (r *Reminder) renderList() string {
	if r.Data == nil {
		return ""
	}
	color := r.Ctx.Config.ListColors[r.Data.List]
	if color != "" {
		return lipgloss.NewStyle().Foreground(lipgloss.Color(color)).Render(r.Data.List)
	}
	return r.getTextStyle().Render(r.Data.List)
}

func (r *Reminder) renderDueIn() string {
	if r.Data == nil {
		return ""
	}
	dueInOutput := utils.TimeUntil(r.Data.DueDate)
	return r.getTextStyle().Foreground(r.Ctx.Theme.FaintText).Render(dueInOutput)
}

func (r *Reminder) renderDate() string {
	if r.Data == nil {
		return ""
	}
	return r.getTextStyle().Render(r.Data.DueDate.Format("15:04, 02 Jan 2006"))
}

func (r *Reminder) renderPriority() string {
	if r.Data == nil {
		return ""
	}
	return r.getTextStyle().Render(strconv.Itoa(r.Data.Priority))
}

func (r *Reminder) renderCompleted() string {
	if r.Data == nil {
		return ""
	}
	if r.Data.IsCompleted {
		return r.getTextStyle().Foreground(r.Ctx.Theme.SuccessText).Render("✓")
	}
	return r.getTextStyle().Foreground(r.Ctx.Theme.FaintText).Render("○")
}

func (r *Reminder) ToTableRow(isSelected bool) table.Row {
	return table.Row{
		r.renderTitle(),
		r.renderList(),
		r.renderDueIn(),
		r.renderDate(),
		r.renderPriority(),
		r.renderCompleted(),
	}
}