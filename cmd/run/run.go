package run

import (
	"fmt"

	"github.com/charmbracelet/log"
	"github.com/joshbeard/walsh/internal/cli"
	"github.com/joshbeard/walsh/internal/config"
	"github.com/joshbeard/walsh/internal/scheduler"
	"github.com/joshbeard/walsh/internal/session"
	"github.com/joshbeard/walsh/internal/util"
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	var cfg config.Config

	cmd := &cobra.Command{
		Use:   "run [flags] [sources...]",
		Short: "run the wallpaper manager",
		Run: func(cmd *cobra.Command, args []string) {
			Run(cmd, args, &cfg)
		},
	}

	setFlags(cmd, &cfg)

	return cmd
}

func setFlags(cmd *cobra.Command, opts *config.Config) {
	cmd.Flags().StringVarP(&opts.Display, "display", "d", "",
		"display to use for operations")
	cmd.Flags().StringVarP(&opts.List, "list", "l", "",
		"set wallpaper from list")
	cmd.Flags().BoolVarP(&opts.NoTrack, "no-track", "n", false,
		"do not track wallpaper")
	cmd.Flags().BoolVarP(&opts.IgnoreHistory, "ignore-history", "i", false,
		"ignore the history when selecting a random image")
	cmd.Flags().IntVarP(&opts.Interval, "interval", "t", 0,
		"set interval for changing wallpapers")
	cmd.Flags().BoolVarP(&opts.SystemTray.Enabled, "tray", "", false,
		"show the system tray")
}

func Run(cmd *cobra.Command, args []string, cfg *config.Config) {
	// Load config
	loaded, err := config.Load("")
	if err != nil {
		log.Fatal(fmt.Errorf("error loading config: %w", err))
	}

	// merge the CLI args with the loaded config. CLI args take precedence.
	cfg, err = loaded.Merge(cfg)
	if err != nil {
		log.Fatal(fmt.Errorf("error merging config: %w", err))
	}

	display, _, err := cli.Setup(cmd, args)
	if err != nil {
		log.Fatal(err)
	}

	sess := session.Current

	err = util.Retry(
		cfg.MaxRetries,
		cfg.RetryInterval,
		func() error {
			return sess.SetWallpaper(display)
		})
	if err != nil {
		log.Fatal(err)
	}

	// Initialize the scheduler
	task := scheduler.New(sess, display)
	err = task.Start()
	if err != nil {
		log.Fatal(err)
	}
}
