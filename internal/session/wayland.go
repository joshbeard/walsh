package session

import (
	"fmt"
	"strings"

	"github.com/joshbeard/walsh/internal/util"
)

var defaultWaylandSetCmds = []string{
	`swww img '{{path}}' --outputs '{{display}}'`,
	// `hyprctl hyprpaper wallpaper "{{display}},{{path}}"`,
	// `swaybg -i '{{path}}' --output '{{display}}'`,
}

// findDisplayLine finds the line in the `swww query` output that
// corresponds to the specified display.
func findDisplayLine(output, displayName string) (string, error) {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, displayName) {
			return line, nil
		}
	}
	return "", fmt.Errorf("no wallpaper found for display %s", displayName)
}

func getSwwwWallpaper(display, _ Display) (string, error) {
	result, err := util.RunCmd("swww query")
	if err != nil {
		return "", fmt.Errorf("failed to query swww: %w", err)
	}

	line, err := findDisplayLine(result, display.Name)
	if err != nil {
		return "", err
	}

	// Get the path from the string:
	// e.g. eDP-1: 1920x1200, scale: 1, currently displaying: image: /tmp/unsplash-zG8VFOg7wgo.jpg
	parts := strings.Split(line, "image: ")
	if len(parts) < 2 {
		return "", fmt.Errorf("no image found for display %s", display.ID)
	}

	return strings.TrimSpace(parts[1]), nil
}

func setWaylandWallpaper(path string, display Display, customCmd string) error {
	var err error
	cmd := ""
	if customCmd != "" {
		cmd = parseSetCmd(customCmd, path, display.Name)
	} else {
		cmd, err = getSetCmd(defaultWaylandSetCmds, path, display.Name)
		if err != nil {
			return fmt.Errorf("error getting wallpaper set command: %w", err)
		}
	}

	if _, err = util.RunCmd(cmd); err != nil {
		return fmt.Errorf("error setting wallpaper: %w", err)
	}

	return nil
}
