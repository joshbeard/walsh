package session

// TODO: macOS support.

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/joshbeard/walsh/internal/util"
)

type macos struct{}

var _ SessionProvider = &macos{}

// isMacOS checks if the current session is macOS.
func isMacOS() bool {
	_, err := os.Stat("/System/Library/CoreServices/SystemVersion.plist")
	return err == nil
}

func (m macos) SetWallpaper(path string, display Display) error {
	osascript := fmt.Sprintf(
		`osascript -e 'tell application "System Events" to set picture of desktop %s to "%s"'`,
		display.Name,
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
	for i, display := range spDisplays {
		displayMap, ok := display.(map[string]interface{})
		if !ok {
			log.Fatalf("Error asserting display as map")
		}
		ndrvs, ok := displayMap["spdisplays_ndrvs"].([]interface{})
		if !ok {
			log.Fatalf("Error asserting spdisplays_ndrvs as array")
		}
		fmt.Printf("Display %d has %d items in 'spdisplays_ndrvs'\n", i+1, len(ndrvs))

		for ii := range ndrvs {
			displays = append(displays, Display{Name: fmt.Sprintf("%d", ii+1)})
		}
	}

	return displays, nil
}

func (m macos) GetCurrentWallpaper(display, _ Display) (string, error) {
	// tell application "System Events" to get picture of desktop 2
	osascript := fmt.Sprintf(
		`osascript -e 'tell application "System Events" to get picture of desktop %s'`,
		display.Name,
	)

	results, err := util.RunCmd(osascript)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(results), nil
}
