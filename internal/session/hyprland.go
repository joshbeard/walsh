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

// SetWallpaper sets the wallpaper for the specified display in a Hyprland
// session.
func (h hyprland) SetWallpaper(path string, display Display) error {
	return setWaylandWallpaper(path, display)
}

// GetDisplays returns a list of displays in a Hyprland session.
// This uses the `hyprctl monitors` command to get a list of displays.
func (h hyprland) GetDisplays() ([]Display, error) {
	result, err := util.RunCmd(`hyprctl monitors`)
	if err != nil {
		return nil, fmt.Errorf("failed to run hyprctl monitors: %w", err)
	}

	return h.parseDisplays(result)
}

// GetCurrentWallpaper returns the current wallpaper for the specified display
// in a Hyprland session. This uses the `swww query` command to get the current
// wallpaper.
func (h hyprland) GetCurrentWallpaper(display, _ Display) (string, error) {
	return getSwwwWallpaper(display, Display{})
}

// parseDisplays parses the output of `hyprctl monitors` and returns a list of
// displays in their struct form.
func (h hyprland) parseDisplays(output string) ([]Display, error) {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	var displays []Display
	re := regexp.MustCompile(`^Monitor (\S+) \(ID (\d+)\):`)

	idx := 0
	for _, line := range lines {
		if matches := re.FindStringSubmatch(line); matches != nil {
			displays = append(displays, Display{
				Index: idx,
				Name:  matches[1],
			})

			idx++
		}
	}

	if len(displays) == 0 {
		return nil, fmt.Errorf("no displays found")
	}

	log.Debugf("Displays: %+v", displays)
	return displays, nil
}
