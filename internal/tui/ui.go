package tui

import (
	"fmt"
	"os"
	"reflect"
	"runtime/debug"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	log "github.com/charmbracelet/log"
	"github.com/cli/go-gh/v2/pkg/browser"
	zone "github.com/lrstanley/bubblezone"

	"github.com/dlvhdr/reminders-dashboard/v4/internal/config"
	"github.com/dlvhdr/reminders-dashboard/v4/internal/data"

	"github.com/dlvhdr/reminders-dashboard/v4/internal/tui/common"
	"github.com/dlvhdr/reminders-dashboard/v4/internal/tui/components/footer"
	"github.com/dlvhdr/reminders-dashboard/v4/internal/tui/components/reminderssection"
	"github.com/dlvhdr/reminders-dashboard/v4/internal/tui/components/section"
	"github.com/dlvhdr/reminders-dashboard/v4/internal/tui/components/sidebar"
	"github.com/dlvhdr/reminders-dashboard/v4/internal/tui/components/tabs"
	"github.com/dlvhdr/reminders-dashboard/v4/internal/tui/constants"
	"github.com/dlvhdr/reminders-dashboard/v4/internal/tui/context"
	"github.com/dlvhdr/reminders-dashboard/v4/internal/tui/keys"
	"github.com/dlvhdr/reminders-dashboard/v4/internal/tui/theme"
)

type Model struct {
	keys          *keys.KeyMap
	sidebar       sidebar.Model
	currSectionId int
	footer        footer.Model
	reminders     []section.Section
	tabs          tabs.Model
	ctx           *context.ProgramContext
	taskSpinner   spinner.Model
	tasks         map[string]context.Task
}

func NewModel(location config.Location) Model {
	taskSpinner := spinner.Model{Spinner: spinner.Dot}
	m := Model{
		keys:        keys.Keys,
		sidebar:     sidebar.NewModel(),
		taskSpinner: taskSpinner,
		tasks:       map[string]context.Task{},
	}

	version := "dev"
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Sum != "" {
		version = info.Main.Version
	}

	m.ctx = &context.ProgramContext{
		RepoPath:   location.RepoPath,
		ConfigFlag: location.ConfigFlag,
		Version:    version,
		StartTask: func(task context.Task) tea.Cmd {
			log.Debug("Starting task", "id", task.Id)
			task.StartTime = time.Now()
			m.tasks[task.Id] = task
			rTask := m.renderRunningTask()
			m.footer.SetRightSection(rTask)
			return m.taskSpinner.Tick
		},
	}

	m.taskSpinner.Style = lipgloss.NewStyle().
		Background(m.ctx.Theme.SelectedBackground)

	m.footer = footer.NewModel(m.ctx)
	m.tabs = tabs.NewModel(m.ctx)

	return m
}

