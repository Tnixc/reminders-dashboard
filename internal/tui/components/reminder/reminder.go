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
		style = style.Bold(true).Foreground(r.Ctx.Theme.PrimaryText).Background(r.Ctx.Theme.SelectedBackground)
	}
	title := style.Render(r.Data.Title)
	list := r.renderList(isSelected)
	return lipgloss.JoinVertical(lipgloss.Left, title, list)
}

func (r *Reminder) renderList(isSelected bool) string {
	if r.Data == nil {
		return ""
	}
	
	// Create badge style with colored background and inverted text
	badgeStyle := lipgloss.NewStyle().
		Padding(0, 1).
		Bold(true)
	
	color := r.Ctx.Config.ListColors[r.Data.List]
	if color != "" {
		badgeStyle = badgeStyle.
			Background(lipgloss.Color(color)).
			Foreground(r.Ctx.Theme.InvertedText)
	} else {
		badgeStyle = badgeStyle.
			Background(r.Ctx.Theme.FaintBorder).
			Foreground(r.Ctx.Theme.PrimaryText)
	}
	
	badge := badgeStyle.Render(r.Data.List)
	
	// If selected, wrap in a container with selected background to fill the cell
	if isSelected {
		containerStyle := lipgloss.NewStyle().Background(r.Ctx.Theme.SelectedBackground)
		return containerStyle.Render(badge)
	}
	
	return badge
}

func (r *Reminder) renderDueIn(isSelected bool) string {
	if r.Data == nil {
		return ""
	}
	dueInOutput := utils.TimeUntil(r.Data.DueDate)
	urgencyColor := r.getUrgencyColor(r.Data.DueDate)
	
	dueInStyle := r.getTextStyle().Foreground(urgencyColor).Bold(true)
	if isSelected {
		dueInStyle = dueInStyle.Background(r.Ctx.Theme.SelectedBackground)
	}
	dueIn := dueInStyle.Render(dueInOutput)
	
	date := r.renderDate(isSelected)
	return lipgloss.JoinVertical(lipgloss.Left, dueIn, date)
}

func (r *Reminder) renderDate(isSelected bool) string {
	if r.Data == nil {
		return ""
	}
	dateStyle := r.getTextStyle()
	if isSelected {
		dateStyle = dateStyle.Background(r.Ctx.Theme.SelectedBackground)
	}
	return dateStyle.Render(r.Data.DueDate.Format("15:04, 02 Jan 2006"))
}

func (r *Reminder) renderPriority(isSelected bool) string {
	if r.Data == nil {
		return ""
	}
	priorityStyle := r.getTextStyle()
	if isSelected {
		priorityStyle = priorityStyle.Background(r.Ctx.Theme.SelectedBackground)
	}
	return priorityStyle.Render(strconv.Itoa(r.Data.Priority))
}

func (r *Reminder) renderCompleted(isSelected bool) string {
	if r.Data == nil {
		return ""
	}
	completedStyle := r.getTextStyle()
	if isSelected {
		completedStyle = completedStyle.Background(r.Ctx.Theme.SelectedBackground)
	}
	if r.Data.IsCompleted {
		return completedStyle.Foreground(r.Ctx.Theme.SuccessText).Render("✓")
	}
	return completedStyle.Foreground(r.Ctx.Theme.FaintText).Render("○")
}

func (r *Reminder) ToTableRow(isSelected bool) table.Row {
	return table.Row{
		r.renderTitle(isSelected),
		r.renderDueIn(isSelected),
		r.renderPriority(isSelected),
		r.renderCompleted(isSelected),
	}
}