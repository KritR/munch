package ui

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"

	munchctx "github.com/krithikr/munch/internal/context"
	"github.com/krithikr/munch/internal/protocol"
	"github.com/krithikr/munch/internal/suggest"
)

const debounceDelay = 250 * time.Millisecond

type Selection struct {
	Action  protocol.Action
	Command string
}

type generateResultMsg struct {
	version     int
	prompt      string
	suggestions []protocol.Suggestion
	err         error
}

type debounceReadyMsg struct {
	version int
}

type selectorModel struct {
	engine      suggest.Engine
	ctx         munchctx.Normalized
	safetyLevel string
	styles      uiStyles
	executeHint string

	input         textinput.Model
	spinner       spinner.Model
	selected      int
	confirming    bool
	pendingAction protocol.Action
	selection     Selection
	width         int
	height        int
	loading       bool
	err           error
	version       int
	suggestions   []protocol.Suggestion
}

type uiStyles struct {
	promptStyle     lipgloss.Style
	hintStyle       lipgloss.Style
	rowStyle        lipgloss.Style
	activeStyle     lipgloss.Style
	descStyle       lipgloss.Style
	activeDescStyle lipgloss.Style
	errorStyle      lipgloss.Style
	accentColor     lipgloss.AdaptiveColor
}

func newUIStyles(r *lipgloss.Renderer) uiStyles {
	primaryColor := lipgloss.AdaptiveColor{Light: "#1D4ED8", Dark: "#7DD3FC"}
	mutedColor := lipgloss.AdaptiveColor{Light: "#6B7280", Dark: "#94A3B8"}
	selectedText := lipgloss.AdaptiveColor{Light: "#111827", Dark: "#F8FAFC"}
	selectedBg := lipgloss.AdaptiveColor{Light: "#BFDBFE", Dark: "#1E3A8A"}
	selectedDesc := lipgloss.AdaptiveColor{Light: "#334155", Dark: "#DBEAFE"}
	commandColor := lipgloss.AdaptiveColor{Light: "#0F172A", Dark: "#F8FAFC"}
	descriptionCol := lipgloss.AdaptiveColor{Light: "#64748B", Dark: "#94A3B8"}
	errorColor := lipgloss.AdaptiveColor{Light: "#B91C1C", Dark: "#FCA5A5"}
	accentColor := lipgloss.AdaptiveColor{Light: "#2563EB", Dark: "#93C5FD"}
	return uiStyles{
		promptStyle: r.NewStyle().Bold(true).Foreground(primaryColor),
		hintStyle:   r.NewStyle().Foreground(mutedColor),
		rowStyle:    r.NewStyle().PaddingLeft(2).Foreground(commandColor).Bold(true),
		activeStyle: r.NewStyle().
			Bold(true).
			Foreground(selectedText).
			Background(selectedBg).
			PaddingLeft(2).
			PaddingRight(2),
		descStyle: r.NewStyle().
			Foreground(descriptionCol).
			Italic(true).
			PaddingLeft(4),
		activeDescStyle: r.NewStyle().
			Foreground(selectedDesc).
			Italic(true).
			Background(selectedBg).
			PaddingLeft(4).
			PaddingRight(2),
		errorStyle:  r.NewStyle().Foreground(errorColor).Bold(true),
		accentColor: accentColor,
	}
}

