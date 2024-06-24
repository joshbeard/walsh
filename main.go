package main

import (
	"fmt"
	"os"

	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"

	"github.com/joshbeard/walsh/cmd/blacklist"
	"github.com/joshbeard/walsh/cmd/download"
	"github.com/joshbeard/walsh/cmd/list"
	"github.com/joshbeard/walsh/cmd/set"
	"github.com/joshbeard/walsh/cmd/view"
)

// Set at build time
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func Command() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "walsh",
		Short: "walsh is a tool for managing wallpapers",
		Version: fmt.Sprintf("%s, commit %s, built at %s",
			version, commit, date) +
			"\nhttps://github.com/joshbeard/walsh" +
			"\nCopyright (c) 2024 Josh Beard\n" +
			"0BSD License <https://spdx.org/licenses/0BSD.html>",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			logLevel, _ := cmd.Flags().GetString("log-level")
			logFile, _ := cmd.Flags().GetString("log-file")
			if _, _, err := setupLogger(logLevel, logFile); err != nil {
				return err
			}

			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			log.Debug("No subcommand provided, running 'set'")
			cmd.SetArgs([]string{"set"})
			if err := cmd.Execute(); err != nil {
				log.Fatal(err)
			}
		},
	}

	rootCmd.AddCommand(blacklist.Command())
	rootCmd.AddCommand(list.Command())
	rootCmd.AddCommand(set.Command())
	rootCmd.AddCommand(download.Cmd)
	rootCmd.AddCommand(view.Command())
	rootCmd.AddCommand(list.AddCommand())

	rootCmd.PersistentFlags().StringP("config", "c", "", "path to config file")
	rootCmd.PersistentFlags().StringP("display", "d", "",
		"display to use for operations")
	rootCmd.PersistentFlags().StringP("log-level", "L", "info",
		"log level (debug, info, warn, error)")
	rootCmd.PersistentFlags().StringP("log-file", "", "",
		"log file (default is stderr)")

	return rootCmd
}

func main() {
	if err := Command().Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func setupLogger(level, file string) (*log.Logger, *os.File, error) {
	logH := os.Stderr
	if file != "" {
		var err error
		logH, err = os.OpenFile(file, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
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
