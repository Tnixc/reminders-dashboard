package reminderssection

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/dlvhdr/gh-dash/v4/internal/config"
	"github.com/dlvhdr/gh-dash/v4/internal/data"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/reminder"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/section"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/table"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/constants"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/context"
	"github.com/dlvhdr/gh-dash/v4/internal/utils"
)

const SectionType = "reminder"

type Model struct {
	section.BaseModel
	Reminders []data.Reminder
}

func NewModel(
	id int,
	ctx *context.ProgramContext,
	cfg config.RemindersSectionConfig,
	lastUpdated time.Time,
	createdAt time.Time,
) Model {
	m := Model{}
	m.BaseModel = section.NewModel(
		ctx,
		section.NewSectionOptions{
			Id:          id,
			Config:      cfg.ToSectionConfig(),
			Type:        SectionType,
			Columns:     GetSectionColumns(cfg, ctx),
			Singular:    m.GetItemSingularForm(),
			Plural:      m.GetItemPluralForm(),
			LastUpdated: lastUpdated,
			CreatedAt:   createdAt,
		},
	)
	m.Reminders = []data.Reminder{}

	return m
}

func (m *Model) Update(msg tea.Msg) (section.Section, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:

		if m.IsSearchFocused() {
			switch msg.Type {
			case tea.KeyCtrlC, tea.KeyEsc:
				m.SearchBar.SetValue(m.SearchValue)
				blinkCmd := m.SetIsSearching(false)
				return m, blinkCmd

			case tea.KeyEnter:
				m.SearchValue = m.SearchBar.Value()
				m.SetIsSearching(false)
				m.ResetRows()
				return m, tea.Batch(m.FetchNextPageSectionRows()...)
			}

			break
		}

		// No prompt confirmation for reminders

	case SectionRemindersFetchedMsg:
		if m.LastFetchTaskId == msg.TaskId {
			if m.PageInfo != nil {
				m.Reminders = append(m.Reminders, msg.Reminders...)
			} else {
				m.Reminders = msg.Reminders
			}
			m.TotalCount = msg.TotalCount
			m.PageInfo = &msg.PageInfo
			m.SetIsLoading(false)
			m.Table.SetRows(m.BuildRows())
			m.Table.UpdateLastUpdated(time.Now())
			m.UpdateTotalItemsCount(m.TotalCount)
		}
	}

	search, searchCmd := m.SearchBar.Update(msg)
	m.Table.SetRows(m.BuildRows())
	m.SearchBar = search

	prompt, promptCmd := m.PromptConfirmationBox.Update(msg)
	m.PromptConfirmationBox = prompt

	table, tableCmd := m.Table.Update(msg)
	m.Table = table

	return m, tea.Batch(cmd, searchCmd, promptCmd, tableCmd)
}

func GetSectionColumns(
	cfg config.RemindersSectionConfig,
	ctx *context.ProgramContext,
) []table.Column {
	dLayout := ctx.Config.Defaults.Layout.Reminders
	sLayout := cfg.Layout

	titleLayout := config.MergeColumnConfigs(dLayout.Title, sLayout.Title)
	listLayout := config.MergeColumnConfigs(dLayout.List, sLayout.List)
	dueInLayout := config.MergeColumnConfigs(dLayout.DueIn, sLayout.DueIn)
	dateLayout := config.MergeColumnConfigs(dLayout.Date, sLayout.Date)
	priorityLayout := config.MergeColumnConfigs(dLayout.Priority, sLayout.Priority)
	completedLayout := config.MergeColumnConfigs(dLayout.Completed, sLayout.Completed)

	return []table.Column{
		{
			Title:  "Title",
			Grow:   utils.BoolPtr(true),
			Hidden: titleLayout.Hidden,
		},
		{
			Title:  "List",
			Width:  listLayout.Width,
			Hidden: listLayout.Hidden,
		},
		{
			Title:  "Due In",
			Width:  dueInLayout.Width,
			Hidden: dueInLayout.Hidden,
		},
		{
			Title:  "Date",
			Width:  dateLayout.Width,
			Hidden: dateLayout.Hidden,
		},
		{
			Title:  "Priority",
			Width:  priorityLayout.Width,
			Hidden: priorityLayout.Hidden,
		},
		{
			Title:  "Completed",
			Width:  completedLayout.Width,
			Hidden: completedLayout.Hidden,
		},
	}
}

func (m Model) BuildRows() []table.Row {
	var rows []table.Row
	currItem := m.Table.GetCurrItem()
	for i, currReminder := range m.Reminders {
		i := i
		reminderModel := reminder.Reminder{Ctx: m.Ctx, Data: &currReminder, Columns: m.Table.Columns}
		rows = append(
			rows,
			reminderModel.ToTableRow(currItem == i),
		)
	}

	if rows == nil {
		rows = []table.Row{}
	}

	return rows
}

