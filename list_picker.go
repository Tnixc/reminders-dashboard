package main

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type listItem struct {
	name    string
	enabled bool
}

type listPicker struct {
	items  []listItem
	cursor int
	width  int
	height int
}

type filterChangeMsg struct {
	enabledLists []string
}

func newListPicker(lists []string) listPicker {
	items := make([]listItem, len(lists))
	for i, name := range lists {
		items[i] = listItem{
			name:    name,
			enabled: true, // All enabled by default
		}
	}

	return listPicker{
		items:  items,
		cursor: 0,
	}
}

func (lp listPicker) Init() tea.Cmd {
	return nil
}

func (lp listPicker) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		lp.width = msg.Width
		lp.height = msg.Height
		return lp, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if lp.cursor > 0 {
				lp.cursor--
			}
		case "down", "j":
			if lp.cursor < len(lp.items)-1 {
				lp.cursor++
			}
		case " ", "enter":
			// Toggle the current item
			if lp.cursor < len(lp.items) {
				lp.items[lp.cursor].enabled = !lp.items[lp.cursor].enabled
				// Notify parent of the change
				return lp, func() tea.Msg {
					return filterChangeMsg{enabledLists: lp.getEnabledLists()}
				}
			}
		case "a":
			// Enable all
			for i := range lp.items {
				lp.items[i].enabled = true
			}
			return lp, func() tea.Msg {
				return filterChangeMsg{enabledLists: lp.getEnabledLists()}
			}
		case "n":
			// Disable all
			for i := range lp.items {
				lp.items[i].enabled = false
			}
			return lp, func() tea.Msg {
				return filterChangeMsg{enabledLists: lp.getEnabledLists()}
			}
		}
	}
	return lp, nil
}

func (lp listPicker) View() string {
	if lp.width == 0 || lp.height == 0 {
		return ""
	}

	// Account for appStyle padding (1 vertical, 2 horizontal each side)
	h, _ := appStyle.GetFrameSize()
	maxWidth := lp.width - h
	if maxWidth < 10 {
		return appStyle.Render("Too narrow")
	}

	var output string

	// Title - account for padding (2 chars total)
	titleText := "Filter Lists"
	titlePadding := 2
	if len(titleText)+titlePadding > maxWidth {
		titleText = titleText[:maxWidth-titlePadding-3] + "..."
	}

	titleStyle := lipgloss.NewStyle().
		Foreground(theme.Bg()).
		Background(theme.Blue()).
		Padding(0, 1).
		MaxWidth(maxWidth)

	output += titleStyle.Render(titleText) + "\n"
	output += "\n"

	// Items - each line: "› [✔] Name" (7 chars for markers + name)
	markerWidth := 7 // "› [✔] "
	maxNameWidth := maxWidth - markerWidth
	if maxNameWidth < 3 {
		maxNameWidth = 3
	}

	itemStyle := lipgloss.NewStyle().Foreground(theme.Fg())
	cursorStyle := lipgloss.NewStyle().Foreground(theme.BrightCyan())

	for i, item := range lp.items {
		cursor := " "
		if i == lp.cursor {
			cursor = "›"
		}

		checkbox := "[ ]"
		if item.enabled {
			checkbox = "[✔]"
		}

		// Truncate name if needed
		name := item.name
		if len(name) > maxNameWidth {
			if maxNameWidth > 3 {
				name = name[:maxNameWidth-3] + "..."
			} else {
				name = name[:maxNameWidth]
			}
		}

		line := fmt.Sprintf("%s %s %s", cursor, checkbox, name)

		if i == lp.cursor {
			output += cursorStyle.Render(line) + "\n"
		} else {
			output += itemStyle.Render(line) + "\n"
		}
	}

	// Help text
	output += "\n"
	helpText := "↑/↓ space a/n"
	if len(helpText) > maxWidth {
		helpText = helpText[:maxWidth-3] + "..."
	}
	helpStyle := lipgloss.NewStyle().Foreground(theme.BrightBlack())
	output += helpStyle.Render(helpText)

	return appStyle.Render(output)
}

func (lp listPicker) getEnabledLists() []string {
	var enabled []string
	for _, item := range lp.items {
		if item.enabled {
			enabled = append(enabled, item.name)
		}
	}
	return enabled
}
