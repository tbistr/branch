package cmd

import (
	"fmt"
	"io"
	"os"
	"os/exec"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/creack/pty"
	"github.com/spf13/cobra"
	"github.com/tbistr/branch/view"
)

var (
	// Used for flags.
	dumpDefault bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "branch",
	Short: "branch is a command line demultiplexer with multiple output windows",
	Long: `branch is a command line demultiplexer with multiple output windows.
It reads from stdin and sends it to multiple commands.
Each command is displayed in a separate window.

Example:
  tail -f /var/log/syslog | branch 'grep -i error' 'grep -i warn' 'grep -i fail'
`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: rootCmdRun,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// --default
	rootCmd.Flags().BoolVarP(&dumpDefault, "default", "d", false, "Dump lines that do not match any filter")
}

func rootCmdRun(cmd *cobra.Command, args []string) {
	// check if pipe is connected
	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) != 0 {
		fmt.Println("The command is intended to work with pipes.")
		return
	}

	type cmder struct {
		cmd   *exec.Cmd
		title string
	}

	cmders := make([]cmder, 0, len(args))
	for _, arg := range args {
		cmders = append(cmders, cmder{exec.Command("sh", "-c", arg), arg})
	}
	if dumpDefault {
		cmders = append(cmders, cmder{exec.Command("sh", "-c", "cat"), "default"})
	}

	ws := make([]*view.Window, 0, len(cmders))
	stdinWriters := make([]io.WriteCloser, 0, len(cmders))
	for _, c := range cmders {
		ptyFD, _ := pty.Start(c.cmd)
		defer ptyFD.Close()
		echoOff(ptyFD)

		stdinWriters = append(stdinWriters, ptyFD)
		ws = append(ws, view.NewWindow(c.title, ptyFD))
	}

	// stdin multiplexer
	go func() {
		noClosers := make([]io.Writer, 0, len(stdinWriters))
		for _, w := range stdinWriters {
			noClosers = append(noClosers, w)
		}
		io.Copy(io.MultiWriter(noClosers...), os.Stdin)
	}()

	p := tea.NewProgram(
		view.Branch{Windows: ws},
		tea.WithAltScreen(),
		// tea.WithMouseCellMotion(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Println("could not run program:", err)
		os.Exit(1)
	}
	for _, c := range cmders {
		c.cmd.Process.Kill()
	}
}
