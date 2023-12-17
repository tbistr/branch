package view

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type TextViewModel struct {
	inputs      []<-chan string
	lss         [][]string
	h, w, eachW int
}

func New(inputs []<-chan string) TextViewModel {
	m := TextViewModel{inputs: inputs}
	m.lss = make([][]string, len(m.inputs))
	return m
}

type initMsg struct{}

func (m TextViewModel) Init() tea.Cmd {
	return func() tea.Msg {
		return initMsg{}
	}
}

type scanMsg struct {
	index int
	text  string
}

func cmdRead(i int, input <-chan string) tea.Cmd {
	return func() tea.Msg {
		text := <-input
		return scanMsg{
			index: i,
			text:  text,
		}
	}
}

func (m TextViewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case initMsg:
		for i, input := range m.inputs {
			cmds = append(cmds, cmdRead(i, input))
		}

	case scanMsg:
		m.pushLine(msg.index, msg.text)
		cmd = cmdRead(msg.index, m.inputs[msg.index])

	case tea.WindowSizeMsg:
		m.h = msg.Height
		m.w = msg.Width
		m.eachW = (m.w - len(m.inputs) - 1) / len(m.inputs)
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
	views := make([]string, 0, len(m.lss)*2)
	for _, ls := range m.lss {
		views = append(views,
			border,
			windowStyle.Render(lipgloss.JoinVertical(lipgloss.Left, tail(ls, m.h)...)),
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
	return s[len(s)-n:]
}

func (m *TextViewModel) pushLine(i int, line string) {
	ls := m.lss[i]
	if len(ls) > 100 {
		ls = ls[1:]
	}
	ls = append(ls, line)

	m.lss[i] = ls
}
