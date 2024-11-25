package cfg

import (
	"fmt"

	"github.com/charmbracelet/log"
	"github.com/joshbeard/walsh/internal/config"
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	return &cobra.Command{
		Use:   "cfg [flags]",
		Short: "output the default configuration to stdout",
		Run: func(cmd *cobra.Command, args []string) {
			cfg, err := config.DefaultYAML()
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println(cfg)
		},
	}
}
