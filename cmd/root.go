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
	greps []string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "branch",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
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
	rootCmd.Flags().StringArrayVar(&greps, "grep", []string{}, "Filters to apply to the output")
	rootCmd.MarkFlagRequired("grep")
}

func rootCmdRun(cmd *cobra.Command, args []string) {
	type grepper struct {
		filter string
		c      chan string
	}

	greppers := make([]grepper, 0)
	for _, text := range greps {
		c := make(chan string)
		greppers = append(greppers, grepper{
			filter: text,
			c:      c,
		})
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
			shouldDefault := true
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

	cs := make([]<-chan string, 0)
	for _, grepper := range greppers {
		cs = append(cs, grepper.c)
	}
	cs = append(cs, defaultC)

	p := tea.NewProgram(
		view.New(cs),
		tea.WithAltScreen(), // use the full size of the terminal in its "alternate screen buffer"
		// tea.WithMouseCellMotion(), // turn on mouse support so we can track the mouse wheel
	)

	if _, err := p.Run(); err != nil {
		fmt.Println("could not run program:", err)
		os.Exit(1)
	}
}
