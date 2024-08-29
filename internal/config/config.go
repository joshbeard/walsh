package config

import (
	"fmt"
	"strings"

	"github.com/adrg/xdg"
	"github.com/charmbracelet/log"
	"github.com/joshbeard/walsh/internal/util"
	"gopkg.in/yaml.v3"
)

const (
	MaxInterval = 31536000
	MinInterval = 10
)

type Config struct {
	Sources                 []string         `yaml:"sources"`
	ListsDir                string           `yaml:"lists_dir"`
	BlacklistFile           string           `yaml:"blacklist"`
	HistoryFile             string           `yaml:"history"`
	CurrentFile             string           `yaml:"current"`
	HistorySize             int              `yaml:"history_size"`
	CacheDir                string           `yaml:"cache_dir"`
	CacheSize               int              `yaml:"cache_size"`
	DownloadDest            string           `yaml:"download_dest"`
	DeleteBlacklistedImages bool             `yaml:"delete_blacklisted_images"`
	SetCommand              string           `yaml:"set_command"`
	ViewCommand             string           `yaml:"view_command"`
	LogLevel                string           `yaml:"log_level"`
	LogFile                 string           `yaml:"log_file"`
	ConfigFile              string           `yaml:"config_file"`
	List                    string           `yaml:"list"`
	NoTrack                 bool             `yaml:"no_track"`
	IgnoreHistory           bool             `yaml:"ignore_history"`
	Display                 string           `yaml:"display"`
	Interval                int              `yaml:"interval"`
	ShowTray                bool             `yaml:"enable_tray"`
	MenuIntervals           []RotateInterval `yaml:"menu_intervals"`
	Once                    bool             `yaml:"-"`
}

type RotateInterval int

