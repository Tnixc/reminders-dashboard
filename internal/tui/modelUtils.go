package tui

import (
	"bytes"
	"errors"
	"fmt"
	"maps"
	"os"
	"os/exec"
	"text/template"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	log "github.com/charmbracelet/log"

	"github.com/dlvhdr/gh-dash/v4/internal/config"
	"github.com/dlvhdr/gh-dash/v4/internal/data"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/section"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/constants"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/context"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/markdown"
)

func (m *Model) getCurrSection() section.Section {
	sections := m.getCurrentViewSections()
	if len(sections) == 0 || m.currSectionId >= len(sections) {
		return nil
	}
	return sections[m.currSectionId]
}

func (m *Model) getCurrRowData() data.RowData {
	section := m.getCurrSection()
	if section == nil {
		return nil
	}
	return section.GetCurrRow()
}

func (m *Model) getSectionAt(id int) section.Section {
	sections := m.getCurrentViewSections()
	if len(sections) <= id {
		return nil
	}
	return sections[id]
}

func (m *Model) getPrevSectionId() int {
	return max(0, (m.currSectionId - 1))
}

func (m *Model) getNextSectionId() int {
	return min((m.currSectionId + 1), len(m.ctx.GetViewSectionsConfig())-1)
}

type IssueCommandTemplateInput struct {
	RepoName    string
	RepoPath    string
	IssueNumber int
	HeadRefName string
}

func (m *Model) executeKeybinding(key string) tea.Cmd {
	currRowData := m.getCurrRowData()

	for _, keybinding := range m.ctx.Config.Keybindings.Universal {
		if keybinding.Key != key {
			continue
		}

		log.Debug("executing keybind", "key", keybinding.Key, "command", keybinding.Command)
		return m.runCustomUniversalCommand(keybinding.Command)
	}

	switch m.ctx.View {
	case config.RemindersView:
		for _, keybinding := range m.ctx.Config.Keybindings.Reminders {
			if keybinding.Key != key || keybinding.Command == "" {
				continue
			}

			log.Debug("executing keybind", "key", keybinding.Key, "command", keybinding.Command)

			switch data := currRowData.(type) {
			case *data.Reminder:
				return m.runCustomReminderCommand(keybinding.Command, data)
			}
		}
	default:
		// Not a valid case - ignore it
	}

	return nil
}

// runCustomCommand executes a user-defined command.
// commandTemplate is a template string that will be parsed with the input data.
// contextData is a map of key-value pairs of data specific to the context the command is being run in.
func (m *Model) runCustomCommand(commandTemplate string, contextData *map[string]any) tea.Cmd {
	// A generic map is a pretty easy & flexible way to populate a template if there's no pressing need
	// for structured data, existing structs, etc. Especially if holes in the data are expected.
	// Common data shared across contexts could be set here.
	input := map[string]any{}

	// Merge data specific to the context the command is being run in onto any common data, overwriting duplicate keys.
	if contextData != nil {
		maps.Copy(input, *contextData)
	}

	// No RepoPaths for reminders

	cmd, err := template.New("keybinding_command").Parse(commandTemplate)
	if err != nil {
		log.Fatal("Failed parse keybinding template", "error", err)
	}

	// Set the command to error out if required input (e.g. RepoPath) is missing
	cmd = cmd.Option("missingkey=error")

	var buff bytes.Buffer
	err = cmd.Execute(&buff, input)
	if err != nil {
		return func() tea.Msg {
			return constants.ErrMsg{Err: fmt.Errorf("failed to parsetemplate %s", commandTemplate)}
		}
	}
	return m.executeCustomCommand(buff.String())
}

func (m *Model) runCustomReminderCommand(commandTemplate string, reminderData *data.Reminder) tea.Cmd {
	return m.runCustomCommand(commandTemplate,
		&map[string]any{
			"ReminderId":    reminderData.Id,
			"ReminderTitle": reminderData.Title,
		})
}

func (m *Model) runCustomUniversalCommand(commandTemplate string) tea.Cmd {
	input := map[string]any{"RepoPath": m.ctx.RepoPath}
	return m.runCustomCommand(commandTemplate, &input)
}

type execProcessFinishedMsg struct{}

func (m *Model) executeCustomCommand(cmd string) tea.Cmd {
	log.Debug("executing custom command", "cmd", cmd)
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "sh"
	}
	c := exec.Command(shell, "-c", cmd)
	return tea.ExecProcess(c, func(err error) tea.Msg {
		if err != nil {
			mdRenderer := markdown.GetMarkdownRenderer(m.ctx.ScreenWidth)
			md, mdErr := mdRenderer.Render(fmt.Sprintf("While running: `%s`", cmd))
			if mdErr != nil {
				return constants.ErrMsg{Err: mdErr}
			}
			return constants.ErrMsg{Err: errors.New(
				lipgloss.JoinVertical(lipgloss.Left,
					fmt.Sprintf("Whoops, got an error: %s", err),
					md,
				),
			)}
		}
		return execProcessFinishedMsg{}
	})
}

func (m *Model) notify(text string) tea.Cmd {
	id := fmt.Sprint(time.Now().Unix())
	startCmd := m.ctx.StartTask(
		context.Task{
			Id:           id,
			StartText:    text,
			FinishedText: text,
			State:        context.TaskStart,
		})

	finishCmd := func() tea.Msg {
		return constants.TaskFinishedMsg{
			TaskId: id,
		}
	}

	return tea.Sequence(startCmd, finishCmd)
}

func (m *Model) notifyErr(text string) tea.Cmd {
	id := fmt.Sprint(time.Now().Unix())
	startCmd := m.ctx.StartTask(
		context.Task{
			Id:           id,
			StartText:    text,
			FinishedText: text,
			State:        context.TaskStart,
		})

	finishCmd := func() tea.Msg {
		return constants.TaskFinishedMsg{
			TaskId: id,
			Err:    errors.New(text),
		}
	}

	return tea.Sequence(startCmd, finishCmd)
}
