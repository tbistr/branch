package view

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	LINE_BUFFER_CAP = 10000
)

// Window represents a single window in the view.
type Window struct {
	Title         string
	id            int // id of the window (used to avoid channel conflicts)
	Input         <-chan string
	buf           []string // line buffer
	height, width int
}

var idIota int = 0

// NewWindow creates a new window model.
func NewWindow(title string, input <-chan string) *Window {
	idIota += 1
	return &Window{
		Title: title,
		id:    idIota,
		Input: input,
		buf:   make([]string, 0, LINE_BUFFER_CAP),
	}
}

// pushLine adds a line to the buffer. If the buffer is full, it removes the oldest line.
func (w *Window) pushLine(line string) {
	if len(w.buf) >= LINE_BUFFER_CAP {
		w.buf = w.buf[1:]
	}
	w.buf = append(w.buf, line)
}

// Init initializes the window model.
func (w Window) Init() tea.Cmd {
	return cmdRead(w.id, w.Input)
}

// scanMsg represents a message from the input channel.
type scanMsg struct {
	id   int // which window the message is for
	text string
}

// cmdRead returns a command that waits for input from the channel and sends a message.
func cmdRead(id int, input <-chan string) tea.Cmd {
	return func() tea.Msg {
		text := <-input
		return scanMsg{id, text}
	}
}

// windowSizeMsg is the tea.WindowSizeMsg for the child window.
type windowSizeMsg struct {
	Width, Height int
}

// Update updates the window state based on the received message.
func (w Window) Update(msg tea.Msg) (Window, tea.Cmd) {
	switch msg := msg.(type) {
	case scanMsg:
		if msg.id != w.id {
			return w, nil
		}
		w.pushLine(msg.text)
		return w, cmdRead(msg.id, w.Input)

	case windowSizeMsg:
		w.height = msg.Height
		w.width = msg.Width
	}
	return w, nil
}

// View renders the window model.
func (w Window) View() string {
	withWidth := lipgloss.NewStyle().Width(w.width)
	titleStyle := withWidth.MaxHeight(1).
		Bold(true).
		Reverse(true).
		Align(lipgloss.Center)

	contentStyle := withWidth.Height(w.height - 1).MaxHeight(w.height - 1)

	title := titleStyle.Render(w.Title)
	content := contentStyle.Render(lipgloss.JoinVertical(lipgloss.Left, tail(w.buf, w.height-1)...))

	return lipgloss.JoinVertical(lipgloss.Left, title, content)
}

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

// tail returns the last n elements of a slice.
func tail(s []string, n int) []string {
	if len(s) <= n {
		return s
	}
	return s[len(s)-n-1:]
}
