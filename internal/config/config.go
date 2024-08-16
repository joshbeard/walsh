package config

import (
	"github.com/adrg/xdg"
	"github.com/charmbracelet/log"
	"github.com/joshbeard/walsh/internal/util"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Sources                 []string `yaml:"sources"`
	ListsDir                string   `yaml:"lists_dir"`
	BlacklistFile           string   `yaml:"blacklist"`
	HistoryFile             string   `yaml:"history"`
	CurrentFile             string   `yaml:"current"`
	HistorySize             int      `yaml:"history_size"`
	CacheDir                string   `yaml:"cache_dir"`
	CacheSize               int      `yaml:"cache_size"`
	DownloadDest            string   `yaml:"download_dest"`
	Interval                int      `yaml:"interval"`
	DeleteBlacklistedImages bool     `yaml:"delete_blacklisted_images"`
	SetCommand              string   `yaml:"set_command"`
	ViewCommand             string   `yaml:"view_command"`
}

type CLIFlags struct {
	LogLevel   string
	ConfigFile string
	Display    string
}

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

	err = cfg.createDirs()
	if err != nil {
		return nil, err
	}

	return cfg, nil
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

func (c Config) createDirs() error {
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
	}
}

func applyDefaults(cfg *Config, defaults *Config) {
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
}
