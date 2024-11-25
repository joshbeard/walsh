package set

import (
	"fmt"

	"github.com/charmbracelet/log"
	"github.com/joshbeard/walsh/internal/cli"
	"github.com/joshbeard/walsh/internal/config"
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
			"  walsh s 1 path/to/images",
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

	display, err := cli.Setup(cmd, args)
	if err != nil {
		log.Fatal(err)
	}

	err = util.Retry(
		cfg.MaxRetries,
		cfg.RetryInterval,
		func() error {
			return session.SetWallpaper(display)
		})
	if err != nil {
		log.Fatal(err)
	}
}
