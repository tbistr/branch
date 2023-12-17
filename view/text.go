package view

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	LINE_BUFFER_SIZE = 1000
)

type Window struct {
	Title string
	Input <-chan string
	lines []string
}

func (w *Window) pushLine(line string) {
	if len(w.lines) > LINE_BUFFER_SIZE {
		w.lines = w.lines[1:]
	}
	w.lines = append(w.lines, line)
}

type TextViewModel struct {
	Windows     []*Window
	h, w, eachW int
}

type initMsg struct{}

func (m TextViewModel) Init() tea.Cmd {
	return func() tea.Msg {
		return initMsg{}
	}
}

type scanMsg struct {
	i    int
	text string
}

func cmdRead(i int, input <-chan string) tea.Cmd {
	return func() tea.Msg {
		text := <-input
		return scanMsg{i, text}
	}
}

func (m TextViewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case initMsg:
		for i, w := range m.Windows {
			cmds = append(cmds, cmdRead(i, w.Input))
		}

	case scanMsg:
		m.Windows[msg.i].pushLine(msg.text)
		cmd = cmdRead(msg.i, m.Windows[msg.i].Input)

	case tea.WindowSizeMsg:
		m.h = msg.Height
		m.w = msg.Width
		m.eachW = (m.w - len(m.Windows) - 1) / len(m.Windows)
	case tea.KeyMsg:
		if msg.Type == tea.KeyCtrlC {
			cmd = tea.Quit
		}
	}

	if len(cmds) > 0 {
		cmd = tea.Batch(cmds...)
	}
	return m, cmd
}

func (m TextViewModel) View() string {
	border := lipgloss.NewStyle().Render(strings.Repeat("|\n", max(m.h-1, 0)) + "|")

	windowStyle := lipgloss.NewStyle().
		Width(m.eachW).MaxWidth(m.eachW).
		Height(m.h).MaxHeight(m.h)
	titleStyle := lipgloss.NewStyle().
		Width(m.eachW).MaxWidth(m.eachW).
		Bold(true).
		Reverse(true).
		Align(lipgloss.Center)

	views := make([]string, 0, len(m.Windows)*2)
	for _, w := range m.Windows {
		title := titleStyle.Render(w.Title)
		windowView := lipgloss.JoinVertical(lipgloss.Left,
			title,
			lipgloss.JoinVertical(lipgloss.Left, tail(w.lines, m.h-1)...))
		views = append(views,
			border,
			windowStyle.Render(windowView),
		)
	}
	views = append(views, border)

	return lipgloss.JoinHorizontal(lipgloss.Top, views...)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func tail(s []string, n int) []string {
	if len(s) <= n {
		return s
	}
	return s[len(s)-n-1:]
}
