package set

import (
	"fmt"

	"github.com/charmbracelet/log"
	"github.com/joshbeard/walsh/internal/cli"
	"github.com/joshbeard/walsh/internal/config"
	"github.com/joshbeard/walsh/internal/daemon"
	"github.com/joshbeard/walsh/internal/session"
	"github.com/joshbeard/walsh/internal/util"
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	var cfg config.Config

	cmd := &cobra.Command{
		Use:     "set [flags] [sources...]",
		Aliases: []string{"s"},
		Short:   "set wallpapers (default command)",
		Long: "Set a random wallpaper from the provided sources, from a list, or " +
			"directly from a file.\n\n" +
			"Wallpapers can be set once or at a regular interval and specific displays " +
			"can be targeted.",
		Example: "  walsh set -d 0\n" +
			"  walsh set -d 1 path/to/images\n" +
			"  walsh s 0\n" +
			"  walsh s 1 path/to/images\n" +
			"  walsh set --interval 60 -d 0\n" +
			"  walsh set --interval 60 -d 0 --tray\n" +
			"  walsh set --once (default behavior; override config file)",
		Run: func(cmd *cobra.Command, args []string) {
			Run(cmd, args, &cfg)
		},
	}

	SetFlags(cmd, &cfg)

	return cmd
}

func SetFlags(cmd *cobra.Command, opts *config.Config) {
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
	cmd.Flags().BoolVarP(&opts.ShowTray, "tray", "", false,
		"show the system tray")
	cmd.Flags().BoolVarP(&opts.Once, "once", "", false,
		"set wallpaper once and exit. This overrides the config file when interval is set")
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

	if cfg.ShowTray && cfg.Once {
		log.Debug("Tray is enabled, but 'once' is set. Disabling tray.")
	}

	if cfg.Interval > 0 && cfg.Once {
		log.Debug("Interval is set, but 'once' is set. Disabling interval.")
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

	if (cfg.Interval <= 0 && !cfg.ShowTray) || cfg.Once {
		log.Debug("Setting wallpaper once")
		return
	}

	// Initialize the daemon
	d := daemon.NewDaemon(sess, display)

	// Start the daemon
	err = d.Start()
	if err != nil {
		log.Fatal(err)
	}

	log.Info("Daemon started")
}
