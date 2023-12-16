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
	Run: func(cmd *cobra.Command, args []string) {

		filters := []string{"python", "x86", "ssh"}

		type grepper struct {
			filter string
			c      chan string
		}

		greppers := make([]grepper, 0)

		for _, text := range filters {
			c := make(chan string)
			greppers = append(greppers, grepper{
				filter: text,
				c:      c,
			})
		}
		defaultC := make(chan string)

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
	},
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
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.branch.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}