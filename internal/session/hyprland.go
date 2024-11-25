package session

import (
	"encoding/json"
	"fmt"

	"github.com/charmbracelet/log"
	"github.com/joshbeard/walsh/internal/config"
	"github.com/joshbeard/walsh/internal/util"
)

type hyprland struct {
	cfg *config.Config
}

var _ SessionProvider = &hyprland{}

func NewHyprland(cfg *config.Config) SessionProvider {
	return &hyprland{cfg: cfg}
}

// SetWallpaper sets the wallpaper for the specified display in a Hyprland
// session.
func (h hyprland) SetWallpaper(path string, display Display) error {
	return setWaylandWallpaper(path, display, h.cfg.SetCommand)
}

func (h hyprland) getInstance() (string, error) {
	instances, err := util.RunCmd(`hyprctl -j instances`)
	if err != nil {
		return "", fmt.Errorf("failed to run hyprctl -j instances: %w", err)
	}
	instancesJSON := []map[string]interface{}{}
	err = json.Unmarshal([]byte(instances), &instancesJSON)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal hyprctl -j instances: %w", err)
	}

	// Assume instance 0 for now
	instance := instancesJSON[0]
	instanceID := instance["instance"].(string)

	log.Debug("found hyprland instance", "instance", instanceID)

	return instanceID, nil
}

// GetDisplays returns a list of displays in a Hyprland session.
// This uses the `hyprctl monitors` command to get a list of displays.
func (h hyprland) GetDisplays() ([]Display, error) {
	instance, err := h.getInstance()
	if err != nil {
		return nil, fmt.Errorf("failed to get instance: %w", err)
	}

	monitorCmd := fmt.Sprintf("hyprctl -i %s -j monitors", instance)
	result, err := util.RunCmd(monitorCmd)
	if err != nil {
		return nil, fmt.Errorf("failed to run hyprctl monitors: %w", err)
	}

	return parseWaylandDisplays(result)
}

// GetCurrentWallpaper returns the current wallpaper for the specified display
// in a Hyprland session. This uses the `swww query` command to get the current
// wallpaper.
func (h hyprland) GetCurrentWallpaper(display, _ Display) (string, error) {
	return getSwwwWallpaper(display, Display{})
}
