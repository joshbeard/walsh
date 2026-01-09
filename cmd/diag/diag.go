package diag

import (
	"fmt"
	"runtime"

	"github.com/charmbracelet/log"
	"github.com/joshbeard/walsh/internal/cli"
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "diag",
		Aliases: []string{"diagnostics", "info"},
		Short:   "display diagnostic information",
		Long:    "Display diagnostic information about displays and system configuration",
		Run: func(cmd *cobra.Command, args []string) {
			_, sess, err := cli.Setup(cmd, args)
			if err != nil {
				log.Fatal(err)
			}

			displays := sess.Displays()

			fmt.Println("╔════════════════════════════════════════════════════════╗")
			fmt.Println("║                  Walsh Diagnostics                     ║")
			fmt.Println("╚════════════════════════════════════════════════════════╝")
			fmt.Println()

			fmt.Printf("Operating System: %s\n", runtime.GOOS)
			fmt.Printf("Architecture:     %s\n", runtime.GOARCH)
			fmt.Println()

			fmt.Printf("Detected Displays: %d\n", len(displays))
			fmt.Println()

			if len(displays) == 0 {
				fmt.Println("⚠️  No displays detected")
				return
			}

			fmt.Println("┌────────┬──────────────┬─────────────────────────────────┐")
			fmt.Println("│ Index  │ Name         │ Current Wallpaper               │")
			fmt.Println("├────────┼──────────────┼─────────────────────────────────┤")

			for _, display := range displays {
				currentPath := display.Current.Path
				if currentPath == "" {
					currentPath = "(none)"
				}
				if len(currentPath) > 30 {
					currentPath = "..." + currentPath[len(currentPath)-27:]
				}
				fmt.Printf("│ %-6d │ %-12s │ %-31s │\n", display.Index, display.Name, currentPath)
			}

			fmt.Println("└────────┴──────────────┴─────────────────────────────────┘")
			fmt.Println()

			fmt.Println("Configuration:")
			fmt.Printf("  Config Dir:    %s\n", sess.Config().ListsDir)
			fmt.Printf("  Cache Dir:     %s\n", sess.Config().CacheDir)
			fmt.Printf("  Cache Size:    %d\n", sess.Config().CacheSize)
			fmt.Printf("  History Size:  %d\n", sess.Config().HistorySize)
			fmt.Printf("  Sources:       %d configured\n", len(sess.Config().Sources))
			fmt.Println()
		},
	}

	return cmd
}
