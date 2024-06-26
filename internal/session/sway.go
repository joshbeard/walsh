package session

import (
	"encoding/json"
	"fmt"

	"github.com/charmbracelet/log"
	"github.com/joshbeard/walsh/internal/config"
	"github.com/joshbeard/walsh/internal/util"
)

type sway struct {
	cfg *config.Config
}

var _ SessionProvider = sway{}

func NewSway(cfg *config.Config) SessionProvider {
	return sway{cfg: cfg}
}

// SetWallpaper sets the wallpaper for the specified display in a Hyprland
// session.
func (s sway) SetWallpaper(path string, display Display) error {
	return setWaylandWallpaper(path, display, s.cfg.SetCommand)
}

// GetDisplays returns a list of displays in a Hyprland session.
// This uses the `hyprctl monitors` command to get a list of displays.
func (s sway) GetDisplays() ([]Display, error) {
	result, err := util.RunCmd(`swaymsg -t get_outputs`)
	if err != nil {
		return nil, fmt.Errorf("failed to run sway monitors: %w", err)
	}

	return s.parseDisplays(result)
}

// GetCurrentWallpaper returns the current wallpaper for the specified display
// in a Hyprland session. This uses the `swww query` command to get the current
// wallpaper.
func (s sway) GetCurrentWallpaper(display, _ Display) (string, error) {
	return getSwwwWallpaper(display, Display{})
}

func (s sway) parseDisplays(result string) ([]Display, error) {
	jsonObj := []map[string]interface{}{}
	err := json.Unmarshal([]byte(result), &jsonObj)
	if err != nil {
		return nil, fmt.Errorf("failed to parse json: %w", err)
	}

	displays := make([]Display, 0, len(jsonObj))
	for i, value := range jsonObj {
		name := value["name"]

		log.Warnf("i: %d, name: %s", i, name)
		displays = append(displays, Display{
			Index: i,
			Name:  name.(string),
		})
	}

	log.Warnf("found %d displays: %+v", len(displays), displays)

	return displays, nil
}
