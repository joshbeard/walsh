package set

import (
	"time"

	"github.com/charmbracelet/log"
	"github.com/joshbeard/walsh/internal/cli"
	"github.com/joshbeard/walsh/internal/session"
	"github.com/spf13/cobra"
)

type setOptions struct {
	list          string
	noTrack       bool
	ignoreHistory bool
	srcs          []string
	display       string
	interval      int
}

func Command() *cobra.Command {
	var opts setOptions

	cmd := &cobra.Command{
		Use:     "set [flags] [sources...]",
		Aliases: []string{"s"},
		Short:   "set wallpapers (default command)",
		Long: "Set a random wallpaper from the provided sources, from a list, or " +
			"directly from a file.\n\n",
		Example: "  walsh set -d 0\n" +
			"  walsh set -d 1 path/to/images\n" +
			"  walsh s 0\n" +
			"  walsh s 1 path/to/images\n" +
			"  walsh set --interval 60 -d 0",
		Run: func(cmd *cobra.Command, args []string) {
			display, sess, err := cli.Setup(cmd, args)
			if err != nil {
				log.Fatal(err)
			}
			opts.display = display

			err = setWallpaper(sess, opts)
			if err != nil {
				log.Fatalf("Error: %v", err)
			}
		},
	}

	cmd.Flags().StringVarP(&opts.list, "list", "l", "",
		"set wallpaper from list")
	cmd.Flags().BoolVarP(&opts.noTrack, "no-track", "n", false,
		"do not track wallpaper")
	cmd.Flags().BoolVarP(&opts.ignoreHistory, "ignore-history", "i", false,
		"ignore the history when selecting a random image")
	cmd.Flags().IntVarP(&opts.interval, "interval", "t", 0,
		"set interval for changing wallpapers")

	return cmd
}

func setWallpaper(sess *session.Session, opts setOptions) error {
	err := sess.SetWallpaper(opts.srcs, opts.display)
	if err != nil {
		log.Errorf("Error setting wallpaper: %s", err)
		return err
	}

	if opts.interval <= 0 {
		return nil
	}

	ticker := time.NewTicker(time.Duration(opts.interval) * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		if err := sess.SetWallpaper(opts.srcs, opts.display); err != nil {
			log.Errorf("Error setting wallpaper: %s", err)
			return err
		}
		log.Infof("Next wallpaper change in %d seconds", opts.interval)
	}

	return nil
}
