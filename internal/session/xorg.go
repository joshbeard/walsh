package session

// TODO: xorg support

import (
	"fmt"
	"strings"

	"github.com/joshbeard/walsh/internal/config"
	"github.com/joshbeard/walsh/internal/util"
)

type xorg struct {
	cfg *config.Config
}

var _ SessionProvider = &xorg{}

var defaultXorgSetCmds = []string{
	`nitrogen --head={{display}} --set-zoom-fill -- '{{path}}'`,
	`feh --bg-fill --no-xinerama --display {{display}} '{{path}}'`,
	`xwallpaper --output {{display}} --zoom '{{path}}'`,
	`xsetbg -display {{display}} '{{path}}'`,
}

func NewXorg(cfg *config.Config) SessionProvider {
	return &xorg{cfg: cfg}
}

func (x xorg) SetWallpaper(path string, display Display) error {
	if x.cfg.SetCommand != "" {
		cmd := parseSetCmd(x.cfg.SetCommand, path, display.ID)
		_, err := util.RunCmd(cmd)
		if err != nil {
			return err
		}

		return nil
	}

	cmd, err := getSetCmd(defaultXorgSetCmds, path, display.ID)
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
	return getXrandrDisplays()
}

func (x xorg) GetCurrentWallpaper(display, current Display) (string, error) {
	return current.Current.Path, nil
}

func getXrandrDisplays() ([]Display, error) {
	cmd := `xrandr --listactivemonitors | grep "^ " | awk '{print $1}' | cut -d':' -f1`
	output, err := util.RunCmd(cmd)
	if err != nil {
		return nil, err
	}
	// Trim any leading/trailing whitespace
	output = strings.TrimSpace(output)

	lines := strings.Split(output, "\n")

	displays := make([]Display, 0, len(lines))
	for i, line := range lines {
		if line == "" {
			continue
		}

		displays = append(displays, Display{
			Index: i,
			ID:    fmt.Sprintf("%d", i),
			Name:  line,
		})
	}

	return displays, nil
}
