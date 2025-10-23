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
	tui            boxer.Boxer
	showRightPane  bool
	focusedPane    string // "left", "right", or "multi"
	lastWindowSize tea.WindowSizeMsg
	picker         listPicker // Persistent picker state
	viewMode       string     // "single" or "multi"

	// View-specific models
	singleList listModel           // For single view mode
	multiView  multiColumnView     // For multi-column view mode
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

	// Get all enabled lists for initial state
	enabledLists := picker.getEnabledLists()

	// Create the single list model
	singleListModel := newListModel()

	// Create the multi-column view
	multiColumnViewModel := newMultiColumnView(enabledLists)
	multiColumnViewModel.loadItems()

	// Layout tree definition
	m := model{
		tui:           boxer.Boxer{},
		showRightPane: false, // Sidebar closed by default
		focusedPane:   leftAddr, // Focus left pane
		picker:        picker,
		viewMode:      "single", // Start in single view mode
		singleList:    singleListModel,
		multiView:     multiColumnViewModel,
	}

	// Create models for single view
	left := listModelHolder{m: singleListModel}

	m.tui.LayoutTree = m.buildLayoutTree(left, m.picker)

	return m
}

func (m *model) buildLayoutTree(left tea.Model, right listPicker) boxer.Node {
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
		isListFiltering := false
		if m.viewMode == "single" {
			leftModel, leftOk := m.tui.ModelMap[leftAddr]
			if leftOk {
				if holder, ok := leftModel.(listModelHolder); ok {
					if holder.m.list.FilterState() == list.Filtering {
						isListFiltering = true
					}
				}
			}
		} else {
			// In multi-view, check the multiView's filtering state
			isListFiltering = m.multiView.filtering
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
				m.focusedPane = leftAddr // Focus left pane (regardless of view mode)
			}

			// Rebuild layout (both views use bubbleboxer now)
			m.rebuildLayout()
			// Trigger a resize to update the layout
			if m.lastWindowSize.Width > 0 {
				m.tui.UpdateSize(m.lastWindowSize)
			}
			return m, nil
		}

		// Toggle view mode with 'v' key (disabled during filtering)
		if msg.String() == "v" && !isListFiltering {
			if m.viewMode == "single" {
				m.viewMode = "multi"
				m.focusedPane = leftAddr // Still use leftAddr for multi-view
				// Reload items in multi-view
				m.multiView.loadItems()
			} else {
				m.viewMode = "single"
				m.focusedPane = leftAddr
			}

			// Rebuild layout to switch between models
			m.rebuildLayout()
			if m.lastWindowSize.Width > 0 {
				m.tui.UpdateSize(m.lastWindowSize)
			}
			return m, nil
		}

	case tea.WindowSizeMsg:
		m.lastWindowSize = msg
		m.tui.UpdateSize(msg)

		// Also update multi-column view
		m.multiView, _ = m.multiView.Update(msg)

		// Update picker size for sidebar
		rightWidth := msg.Width / 3
		m.picker.width = rightWidth
		m.picker.height = msg.Height

		return m, nil

	case filterChangeMsg:
		if m.viewMode == "single" {
			// Filter changed - reload the left pane
			var cmd tea.Cmd
			m.editModel(leftAddr, func(v tea.Model) (tea.Model, error) {
				holder := v.(listModelHolder)
				cmd = holder.m.reloadWithFilter(msg.enabledLists)
				return holder, nil
			})
			return m, cmd
		} else {
			// Multi-column view - update enabled lists
			m.multiView.updateEnabledLists(msg.enabledLists)

			// Update the holder in the model map
			m.editModel(leftAddr, func(v tea.Model) (tea.Model, error) {
				holder := v.(multiColumnViewHolder)
				holder.m = m.multiView
				return holder, nil
			})
			return m, nil
		}
	}

	// Only send keyboard input to the focused pane
	var cmd tea.Cmd
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		// Route keyboard input based on view mode and focus
		if m.focusedPane == rightAddr && m.showRightPane {
			// Picker is focused - check if anything is filtering first
			isListFiltering := false
			if m.viewMode == "single" {
				leftModel, leftOk := m.tui.ModelMap[leftAddr]
				if leftOk {
					if holder, ok := leftModel.(listModelHolder); ok {
						if holder.m.list.FilterState() == list.Filtering {
							isListFiltering = true
						}
					}
				}
			} else {
				// Multi-view has its own filtering
				isListFiltering = m.multiView.filtering
			}

			// Only update picker if nothing is filtering
			if !isListFiltering {
				updatedPicker, pickerCmd := m.picker.Update(keyMsg)
				m.picker = updatedPicker.(listPicker)
				cmds = append(cmds, pickerCmd)

				// Also update in the boxer model map
				m.editModel(rightAddr, func(v tea.Model) (tea.Model, error) {
					return m.picker, nil
				})
			}
		} else if m.focusedPane == leftAddr {
			// Left pane focused - could be single list or multi-column view
			if m.viewMode == "single" {
				// Single view mode - route to list
				m.editModel(leftAddr, func(v tea.Model) (tea.Model, error) {
					v, cmd = v.Update(keyMsg)
					cmds = append(cmds, cmd)
					return v, nil
				})
			} else {
				// Multi-column view - route through editModel to update holder
				m.editModel(leftAddr, func(v tea.Model) (tea.Model, error) {
					holder := v.(multiColumnViewHolder)
					holder.m, cmd = holder.m.Update(keyMsg)
					cmds = append(cmds, cmd)
					// Update the main model's copy too
					m.multiView = holder.m
					return holder, nil
				})
			}
		}
	} else {
		// Non-keyboard messages go to all panes
		m.editModel(leftAddr, func(v tea.Model) (tea.Model, error) {
			v, cmd = v.Update(msg)
			cmds = append(cmds, cmd)

			// If multi-view, sync back to main model
			if m.viewMode == "multi" {
				if holder, ok := v.(multiColumnViewHolder); ok {
					m.multiView = holder.m
				}
			}
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
	// Store picker dimensions before clearing
	pickerWidth := m.picker.width
	pickerHeight := m.picker.height

	// Clear the model map
	m.tui.ModelMap = make(map[string]tea.Model)

	// Restore picker dimensions
	m.picker.width = pickerWidth
	m.picker.height = pickerHeight

	if m.viewMode == "single" {
		// Single view - use list model
		leftHolder := listModelHolder{m: m.singleList}
		m.tui.LayoutTree = m.buildLayoutTree(leftHolder, m.picker)
	} else {
		// Multi-column view - use multi-column model
		multiHolder := multiColumnViewHolder{m: m.multiView}
		m.tui.LayoutTree = m.buildLayoutTree(multiHolder, m.picker)
	}

	// Update picker size if sidebar is shown and we have window size
	if m.showRightPane && m.lastWindowSize.Width > 0 {
		rightWidth := m.lastWindowSize.Width / 3
		m.picker.width = rightWidth
		m.picker.height = m.lastWindowSize.Height
	}
}

func (m model) View() string {
	// Always use bubbleboxer - it handles both single and multi view
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

// multiColumnViewHolder wraps multiColumnView to satisfy tea.Model
type multiColumnViewHolder struct {
	m multiColumnView
}

func (mc multiColumnViewHolder) Init() tea.Cmd {
	return mc.m.Init()
}

func (mc multiColumnViewHolder) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	m, cmd := mc.m.Update(msg)
	mc.m = m
	return mc, cmd
}

func (mc multiColumnViewHolder) View() string {
	return mc.m.View()
}


func main() {
	rand.Seed(time.Now().UTC().UnixNano())

	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
