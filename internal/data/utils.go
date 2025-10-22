package data

import (
	"time"

	"github.com/charmbracelet/lipgloss"

	"github.com/dlvhdr/reminders-dashboard/v4/internal/tui/constants"
	"github.com/dlvhdr/reminders-dashboard/v4/internal/tui/theme"
)

type RowData interface {
	GetRepoNameWithOwner() string
	GetTitle() string
	GetNumber() int
	GetUrl() string
	GetUpdatedAt() time.Time
}

func IsStatusWaiting(status string) bool {
	return status == "PENDING" ||
		status == "QUEUED" ||
		status == "IN_PROGRESS" ||
		status == "WAITING"
}

func IsConclusionASkip(conclusion string) bool {
	return conclusion == "SKIPPED"
}

func IsConclusionAFailure(conclusion string) bool {
	return conclusion == "FAILURE" || conclusion == "TIMED_OUT" || conclusion == "STARTUP_FAILURE"
}

func IsConclusionASuccess(conclusion string) bool {
	return conclusion == "SUCCESS"
}

func GetAuthorRoleIcon(role string, theme theme.Theme) string {
	// https://docs.github.com/en/graphql/reference/enums#commentauthorassociation
	switch role {
	case "FIRST_TIMER", "FIRST_TIME_CONTRIBUTOR", "NONE":
		return lipgloss.NewStyle().Foreground(theme.SuccessText).Render(constants.NewContributorIcon)
	case "COLLABORATOR":
		return lipgloss.NewStyle().Foreground(theme.WarningText).Render(constants.CollaboratorIcon)
	case "CONTRIBUTOR":
		return lipgloss.NewStyle().Foreground(theme.SecondaryText).Render(constants.ContributorIcon)
	case "MEMBER":
		return lipgloss.NewStyle().Foreground(theme.PrimaryText).Render(constants.MemberIcon)
	case "OWNER":
		return lipgloss.NewStyle().Foreground(theme.PrimaryText).Render(constants.OwnerIcon)
	default:
		return lipgloss.NewStyle().Foreground(theme.FaintText).Render(constants.UnknownRoleIcon)
	}
}
