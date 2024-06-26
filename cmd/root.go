package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"github.com/tbistr/branch/view"
)

var (
	// Used for flags.
	greps       []string
	dumpDefault bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "branch",
	Short: "branch is a grep-like tool with multiple output windows",
	Long: `branch is a grep-like tool with multiple output windows.
It is intended to be used with pipes.

For example, you can use it like this:
$ cat /var/log/syslog | branch --grep=error --grep=warning --default

The above command will show you 3 windows:
1. error
2. warning
3. default(= lines that do not match any filter)
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
	// --grep=hoge,huga
	rootCmd.Flags().StringArrayVar(&greps, "grep", []string{}, "Filters to apply to the output")
	rootCmd.MarkFlagRequired("grep")
	// --default
	rootCmd.Flags().BoolVarP(&dumpDefault, "default", "d", false, "Dump lines that do not match any filter")
}

func rootCmdRun(cmd *cobra.Command, args []string) {
	type grepper struct {
		filter string
		c      chan string
	}

	greppers := make([]grepper, 0)
	for _, filter := range greps {
		c := make(chan string)
		greppers = append(greppers, grepper{filter, c})
	}
	defaultC := make(chan string)

	// check if pipe is connected
	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) != 0 {
		fmt.Println("The command is intended to work with pipes.")
		return
	}

	stdinScanner := bufio.NewScanner(os.Stdin)
	go func() {
		for stdinScanner.Scan() {
			line := stdinScanner.Text()
			shouldDefault := dumpDefault
			for _, g := range greppers {
				if strings.Contains(line, g.filter) {
					g.c <- stdinScanner.Text()
					shouldDefault = false
				}
			}
			if shouldDefault {
				defaultC <- stdinScanner.Text()
			}
		}
	}()

	ws := make([]*view.Window, 0)
	for _, grepper := range greppers {
		ws = append(ws, view.NewWindow(fmt.Sprintf("grep by %s", grepper.filter), grepper.c))
	}
	if dumpDefault {
		ws = append(ws, view.NewWindow("default", defaultC))
	}

	p := tea.NewProgram(
		view.Branch{Windows: ws},
		tea.WithAltScreen(),
		// tea.WithMouseCellMotion(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Println("could not run program:", err)
		os.Exit(1)
	}
}
