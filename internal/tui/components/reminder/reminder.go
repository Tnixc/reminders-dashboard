package reminder

import (
	"strconv"
	"time"

	"github.com/charmbracelet/lipgloss"

	"github.com/dlvhdr/reminders-dashboard/v4/internal/data"
	"github.com/dlvhdr/reminders-dashboard/v4/internal/tui/components"
	"github.com/dlvhdr/reminders-dashboard/v4/internal/tui/components/table"
	"github.com/dlvhdr/reminders-dashboard/v4/internal/tui/context"
	"github.com/dlvhdr/reminders-dashboard/v4/internal/utils"
)

type Reminder struct {
	Ctx     *context.ProgramContext
	Data    *data.Reminder
	Columns []table.Column
}

func (r *Reminder) getTextStyle() lipgloss.Style {
	return components.GetIssueTextStyle(r.Ctx)
}

func (r *Reminder) getUrgencyColor(dueDate time.Time) lipgloss.AdaptiveColor {
	now := time.Now()
	if dueDate.Before(now) {
		// Overdue - red
		return r.Ctx.Theme.ErrorText
	}

	diff := dueDate.Sub(now)
	hours := diff.Hours()

	if hours <= 24 {
		// Due within 24 hours - red
		return r.Ctx.Theme.ErrorText
	} else if hours <= 48 {
		// Due within 48 hours - orange
		return r.Ctx.Theme.WarningText
	} else if hours <= 168 {
		// Due within 1 week - yellow (using warning text as yellow)
		return lipgloss.AdaptiveColor{Light: "003", Dark: "011"}
	}

	// Due later - grey
	return r.Ctx.Theme.FaintText
}

func (r *Reminder) renderTitle(isSelected bool) string {
	if r.Data == nil {
		return ""
	}
	style := r.getTextStyle()
	if isSelected {
		style = style.Bold(true).Foreground(r.Ctx.Theme.PrimaryText)
	}
	return style.Render(r.Data.Title)
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
	urgencyColor := r.getUrgencyColor(r.Data.DueDate)
	return r.getTextStyle().Foreground(urgencyColor).Bold(true).Render(dueInOutput)
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
		r.renderTitle(isSelected),
		r.renderList(),
		r.renderDueIn(),
		r.renderDate(),
		r.renderPriority(),
		r.renderCompleted(),
	}
}