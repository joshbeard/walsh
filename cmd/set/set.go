package set

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"fyne.io/systray"
	"github.com/charmbracelet/log"
	"github.com/joshbeard/walsh/internal/cli"
	"github.com/joshbeard/walsh/internal/config"
	"github.com/joshbeard/walsh/internal/tray"
	"github.com/spf13/cobra"
)

var cancelFunc context.CancelFunc

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
			Run(cmd, args, cfg)
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

func Run(cmd *cobra.Command, args []string, cfg config.Config) {
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

	ctx, cancel := context.WithCancel(context.Background())
	cancelFunc = cancel

	if (cfg.Interval <= 0 && !cfg.ShowTray) || cfg.Once {
		if err := setWallpaper(cmd, args, &cfg); err != nil {
			log.Fatalf("Error: %v", err)
		}

		return
	}

	go func() {
		if err := setWallpaperWithContext(ctx, cmd, args, &cfg); err != nil {
			log.Fatalf("Error: %v", err)
		}
	}()

	if cfg.ShowTray {
		systray.Run(tray.OnReady, quit)
	}

	// Handle OS signals for clean shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	select {
	case <-ctx.Done():
	case <-sigChan:
		quit()
	}
}

func quit() {
	if cancelFunc != nil {
		cancelFunc()
	}
	systray.Quit()
	log.Info("Exiting...")
}

func setWallpaperWithContext(ctx context.Context, cmd *cobra.Command, args []string, cfg *config.Config) error {
	for {
		select {
		case <-ctx.Done():
			log.Info("Stopping wallpaper setting due to context cancellation")
			return nil
		default:
			if err := setWallpaper(cmd, args, cfg); err != nil {
				return err
			}
		}
	}
}

func setWallpaper(cmd *cobra.Command, args []string, cfg *config.Config) error {
	maxRetries := 3                  // Maximum number of retries
	retryInterval := 2 * time.Second // Interval between retries

	retry := func(operation func() error) error {
		var err error
		for i := 0; i < maxRetries; i++ {
			err = operation()
			if err == nil {
				return nil
			}
			log.Errorf("Error encountered: %s. Retrying in %v...", err, retryInterval)
			time.Sleep(retryInterval)
		}
		return err
	}

	err := retry(func() error {
		display, sess, err := cli.Setup(cmd, args)
		if err != nil {
			return err
		}
		cfg.Display = display
		return sess.SetWallpaper(cfg.Sources, cfg.Display)
	})
	if err != nil {
		log.Fatal(err)
		return err
	}

	if cfg.Interval <= 0 || cfg.Once {
		return nil
	}

	ticker := time.NewTicker(time.Duration(cfg.Interval) * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		err := retry(func() error {
			display, sess, err := cli.Setup(cmd, args)
			if err != nil {
				return err
			}
			cfg.Display = display
			return sess.SetWallpaper(cfg.Sources, cfg.Display)
		})
		if err != nil {
			log.Fatal(err)
			return err
		}
		log.Infof("Next wallpaper change in %d seconds", cfg.Interval)
	}

	return nil
}
