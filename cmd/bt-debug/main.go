package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

type model struct {
	input    textinput.Model
	spinner  spinner.Model
	selected int
	lastKey  string
	summary  string
}

var rows = []struct {
	Command     string
	Description string
}{
	{Command: "rg --files -t go", Description: "Use ripgrep to enumerate Go files quickly."},
	{Command: "find . -type f -name \"*.go\"", Description: "Use find to list Go files recursively."},
	{Command: "git ls-files \"*.go\"", Description: "Show tracked Go files in the repository."},
}

var (
	primaryColor   = lipgloss.AdaptiveColor{Light: "#1D4ED8", Dark: "#7DD3FC"}
	mutedColor     = lipgloss.AdaptiveColor{Light: "#6B7280", Dark: "#94A3B8"}
	selectedText   = lipgloss.AdaptiveColor{Light: "#111827", Dark: "#F8FAFC"}
	selectedBg     = lipgloss.AdaptiveColor{Light: "#BFDBFE", Dark: "#1E3A8A"}
	selectedDesc   = lipgloss.AdaptiveColor{Light: "#334155", Dark: "#DBEAFE"}
	commandColor   = lipgloss.AdaptiveColor{Light: "#0F172A", Dark: "#F8FAFC"}
	descriptionCol = lipgloss.AdaptiveColor{Light: "#64748B", Dark: "#94A3B8"}
	accentColor    = lipgloss.AdaptiveColor{Light: "#2563EB", Dark: "#93C5FD"}

	titleStyle = lipgloss.NewStyle().Bold(true).Foreground(primaryColor)
	infoStyle  = lipgloss.NewStyle().Foreground(mutedColor)
	rowStyle   = lipgloss.NewStyle().PaddingLeft(2).Foreground(commandColor).Bold(true)
	descStyle  = lipgloss.NewStyle().PaddingLeft(4).Foreground(descriptionCol).Italic(true)
	activeRow  = lipgloss.NewStyle().
			Bold(true).
			Foreground(selectedText).
			Background(selectedBg).
			PaddingLeft(2).
			PaddingRight(2)
	activeDesc = lipgloss.NewStyle().
			Foreground(selectedDesc).
			Background(selectedBg).
			PaddingLeft(4).
			PaddingRight(2).
			Italic(true)
)

func main() {
	input := textinput.New()
	input.SetValue("find me all go files in this folder")
	input.Focus()
	input.CursorEnd()
	input.CharLimit = 2000
	input.Prompt = "> "
	input.PromptStyle = lipgloss.NewStyle().Foreground(primaryColor).Bold(true)
	input.TextStyle = lipgloss.NewStyle().Foreground(commandColor)
	input.Cursor.Style = lipgloss.NewStyle().Foreground(primaryColor).Bold(true)

	spin := spinner.New()
	spin.Spinner = spinner.Dot
	spin.Style = lipgloss.NewStyle().Foreground(primaryColor)

	p := tea.NewProgram(model{
		input:   input,
		spinner: spin,
		lastKey: "<none>",
		summary: debugSummary(),
	})

	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(textinput.Blink, m.spinner.Tick)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	case tea.KeyMsg:
		m.lastKey = describeKey(msg)
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			return m, tea.Quit
		case "up":
			if m.selected > 0 {
				m.selected--
			}
			return m, nil
		case "down":
			if m.selected < len(rows)-1 {
				m.selected++
			}
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m model) View() string {
	lines := []string{
		titleStyle.Render("Bubble Tea Debug"),
		"",
		infoStyle.Render(m.summary),
		"",
		titleStyle.Render(m.input.View()),
		"",
		lipgloss.JoinHorizontal(lipgloss.Left, m.spinner.View(), " ", infoStyle.Render("Spinner should be colored")),
		"",
		titleStyle.Render("Last Key"),
		infoStyle.Render(m.lastKey),
		"",
	}

	for i, row := range rows {
		if i == m.selected {
			lines = append(lines, renderSelected(row.Command, row.Description))
		} else {
			lines = append(lines, rowStyle.Render(row.Command))
			lines = append(lines, descStyle.Render(row.Description))
		}
	}

	lines = append(lines,
		"",
		termenv.String("termenv styled sample").Foreground(termenv.ANSIBrightBlue).Bold().String(),
		"",
		infoStyle.Render("Type to edit. Up/down moves selection. q/esc exits."),
	)

	return strings.Join(lines, "\n")
}

func describeKey(msg tea.KeyMsg) string {
	return fmt.Sprintf(
		"string=%q  type=%v  runes=%q  alt=%t  paste=%t",
		msg.String(),
		msg.Type,
		string(msg.Runes),
		msg.Alt,
		msg.Paste,
	)
}

func renderSelected(command, description string) string {
	block := lipgloss.JoinVertical(
		lipgloss.Left,
		activeRow.Render(command),
		activeDesc.Render(description),
	)
	return lipgloss.NewStyle().
		BorderStyle(lipgloss.ThickBorder()).
		BorderLeft(true).
		BorderForeground(accentColor).
		Render(block)
}

func debugSummary() string {
	return fmt.Sprintf(
		"TERM=%s  COLORTERM=%s  TERM_PROGRAM=%s  NO_COLOR=%s  profile=%s",
		os.Getenv("TERM"),
		os.Getenv("COLORTERM"),
		os.Getenv("TERM_PROGRAM"),
		os.Getenv("NO_COLOR"),
		profileName(termenv.EnvColorProfile()),
	)
}

func profileName(p termenv.Profile) string {
	switch p {
	case termenv.TrueColor:
		return "truecolor"
	case termenv.ANSI256:
		return "256"
	case termenv.ANSI:
		return "ansi"
	case termenv.Ascii:
		return "ascii"
	default:
		return "unknown"
	}
}
