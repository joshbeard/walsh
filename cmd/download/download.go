package download

import (
	"fmt"

	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:     "download",
	Aliases: []string{"dl"},
	Short:   "download wallpapers",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("This is the download command")
	},
}