func (m *Model) initScreen() tea.Msg {
	showError := func(err error) {
		styles := log.DefaultStyles()
		styles.Key = lipgloss.NewStyle().
			Foreground(lipgloss.Color("1")).
			Bold(true)
		styles.Separator = lipgloss.NewStyle()

		logger := log.New(os.Stderr)
		logger.SetStyles(styles)
		logger.SetTimeFormat(time.RFC3339)
		logger.SetReportTimestamp(true)
		logger.SetPrefix("Reading config file")
		logger.SetReportCaller(true)

		logger.
			Fatal(
				"failed parsing config file",
				"location",
				m.ctx.ConfigFlag,
				"err",
				err,
			)
	}

	cfg, err := config.ParseConfig(config.Location{RepoPath: m.ctx.RepoPath, ConfigFlag: m.ctx.ConfigFlag})
	if err != nil {
		showError(err)
		return initMsg{Config: cfg}
	}

	// Debug logging to verify config was loaded correctly
	log.Debug("Config loaded successfully")
	log.Debug("Config file path", "path", m.ctx.ConfigFlag)
	log.Debug("Confirm quit", "value", cfg.ConfirmQuit)
	log.Debug("Smart filtering at launch", "value", cfg.SmartFilteringAtLaunch)
	log.Debug("Reminders limit", "value", cfg.Defaults.RemindersLimit)
	log.Debug("Preview open", "value", cfg.Defaults.Preview.Open)
	log.Debug("Preview width", "value", cfg.Defaults.Preview.Width)
	log.Debug("Refetch interval minutes", "value", cfg.Defaults.RefetchIntervalMinutes)
	log.Debug("Date format", "value", cfg.Defaults.DateFormat)
	log.Debug("Number of reminder sections", "count", len(cfg.RemindersSections))
	for i, section := range cfg.RemindersSections {
		log.Debug("Reminder section", "index", i, "title", section.Title, "filters", section.Filters)
	}
	if cfg.Theme != nil {
		log.Debug("Theme UI sections show count", "value", cfg.Theme.Ui.SectionsShowCount)
		log.Debug("Theme UI table show separator", "value", cfg.Theme.Ui.Table.ShowSeparator)
		log.Debug("Theme UI table compact", "value", cfg.Theme.Ui.Table.Compact)
	}
	log.Debug("Number of list colors", "count", len(cfg.ListColors))
	for list, color := range cfg.ListColors {
		log.Debug("List color", "list", list, "color", color)
	}

	err = keys.Rebind(
		cfg.Keybindings.Universal,
		cfg.Keybindings.Reminders,
	)
	if err != nil {
		showError(err)
	}

	return initMsg{Config: cfg}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(m.initScreen, tea.EnterAltScreen)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd         tea.Cmd
		tabsCmd     tea.Cmd
		sidebarCmd  tea.Cmd
		footerCmd   tea.Cmd
		cmds        []tea.Cmd
		currSection = m.getCurrSection()
		currRowData = m.getCurrRowData()
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		log.Debug("Key pressed", "key", msg.String())
		m.ctx.Error = nil

		if currSection != nil && (currSection.IsSearchFocused() ||
			currSection.IsPromptConfirmationFocused()) {
			cmd = m.updateSection(currSection.GetId(), currSection.GetType(), msg)
			return m, cmd
		}

		switch {
		case m.isUserDefinedKeybinding(msg):
			cmd = m.executeKeybinding(msg.String())
			return m, cmd

		case key.Matches(msg, m.keys.PrevSection):
			prevSection := m.getSectionAt(m.getPrevSectionId())
			if prevSection != nil {
				m.setCurrSectionId(prevSection.GetId())
				m.onViewedRowChanged()
			}

		case key.Matches(msg, m.keys.NextSection):
			nextSectionId := m.getNextSectionId()
			nextSection := m.getSectionAt(nextSectionId)
			if nextSection != nil {
				m.setCurrSectionId(nextSection.GetId())
				m.onViewedRowChanged()
			}

		case key.Matches(msg, m.keys.Down):
			prevRow := currSection.CurrRow()
			nextRow := currSection.NextRow()
			if prevRow != nextRow && nextRow == currSection.NumRows()-1 {
				cmds = append(cmds, currSection.FetchNextPageSectionRows()...)
			}
			cmd = m.onViewedRowChanged()

		case key.Matches(msg, m.keys.Up):
			currSection.PrevRow()
			m.onViewedRowChanged()

		case key.Matches(msg, m.keys.FirstLine):
			currSection.FirstItem()
			cmd = m.onViewedRowChanged()

		case key.Matches(msg, m.keys.LastLine):
			if currSection.CurrRow()+1 < currSection.NumRows() {
				cmds = append(cmds, currSection.FetchNextPageSectionRows()...)
			}
			currSection.LastItem()
			cmd = m.onViewedRowChanged()

		case key.Matches(msg, m.keys.TogglePreview):
			m.sidebar.IsOpen = !m.sidebar.IsOpen
			m.syncMainContentWidth()

		case key.Matches(msg, m.keys.Refresh):
			currSection.ResetFilters()
			currSection.ResetRows()
			m.syncSidebar()
			currSection.SetIsLoading(true)
			cmds = append(cmds, currSection.FetchNextPageSectionRows()...)

		case key.Matches(msg, m.keys.RefreshAll):
			newSections, fetchSectionsCmds := m.fetchAllViewSections()
			m.setCurrentViewSections(newSections)
			cmds = append(cmds, fetchSectionsCmds)

		case key.Matches(msg, m.keys.Redraw):
			// can't find a way to just ask to send bubbletea's internal repaintMsg{},
			// so this seems like the lightest-weight alternative
			return m, tea.Batch(tea.ExitAltScreen, tea.EnterAltScreen)

		case key.Matches(msg, m.keys.Search):
			if currSection != nil {
				cmd = currSection.SetIsSearching(true)
				return m, cmd
			}

		case key.Matches(msg, m.keys.Help):
			if !m.footer.ShowAll {
				m.ctx.MainContentHeight = m.ctx.MainContentHeight +
					common.FooterHeight - common.ExpandedHelpHeight
			} else {
				m.ctx.MainContentHeight = m.ctx.MainContentHeight +
					common.ExpandedHelpHeight - common.FooterHeight
			}

		case key.Matches(msg, m.keys.CopyNumber):
			var cmd tea.Cmd
			if currRowData == nil || reflect.ValueOf(currRowData).IsNil() {
				cmd = m.notifyErr("Current selection isn't associated with a PR/Issue")
				return m, cmd
			}
			number := fmt.Sprint(currRowData.GetNumber())
			err := clipboard.WriteAll(number)
			if err != nil {
				cmd = m.notifyErr(fmt.Sprintf("Failed copying to clipboard %v", err))
			} else {
				cmd = m.notify(fmt.Sprintf("Copied %s to clipboard", number))
			}
			return m, cmd

		case key.Matches(msg, m.keys.CopyUrl):
			var cmd tea.Cmd
			if currRowData == nil || reflect.ValueOf(currRowData).IsNil() {
				cmd = m.notifyErr("Current selection isn't associated with a PR/Issue")
				return m, cmd
			}
			url := currRowData.GetUrl()
			err := clipboard.WriteAll(url)
			if err != nil {
				cmd = m.notifyErr(fmt.Sprintf("Failed copying to clipboard %v", err))
			} else {
				cmd = m.notify(fmt.Sprintf("Copied %s to clipboard", url))
			}
			return m, cmd

		case key.Matches(msg, m.keys.Quit):
			if m.ctx.Config.ConfirmQuit {
				m.footer, cmd = m.footer.Update(msg)
				return m, cmd
			}
			cmd = tea.Quit

		case m.ctx.View == config.RemindersView:
			// No specific key handlers for reminders yet
		}

	case initMsg:
		m.ctx.Config = &msg.Config
		m.ctx.Theme = theme.ParseTheme(m.ctx.Config)
		m.ctx.Styles = context.InitStyles(m.ctx.Theme)
		m.ctx.View = m.ctx.Config.Defaults.View
		m.currSectionId = m.getCurrentViewDefaultSection()
		m.sidebar.IsOpen = msg.Config.Defaults.Preview.Open
		m.syncMainContentWidth()

		newSections, fetchSectionsCmds := m.fetchAllViewSections()
		m.setCurrentViewSections(newSections)
		m.tabs.SetCurrSectionId(1)
		cmds = append(cmds, fetchSectionsCmds, m.tabs.Init(),
			m.doRefreshAtInterval(), m.doUpdateFooterAtInterval())

	case intervalRefresh:
		newSections, fetchSectionsCmds := m.fetchAllViewSections()
		m.setCurrentViewSections(newSections)
		cmds = append(cmds, fetchSectionsCmds, m.doRefreshAtInterval())

	case userFetchedMsg:
		m.ctx.User = msg.user

	case constants.TaskFinishedMsg:
		task, ok := m.tasks[msg.TaskId]
		if ok {
			log.Debug("Task finished", "id", task.Id)
			if msg.Err != nil {
				log.Error("Task finished with error", "id", task.Id, "err", msg.Err)
				task.State = context.TaskError
				task.Error = msg.Err
			} else {
				task.State = context.TaskFinished
			}
			now := time.Now()
			task.FinishedTime = &now
			m.tasks[msg.TaskId] = task
			clear := tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
				return constants.ClearTaskMsg{TaskId: msg.TaskId}
			})
			cmds = append(cmds, clear)

			scmd := m.updateSection(msg.SectionId, msg.SectionType, msg.Msg)
			cmds = append(cmds, scmd)

			syncCmd := m.syncSidebar()
			cmds = append(cmds, syncCmd)
		}

	case spinner.TickMsg:
		if len(m.tasks) > 0 {
			taskSpinner, internalTickCmd := m.taskSpinner.Update(msg)
			m.taskSpinner = taskSpinner
			rTask := m.renderRunningTask()
			m.footer.SetRightSection(rTask)
			cmd = internalTickCmd
		}

	case constants.ClearTaskMsg:
		m.footer.SetRightSection("")
		delete(m.tasks, msg.TaskId)

	case section.SectionMsg:
		cmd = m.updateRelevantSection(msg)

		if msg.Id == m.currSectionId {
			m.onViewedRowChanged()
		}

	case execProcessFinishedMsg, tea.FocusMsg:
		if currSection != nil {
			cmds = append(cmds, currSection.FetchNextPageSectionRows()...)
		}

	case tea.MouseMsg:
		if msg.Action != tea.MouseActionRelease || msg.Button != tea.MouseButtonLeft {
			return m, nil
		}
		if zone.Get("donate").InBounds(msg) {
			log.Debug("Donate clicked", "msg", msg)
			openCmd := func() tea.Msg {
				b := browser.New("", os.Stdout, os.Stdin)
				err := b.Browse("https://github.com/sponsors/dlvhdr")
				if err != nil {
					return constants.ErrMsg{Err: err}
				}
				return nil
			}
			cmds = append(cmds, openCmd)
		}

	case tea.WindowSizeMsg:
		m.onWindowSizeChanged(msg)

	case updateFooterMsg:
		m.footer, cmd = m.footer.Update(msg)
		cmds = append(cmds, cmd, m.doUpdateFooterAtInterval())

	case constants.ErrMsg:
		m.ctx.Error = msg.Err
	}

	m.syncProgramContext()

	m.sidebar, sidebarCmd = m.sidebar.Update(msg)

	m.footer, footerCmd = m.footer.Update(msg)
	if currSection != nil {
		if currSection.IsPromptConfirmationFocused() {
			m.footer.SetLeftSection(currSection.GetPromptConfirmation())
		}

		if !currSection.IsPromptConfirmationFocused() {
			m.footer.SetLeftSection(currSection.GetPagerContent())
		}
	}

	tm, tabsCmd := m.tabs.Update(msg)
	m.tabs = tm.(tabs.Model)

	sectionCmd := m.updateCurrentSection(msg)
	cmds = append(
		cmds,
		cmd,
		tabsCmd,
		sidebarCmd,
		footerCmd,
		sectionCmd,
	)

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	if m.ctx.Config == nil {
		return lipgloss.Place(m.ctx.ScreenWidth, m.ctx.ScreenHeight, lipgloss.Center, lipgloss.Center, "Reading config...")
	}

	s := strings.Builder{}
	if m.ctx.View == config.RemindersView {
		s.WriteString(m.tabs.View())
	}
	s.WriteString("\n")
	content := "No sections defined"
	currSection := m.getCurrSection()
	if currSection != nil {
		content = lipgloss.JoinHorizontal(
			lipgloss.Top,
			m.getCurrSection().View(),
			m.sidebar.View(),
		)
	}
	s.WriteString(content)
	s.WriteString("\n")
	if m.ctx.Error != nil {
		s.WriteString(
			m.ctx.Styles.Common.ErrorStyle.
				Width(m.ctx.ScreenWidth).
				Render(fmt.Sprintf("%s %s",
					m.ctx.Styles.Common.FailureGlyph,
					lipgloss.NewStyle().
						Foreground(m.ctx.Theme.ErrorText).
						Render(m.ctx.Error.Error()),
				)),
		)
	} else {
		s.WriteString(m.footer.View())
	}

	return zone.Scan(s.String())
}

