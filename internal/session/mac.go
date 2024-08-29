package session

// TODO: macOS support.

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/joshbeard/walsh/internal/config"
	"github.com/joshbeard/walsh/internal/util"
)

type macos struct {
	cfg *config.Config
}

var _ SessionProvider = &macos{}

func NewMacOS(cfg *config.Config) SessionProvider {
	return &macos{cfg: cfg}
}

// isMacOS checks if the current session is macOS.
func isMacOS() bool {
	return util.FileExists("/System/Library/CoreServices/SystemVersion.plist")
}

func (m macos) SetWallpaper(path string, display Display) error {
	osascript := fmt.Sprintf(
		`osascript -e 'tell application "System Events" to set picture of desktop %s to "%s"'`,
		display.ID,
		path,
	)

	_, err := util.RunCmd(osascript)
	if err != nil {
		return err
	}

	return nil
}

func (m macos) GetDisplays() ([]Display, error) {
	cmd := "system_profiler SPDisplaysDataType -json"
	results, err := util.RunCmd(cmd)
	if err != nil {
		return nil, err
	}

	var data map[string]interface{}
	err = json.Unmarshal([]byte(results), &data)
	if err != nil {
		log.Fatalf("Error unmarshalling JSON: %v", err)
	}

	spDisplays, ok := data["SPDisplaysDataType"].([]interface{})
	if !ok {
		log.Fatalf("Error asserting SPDisplaysDataType as array")
	}

	var displays []Display
	for _, display := range spDisplays {
		displayMap, ok := display.(map[string]interface{})
		if !ok {
			log.Fatalf("Error asserting display as map")
		}

		ndrvs, ok := displayMap["spdisplays_ndrvs"].([]interface{})
		if !ok {
			log.Fatalf("Error asserting spdisplays_ndrvs as array")
		}

		for ii := range ndrvs {
			// name is from the _name
			// name := fmt.Sprintf("%d", ii+1)
			// name := displayMap["_name"].(string)
			name := ndrvs[ii].(map[string]interface{})["_name"].(string)
			id := ndrvs[ii].(map[string]interface{})["_spdisplays_displayID"].(string)
			displays = append(displays, Display{ID: id, Index: ii + 1, Name: name})
		}

		log.Debugf("Found %d displays", len(ndrvs))
	}

	return displays, nil
}

func (m macos) GetCurrentWallpaper(display, _ Display) (string, error) {
	osascript := fmt.Sprintf(
		`osascript -e 'tell application "System Events" to get picture of desktop %s'`,
		display.ID,
	)

	results, err := util.RunCmd(osascript)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(results), nil
}
