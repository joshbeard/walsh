package blacklist

import (
	"fmt"

	"github.com/charmbracelet/log"
	"github.com/joshbeard/walsh/internal/cli"
	"github.com/joshbeard/walsh/internal/session"
	"github.com/joshbeard/walsh/internal/source"
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	opts := struct {
		delete bool
	}{}

	cmd := &cobra.Command{
		Use:     "blacklist",
		Aliases: []string{"bl"},
		Short:   "blacklist wallpapers",
		Long: "Blacklist a wallpaper.\n\n" +
			"Add the current wallpaper on a specific display to the blacklist or " +
			"optionally provide a path to a specific image file to blacklist.\n\n" +
			"Blacklisted images will not be set as wallpapers.",
		Example: "  blacklist current wallpaper on display 0:\n" +
			"    walsh bl 0\n\n" +
			"  blacklist a specific image file:\n" +
			"    walsh bl path/to/image.jpg\n\n" +
			"  blacklist wallpaper and remove the file:\n" +
			"    walsh bl --rm 0",
		Run: func(cmd *cobra.Command, args []string) {
			displayArg, err := cli.Setup(cmd, args)
			if err != nil {
				log.Fatal(err)
			}

			if len(args) == 0 {
				log.Fatal("no display or image provided")
			}

			if err = Blacklist(displayArg); err != nil {
				log.Fatal(err)
			}
		},
	}

	cmd.Flags().BoolVarP(&opts.delete, "rm", "", false,
		"delete the image from the source")

	return cmd
}

func Blacklist(displayArg string) error {
	log.Infof("getting current image for display %s", displayArg)

	// Get current
	current, err := session.GetCurrentWallpaper(displayArg)
	if err != nil {
		log.Fatal(err)
	}

	// Write to blacklist
	img := source.Image{
		Path: current,
	}
	log.Warnf("blacklisting image %s", current)
	err = session.WriteList(session.Config().BlacklistFile, img)
	if err != nil {
		log.Fatal(err)
	}

	// Set new wallpaper
	err = session.SetWallpaper(displayArg)
	if err != nil {
		return fmt.Errorf("error setting wallpaper: %w", err)
	}
	// TODO: argument for deleting file

	return nil
}
