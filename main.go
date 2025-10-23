package main

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	boxer "github.com/treilik/bubbleboxer"
)

const (
	leftAddr  = "left"
	rightAddr = "right"
)

type model struct {
	tui              boxer.Boxer
	showRightPane    bool
	focusedPane      string // "left" or "right"
	lastWindowSize   tea.WindowSizeMsg
	picker           listPicker // Persistent picker state
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

	// Get unique lists for the picker
	lists, err := getUniqueLists()
	if err != nil {
		lists = []string{} // Fallback to empty if error
	}

	// Create picker (will be stored in model for persistence)
	picker := newListPicker(lists)

	// Layout tree definition
	m := model{
		tui:           boxer.Boxer{},
		showRightPane: true,
		focusedPane:   rightAddr, // Focus sidebar when open
		picker:        picker,
	}

	// Create models
	left := listModelHolder{m: newListModel()}

	m.tui.LayoutTree = m.buildLayoutTree(left, m.picker)

	return m
}

func (m *model) buildLayoutTree(left listModelHolder, right listPicker) boxer.Node {
	if m.showRightPane {
		// Two-pane layout with separator
		return boxer.Node{
			VerticalStacked: false, // horizontal split
			SizeFunc: func(_ boxer.Node, widthOrHeight int) []int {
				// Split width 66/33
				ratio := widthOrHeight / 3
				return []int{2 * ratio, widthOrHeight - (2 * ratio)}
			},
			Children: []boxer.Node{
				stripErr(m.tui.CreateLeaf(leftAddr, left)),
				stripErr(m.tui.CreateLeaf(rightAddr, right)),
			},
		}
	} else {
		// Single pane layout - just the left side
		return stripErr(m.tui.CreateLeaf(leftAddr, left))
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if key.Matches(msg, quitKeys) {
			return m, tea.Quit
		}

		// Check if list is filtering before processing keybinds
		leftModel, leftOk := m.tui.ModelMap[leftAddr]
		isListFiltering := false
		if leftOk {
			holder := leftModel.(listModelHolder)
			if holder.m.list.FilterState() == list.Filtering {
				isListFiltering = true
			}
		}

		// Tab to switch focus between panes (disabled during filtering)
		if msg.String() == "tab" && m.showRightPane && !isListFiltering {
			if m.focusedPane == leftAddr {
				m.focusedPane = rightAddr
			} else {
				m.focusedPane = leftAddr
			}
			return m, nil
		}

		// Toggle right pane with 's' key (disabled during filtering)
		if msg.String() == "s" && !isListFiltering {
			m.showRightPane = !m.showRightPane
			if m.showRightPane {
				m.focusedPane = rightAddr // Focus sidebar when shown
			} else {
				m.focusedPane = leftAddr // Focus list when sidebar hidden
			}
			m.rebuildLayout()
			// Trigger a resize to update the layout
			if m.lastWindowSize.Width > 0 {
				m.tui.UpdateSize(m.lastWindowSize)
			}
			return m, nil
		}

	case tea.WindowSizeMsg:
		m.lastWindowSize = msg
		m.tui.UpdateSize(msg)

	case filterChangeMsg:
		// Filter changed - reload the left pane
		m.editModel(leftAddr, func(v tea.Model) (tea.Model, error) {
			holder := v.(listModelHolder)
			holder.m.reloadWithFilter(msg.enabledLists)
			return holder, nil
		})
		return m, nil
	}

	// Only send keyboard input to the focused pane
	var cmd tea.Cmd
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		// Route keyboard input only to focused pane
		if m.focusedPane == leftAddr {
			m.editModel(leftAddr, func(v tea.Model) (tea.Model, error) {
				v, cmd = v.Update(keyMsg)
				cmds = append(cmds, cmd)
				return v, nil
			})
		} else if m.focusedPane == rightAddr && m.showRightPane {
			// Check if list is filtering - if so, don't send keys to picker
			leftModel, leftOk := m.tui.ModelMap[leftAddr]
			isListFiltering := false
			if leftOk {
				holder := leftModel.(listModelHolder)
				if holder.m.list.FilterState() == list.Filtering {
					isListFiltering = true
				}
			}

			// Only update picker if list is not filtering
			if !isListFiltering {
				updatedPicker, pickerCmd := m.picker.Update(keyMsg)
				m.picker = updatedPicker.(listPicker)
				cmds = append(cmds, pickerCmd)

				// Also update in the boxer model map
				m.editModel(rightAddr, func(v tea.Model) (tea.Model, error) {
					return m.picker, nil
				})
			}
		}
	} else {
		// Non-keyboard messages go to all panes
		m.editModel(leftAddr, func(v tea.Model) (tea.Model, error) {
			v, cmd = v.Update(msg)
			cmds = append(cmds, cmd)
			return v, nil
		})

		if m.showRightPane {
			// Update picker with non-keyboard messages
			updatedPicker, pickerCmd := m.picker.Update(msg)
			m.picker = updatedPicker.(listPicker)
			cmds = append(cmds, pickerCmd)

			// Also update in the boxer model map
			m.editModel(rightAddr, func(v tea.Model) (tea.Model, error) {
				return m.picker, nil
			})
		}
	}

	return m, tea.Batch(cmds...)
}

func (m *model) rebuildLayout() {
	// Get existing left model
	left, leftOk := m.tui.ModelMap[leftAddr]

	// Clear the model map
	m.tui.ModelMap = make(map[string]tea.Model)

	// Rebuild with existing models
	var leftHolder listModelHolder

	if leftOk {
		leftHolder = left.(listModelHolder)
	} else {
		leftHolder = listModelHolder{m: newListModel()}
	}

	// Use the persistent picker from model
	m.tui.LayoutTree = m.buildLayoutTree(leftHolder, m.picker)
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


func main() {
	rand.Seed(time.Now().UTC().UnixNano())

	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
