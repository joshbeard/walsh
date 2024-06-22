package session

// TODO: xorg support

import (
	"strings"

	"github.com/joshbeard/walsh/internal/util"
)

type xorg struct{}

var _ SessionProvider = &xorg{}

var defaultXorgSetCmds = []string{
	`nitrogen --head={{display}} --set-zoom-fill -- '{{path}}'`,
	`feh --bg-fill --no-xinerama --display {{display}} '{{path}}'`,
	`xwallpaper --output {{display}} --zoom '{{path}}'`,
	`xsetbg -display {{display}} '{{path}}'`,
}

func (x xorg) SetWallpaper(path string, display Display) error {
	cmd, err := getSetCmd(defaultXorgSetCmds, path, display.Name)
	if err != nil {
		return err
	}

	_, err = util.RunCmd(cmd)
	if err != nil {
		return err
	}

	return nil
}

func (x xorg) GetDisplays() ([]Display, error) {
	cmd := `xrandr --listactivemonitors | grep "^ " | awk '{print $1}'`
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

func (x xorg) GetCurrentWallpaper(display Display) (string, error) {
	return "", nil
}