type initMsg struct {
	Config config.Config
}

func (m *Model) setCurrSectionId(newSectionId int) {
	m.currSectionId = newSectionId
	m.tabs.SetCurrSectionId(newSectionId)
}

func (m *Model) onViewedRowChanged() tea.Cmd {
	cmd := m.syncSidebar()
	m.sidebar.ScrollToTop()
	return cmd
}

func (m *Model) onWindowSizeChanged(msg tea.WindowSizeMsg) {
	log.Debug("window size changed", "width", msg.Width, "height", msg.Height)
	m.footer.SetWidth(msg.Width)
	m.ctx.ScreenWidth = msg.Width
	m.ctx.ScreenHeight = msg.Height
	if m.footer.ShowAll {
		m.ctx.MainContentHeight = msg.Height - common.TabsHeight - common.ExpandedHelpHeight
	} else {
		m.ctx.MainContentHeight = msg.Height - common.TabsHeight - common.FooterHeight
	}
	m.syncMainContentWidth()
}

func (m *Model) syncProgramContext() {
	for _, section := range m.getCurrentViewSections() {
		section.UpdateProgramContext(m.ctx)
	}
	m.tabs.UpdateProgramContext(m.ctx)
	m.footer.UpdateProgramContext(m.ctx)
	m.sidebar.UpdateProgramContext(m.ctx)
}

