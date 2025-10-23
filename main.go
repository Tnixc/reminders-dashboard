package main

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

type errMsg error

type model struct {
	list     listModel
	quitting bool
	err      error
	width    int
	height   int
}

var quitKeys = key.NewBinding(
	key.WithKeys("q", "esc", "ctrl+c"),
	key.WithHelp("", "press q to quit"),
)

func initialModel() model {
	return model{
		list: newListModel(),
	}
}

func (m model) Init() tea.Cmd {
	return m.list.Init()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyMsg:
		if key.Matches(msg, quitKeys) {
			m.quitting = true
			return m, tea.Quit
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case errMsg:
		m.err = msg
		return m, nil
	}

	// Pass message to list
	var cmd tea.Cmd
	updatedList, cmd := m.list.Update(msg)
	m.list = updatedList.(listModel)
	return m, cmd
}

func (m model) View() string {
	if m.err != nil {
		return m.err.Error()
	}

	// Don't render until we have dimensions
	if m.width == 0 || m.height == 0 {
		return ""
	}

	return m.list.View()
}

func main() {
	rand.Seed(time.Now().UTC().UnixNano())

	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
