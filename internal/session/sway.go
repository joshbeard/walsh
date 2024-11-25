package session

import (
	"fmt"

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

	return parseWaylandDisplays(result)
}

// GetCurrentWallpaper returns the current wallpaper for the specified display
// in a Hyprland session. This uses the `swww query` command to get the current
// wallpaper.
func (s sway) GetCurrentWallpaper(display, _ Display) (string, error) {
	return getSwwwWallpaper(display, Display{})
}