func (m *Model) updateSection(id int, sType string, msg tea.Msg) (cmd tea.Cmd) {
	var updatedSection section.Section
	switch sType {
	case reminderssection.SectionType:
		updatedSection, cmd = m.reminders[id].Update(msg)
		m.reminders[id] = updatedSection
	}

	return cmd
}

func (m *Model) updateRelevantSection(msg section.SectionMsg) (cmd tea.Cmd) {
	return m.updateSection(msg.Id, msg.Type, msg)
}

func (m *Model) updateCurrentSection(msg tea.Msg) (cmd tea.Cmd) {
	section := m.getCurrSection()
	if section == nil {
		return nil
	}
	return m.updateSection(section.GetId(), section.GetType(), msg)
}

func (m *Model) syncMainContentWidth() {
	sideBarOffset := 0
	if m.sidebar.IsOpen {
		sideBarOffset = m.ctx.Config.Defaults.Preview.Width
	}
	m.ctx.MainContentWidth = m.ctx.ScreenWidth - sideBarOffset
}

func (m *Model) syncSidebar() tea.Cmd {
	currRowData := m.getCurrRowData()

	if currRowData == nil {
		m.sidebar.SetContent("")
		return nil
	}

	switch row := currRowData.(type) {
	case *data.Reminder:
		// For now, just show basic info
		content := fmt.Sprintf("Title: %s\nList: %s\nDue: %s\nPriority: %s\nDone: %t",
			row.Title, row.List, row.DueDate.Format("2006-01-02"), strconv.Itoa(row.Priority), row.IsCompleted)
		m.sidebar.SetContent(content)
	}

	return nil
}

