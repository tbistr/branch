package view

import (
	"bufio"
	"fmt"
	"io"

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
	Input         io.Reader
	scanner       *bufio.Scanner
	buf           []string // line buffer
	viewB, viewE  int      // index of the first and last line to display
	height, width int
}

var idIota int = 0

// NewWindow creates a new window model.
func NewWindow(title string, input io.Reader) *Window {
	idIota += 1
	scanner := bufio.NewScanner(input)
	return &Window{
		Title:   title,
		id:      idIota,
		Input:   input,
		scanner: scanner,
		buf:     make([]string, 0, LINE_BUFFER_CAP),
	}
}

// scrollUp scrolls the window up by one line.
func (w *Window) scrollUp() {
	if w.viewE-w.viewB < w.height {
		w.viewB = 0
		w.viewE = len(w.buf)
	}
	if w.viewB > 0 {
		w.viewB -= 1
		w.viewE -= 1
	}
}

// scrollDown scrolls the window down by one line.
func (w *Window) scrollDown() {
	if w.viewE-w.viewB < w.height {
		w.viewB = 0
		w.viewE = len(w.buf)
	}
	if w.viewE < len(w.buf) {
		w.viewB += 1
		w.viewE += 1
	}
}

// pushLine adds a line to the buffer. If the buffer is full, it removes the oldest line.
func (w *Window) pushLine(line string) {
	if len(w.buf) >= LINE_BUFFER_CAP {
		w.buf = w.buf[1:]
	}
	w.buf = append(w.buf, line)
	w.scrollDown()
}

func (w *Window) changeSize(width, height int) {
	w.width = width
	w.height = height

	// Adjust the view to the new size
	if len(w.buf) < height {
		w.viewB = 0
		w.viewE = len(w.buf)
	} else {
		w.viewB = len(w.buf) - height
		w.viewE = len(w.buf)
	}
}

type scrollUpMsg struct{}

type scrollDownMsg struct{}

// scanMsg represents a message from the input channel.
type scanMsg struct {
	id   int // which window the message is for
	text string
}

// cmdRead returns a command that waits for input from the channel and sends a message.
func cmdRead(id int, scanner *bufio.Scanner) tea.Cmd {
	return func() tea.Msg {
		text := ""
		if scanner.Scan() {
			text = scanner.Text()
		}
		return scanMsg{id, text}
	}
}

// windowSizeMsg is the tea.WindowSizeMsg for the child window.
type windowSizeMsg struct {
	Width, Height int
}

// Init initializes the window model.
func (w Window) Init() tea.Cmd {
	return cmdRead(w.id, w.scanner)
}

// Update updates the window state based on the received message.
func (w Window) Update(msg tea.Msg) (Window, tea.Cmd) {
	switch msg := msg.(type) {
	case scrollUpMsg:
		w.scrollUp()
	case scrollDownMsg:
		w.scrollDown()

	case scanMsg:
		if msg.id != w.id {
			return w, nil
		}
		w.pushLine(msg.text)
		return w, cmdRead(msg.id, w.scanner)

	case windowSizeMsg:
		w.changeSize(msg.Width, msg.Height)
		w.pushLine(fmt.Sprintln("Resized to ", msg.Width, "x", msg.Height))
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
	content := contentStyle.Render(lipgloss.JoinVertical(lipgloss.Left, w.buf[w.viewB:w.viewE]...))

	return lipgloss.JoinVertical(lipgloss.Left, title, content)
}