func Load(path string) (*Config, error) {
	path, err := resolveFilePath(path)
	if err != nil {
		return nil, err
	}

	var cfg *Config
	if !util.FileExists(path) {
		log.Warnf("Creating new config file at %s", path)
		cfg, err = createNewConfig(path)
		if err != nil {
			return nil, err
		}

		return cfg, nil
	}

	cfgData, err := util.OpenFile(path)
	if err != nil {
		return nil, err
	}

	cfg, err = unmarshalConfig(cfgData)
	if err != nil {
		return nil, err
	}

	applyDefaults(cfg, defaultConfig())

	err = cfg.Validate()
	if err != nil {
		return nil, err
	}

	err = cfg.createDirs()
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

// Merge merges two Config structs, with the second argument taking precedence.
func (c Config) Merge(other Config) (Config, error) {
	merged := c

	// Merge simple fields
	if other.ListsDir != "" {
		merged.ListsDir = other.ListsDir
	}
	if other.BlacklistFile != "" {
		merged.BlacklistFile = other.BlacklistFile
	}
	if other.HistoryFile != "" {
		merged.HistoryFile = other.HistoryFile
	}
	if other.CurrentFile != "" {
		merged.CurrentFile = other.CurrentFile
	}
	if other.HistorySize != 0 {
		merged.HistorySize = other.HistorySize
	}
	if other.CacheDir != "" {
		merged.CacheDir = other.CacheDir
	}
	if other.CacheSize != 0 {
		merged.CacheSize = other.CacheSize
	}
	if other.DownloadDest != "" {
		merged.DownloadDest = other.DownloadDest
	}
	if other.Interval != 0 {
		merged.Interval = other.Interval
	}
	if other.SetCommand != "" {
		merged.SetCommand = other.SetCommand
	}
	if other.ViewCommand != "" {
		merged.ViewCommand = other.ViewCommand
	}
	if other.LogLevel != "" {
		merged.LogLevel = other.LogLevel
	}
	if other.LogFile != "" {
		merged.LogFile = other.LogFile
	}
	if other.ConfigFile != "" {
		merged.ConfigFile = other.ConfigFile
	}
	if other.List != "" {
		merged.List = other.List
	}
	if other.Once {
		merged.Once = other.Once
	}

	merged.NoTrack = merged.NoTrack || other.NoTrack
	merged.IgnoreHistory = merged.IgnoreHistory || other.IgnoreHistory
	if other.Display != "" {
		merged.Display = other.Display
	}
	merged.ShowTray = merged.ShowTray || other.ShowTray

	// Merge boolean fields
	merged.DeleteBlacklistedImages = merged.DeleteBlacklistedImages || other.DeleteBlacklistedImages

	// Merge slices
	if len(other.Sources) > 0 {
		merged.Sources = other.Sources
	}

	return merged, nil
}

func resolveFilePath(path string) (string, error) {
	var err error
	if path == "" {
		path, err = xdg.ConfigFile("walsh/config.yml")
		if err != nil {
			return "", err
		}
	}

	return path, nil
}

func unmarshalConfig(data []byte) (*Config, error) {
	cfg := &Config{}
	err := yaml.Unmarshal(data, cfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) marshalConfig() ([]byte, error) {
	return yaml.Marshal(c)
}

func createNewConfig(path string) (*Config, error) {
	cfg := defaultConfig()
	data, err := cfg.marshalConfig()
	if err != nil {
		return nil, err
	}

	if err = util.WriteFile(path, data); err != nil {
		return nil, err
	}

	if err = cfg.createDirs(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) createDirs() error {
	if !util.FileExists(xdg.DataHome) {
		err := util.MkDir(xdg.DataHome)
		if err != nil {
			return err
		}
	}

	if !util.FileExists(c.ListsDir) {
		err := util.MkDir(c.ListsDir)
		if err != nil {
			return err
		}
	}

	if !util.FileExists(c.CacheDir) {
		err := util.MkDir(c.CacheDir)
		if err != nil {
			return err
		}
	}

	// if !util.FileExists(c.DownloadDest) {
	// 	err := util.MkDir(c.DownloadDest)
	// 	if err != nil {
	// 		return err
	// 	}
	// }

	return nil
}

func defaultConfig() *Config {
	return &Config{
		BlacklistFile: xdg.ConfigHome + "/walsh/blacklist.json",
		CurrentFile:   xdg.DataHome + "/walsh/current.json",
		HistoryFile:   xdg.DataHome + "/walsh/history.json",
		ListsDir:      xdg.DataHome + "/walsh/lists",
		CacheDir:      xdg.CacheHome + "/walsh",
		DownloadDest:  xdg.Home + "/Pictures/Wallpapers",
		HistorySize:   50,
		CacheSize:     50,
		Interval:      0,
		Sources: []string{
			"dir://" + xdg.Home + "/Pictures/Wallpapers",
		},
		MenuIntervals: []RotateInterval{
			60,
			300,
			600,
			1200,
			1800,
			3600,
			7200,
			21600,
			43200,
			86400,
		},
	}
}

func applyDefaults(cfg, defaults *Config) {
	if cfg.HistorySize == 0 {
		cfg.HistorySize = defaults.HistorySize
	}

	if cfg.CacheDir == "" {
		cfg.CacheDir = defaults.CacheDir
	}

	if cfg.CacheSize == 0 {
		cfg.CacheSize = defaults.CacheSize
	}

	if cfg.DownloadDest == "" {
		cfg.DownloadDest = defaults.DownloadDest
	}

	if cfg.BlacklistFile == "" {
		cfg.BlacklistFile = defaults.BlacklistFile
	}

	if cfg.CurrentFile == "" {
		cfg.CurrentFile = defaults.CurrentFile
	}

	if cfg.HistoryFile == "" {
		cfg.HistoryFile = defaults.HistoryFile
	}

	if cfg.ListsDir == "" {
		cfg.ListsDir = defaults.ListsDir
	}

	if cfg.Sources == nil {
		cfg.Sources = defaults.Sources
	}

	if cfg.MenuIntervals == nil {
		cfg.MenuIntervals = defaults.MenuIntervals
	}
}

func (c Config) Validate() error {
	if c.Interval < 0 {
		return fmt.Errorf("interval must be greater than or equal to 0")
	}

	if c.Interval > MaxInterval {
		return fmt.Errorf("interval must be less than or equal to %d", MaxInterval)
	}

	if c.Interval < MinInterval && c.Interval != 0 {
		return fmt.Errorf("interval must be greater than or equal to %d", MinInterval)
	}

	for _, i := range c.MenuIntervals {
		if i < 0 {
			return fmt.Errorf("menu interval must be greater than or equal to 0")
		}

		if i > MaxInterval {
			return fmt.Errorf("menu interval must be less than or equal to %d", MaxInterval)
		}

		if i < MinInterval && i != 0 {
			return fmt.Errorf("menu interval must be greater than or equal to %d", MinInterval)
		}
	}

	return nil
}

func (r RotateInterval) InList(intervals []RotateInterval) bool {
	for _, i := range intervals {
		if i == r {
			return true
		}
	}

	return false
}

func (r RotateInterval) String() string {
	if r == 0 {
		return "Pause"
	}

	seconds := int(r)
	var result []string

	days := seconds / 86400
	if days > 0 {
		if days == 1 {
			result = append(result, "1 day")
		} else {
			result = append(result, fmt.Sprintf("%d days", days))
		}
		seconds %= 86400
	}

	hours := seconds / 3600
	if hours > 0 {
		if hours == 1 {
			result = append(result, "1 hour")
		} else {
			result = append(result, fmt.Sprintf("%d hours", hours))
		}
		seconds %= 3600
	}

	minutes := seconds / 60
	if minutes > 0 {
		if minutes == 1 {
			result = append(result, "1 minute")
		} else {
			result = append(result, fmt.Sprintf("%d minutes", minutes))
		}
		seconds %= 60
	}

	if seconds > 0 {
		if seconds == 1 {
			result = append(result, "1 second")
		} else {
			result = append(result, fmt.Sprintf("%d seconds", seconds))
		}
	}

	return strings.Join(result, " ")
}
