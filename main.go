package main

import (
	"encoding/base64"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/log"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"golang.org/x/exp/rand"

	"github.com/joshbeard/walsh/cmd/blacklist"
	"github.com/joshbeard/walsh/cmd/download"
	"github.com/joshbeard/walsh/cmd/list"
	"github.com/joshbeard/walsh/cmd/run"
	"github.com/joshbeard/walsh/cmd/set"
	"github.com/joshbeard/walsh/cmd/view"
	"github.com/joshbeard/walsh/internal/config"
)

// Set at build time
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

var banner = []string{
	`ICAgICAgICAgICAgICAgICAgICAgICAgIGQ4YiAgICAgICAgICBkOGIKICAgICAgICAgICAgIC`,
	`AgICAgICAgICAgIDg4UCAgICAgICAgICA/ODgKICAgICAgICAgICAgICAgICAgICAgICAgZDg4`,
	`ICAgICAgICAgICAgODhiCj84OCAgIGQ4UCAgZDhQIGQ4ODhiOGIgIDg4OCAgIC5kODg4YiwgID`,
	`g4ODg4OGIKZDg4ICBkOFAnIGQ4UCdkOFAnID84OCAgPzg4ICAgPzhiLCAgICAgODhQIGA/OGIK`,
	`PzhiICw4OGIgLDg4JyA4OGIgICw4OGIgIDg4YiAgICBgPzhiICBkODggICA4OFAKYD84ODhQJz`,
	`g4OFAnICBgPzg4UCdgODhiICA4OGJgPzg4OFAnIGQ4OCcgICA4OGIK`,
}

func main() {
	if err := Command().Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func Command() *cobra.Command {
	var cfg config.Config
	rootCmd := &cobra.Command{
		Use:   "walsh",
		Short: "walsh is a tool for managing wallpapers",
		Long: renderBanner() +
			"\nwalsh is a tool for managing wallpapers",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {

			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			if v, _ := cmd.Flags().GetBool("version"); v {
				fmt.Println(renderVersion())
				os.Exit(0)
			}

			log.Debug("No subcommand provided, running 'set'")
			set.Run(cmd, args, &cfg)
		},
	}

	rootCmd.AddCommand(blacklist.Command())
	rootCmd.AddCommand(list.Command())
	rootCmd.AddCommand(set.Command())
	rootCmd.AddCommand(download.Command())
	rootCmd.AddCommand(view.Command())
	rootCmd.AddCommand(list.AddCommand())
	rootCmd.AddCommand(run.Command())

	rootCmd.PersistentFlags().StringVarP(&cfg.ConfigFile, "config", "c", "",
		"path to config file")
	rootCmd.PersistentFlags().StringVarP(&cfg.LogLevel, "log-level", "L", "",
		"log level (debug, info, warn, error)")
	rootCmd.PersistentFlags().StringVarP(&cfg.LogFile, "log-file", "F", "",
		"log file (default is stderr)")
	rootCmd.PersistentFlags().BoolP("version", "V", false,
		"print version information")

	set.SetFlags(rootCmd, &cfg)

	return rootCmd
}

func renderBanner() string {
	str := strings.Join(banner, "")
	decoded, err := base64.StdEncoding.DecodeString(str)
	if err != nil {
		return "walsh"
	}

	rndColor := randomColor()

	return rndColor(string(decoded))
}

func renderVersion() string {
	str := renderBanner()

	str += "\nVersion " + color.GreenString(version)
	str += " | Commit " + color.GreenString(commit)
	str += " | Date " + color.GreenString(date)
	str += "\n\nhttps://github.com/joshbeard/walsh\n"
	str += "Copyright (c) 2024 Josh Beard | 0BSD License"

	return str
}

func randomColor() func(a ...interface{}) string {
	rand.Seed(uint64(time.Now().UnixNano()))

	colors := []func(a ...interface{}) string{
		color.New(color.FgRed).SprintFunc(),
		color.New(color.FgGreen).SprintFunc(),
		color.New(color.FgYellow).SprintFunc(),
		color.New(color.FgBlue).SprintFunc(),
		color.New(color.FgMagenta).SprintFunc(),
		color.New(color.FgCyan).SprintFunc(),
		color.New(color.FgWhite).SprintFunc(),
		color.New(color.FgHiRed).SprintFunc(),
		color.New(color.FgHiGreen).SprintFunc(),
		color.New(color.FgHiYellow).SprintFunc(),
		color.New(color.FgHiBlue).SprintFunc(),
		color.New(color.FgHiMagenta).SprintFunc(),
		color.New(color.FgHiCyan).SprintFunc(),
		color.New(color.FgHiWhite).SprintFunc(),
	}

	return colors[rand.Intn(len(colors))]
}
