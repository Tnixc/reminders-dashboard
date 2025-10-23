package main

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	tint "github.com/lrstanley/bubbletint"
)

var (
	// Theme for the app
	theme = tint.TintSerendipityMidnight

	appStyle = lipgloss.NewStyle().Padding(1, 2)

	titleStyle = lipgloss.NewStyle().
			Foreground(theme.Bg()).
			Background(theme.Blue()).
			Padding(0, 1)

	statusMessageStyle = lipgloss.NewStyle().
				Foreground(theme.BrightGreen()).
				Render
)

type item struct {
	title       string
	description string
	listName    string
	color       string
	urgencyText string
	urgencyColor string
}

func (i item) Title() string {
	// Return plain text with bullet, no colors (colors applied in delegate)
	if i.color != "" {
		return "● " + i.title
	}
	return i.title
}

func (i item) Description() string {
	// Return plain text (colors applied in delegate)
	if i.urgencyText != "" {
		return i.urgencyText + " • " + i.description
	}
	return i.description
}

func (i item) FilterValue() string {
	return i.title
}

// Helper to convert urgency color names to theme colors
func urgencyColorToTheme(colorName string) lipgloss.TerminalColor {
	switch colorName {
	case "red":
		return theme.Red()
	case "orange":
		return theme.Yellow()
	case "yellow":
		return theme.White()
	default:
		// Use subtle/dimmed color for non-urgent items
		return theme.BrightBlack()
	}
}

type listKeyMap struct {
	toggleSpinner    key.Binding
	toggleTitleBar   key.Binding
	toggleStatusBar  key.Binding
	togglePagination key.Binding
	toggleHelpMenu   key.Binding
	insertItem       key.Binding
}

func newListKeyMap() *listKeyMap {
	return &listKeyMap{
		insertItem: key.NewBinding(
			key.WithKeys("a"),
			key.WithHelp("a", "add item"),
		),
		toggleSpinner: key.NewBinding(
			key.WithKeys("s"),
			key.WithHelp("s", "toggle spinner"),
		),
		toggleTitleBar: key.NewBinding(
			key.WithKeys("T"),
			key.WithHelp("T", "toggle title"),
		),
		toggleStatusBar: key.NewBinding(
			key.WithKeys("S"),
			key.WithHelp("S", "toggle status"),
		),
		togglePagination: key.NewBinding(
			key.WithKeys("P"),
			key.WithHelp("P", "toggle pagination"),
		),
		toggleHelpMenu: key.NewBinding(
			key.WithKeys("H"),
			key.WithHelp("H", "toggle help"),
		),
	}
}

type listModel struct {
	list          list.Model
	itemGenerator *randomItemGenerator
	keys          *listKeyMap
	delegateKeys  *delegateKeyMap
	commonHelp    commonHelp
	width         int
	height        int
}

