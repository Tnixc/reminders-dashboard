package main

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	boxer "github.com/treilik/bubbleboxer"
)

const (
	leftAddr  = "left"
	rightAddr = "right"
)

type model struct {
	tui boxer.Boxer
}

var quitKeys = key.NewBinding(
	key.WithKeys("q", "ctrl+c"),
	key.WithHelp("", "press q to quit"),
)

func stripErr(n boxer.Node, _ error) boxer.Node {
	return n
}

func initialModel() model {
	// Style the separator with a subtle theme color
	separatorStyle := lipgloss.NewStyle().
		Foreground(theme.BrightBlack())
	boxer.HorizontalSeparator = separatorStyle.Render("│")

	// Create models
	left := listModelHolder{m: newListModel()}
	right := stringer("Hello World")

	// Layout tree definition
	m := model{tui: boxer.Boxer{}}
	m.tui.LayoutTree = boxer.Node{
		VerticalStacked: false, // horizontal split
		SizeFunc: func(_ boxer.Node, widthOrHeight int) []int {
			// Split width 66/33
			ratio := widthOrHeight / 3
			return []int{2 * ratio, widthOrHeight - ratio}
		},
		Children: []boxer.Node{
			stripErr(m.tui.CreateLeaf(leftAddr, left)),
			stripErr(m.tui.CreateLeaf(rightAddr, right)),
		},
	}

	return m
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if key.Matches(msg, quitKeys) {
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		m.tui.UpdateSize(msg)
	}

	// Update the list model
	var cmd tea.Cmd
	m.editModel(leftAddr, func(v tea.Model) (tea.Model, error) {
		v, cmd = v.Update(msg)
		return v, nil
	})

	return m, cmd
}

func (m model) View() string {
	return m.tui.View()
}

func (m *model) editModel(addr string, edit func(tea.Model) (tea.Model, error)) error {
	if edit == nil {
		return fmt.Errorf("no edit function provided")
	}
	v, ok := m.tui.ModelMap[addr]
	if !ok {
		return fmt.Errorf("no model with address '%s' found", addr)
	}
	v, err := edit(v)
	if err != nil {
		return err
	}
	m.tui.ModelMap[addr] = v
	return nil
}

// listModelHolder wraps listModel to satisfy tea.Model
type listModelHolder struct {
	m listModel
}

func (l listModelHolder) Init() tea.Cmd {
	return l.m.Init()
}

func (l listModelHolder) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	m, cmd := l.m.Update(msg)
	l.m = m.(listModel)
	return l, cmd
}

func (l listModelHolder) View() string {
	return l.m.View()
}

// stringer is a simple string model with padding
type stringer string

func (s stringer) String() string {
	return string(s)
}

func (s stringer) Init() tea.Cmd                           { return nil }
func (s stringer) Update(msg tea.Msg) (tea.Model, tea.Cmd) { return s, nil }
func (s stringer) View() string {
	// Apply the same padding as the list
	return appStyle.Render(s.String())
}

func main() {
	rand.Seed(time.Now().UTC().UnixNano())

	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
