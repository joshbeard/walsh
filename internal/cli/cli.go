package cli

import (
	"fmt"
	"os"

	"github.com/charmbracelet/log"
	"github.com/joshbeard/walsh/internal/config"
	"github.com/joshbeard/walsh/internal/session"
	"github.com/joshbeard/walsh/internal/util"
	"github.com/spf13/cobra"
)

func Setup(cmd *cobra.Command, args []string) (string, *session.Session, error) {
	// Load config
	cfg, err := config.Load("")
	if err != nil {
		return "", nil, fmt.Errorf("error loading config: %w", err)
	}

	if _, _, err := setupLogger(cfg.LogLevel, cfg.LogFile); err != nil {
		return "", nil, fmt.Errorf("error setting up logger: %w", err)
	}

	// Create session
	sess, err := session.NewSession(cfg)
	if err != nil {
		return "", nil, fmt.Errorf("error creating session: %w", err)
	}

	displays := sess.Displays()
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

	return display, sess, nil
}

func setupLogger(level, file string) (*log.Logger, *os.File, error) {
	logH := os.Stderr
	if file != "" {
		var err error
		logH, err = os.OpenFile(file, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to open log file: %w", err)
		}
	}

	if level == "" {
		logH, err := os.Open(os.DevNull)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to open log file: %w", err)
		}

		logger := log.New(logH)
		log.SetDefault(logger)

		return logger, logH, nil
	}

	logger := log.New(logH)

	logLevel, err := log.ParseLevel(level)
	if err != nil {
		return nil, nil, fmt.Errorf("could not parse log level: %w", err)
	}

	logger.SetPrefix("walsh")
	logger.SetOutput(logH)
	logger.SetReportCaller(true)
	logger.SetReportTimestamp(true)
	logger.SetLevel(logLevel)

	log.SetLevel(logLevel)
	log.SetDefault(logger)

	return logger, logH, nil
}
