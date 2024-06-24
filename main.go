package main

import (
	"encoding/base64"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/fatih/color"
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

var banner = []string{
	`ICAgICAgICAgICAgICAgICAgICAgICAgICBkOGIgICAgICAgICAgZDhiCiAgICAgICAgICAgICA`,
	`gICAgICAgICAgICAgODhQICAgICAgICAgID84OAogICAgICAgICAgICAgICAgICAgICAgICAgZD`,
	`g4ICAgICAgICAgICAgODhiCiA/ODggICBkOFAgIGQ4UCBkODg4YjhiICA4ODggICAuZDg4OGIsI`,
	`CA4ODg4ODhiCiBkODggIGQ4UCcgZDhQJ2Q4UCcgPzg4ICA/ODggICA/OGIsICAgICA4OFAgYD84`,
	`YgogPzhiICw4OGIgLDg4JyA4OGIgICw4OGIgIDg4YiAgICBgPzhiICBkODggICA4OFAKIGA/ODg`,
	`4UCc4ODhQJyAgYD84OFAnYDg4YiAgODhiYD84ODhQJyBkODgnICAgODhiCg==`,
}

func Command() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "walsh",
		Short: "walsh is a tool for managing wallpapers",
		Long: renderBanner() +
			"\nwalsh is a tool for managing wallpapers",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			logLevel, _ := cmd.Flags().GetString("log-level")
			logFile, _ := cmd.Flags().GetString("log-file")
			if _, _, err := setupLogger(logLevel, logFile); err != nil {
				return err
			}

			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			if v, _ := cmd.Flags().GetBool("version"); v {
				fmt.Println(renderVersion())
				os.Exit(0)
			}

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
	rootCmd.AddCommand(download.Command())
	rootCmd.AddCommand(view.Command())
	rootCmd.AddCommand(list.AddCommand())

	rootCmd.PersistentFlags().StringP("config", "c", "", "path to config file")
	rootCmd.PersistentFlags().StringP("display", "d", "",
		"display to use for operations")
	rootCmd.PersistentFlags().StringP("log-level", "L", "info",
		"log level (debug, info, warn, error)")
	rootCmd.PersistentFlags().StringP("log-file", "", "",
		"log file (default is stderr)")
	rootCmd.PersistentFlags().BoolP("version", "V", false, "print version information")

	return rootCmd
}

func main() {
	if err := Command().Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func renderBanner() string {
	str := strings.Join(banner, "")
	decoded, err := base64.StdEncoding.DecodeString(str)
	if err != nil {
		return "walsh"
	}

	return color.RedString(string(decoded))
}

func renderVersion() string {
	str := renderBanner()

	str += "\nVersion " + color.GreenString(version)
	str += " | Commit " + color.GreenString(commit)
	str += " | Date " + color.GreenString(date)
	str += "\n\nhttps://github.com/joshbeard/walsh\n"
	str += "Copyright (c) 2024 Josh Beard\n"
	str += "0BSD License"

	return str
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
