package session

// TODO: macOS support.

import (
	"fmt"
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
	// osascript := `osascript -e 'tell application "System Events" to set picture of every desktop to "%s"'`
	// tell application "System Events" to set picture of desktop 1 to "/path/to/image.jpg"
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
	cmd := "system_profiler SPDisplaysDataType | grep Resolution | awk '{print $2}'"
	results, err := util.RunCmd(cmd)
	if err != nil {
		return nil, err
	}

	// Results in slice by splitting on newline
	lines := strings.Split(results, "\n")

	var displays []Display
	for _, line := range lines {
		displays = append(displays, Display{Name: line})
	}

	return displays, nil
}

func (m macos) GetCurrentWallpaper(display Display) (string, error) {
	return "", nil
}
