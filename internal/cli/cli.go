package cli

import (
	"fmt"

	"github.com/joshbeard/walsh/internal/config"
	"github.com/joshbeard/walsh/internal/logger"
	"github.com/joshbeard/walsh/internal/session"
	"github.com/joshbeard/walsh/internal/util"
	"github.com/spf13/cobra"
)

func Setup(cmd *cobra.Command, args []string) (string, error) {
	// Load config
	cfg, err := config.Load("")
	if err != nil {
		return "", fmt.Errorf("error loading config: %w", err)
	}

	if _, _, err := logger.Setup(cfg.LogLevel, cfg.LogFile); err != nil {
		return "", fmt.Errorf("error setting up logger: %w", err)
	}

	// Create session
	err = session.NewSession(cfg)
	if err != nil {
		return "", fmt.Errorf("error creating session: %w", err)
	}

	displays := session.Displays()
	display, _ := cmd.Flags().GetString("display")
	if len(args) > 0 && display == "" {
		// If the argument is a digit, assume it's a display. If it's a display
		// name, use it. Otherwise, assume it's a source.
		matchName := false
		for _, d := range displays {
			if d.Name == args[0] {
				matchName = true
				break
			}
		}

		if matchName || util.IsNumber(args[0]) {
			display = args[0]
		}
	}

	return display, nil
}
