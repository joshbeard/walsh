package list

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/joshbeard/walsh/internal/cli"
	"github.com/joshbeard/walsh/internal/util"
	"github.com/spf13/cobra"
)

var (
	listName string
	index    int
)

func Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"l"},
		Short:   "manage wallpaper lists",
		Long:    "Manage wallpaper lists",
	}

	cmd.AddCommand(AddCommand())
	cmd.AddCommand(ViewCommand())
	cmd.AddCommand(lsCommand())
	cmd.AddCommand(editCommand())
	cmd.PersistentFlags().StringVarP(&listName, "list", "l", "", "list name")

	return cmd
}

func lsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "ls",
		Short:   "view wallpaper lists",
		Aliases: []string{"list", "show"},
		Example: "  List all lists:\n" +
			"    walsh list ls\n\n" +
			"  List wallpapers in a specific list:\n" +
			"    walsh list ls mylist",
		Run: func(cmd *cobra.Command, args []string) {
			// List name is provided using the flag or it's the last argument.
			if listName == "" && len(args) > 0 {
				listName = args[len(args)-1]
				args = args[:len(args)-1]
			}

			_, sess, err := cli.Setup(cmd, args)
			if err != nil {
				log.Fatal(err)
			}

			if listName == "" {
				path := sess.Config().ListsDir
				files, err := os.ReadDir(path)
				if err != nil {
					log.Fatal(err)
				}

				printBanner(fmt.Sprintf("Lists (%d)", len(files)))
				for _, file := range files {
					fmt.Println(strings.TrimSuffix(file.Name(), ".json"))
				}

				return
			}

			path := filepath.Join(sess.Config().ListsDir, listName+".json")
			list, err := sess.ReadList(path)
			if err != nil {
				log.Fatal(err)
			}

			printBanner(fmt.Sprintf("%s (%d)", listName, len(list)))
			for i, wp := range list {
				fmt.Printf("%d: %s\n", i+1, wp.Source)
			}
		},
	}

	return cmd
}

func AddCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "add",
		Aliases: []string{"a"},
		Short:   "add wallpaper to list",
		Long: "Add the current wallpaper on a specific display to a list.\n\n" +
			"Lists can be used for managing collections of wallpapers.\n\n",
		Example: "  walsh list add --display 0 --list mylist\n" +
			"  walsh list add 0 mylist\n" +
			"  walsh list add -d 0 -l mylist\n" +
			"  walsh add 0 mylist\n" +
			"  walsh a 0 nature",
		Run: func(cmd *cobra.Command, args []string) {
			// List name is provided using the flag or it's the last argument.
			if listName == "" {
				listName = args[len(args)-1]
				args = args[:len(args)-1]
			}

			displayArg, sess, err := cli.Setup(cmd, args)
			if err != nil {
				log.Fatal(err)
			}

			// Read current file
			currentFile, err := sess.ReadCurrent()
			if err != nil {
				log.Fatal(err)
			}

			// Get display's current wallpaper
			display, err := currentFile.Display(displayArg)
			if err != nil {
				log.Fatal(err)
			}

			path := filepath.Join(sess.Config().ListsDir, listName+".json")
			log.Infof("Adding %s to list %s", display.Current.Path, path)
			err = sess.WriteList(path, display.Current)
			if err != nil {
				log.Fatal(err)
			}
		},
	}

	cmd.PersistentFlags().StringVarP(&listName, "list", "l", "", "list name")
	return cmd
}

// open file in $EDITOR
func editCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "edit",
		Aliases: []string{"e"},
		Short:   "edit wallpaper list",
		Example: "  walsh list edit --list mylist\n" +
			"  walsh list edit mylist\n" +
			"  walsh l e mylist",
		Run: func(cmd *cobra.Command, args []string) {
			// List name is provided using the flag or it's the last argument.
			if listName == "" && len(args) > 0 {
				listName = args[len(args)-1]
				args = args[:len(args)-1]
			}

			if listName == "" {
				log.Fatal("list name is required")
			}

			_, sess, err := cli.Setup(cmd, args)
			if err != nil {
				log.Fatal(err)
			}

			editor := os.Getenv("EDITOR")
			if editor == "" {
				log.Fatal("EDITOR environment variable not set")
			}

			path := filepath.Join(sess.Config().ListsDir, listName+".json")
			edit := exec.Command(editor, path)
			edit.Stdout = os.Stdout
			edit.Stdin = os.Stdin
			edit.Stderr = os.Stderr

			if err := edit.Run(); err != nil {
				log.Fatal(err)
			}
		},
	}

	return cmd
}

func ViewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "view",
		Aliases: []string{"v"},
		Short:   "view wallpaper in list",
		Example: "  walsh list view --list mylist --index 1\n" +
			"  walsh l v mylist 1",
		Run: func(cmd *cobra.Command, args []string) {
			// List name is provided using the flag or it's the last argument.
			if listName == "" {
				listName = args[len(args)-1]
				args = args[:len(args)-1]
			}

			_, sess, err := cli.Setup(cmd, args)
			if err != nil {
				log.Fatal(err)
			}

			path := filepath.Join(sess.Config().ListsDir, listName+".json")
			list, err := sess.ReadList(path)
			if err != nil {
				log.Fatal(err)
			}

			index, err := cmd.Flags().GetInt("index")
			if err != nil {
				log.Fatal(err)
			}

			if index < 1 || index > len(list) {
				log.Fatal("index out of range")
			}

			selected := list[index-1]
			log.Infof("Viewing %s", selected.Source)
			// Check if the image is already cached
			if selected.Path != "" {
				if util.FileExists(selected.Path) {
					if err = sess.View(selected.Path); err != nil {
						log.Fatal(err)
					}
					return
				}
			}

			fmt.Println("Image not found locally, downloading...")
			// TODO: Download the image
		},
	}

	cmd.Flags().IntVarP(&index, "index", "i", 0, "index of wallpaper in list")

	return cmd
}

func printBanner(text string) {
	border := "═"
	cornerTL := "╔"
	cornerTR := "╗"
	cornerBL := "╚"
	cornerBR := "╝"
	vertical := "║"

	width := len(text) + 4

	fmt.Println(cornerTL + strings.Repeat(border, width) + cornerTR)
	fmt.Println(vertical + "  " + text + "  " + vertical)
	fmt.Println(cornerBL + strings.Repeat(border, width) + cornerBR)
}
