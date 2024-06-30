package view

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Branch is a model that represents the entire view of the application.
type Branch struct {
	Windows                  []*Window
	height, width, eachWidth int
}

// Init initializes the view model.
func (m Branch) Init() tea.Cmd {
	cmds := make([]tea.Cmd, 0, len(m.Windows))
	for _, w := range m.Windows {
		cmds = append(cmds, w.Init())
	}
	return tea.Batch(cmds...)
}

// Update updates the model state based on the received message.
func (m Branch) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.height = msg.Height
		m.width = msg.Width
		m.eachWidth = (m.width - len(m.Windows) - 1) / len(m.Windows)
		cmds = append(cmds, func() tea.Msg {
			return windowSizeMsg{m.eachWidth, m.height}
		})

	case tea.KeyMsg:
		if msg.Type == tea.KeyCtrlC {
			return m, tea.Quit
		}
		if msg.Type == tea.KeyUp {
			cmds = append(cmds, func() tea.Msg {
				return scrollUpMsg{}
			})
		}
		if msg.Type == tea.KeyDown {
			cmds = append(cmds, func() tea.Msg {
				return scrollDownMsg{}
			})
		}
	}

	// Update all sub windows
	for i, w := range m.Windows {
		newWindow, cmd := w.Update(msg)
		m.Windows[i] = &newWindow
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// View renders the view model.
func (m Branch) View() string {
	border := lipgloss.NewStyle().Render(strings.Repeat("|\n", max(m.height-1, 0)) + "|")

	views := make([]string, 0, len(m.Windows)*2)
	for _, w := range m.Windows {
		views = append(views,
			border,
			w.View(),
		)
	}
	views = append(views, border)

	return lipgloss.JoinHorizontal(lipgloss.Top, views...)
}

// max returns the maximum of two integers.
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
