package session

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/joshbeard/walsh/internal/util"
)

type hyprland struct{}

var _ SessionProvider = &hyprland{}

var defaultWLRootsSetCmds = []string{
	`swww img '{{path}}' --outputs '{{display}}'`,
	// `hyprctl hyprpaper wallpaper "{{display}},{{path}}"`,
	// `swaybg -i '{{path}}' --output '{{display}}'`,
}

// SetWallpaper sets the wallpaper for the specified display in a Hyprland
// session.
func (h hyprland) SetWallpaper(path string, display Display) error {
	cmd, err := getSetCmd(defaultWLRootsSetCmds, path, display.Name)
	if err != nil {
		return fmt.Errorf("error setting wallpaper: %w", err)
	}

	if _, err = util.RunCmd(cmd); err != nil {
		return fmt.Errorf("error setting wallpaper: %w", err)
	}

	return nil
}

// GetDisplays returns a list of displays in a Hyprland session.
// This uses the `hyprctl monitors` command to get a list of displays.
func (h hyprland) GetDisplays() ([]Display, error) {
	result, err := util.RunCmd(`hyprctl monitors`)
	if err != nil {
		return nil, fmt.Errorf("failed to run hyprctl monitors: %w", err)
	}

	return parseDisplays(result)
}

// GetCurrentWallpaper returns the current wallpaper for the specified display
// in a Hyprland session. This uses the `swww query` command to get the current
// wallpaper.
func (h hyprland) GetCurrentWallpaper(display Display) (string, error) {
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
		return "", fmt.Errorf("no image found for display %s", display.Name)
	}

	return strings.TrimSpace(parts[1]), nil
}

// parseDisplays parses the output of `hyprctl monitors` and returns a list of
// displays in their struct form.
func parseDisplays(output string) ([]Display, error) {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	var displays []Display
	re := regexp.MustCompile(`^Monitor (\S+) \(ID (\d+)\):`)

	for i, line := range lines {
		if matches := re.FindStringSubmatch(line); matches != nil {
			displays = append(displays, Display{
				Index: i,
				Name:  matches[1],
			})
		}
	}

	if len(displays) == 0 {
		return nil, fmt.Errorf("no displays found")
	}

	log.Debugf("Displays: %+v", displays)
	return displays, nil
}

// findDisplayLine finds the line in the output that contains the display name
// in a Hyprland session when using `hyprctl monitors`.
func findDisplayLine(output, displayName string) (string, error) {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, displayName) {
			return line, nil
		}
	}
	return "", fmt.Errorf("no wallpaper found for display %s", displayName)
}