func (m *Model) fetchAllViewSections() ([]section.Section, tea.Cmd) {
	cmds := make([]tea.Cmd, 0)
	cmds = append(cmds, m.tabs.SetAllLoading()...)

	switch m.ctx.View {
	case config.RemindersView:
		s, reminderCmds := reminderssection.FetchAllSections(m.ctx, m.reminders)
		cmds = append(cmds, reminderCmds)
		return s, tea.Batch(cmds...)
	default:
		return nil, nil
	}
}

func (m *Model) getCurrentViewSections() []section.Section {
	switch m.ctx.View {
	case config.RemindersView:
		return m.reminders
	default:
		return nil
	}
}

func (m *Model) getCurrentViewDefaultSection() int {
	switch m.ctx.View {
	case config.RemindersView:
		return 1
	default:
		return 1
	}
}

func (m *Model) setCurrentViewSections(newSections []section.Section) {
	if newSections == nil {
		return
	}

	missingSearchSection := len(newSections) == 0 || (len(newSections) > 0 && newSections[0].GetId() != 0)
	s := make([]section.Section, 0)
	if m.ctx.View == config.RemindersView {
		if missingSearchSection {
			search := reminderssection.NewModel(
				0,
				m.ctx,
				config.RemindersSectionConfig{
					Title:   "",
					Filters: "",
				},
				time.Now(),
				time.Now(),
			)
			s = append(s, &search)
		}
		m.reminders = append(s, newSections...)
		newSections = m.reminders
	}

	m.tabs.SetSections(newSections)
}