func SelectSuggestion(prompt string, engine suggest.Engine, ctx munchctx.Normalized, safetyLevel string) (Selection, error) {
	tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err != nil {
		return Selection{}, err
	}
	defer tty.Close()

	renderer := lipgloss.NewRenderer(tty)
	styles := newUIStyles(renderer)
	executeHint := executeShortcutHint(os.Getenv("TERM_PROGRAM"), os.Getenv("TERM"))

	input := textinput.New()
	input.SetValue(prompt)
	input.Focus()
	input.CursorEnd()
	input.CharLimit = 2000
	input.Prompt = "> "
	input.Width = max(32, len(prompt)+8)
	input.PromptStyle = renderer.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#1D4ED8", Dark: "#7DD3FC"}).Bold(true)
	input.TextStyle = renderer.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#0F172A", Dark: "#F8FAFC"})
	input.Cursor.Style = renderer.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#1D4ED8", Dark: "#7DD3FC"}).Bold(true)

	spin := spinner.New()
	spin.Spinner = spinner.Dot
	spin.Style = renderer.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#1D4ED8", Dark: "#7DD3FC"})

	model := selectorModel{
		engine:      engine,
		ctx:         ctx,
		safetyLevel: safetyLevel,
		styles:      styles,
		executeHint: executeHint,
		input:       input,
		spinner:     spin,
		version:     1,
		loading:     true,
		selection: Selection{
			Action: protocol.ActionCancel,
		},
		pendingAction: protocol.ActionInsert,
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
	clearRenderedUI(tty, result.renderHeight())
	return result.selection, nil
}

func (m selectorModel) Init() tea.Cmd {
	return tea.Batch(textinput.Blink, m.spinner.Tick, scheduleDebounce(m.version))
}

func (m selectorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		if msg.Width > 0 {
			m.input.Width = max(20, msg.Width-4)
		}
		return m, nil
	case spinner.TickMsg:
		if !m.loading {
			return m, nil
		}
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	case debounceReadyMsg:
		if msg.version != m.version {
			return m, nil
		}
		prompt := strings.TrimSpace(m.input.Value())
		m.loading = true
		m.err = nil
		return m, generateSuggestions(msg.version, prompt, m.engine, m.ctx, m.safetyLevel)
	case generateResultMsg:
		if msg.version != m.version {
			return m, nil
		}
		m.loading = false
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		m.err = nil
		m.suggestions = msg.suggestions
		if len(m.suggestions) == 0 {
			m.selected = 0
			return m, nil
		}
		if m.selected >= len(m.suggestions) {
			m.selected = len(m.suggestions) - 1
		}
		return m, nil
	case tea.KeyMsg:
		if m.confirming {
			return m.updateConfirming(msg)
		}
		switch msg.String() {
		case "ctrl+c", "esc":
			m.selection = Selection{Action: protocol.ActionCancel}
			return m, tea.Quit
		case "up":
			if len(m.suggestions) > 0 && m.selected > 0 {
				m.selected--
			}
			return m, nil
		case "down":
			if len(m.suggestions) > 0 && m.selected < len(m.suggestions)-1 {
				m.selected++
			}
			return m, nil
		case "enter":
			return m.prepareSuggestionAction(protocol.ActionInsert)
		case "alt+enter", "ctrl+e":
			return m.prepareSuggestionAction(protocol.ActionExecute)
		}

		prevValue := m.input.Value()
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		m.confirming = false
		if m.input.Value() == prevValue {
			return m, cmd
		}
		m.version++
		m.loading = true
		m.err = nil
		return m, tea.Batch(cmd, scheduleDebounce(m.version))
	}
	return m, nil
}

func (m selectorModel) updateConfirming(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "esc", "n":
		m.confirming = false
		m.pendingAction = protocol.ActionInsert
		return m, nil
	case "y", "enter":
		if len(m.suggestions) == 0 {
			m.confirming = false
			return m, nil
		}
		m.selection = Selection{
			Action:  m.pendingAction,
			Command: m.suggestions[m.selected].Command,
		}
		return m, tea.Quit
	default:
		return m, nil
	}
}

