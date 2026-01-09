package cli

import (
	"fmt"

	"github.com/charmbracelet/log"
	"github.com/joshbeard/walsh/internal/config"
	"github.com/joshbeard/walsh/internal/session"
	"github.com/spf13/cobra"
)

func Setup(cmd *cobra.Command, args []string) (string, *session.Session, error) {
	cfg, err := config.Load("")
	if err != nil {
		return "", nil, fmt.Errorf("error loading config: %w", err)
	}

	sess, err := session.NewSession(cfg)
	if err != nil {
		return "", nil, fmt.Errorf("error creating session: %w", err)
	}

	displays := sess.Displays()
	log.Debugf("Displays: %v", displays)
	display, _ := cmd.Flags().GetString("display")
	if len(args) > 0 && display == "" {
		// Try to resolve the first argument as a display using the optimized GetDisplay method
		_, _, err := sess.GetDisplay(args[0])
		if err == nil {
			// Successfully resolved as a display, use it
			display = args[0]
		}
	}

	// Don't default to any display - if display is empty, SetWallpaper will process all displays
	log.Debugf("displays: %+v; display: %+v", displays, display)

	log.Debugf("Returning displays: %+v; sess: %+v", display, sess)

	return display, sess, nil
}
