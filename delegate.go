package main

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type delegateKeyMap struct {
}

func newDelegateKeyMap() *delegateKeyMap {
	return &delegateKeyMap{}
}

// Additional short help entries. This satisfies the help.KeyMap interface and
// is entirely optional.
func (d delegateKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{}
}

// Additional full help entries. This satisfies the help.KeyMap interface and
// is entirely optional.
func (d delegateKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{},
	}
}

type customItemDelegate struct {
	list.DefaultDelegate
	keys *delegateKeyMap
}

func newItemDelegate(keys *delegateKeyMap) customItemDelegate {
	d := list.NewDefaultDelegate()

	// Customize delegate styles with theme colors - remove Foreground from titles
	// so we can apply it selectively
	d.Styles.NormalTitle = lipgloss.NewStyle().
		Padding(0, 0, 0, 2)

	d.Styles.NormalDesc = lipgloss.NewStyle().
		Foreground(theme.BrightBlack()).
		Padding(0, 0, 0, 2)

	d.Styles.SelectedTitle = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), false, false, false, true).
		BorderForeground(theme.BrightCyan()).
		Padding(0, 0, 0, 1)

	d.Styles.SelectedDesc = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), false, false, false, true).
		BorderForeground(theme.BrightCyan()).
		Foreground(theme.BrightBlack()).
		Padding(0, 0, 0, 1)

	d.Styles.DimmedTitle = lipgloss.NewStyle().
		Foreground(theme.BrightBlack()).
		Padding(0, 0, 0, 2)

	d.Styles.DimmedDesc = lipgloss.NewStyle().
		Foreground(theme.BrightBlack()).
		Padding(0, 0, 0, 2)

	d.Styles.FilterMatch = lipgloss.NewStyle().
		Foreground(theme.Yellow()).
		Underline(true)

	d.UpdateFunc = func(msg tea.Msg, m *list.Model) tea.Cmd {
		// Removed "You chose X" functionality
		return nil
	}

	d.ShortHelpFunc = func() []key.Binding {
		return []key.Binding{}
	}

	d.FullHelpFunc = func() [][]key.Binding {
		return [][]key.Binding{{}}
	}

	return customItemDelegate{
		DefaultDelegate: d,
		keys:            keys,
	}
}

// Render renders the item with custom coloring that works with filtering
func (d customItemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(item)
	if !ok {
		return
	}

	str := i.Title()
	desc := i.Description()

	if m.Width() <= 0 {
		return
	}

	// Determine styles based on item state
	var (
		isSelected  = index == m.Index()
		isFiltering = m.FilterState() == list.Filtering || m.FilterState() == list.FilterApplied
		matches     = m.MatchesForItem(index)
	)

	// Choose base styles
	var titleStyle, descStyle lipgloss.Style
	var titleFg lipgloss.TerminalColor

	if isFiltering && len(matches) == 0 && m.FilterValue() != "" {
		// No match - use dimmed style
		titleStyle = d.Styles.DimmedTitle
		descStyle = d.Styles.DimmedDesc
		titleFg = theme.BrightBlack()
	} else if isSelected {
		titleStyle = d.Styles.SelectedTitle
		descStyle = d.Styles.SelectedDesc
		titleFg = theme.BrightCyan()
	} else {
		titleStyle = d.Styles.NormalTitle
		descStyle = d.Styles.NormalDesc
		titleFg = theme.Fg()
	}

	// Render the title with selective coloring
	var renderedTitle string

	if strings.HasPrefix(str, "● ") && i.color != "" {
		// Has colored bullet
		bulletColor := lipgloss.Color(i.color)
		bullet := lipgloss.NewStyle().Foreground(bulletColor).Render("●")
		titleText := str[len("● "):]

		// Apply filter highlighting if needed
		if len(matches) > 0 {
			titleText = d.applyFilterMatches(titleText, matches, titleFg)
		} else {
			titleText = lipgloss.NewStyle().Foreground(titleFg).Render(titleText)
		}

		// Combine bullet and title
		combined := bullet + " " + titleText

		// Apply padding and border
		if isSelected {
			renderedTitle = lipgloss.NewStyle().
				Border(lipgloss.NormalBorder(), false, false, false, true).
				BorderForeground(theme.BrightCyan()).
				Padding(0, 0, 0, 1).
				Render(combined)
		} else {
			renderedTitle = lipgloss.NewStyle().
				Padding(0, 0, 0, 2).
				Render(combined)
		}
	} else {
		// No bullet, apply filter highlighting to whole title
		if len(matches) > 0 {
			str = d.applyFilterMatches(str, matches, titleFg)
		} else {
			str = lipgloss.NewStyle().Foreground(titleFg).Render(str)
		}
		renderedTitle = titleStyle.Render(str)
	}

	// Render description with selective coloring for urgency
	var renderedDesc string
	if i.urgencyText != "" && strings.HasPrefix(desc, i.urgencyText) {
		urgencyColor := urgencyColorToTheme(i.urgencyColor)

		// Build the description with proper padding and border first, then apply colors inline
		var paddingLeft int
		var borderStr string

		if isSelected {
			paddingLeft = 1
			borderStr = lipgloss.NewStyle().
				Foreground(theme.BrightCyan()).
				Render("│")
		} else {
			paddingLeft = 2
		}

		// Render urgency and rest with colors
		urgencyStyled := lipgloss.NewStyle().Foreground(urgencyColor).Inline(true).Render(i.urgencyText)
		restText := desc[len(i.urgencyText):]
		restStyled := lipgloss.NewStyle().Foreground(theme.BrightBlack()).Inline(true).Render(restText)

		// Combine with proper spacing
		var combined string
		if isSelected {
			combined = borderStr + strings.Repeat(" ", paddingLeft) + urgencyStyled + restStyled
		} else {
			combined = strings.Repeat(" ", paddingLeft) + urgencyStyled + restStyled
		}

		renderedDesc = combined
	} else {
		renderedDesc = descStyle.Render(desc)
	}

	// Ensure we don't exceed the list's width to avoid boxer overflow
	maxW := m.Width()
	renderedTitle = lipgloss.NewStyle().MaxWidth(maxW).Render(renderedTitle)
	renderedDesc = lipgloss.NewStyle().MaxWidth(maxW).Render(renderedDesc)
	fmt.Fprintf(w, "%s\n%s", renderedTitle, renderedDesc)
}

// applyFilterMatches applies yellow highlighting to matched character ranges
func (d customItemDelegate) applyFilterMatches(text string, matches []int, baseFg lipgloss.TerminalColor) string {
	if len(matches) == 0 {
		return lipgloss.NewStyle().Foreground(baseFg).Render(text)
	}

	runes := []rune(text)
	var result strings.Builder
	lastIdx := 0

	// matches is a flat array: [start1, end1, start2, end2, ...]
	for i := 0; i < len(matches)-1; i += 2 {
		start, end := matches[i], matches[i+1]

		// Validate bounds
		if start < 0 || end > len(runes) || start >= end {
			continue
		}

		// Add non-matching text before this match
		if start > lastIdx {
			result.WriteString(lipgloss.NewStyle().Foreground(baseFg).Render(string(runes[lastIdx:start])))
		}

		// Add matching text with highlight
		matchStyle := lipgloss.NewStyle().Foreground(theme.Yellow()).Underline(true)
		result.WriteString(matchStyle.Render(string(runes[start:end])))
		lastIdx = end
	}

	// Add remaining non-matching text
	if lastIdx < len(runes) {
		result.WriteString(lipgloss.NewStyle().Foreground(baseFg).Render(string(runes[lastIdx:])))
	}

	return result.String()
}
