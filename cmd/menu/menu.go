package menu

import (
	"github.com/charmbracelet/log"
	"github.com/joshbeard/walsh/internal/cli"
	"github.com/joshbeard/walsh/internal/menu"
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "menu",
		Aliases: []string{"m"},
		Short:   "run a rofi/dmenu/wofi menu",
		Example: "  walsh menu",
		Run: func(cmd *cobra.Command, args []string) {
			_, err := cli.Setup(cmd, args)
			if err != nil {
				log.Fatal(err)
			}

			menu.Start()

		},
	}

	return cmd
}
