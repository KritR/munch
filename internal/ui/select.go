package ui

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/krithikr/munch/internal/protocol"
)

type Selection struct {
	Action  protocol.Action
	Command string
}

type selectorModel struct {
	prompt      string
	suggestions []protocol.Suggestion
	selected    int
	selection   Selection
	width       int
	height      int
}

var (
	titleStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
	hintStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	rowStyle    = lipgloss.NewStyle().PaddingLeft(2)
	activeStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("230")).
			Background(lipgloss.Color("62")).
			PaddingLeft(1).
			PaddingRight(1)
	descStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("7")).PaddingLeft(4)
)

func SelectSuggestion(prompt string, suggestions []protocol.Suggestion) (Selection, error) {
	tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err != nil {
		return Selection{}, err
	}
	defer tty.Close()

	model := selectorModel{
		prompt:      prompt,
		suggestions: suggestions,
		selection: Selection{
			Action: protocol.ActionCancel,
		},
	}

	program := tea.NewProgram(
		model,
		tea.WithInput(tty),
		tea.WithOutput(tty),
	)

	finalModel, err := program.Run()
	if err != nil {
		return Selection{}, err
	}

	result, ok := finalModel.(selectorModel)
	if !ok {
		return Selection{}, fmt.Errorf("unexpected final UI model type %T", finalModel)
	}
	return result.selection, nil
}

func (m selectorModel) Init() tea.Cmd {
	return nil
}

func (m selectorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			m.selection = Selection{Action: protocol.ActionCancel}
			return m, tea.Quit
		case "up", "k":
			if len(m.suggestions) == 0 {
				return m, nil
			}
			if m.selected > 0 {
				m.selected--
			}
			return m, nil
		case "down", "j":
			if len(m.suggestions) == 0 {
				return m, nil
			}
			if m.selected < len(m.suggestions)-1 {
				m.selected++
			}
			return m, nil
		case "enter":
			if len(m.suggestions) == 0 {
				m.selection = Selection{Action: protocol.ActionCancel}
				return m, tea.Quit
			}
			m.selection = Selection{
				Action:  protocol.ActionInsert,
				Command: m.suggestions[m.selected].Command,
			}
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m selectorModel) View() string {
	if len(m.suggestions) == 0 {
		return lipgloss.JoinVertical(
			lipgloss.Left,
			titleStyle.Render("munch"),
			"",
			fmt.Sprintf("No suggestions for: %s", m.prompt),
			"",
			hintStyle.Render("Press Esc or Enter to cancel."),
		)
	}

	lines := []string{
		titleStyle.Render("munch"),
		fmt.Sprintf("Task: %s", m.prompt),
		"",
	}

	for i, suggestion := range m.suggestions {
		prefix := " "
		commandLine := suggestion.Command
		if i == m.selected {
			prefix = ">"
			commandLine = activeStyle.Render(commandLine)
		}
		lines = append(lines, rowStyle.Render(fmt.Sprintf("%s %d. %s", prefix, i+1, commandLine)))
		lines = append(lines, descStyle.Render(suggestion.Description))
		lines = append(lines, "")
	}

	lines = append(lines, hintStyle.Render("up/down or j/k: move • enter: insert • esc: cancel"))
	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}