func (m *Model) switchSelectedView() config.ViewType {
	// Only reminders view for now
	return config.RemindersView
}

func (m *Model) isUserDefinedKeybinding(msg tea.KeyMsg) bool {
	for _, keybinding := range m.ctx.Config.Keybindings.Universal {
		if keybinding.Builtin == "" && keybinding.Key == msg.String() {
			return true
		}
	}

	if m.ctx.View == config.RemindersView {
		for _, keybinding := range m.ctx.Config.Keybindings.Reminders {
			if keybinding.Builtin == "" && keybinding.Key == msg.String() {
				return true
			}
		}
	}

	return false
}

func (m *Model) renderRunningTask() string {
	tasks := make([]context.Task, 0, len(m.tasks))
	for _, value := range m.tasks {
		tasks = append(tasks, value)
	}
	sort.Slice(tasks, func(i, j int) bool {
		if tasks[i].FinishedTime != nil && tasks[j].FinishedTime == nil {
			return false
		}
		if tasks[j].FinishedTime != nil && tasks[i].FinishedTime == nil {
			return true
		}
		if tasks[j].FinishedTime != nil && tasks[i].FinishedTime != nil {
			return tasks[i].FinishedTime.After(*tasks[j].FinishedTime)
		}

		return tasks[i].StartTime.After(tasks[j].StartTime)
	})
	task := tasks[0]

	var currTaskStatus string
	switch task.State {
	case context.TaskStart:
		currTaskStatus = lipgloss.NewStyle().
			Background(m.ctx.Theme.SelectedBackground).
			Render(
				fmt.Sprintf(
					"%s%s",
					m.taskSpinner.View(),
					task.StartText,
				))
	case context.TaskError:
		currTaskStatus = lipgloss.NewStyle().
			Foreground(m.ctx.Theme.ErrorText).
			Background(m.ctx.Theme.SelectedBackground).
			Render(fmt.Sprintf("%s %s", constants.FailureIcon, task.Error.Error()))
	case context.TaskFinished:
		currTaskStatus = lipgloss.NewStyle().
			Foreground(m.ctx.Theme.SuccessText).
			Background(m.ctx.Theme.SelectedBackground).
			Render(fmt.Sprintf("%s %s", constants.SuccessIcon, task.FinishedText))
	}

	var numProcessing int
	for _, task := range m.tasks {
		if task.State == context.TaskStart {
			numProcessing += 1
		}
	}

	stats := ""
	if numProcessing > 1 {
		stats = lipgloss.NewStyle().
			Foreground(m.ctx.Theme.FaintText).
			Background(m.ctx.Theme.SelectedBackground).
			Render(fmt.Sprintf("[ %d] ", numProcessing))
	}

	return lipgloss.NewStyle().
		Padding(0, 1).
		Height(1).
		Background(m.ctx.Theme.SelectedBackground).
		Render(strings.TrimSpace(lipgloss.JoinHorizontal(lipgloss.Top, stats, currTaskStatus)))
}

type userFetchedMsg struct {
	user string
}

type intervalRefresh time.Time

func (m *Model) doRefreshAtInterval() tea.Cmd {
	if m.ctx.Config.Defaults.RefetchIntervalMinutes == 0 {
		return nil
	}

	return tea.Tick(
		time.Minute*time.Duration(m.ctx.Config.Defaults.RefetchIntervalMinutes),
		func(t time.Time) tea.Msg {
			return intervalRefresh(t)
		},
	)
}

type updateFooterMsg struct{}

func (m *Model) doUpdateFooterAtInterval() tea.Cmd {
	return tea.Tick(
		time.Second*10,
		func(t time.Time) tea.Msg {
			return updateFooterMsg{}
		},
	)
}