func (m *Model) NumRows() int {
	return len(m.Reminders)
}

type SectionRemindersFetchedMsg struct {
	Reminders  []data.Reminder
	TotalCount int
	PageInfo   data.PageInfo
	TaskId     string
}

func (m *Model) GetCurrRow() data.RowData {
	if len(m.Reminders) == 0 {
		return nil
	}
	reminder := m.Reminders[m.Table.GetCurrItem()]
	return &reminder
}

func (m *Model) FetchNextPageSectionRows() []tea.Cmd {
	if m == nil {
		return nil
	}

	if m.PageInfo != nil && !m.PageInfo.HasNextPage {
		return nil
	}

	var cmds []tea.Cmd

	startCursor := time.Now().String()
	if m.PageInfo != nil {
		startCursor = m.PageInfo.StartCursor
	}
	taskId := fmt.Sprintf("fetching_reminders_%d_%s", m.Id, startCursor)
	isFirstFetch := m.LastFetchTaskId == ""
	m.LastFetchTaskId = taskId
	task := context.Task{
		Id:        taskId,
		StartText: fmt.Sprintf(`Fetching reminders for "%s"`, m.Config.Title),
		FinishedText: fmt.Sprintf(
			`Reminders for "%s" have been fetched`,
			m.Config.Title,
		),
		State: context.TaskStart,
		Error: nil,
	}
	startCmd := m.Ctx.StartTask(task)
	cmds = append(cmds, startCmd)

	fetchCmd := func() tea.Msg {
		limit := m.Config.Limit
		if limit == nil {
			limit = &m.Ctx.Config.Defaults.RemindersLimit
		}

		res, err := data.FetchReminders(m.GetFilters(), *limit, m.PageInfo)
		if err != nil {
			return constants.TaskFinishedMsg{
				SectionId:   m.Id,
				SectionType: m.Type,
				TaskId:      taskId,
				Err:         err,
			}
		}

		return constants.TaskFinishedMsg{
			SectionId:   m.Id,
			SectionType: m.Type,
			TaskId:      taskId,
			Msg: SectionRemindersFetchedMsg{
				Reminders:  res.Reminders,
				TotalCount: res.TotalCount,
				PageInfo:   res.PageInfo,
				TaskId:     taskId,
			},
		}
	}
	cmds = append(cmds, fetchCmd)

	m.IsLoading = true
	if isFirstFetch {
		m.SetIsLoading(true)
		cmds = append(cmds, m.Table.StartLoadingSpinner())
	}

	return cmds
}

func (m *Model) ResetRows() {
	m.Reminders = nil
	m.BaseModel.ResetRows()
}

func FetchAllSections(
	ctx *context.ProgramContext,
	reminders []section.Section,
) (sections []section.Section, fetchAllCmd tea.Cmd) {
	fetchRemindersCmds := make([]tea.Cmd, 0, len(ctx.Config.RemindersSections))
	sections = make([]section.Section, 0, len(ctx.Config.RemindersSections))
	for i, sectionConfig := range ctx.Config.RemindersSections {
		sectionModel := NewModel(
			i+1, // 0 is the search section
			ctx,
			sectionConfig,
			time.Now(),
			time.Now(),
		)
		if len(reminders) > 0 && len(reminders) >= i+1 && reminders[i+1] != nil {
			oldSection := reminders[i+1].(*Model)
			sectionModel.Reminders = oldSection.Reminders
			sectionModel.LastFetchTaskId = oldSection.LastFetchTaskId
		}
		sections = append(sections, &sectionModel)
		fetchRemindersCmds = append(
			fetchRemindersCmds,
			sectionModel.FetchNextPageSectionRows()...)
	}
	return sections, tea.Batch(fetchRemindersCmds...)
}

// No assignees for reminders

func (m Model) GetItemSingularForm() string {
	return "Reminder"
}

func (m Model) GetItemPluralForm() string {
	return "Reminders"
}

func (m Model) GetTotalCount() int {
	return m.TotalCount
}

func (m *Model) SetIsLoading(val bool) {
	m.IsLoading = val
	m.Table.SetIsLoading(val)
}

func (m Model) GetPagerContent() string {
	pagerContent := ""
	timeElapsed := utils.TimeElapsed(m.LastUpdated())
	if timeElapsed == "now" {
		timeElapsed = "just now"
	} else {
		timeElapsed = fmt.Sprintf("~%v ago", timeElapsed)
	}
	if m.TotalCount > 0 {
		pagerContent = fmt.Sprintf(
			"%v Updated %v • %v %v/%v (fetched %v)",
			constants.WaitingIcon,
			timeElapsed,
			m.SingularForm,
			m.Table.GetCurrItem()+1,
			m.TotalCount,
			len(m.Table.Rows),
		)
	}
	pager := m.Ctx.Styles.ListViewPort.PagerStyle.Render(pagerContent)
	return pager
}
