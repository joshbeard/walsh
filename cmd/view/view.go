package view

import (
	"github.com/charmbracelet/log"
	"github.com/joshbeard/walsh/internal/cli"
	"github.com/joshbeard/walsh/internal/session"
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "view [flags] [display]",
		Aliases: []string{"v"},
		Short:   "view wallpapers",
		Long:    "View the current wallpaper on a specific display.",
		Example: "  walsh view -d 0\n  walsh v 1",
		Run: func(cmd *cobra.Command, args []string) {
			display, err := cli.Setup(cmd, args)
			if err != nil {
				log.Fatal(err)
			}

			current, err := session.GetCurrentWallpaper(display)
			if err != nil {
				log.Fatal(err)
			}

			if err = session.View(current); err != nil {
				log.Fatal(err)
			}
		},
	}

	return cmd
}
