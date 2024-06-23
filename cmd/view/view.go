package view

import (
	"github.com/charmbracelet/log"
	"github.com/joshbeard/walsh/internal/cli"
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
			display, sess, err := cli.Setup(cmd, args)
			if err != nil {
				log.Fatal(err)
			}

			current, err := sess.GetCurrentWallpaper(display)
			if err != nil {
				log.Fatal(err)
			}

			if err = sess.View(current); err != nil {
				log.Fatal(err)
			}
		},
	}

	return cmd
}