func (m selectorModel) View() string {
	lines := []string{m.styles.promptStyle.Render(m.input.View())}

	if m.err != nil {
		lines = append(lines, m.styles.errorStyle.Render("Error: "+m.err.Error()))
	} else if m.loading {
		lines = append(lines, lipgloss.JoinHorizontal(lipgloss.Left, m.spinner.View(), " ", m.styles.hintStyle.Render("Generating suggestions")))
	}

	if m.confirming && len(m.suggestions) > 0 {
		suggestion := m.suggestions[m.selected]
		reason := suggestion.ConfirmationReason
		if reason == "" {
			reason = "This command requires confirmation."
		}
		actionLabel := "Insert"
		if m.pendingAction == protocol.ActionExecute {
			actionLabel = "Execute"
		}
		lines = append(lines,
			"",
			m.styles.activeStyle.Render("Confirm "+actionLabel),
			reason,
			"",
			m.styles.rowStyle.Render(suggestion.Command),
			"",
			m.styles.hintStyle.Render("enter/y: confirm • esc/n: go back"),
		)
		return lipgloss.JoinVertical(lipgloss.Left, lines...)
	}

	if len(m.suggestions) == 0 {
		if !m.loading {
			lines = append(lines, "", m.styles.hintStyle.Render("No suggestions yet. Keep typing or wait for results."))
		}
		lines = append(lines, "", m.styles.hintStyle.Render("up/down: move • enter: insert • "+m.executeHint+": execute • esc: cancel"))
		return lipgloss.JoinVertical(lipgloss.Left, lines...)
	}

	lines = append(lines, "")
	for i, suggestion := range m.suggestions {
		if i == m.selected {
			lines = append(lines, m.renderSelectedSuggestion(suggestion))
		} else {
			lines = append(lines, m.styles.rowStyle.Render(suggestion.Command))
			if suggestion.Description != "" {
				lines = append(lines, m.styles.descStyle.Render(suggestion.Description))
			}
		}
	}

	lines = append(lines, "", m.styles.hintStyle.Render("type to edit • up/down: move • enter: insert • "+m.executeHint+": execute • esc: cancel"))
	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func scheduleDebounce(version int) tea.Cmd {
	return tea.Tick(debounceDelay, func(time.Time) tea.Msg {
		return debounceReadyMsg{version: version}
	})
}

func generateSuggestions(version int, prompt string, engine suggest.Engine, ctx munchctx.Normalized, safetyLevel string) tea.Cmd {
	return func() tea.Msg {
		suggestions, err := engine.Generate(prompt, ctx, safetyLevel)
		return generateResultMsg{
			version:     version,
			prompt:      prompt,
			suggestions: suggestions,
			err:         err,
		}
	}
}

func (m selectorModel) renderSelectedSuggestion(suggestion protocol.Suggestion) string {
	parts := []string{
		m.styles.activeStyle.Render(suggestion.Command),
	}
	if suggestion.Description != "" {
		parts = append(parts, m.styles.activeDescStyle.Render(suggestion.Description))
	}

	block := lipgloss.JoinVertical(lipgloss.Left, parts...)
	return lipgloss.NewStyle().
		BorderStyle(lipgloss.ThickBorder()).
		BorderLeft(true).
		BorderForeground(m.styles.accentColor).
		Render(block)
}

func (m selectorModel) prepareSuggestionAction(action protocol.Action) (tea.Model, tea.Cmd) {
	if len(m.suggestions) == 0 || m.loading {
		return m, nil
	}
	m.pendingAction = action
	if m.suggestions[m.selected].RequiresConfirmation {
		m.confirming = true
		return m, nil
	}
	m.selection = Selection{
		Action:  action,
		Command: m.suggestions[m.selected].Command,
	}
	return m, tea.Quit
}

func (m selectorModel) renderHeight() int {
	view := m.View()
	if view == "" {
		return 0
	}

	width := m.width
	if width <= 0 {
		width = 80
	}

	total := 0
	for _, line := range strings.Split(view, "\n") {
		lineWidth := ansi.StringWidth(line)
		lineHeight := 1
		if lineWidth > width {
			lineHeight = (lineWidth + width - 1) / width
		}
		total += lineHeight
	}
	return total
}

func clearRenderedUI(f *os.File, height int) {
	if f == nil || height <= 0 {
		return
	}

	for i := 0; i < height; i++ {
		fmt.Fprint(f, "\r\033[2K")
		if i < height-1 {
			fmt.Fprint(f, "\033[1A")
		}
	}
	fmt.Fprint(f, "\r")
}

func executeShortcutHint(termProgram, term string) string {
	switch strings.ToLower(termProgram) {
	case "ghostty", "iterm.app", "wezterm", "vscode", "hyper":
		return "alt+enter/ctrl+e"
	}

	switch strings.ToLower(term) {
	case "xterm-ghostty", "wezterm", "xterm-kitty":
		return "alt+enter/ctrl+e"
	default:
		return "ctrl+e"
	}
}