func newListModel() listModel {
	var (
		itemGenerator randomItemGenerator
		delegateKeys  = newDelegateKeyMap()
		listKeys      = newListKeyMap()
	)

	// Load reminders from command
	items, err := loadReminders()
	if err != nil {
		// Fallback to empty list if reminders command fails
		items = []list.Item{}
	}

	// Setup list
	delegate := newItemDelegate(delegateKeys)
	remindersList := list.New(items, delegate, 0, 0)
	remindersList.Title = "Reminders"
	remindersList.Styles.Title = titleStyle

	// Customize list styles with theme colors
	remindersList.Styles.PaginationStyle = lipgloss.NewStyle().
		Foreground(theme.BrightBlack())

	remindersList.Styles.HelpStyle = lipgloss.NewStyle().
		Foreground(theme.BrightBlack())

	remindersList.Styles.ActivePaginationDot = lipgloss.NewStyle().
		Foreground(theme.BrightCyan()).
		SetString("•")

	remindersList.Styles.InactivePaginationDot = lipgloss.NewStyle().
		Foreground(theme.BrightBlack()).
		SetString("•")

	remindersList.Styles.FilterPrompt = lipgloss.NewStyle().
		Foreground(theme.BrightCyan())

	remindersList.Styles.FilterCursor = lipgloss.NewStyle().
		Foreground(theme.BrightCyan())

	// Customize the help view styles
	remindersList.Help.Styles.ShortKey = lipgloss.NewStyle().
		Foreground(theme.BrightBlack())

	remindersList.Help.Styles.ShortDesc = lipgloss.NewStyle().
		Foreground(theme.Fg())

	remindersList.Help.Styles.ShortSeparator = lipgloss.NewStyle().
		Foreground(theme.BrightBlack())

	remindersList.Help.Styles.FullKey = lipgloss.NewStyle().
		Foreground(theme.BrightCyan())

	remindersList.Help.Styles.FullDesc = lipgloss.NewStyle().
		Foreground(theme.Fg())

	remindersList.Help.Styles.FullSeparator = lipgloss.NewStyle().
		Foreground(theme.BrightBlack())

	remindersList.Help.Styles.Ellipsis = lipgloss.NewStyle().
		Foreground(theme.BrightBlack())

	remindersList.SetShowPagination(true)
	remindersList.SetShowHelp(false) // Disable list's built-in help, we use commonHelp
	remindersList.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{
			listKeys.toggleSpinner,
			listKeys.insertItem,
			listKeys.toggleTitleBar,
			listKeys.toggleStatusBar,
			listKeys.togglePagination,
			listKeys.toggleHelpMenu,
		}
	}

	return listModel{
		list:          remindersList,
		keys:          listKeys,
		delegateKeys:  delegateKeys,
		itemGenerator: &itemGenerator,
		commonHelp:    newCommonHelp(),
	}
}

func (m listModel) Init() tea.Cmd {
	return nil
}

func (m *listModel) reloadWithFilter(enabledLists []string) tea.Cmd {
	items, err := loadRemindersFiltered(enabledLists)
	if err != nil {
		// On error, just keep current items
		return nil
	}

	// SetItems returns a command that needs to be executed
	// This command handles refiltering if a filter is active
	return m.list.SetItems(items)
}

func (m listModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		h, v := appStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v-1) // Reserve 1 line for help

	case tea.KeyMsg:
		// Don't match any of the keys below if we're actively filtering.
		if m.list.FilterState() == list.Filtering {
			break
		}

		switch {
		case key.Matches(msg, m.keys.toggleSpinner):
			cmd := m.list.ToggleSpinner()
			return m, cmd

		case key.Matches(msg, m.keys.toggleTitleBar):
			v := !m.list.ShowTitle()
			m.list.SetShowTitle(v)
			m.list.SetShowFilter(v)
			m.list.SetFilteringEnabled(v)
			return m, nil

		case key.Matches(msg, m.keys.toggleStatusBar):
			m.list.SetShowStatusBar(!m.list.ShowStatusBar())
			return m, nil

		case key.Matches(msg, m.keys.togglePagination):
			m.list.SetShowPagination(!m.list.ShowPagination())
			return m, nil

		case key.Matches(msg, m.keys.toggleHelpMenu):
			m.list.SetShowHelp(!m.list.ShowHelp())
			return m, nil

		case key.Matches(msg, m.keys.insertItem):
			newItem := m.itemGenerator.next()
			insCmd := m.list.InsertItem(0, newItem)
			statusCmd := m.list.NewStatusMessage(statusMessageStyle("Added " + newItem.Title()))
			return m, tea.Batch(insCmd, statusCmd)
		}
	}

	// This will also call our delegate's update function.
	newListModel, cmd := m.list.Update(msg)
	m.list = newListModel
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m listModel) View() string {
	// Render help with large width to avoid wrapping
	helpView := m.commonHelp.View(1000)
	helpHeight := lipgloss.Height(helpView)

	// Compute dynamic list height
	_, v := appStyle.GetFrameSize()
	listHeight := m.height - v - helpHeight
	if listHeight < 0 {
		listHeight = 0
	}

	// Update list size dynamically
	h, _ := appStyle.GetFrameSize()
	listWidth := m.width - h
	m.list.SetSize(listWidth, listHeight)

	listView := m.list.View()

	return appStyle.Render(lipgloss.JoinVertical(lipgloss.Left, listView, helpView))
}
