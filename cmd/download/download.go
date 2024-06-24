package download

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/log"

	"github.com/joshbeard/walsh/internal/cli"
	"github.com/joshbeard/walsh/internal/session"
	"github.com/joshbeard/walsh/internal/source"
	"github.com/joshbeard/walsh/internal/util"
	"github.com/spf13/cobra"
)

type dlOptions struct {
	dest string
}

func Command() *cobra.Command {
	var opts dlOptions
	cmd := &cobra.Command{
		Use:     "download",
		Aliases: []string{"dl"},
		Short:   "download wallpapers",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Specify either 'bing' or 'unsplash' as a subcommand")
		},
	}

	cmd.AddCommand(BingCommand())
	cmd.AddCommand(UnsplashCommand(opts))
	cmd.PersistentFlags().StringP("dest", "t", "", "destination URI for downloaded wallpapers")

	return cmd
}

func BingCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "bing [gosimac args]",
		Aliases: []string{"b"},
		Short:   "download wallpapers from Bing",
		Run: func(cmd *cobra.Command, args []string) {
			sess, dest := commonSetup(cmd, args, dlOptions{})

			// Pass all remaining arguments to gosimac
			run := fmt.Sprintf(`gosimac bing %s`, strings.Join(args, " "))

			result, err := util.RunCmd(run)
			if err != nil {
				log.Fatal(err)
			}

			log.Debugf("Bing result: %s", result)

			processDownloads(sess, dest)
		},
	}

	return cmd
}

func UnsplashCommand(opts dlOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "unsplash [gosimac args]",
		Aliases: []string{"u"},
		Short:   "download wallpapers from Unsplash",
		Run: func(cmd *cobra.Command, args []string) {
			sess, dest := commonSetup(cmd, args, opts)

			// Pass all remaining arguments to gosimac
			run := fmt.Sprintf(`gosimac unsplash %s`, strings.Join(args, " "))

			result, err := util.RunCmd(run)
			if err != nil {
				log.Fatal(err)
			}

			log.Debugf("Unsplash result: %s", result)

			processDownloads(sess, dest)
		},
	}

	return cmd
}

func commonSetup(cmd *cobra.Command, args []string, opts dlOptions) (*session.Session, string) {
	// Ensure 'gosimac' is in the PATH
	if _, err := exec.LookPath("gosimac"); err != nil {
		log.Fatal("gosimac not found in PATH")
	}

	_, sess, err := cli.Setup(cmd, args)
	if err != nil {
		log.Fatal(err)
	}

	dest := sess.Config().DownloadDest
	if opts.dest != "" {
		dest = opts.dest
	}

	if dest == "" {
		log.Fatal("No destination specified")
	}

	return sess, dest
}

func processDownloads(sess *session.Session, dest string) {
	srcs := []string{"dir:///home/josh/Pictures/GoSiMac"}
	images, err := source.GetImages(srcs)
	if err != nil {
		log.Fatal(err)
	}

	log.Debugf("Downloaded images: %+v", images)

	log.Debugf("Filtering blacklisted images")
	blacklist, err := sess.ReadList(sess.Config().BlacklistFile)
	if err != nil {
		log.Fatal("Error reading blacklist: %s", err)
	}
	blacklisted := source.GetMatches(images, blacklist)

	for _, img := range images {
		for _, bl := range blacklisted {
			if img.ShaSum == bl.ShaSum {
				log.Debugf("Blacklisted: %s", img.Path)

				// Remove the blacklisted image from the list
				images = source.RemoveImage(images, img)

				// Delete the file
				if err := os.Remove(img.Path); err != nil {
					log.Errorf("Error removing blacklisted file: %s", err)
				}
			}
		}
	}

	// Move from srcs[0] to dest
	for _, img := range images {
		// Check if the file already exists in the destination (based on basename)
		// If it does, skip it
		if util.FileExists(filepath.Join(dest, filepath.Base(img.Path))) {
			log.Debugf("File already exists in destination: %s", img.Path)

			if err := os.Remove(img.Path); err != nil {
				log.Errorf("Error removing file: %s", err)
			}

			continue
		}

		log.Debugf("Moving %s to %s", img.Path, dest)
		target := filepath.Join(dest, filepath.Base(img.Path))
		if err := os.Rename(img.Path, target); err != nil {
			log.Errorf("Error moving file: %s", err)
		}
	}
}
