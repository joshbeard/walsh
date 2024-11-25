// Package menu provides an interactive menu for use with rofi/wofi/dmenu/etc.
package menu

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/joshbeard/walsh/cmd/blacklist"
	"github.com/joshbeard/walsh/internal/session"
	"github.com/joshbeard/walsh/internal/util"
)

func Start() string {
	rootMenu := []string{
		"Change Wallpaper...",
		"View Wallpaper...",
		"Blacklist Wallpaper...",
	}

	// TODO: Add to list
	// TODO: Set from list

	choice := runMenu(rootMenu)
	handleFuzzyMenuChoice(choice)

	return strings.Join(rootMenu, "\n")
}

func handleFuzzyMenuChoice(choice string) {
	switch choice {
	case "Change Wallpaper...":
		handleDisplaySelection(true, session.SetWallpaper)
	case "View Wallpaper...":
		handleDisplaySelection(false, func(displayID string) error {
			current, err := session.GetCurrentWallpaper(displayID)
			if err != nil {
				return err
			}
			return session.View(current)
		})
	case "Blacklist Wallpaper...":
		handleDisplaySelection(false, func(displayID string) error {
			return blacklist.Blacklist(displayID)
		})
	default:
		log.Infof("unknown choice: %s", choice)
	}
}

func handleDisplaySelection(all bool, action func(string) error) {
	displays := session.TargetDisplays()
	if len(displays) > 1 {
		choices := showDisplayChoices(all)
		displayChoice := runMenu(choices)

		displayID := ""
		if displayChoice != "All Displays" {
			displayID = strings.Split(displayChoice, ":")[0]
		}
		if err := action(displayID); err != nil {
			log.Fatal(err)
		}
	} else {
		if err := action(displays[0].ID); err != nil {
			log.Fatal(err)
		}
	}
}

func runMenu(options []string) string {
	menuStr := strings.Join(options, "\n")
	menuCommand := fmt.Sprintf("echo \"%s\" | %s", menuStr, session.Config().DmenuCommand)
	result, err := util.RunCmd(menuCommand)
	if err != nil {
		log.Fatal(err)
	}

	return strings.TrimSpace(result)
}

func showDisplayChoices(all bool) []string {
	displays := session.TargetDisplays()
	choices := make([]string, 0, len(displays))
	if all {
		choices = append(choices, "All Displays")
	}

	for _, d := range displays {
		label := d.Label
		if label == "" {
			label = d.Name
		}
		choices = append(choices, fmt.Sprintf("%s: %s", d.ID, label))
	}

	return choices
}
