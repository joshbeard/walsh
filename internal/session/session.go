package session

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/log"
	"github.com/joshbeard/walsh/internal/config"
	"github.com/joshbeard/walsh/internal/source"
	"github.com/joshbeard/walsh/internal/util"
)

// MaxRetries is the maximum number of times to retry setting the wallpaper.
// e.g. if failing to download from a remote source or failure to set the
// wallpaper.
const MaxRetries = 6

var This *Session

// defaultViewCmds are the default commands to view an image.
var defaultViewCmds = []string{
	`xdg-open '{{path}}'`,
	`feh --scale-down --auto-zoom '{{path}}'`,
	`eog '{{path}}'`,
	`eom '{{path}}'`,
}

// Session is a struct for managing the desktop session's wallpaper.
type Session struct {
	displays []Display
	svc      SessionProvider
	sessType SessionType
	cfg      *config.Config
	interval time.Duration

	Ctx     context.Context
	Cancel  context.CancelFunc
	SigChan chan os.Signal
	Ticker  *time.Ticker
}

// SessionProvider is an interface for interacting with the desktop session.
// This is used for abstracting the necessary commands to set the wallpaper and
// get the list of displays.
type SessionProvider interface {
	SetWallpaper(path string, display Display) error
	GetDisplays() ([]Display, error)
	GetCurrentWallpaper(display, current Display) (string, error)
}

// CurrentWallpaper is a helper struct for representing the current wallpaper
// state on displays.
type CurrentWallpaper struct {
	Displays []Display `json:"displays"`
}

// Display represents a display that can use a wallpaper.
// The name and index are used to identify the display, and are determined by
// the display's actual identifier (e.g. eDP-1, HDMI-1, etc.) or an index based
// on how they are queried from the system (e.g. 0, 1, 2, etc.).
type Display struct {
	ID      string       `json:"id"`
	Index   int          `json:"index"`
	Name    string       `json:"name"`
	Label   string       `json:"label"`
	Current source.Image `json:"current"`
}

// I expect this would need to change to support more varieties of Wayland
// compositors and Xorg session types.
// We really just need to know what commands to run to
//  1. get the list of displays
//  2. set the wallpaper
//  3. ideally get the current wallpaper on a display, but that's not strictly
//     necessary
type SessionType int

const (
	SessionTypeUnknown SessionType = iota
	SessionTypeX11Unknown
	SessionTypeWayland
	SessionTypeSway
	SessionTypeHyprland
	SessionTypeMacOS
)

// SetWallpaperParams is a struct for setting the wallpaper.
type SetWallpaperParams struct {
	Path    string
	Cmd     string
	Display string
}

// NewSession creates a new session based on the current session type.
// func NewSession(cfg *config.Config) (*Session, error) {
func NewSession(cfg *config.Config) error {
	sessType, err := detect()
	if err != nil {
		return err
	}

	var svc SessionProvider
	switch sessType {
	case SessionTypeHyprland:
		svc = NewHyprland(cfg)
	case SessionTypeX11Unknown:
		svc = NewXorg(cfg)
	case SessionTypeMacOS:
		svc = NewMacOS(cfg)
	case SessionTypeSway:
		svc = NewSway(cfg)
	default:
		return errors.New("could not determine session type")
	}

	display, err := svc.GetDisplays()
	if err != nil {
		log.Errorf("getting displays", "err", err)
		return err
	}

	// If a specific display is requested via config, limit the displays to that
	// display.
	if cfg.Display != "" {
		log.Info("limiting displays to display in config", "display", cfg.Display)
		_, d, err := GetDisplay(cfg.Display)
		if err != nil {
			return err
		}

		display = []Display{d}
	}

	sess := &Session{
		svc:      svc,
		sessType: sessType,
		displays: display,
		cfg:      cfg,
		interval: cfg.Interval,
	}

	This = sess

	return nil
}

func Refresh() error {
	display, err := This.svc.GetDisplays()
	if err != nil {
		return err
	}

	This.displays = display

	return nil
}

// Config returns the session's config.
func Config() *config.Config {
	return This.cfg
}

func Type() SessionType {
	return This.sessType
}

func SetConfig(cfg *config.Config) {
	This.cfg = cfg
}

func SetInterval(interval time.Duration) {
	This.interval = interval
}

func Interval() time.Duration {
	return This.interval
}

// getImages gets images from the sources and filters them based on the
// blacklist and history files.
func getImages(sources []string) ([]source.Image, error) {
	log.Debugf("getting images from sources")
	images, err := source.GetImages(sources)
	if err != nil {
		log.Error("getting images", "err", err)
		return nil, err
	}

	log.Debugf("filtering blacklisted images")
	blacklist, err := ReadList(This.cfg.BlacklistFile)
	if err != nil {
		log.Error("reading blacklist", "err", err)
		return nil, err
	}
	images = source.FilterImages(images, blacklist)

	history, err := ReadList(This.cfg.HistoryFile)
	if err != nil {
		log.Error("reading history", "err", err)
		return nil, err
	}
	// if the number of images is fewer than the history size, don't filter
	if len(images) > This.cfg.HistorySize {
		log.Debug("filtering images in history",
			"count", len(history),
			"history_size", This.cfg.HistorySize)
		images = source.FilterImages(images, history)
	}

	return images, nil
}

// SetWallpaper sets the wallpaper for the session.
func SetWallpaper(disp string) error {
	var err error

	log.Debug("setting wallpaper", "display", disp)

	images, err := getImages(This.cfg.Sources)
	if err != nil {
		return err
	}

	var display Display
	displays := This.displays

	if disp != "" {
		_, display, err = GetDisplay(disp)
		if err != nil {
			return err
		}
		displays = []Display{display}
	}

	var wg sync.WaitGroup
	errChan := make(chan error, len(displays))
	defer close(errChan)
	var mu sync.Mutex

	// Function to process each display
	processDisplay := func(d Display) {
		defer wg.Done()
		for i := 0; i < MaxRetries; i++ {
			var image source.Image
			// Synchronize access to the images slice
			mu.Lock()
			if len(images) > 0 {
				image, err = source.Random(images, This.cfg.CacheDir)
			} else {
				mu.Unlock()
				errChan <- errors.New("no images available")
				return
			}
			mu.Unlock()

			if err != nil {
				log.Error("selecting random image for display", "display", d.ID, "err", err)
				time.Sleep(1 * time.Second)
				continue
			}
			log.Debugf("number of images: %d", len(images))

			err = This.svc.SetWallpaper(image.Path, d)
			if err != nil {
				log.Error("setting wallpaper for display; will retry", "display", d.ID, "err", err)
				time.Sleep(1 * time.Second)
				continue
			}

			mu.Lock()

			// Don't re-use the same image for multiple displays, unless there aren't
			// enough images to go around.
			if len(displays) <= len(images) {
				images = source.RemoveImage(images, image)
			}

			err = WriteCurrent(d, image)
			if err != nil {
				log.Error("saving to history", "display", d.ID, "err", err)
				errChan <- err
				mu.Unlock()
				return
			}

			err = WriteHistory(image)
			if err != nil {
				log.Error("writing to history", "display", d.ID, "err", err)
				errChan <- err
				mu.Unlock()
				return
			}

			mu.Unlock()

			log.Info("set wallpaper", "display", d.ID, "image", image.Path)
			return
		}
		errChan <- errors.New("max retries exceeded")
	}

	for _, d := range displays {
		wg.Add(1)
		go processDisplay(d)
	}

	wg.Wait()

	select {
	case err = <-errChan:
		return err
	default:
	}

	err = cleanupTmpDir()
	if err != nil {
		log.Errorf("cleaning up tmp dir", "err", err)
	}

	return nil
}

// GetDisplay gets a display by index or name.
func GetDisplay(display string) (int, Display, error) {
	// If it's a number, assume it's an index. Otherwise, look up by name.
	if util.IsNumber(display) {
		i, err := strconv.Atoi(display)
		if err != nil {
			return -1, Display{}, err
		}

		// Get display with matching ID
		for _, d := range This.displays {
			if d.ID == display {
				log.Debug("found display by ID", "display", display)
				return i, d, nil
			}
		}

		return -1, Display{}, errors.New("display not found")
	}

	for i, d := range This.displays {
		if d.Name == display {
			log.Debug("found display by name", "display", display)
			return i, d, nil
		}
	}

	return -1, Display{}, errors.New("display not found")
}

// Displays returns the displays for the session.
func Displays() []Display {
	return This.displays
}

func TargetDisplays() []Display {
	// Return displays that are specified in the config.
	if This.cfg.Display != "" {
		log.Warn("limiting displays to display in config", "display", This.cfg.Display)
		_, d, err := GetDisplay(This.cfg.Display)
		if err != nil {
			log.Fatal(err)
		}

		return []Display{d}
	}

	return Displays()
}

// Display returns the display for the session by index or name.
func (c CurrentWallpaper) Display(display string) (Display, error) {
	for _, d := range c.Displays {
		// If the display arg is a number, match the index. otherwise, match the name.
		if d.ID == display || d.Name == display {
			return d, nil
		}
	}

	return Display{}, errors.New("current wallpaper not found for display")
}

// detect the current session type based on the environment and/or
// commands.
func detect() (SessionType, error) {
	xdgCurrentDesktop := os.Getenv("XDG_CURRENT_DESKTOP")
	xdgSessionType := os.Getenv("XDG_SESSION_TYPE")
	xAuthority := os.Getenv("XAUTHORITY")
	i3Socket := os.Getenv("I3SOCK")
	swaySocket := os.Getenv("SWAYSOCK")
	isMac := isMacOS()

	sessionType := SessionTypeUnknown

	switch {
	case isMac:
		sessionType = SessionTypeMacOS
	case xdgCurrentDesktop == "Hyprland":
		sessionType = SessionTypeHyprland
	case xdgSessionType == "wayland" && swaySocket != "":
		sessionType = SessionTypeSway
	case xdgSessionType == "wayland":
		sessionType = SessionTypeWayland
	case xdgSessionType == "x11" || xAuthority != "" || i3Socket != "":
		sessionType = SessionTypeX11Unknown
	}

	return sessionType, nil
}

func (s SessionType) String() string {
	switch s {
	case SessionTypeUnknown:
		return "unknown"
	case SessionTypeX11Unknown:
		return "x11"
	case SessionTypeWayland:
		return "wayland"
	case SessionTypeSway:
		return "sway"
	case SessionTypeHyprland:
		return "hyprland"
	case SessionTypeMacOS:
		return "macos"
	default:
		return "unknown"
	}
}

// parseSetCmd parses a templatized command string for setting a wallpaper.
// These values are replaced:
//   - {{path}}: the path to the image file
//   - {{display}}: the display number
//
// swww img '{{path}}' --outputs '{{display}}'"
func parseSetCmd(cmd, path, display string) string {
	cmd = strings.ReplaceAll(cmd, "{{path}}", path)
	cmd = strings.ReplaceAll(cmd, "{{display}}", display)

	return cmd
}

// getSetCmd determines and returns the set commands for the session. It
// iterates over the default set commands until an available command is
// found.
func getSetCmd(l []string, path, display string) (string, error) {
	for _, cmd := range l {
		if _, err := exec.LookPath(strings.Split(cmd, " ")[0]); err == nil {
			cmd = parseSetCmd(cmd, path, display)

			return cmd, nil
		}
	}

	return "", errors.New("no set command found")
}

// GetCurrentWallpaper gets the current wallpaper for a display.
func GetCurrentWallpaper(display string) (string, error) {
	_, d, err := GetDisplay(display)
	if err != nil {
		return "", fmt.Errorf("failed to get display %s: %w", display, err)
	}

	currentFile, err := ReadCurrent()
	if err != nil {
		log.Fatal(err, "display", display)
	}

	currentDisplay, err := currentFile.Display(display)
	if err != nil {
		log.Fatal(err, "display", display)
	}

	return This.svc.GetCurrentWallpaper(d, currentDisplay)
}

// View opens an image in the default image viewer.
func View(image string) error {
	log.Debug("viewing image", "image", image)
	cmd, err := getViewCommand(image)
	if err != nil {
		return err
	}

	_, err = util.RunCmd(cmd)
	if err != nil {
		return err
	}

	return nil
}

func parseViewCmd(cmd, path string) string {
	return strings.ReplaceAll(cmd, "{{path}}", path)
}

func getViewCommand(image string) (string, error) {
	if This.cfg.ViewCommand != "" {
		return parseViewCmd(This.cfg.ViewCommand, image), nil
	}

	if isMacOS() {
		return fmt.Sprintf("open -a Preview '%s'", image), nil
	}

	for _, cmd := range defaultViewCmds {
		if _, err := exec.LookPath(strings.Split(cmd, " ")[0]); err == nil {
			cmd = parseViewCmd(cmd, image)

			return cmd, nil
		}
	}

	return "", errors.New("no view command found")
}

// WriteCurrent writes the current wallpaper for a given display to the
// s.cfg.HistoryPath file in the format of "display:path".
// e.g. "0:/path/to/image.jpg"
// This only updates the line for the given display, leaving the rest of the
// file unchanged.
func WriteCurrent(display Display, path source.Image) error {
	var err error
	// Update the display's current path
	display.Current = path

	// Ensure the file exists before reading
	if !util.FileExists(This.cfg.CurrentFile) {
		// Create the file if it doesn't exist
		err = os.WriteFile(This.cfg.CurrentFile, []byte("{}"), 0o644)
		if err != nil {
			return fmt.Errorf("failed to create the file: %w", err)
		}
	}

	// Read the existing file content
	fileBytes, err := os.ReadFile(This.cfg.CurrentFile)
	if err != nil {
		return fmt.Errorf("failed to read the file: %w", err)
	}

	// Unmarshal the file content into a CurrentWallpaper struct
	var current CurrentWallpaper
	err = json.Unmarshal(fileBytes, &current)
	if err != nil {
		return fmt.Errorf("failed to unmarshal the file content: %w", err)
	}

	// Update or append the display in the current wallpaper list
	found := false
	for i, d := range current.Displays {
		if d.ID == display.ID {
			found = true
			current.Displays[i] = display
			break
		}
	}
	if !found {
		current.Displays = append(current.Displays, display)
	}

	// Marshal the updated content back to JSON
	updatedBytes, err := json.MarshalIndent(current, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal the updated content: %w", err)
	}

	// Write the updated content back to the file
	err = os.WriteFile(This.cfg.CurrentFile, updatedBytes, 0o644)
	if err != nil {
		return fmt.Errorf("failed to write the updated content to the file: %w", err)
	}

	return nil
}

// ReadCurrent reads the current wallpaper data for the session.
func ReadCurrent() (CurrentWallpaper, error) {
	if !util.FileExists(This.cfg.CurrentFile) {
		return CurrentWallpaper{}, nil
	}

	// Read the file
	fileBytes, err := os.ReadFile(This.cfg.CurrentFile)
	if err != nil {
		return CurrentWallpaper{}, fmt.Errorf("failed to read the file: %w", err)
	}

	// Unmarshal the file content into a CurrentWallpaper struct
	var current CurrentWallpaper
	err = json.Unmarshal(fileBytes, &current)
	if err != nil {
		return CurrentWallpaper{}, fmt.Errorf("failed to unmarshal the file content: %w", err)
	}

	return current, nil
}

func ListLists() ([]string, error) {
	lists := []string{}
	files, err := os.ReadDir(This.cfg.ListsDir)
	if err != nil {
		return lists, err
	}

	for _, file := range files {
		lists = append(lists, strings.TrimSuffix(file.Name(), ".json"))
	}

	return lists, nil
}

// ReadList reads a list of images from a file.
func ReadList(file string) ([]source.Image, error) {
	if !util.FileExists(file) {
		return nil, nil
	}

	// Read the list file
	fileBytes, err := os.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read the file: %w", err)
	}

	// Unmarshal the file content into an ImageList struct
	var list []source.Image
	err = json.Unmarshal(fileBytes, &list)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal the file content: %w", err)
	}

	return list, nil
}

// WriteList writes a list of images to a file.
func WriteList(file string, image source.Image) error {
	var err error
	// Compute the shasum if it's not set
	if image.ShaSum == "" {
		image.ShaSum, err = util.Sha256(image.Path)
		if err != nil {
			return fmt.Errorf("failed to calculate checksum: %w", err)
		}
	}

	list, err := ReadList(file)
	if err != nil {
		return err
	}

	// Check if it's already in the list
	if source.ImageInList(image, list) {
		log.Warn("image already in list", "image", image.Path)
		return nil
	}

	// Append the image to the list
	list = append(list, image)

	// Marshal the updated content back to JSON
	updatedBytes, err := json.MarshalIndent(list, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal the updated content: %w", err)
	}

	// Write the updated content back to the file
	err = os.WriteFile(file, updatedBytes, 0o644)
	if err != nil {
		return fmt.Errorf("failed to write the updated content to the file: %w", err)
	}

	log.Debug("added image list", "list", file, "image", image.Path)

	return nil
}

// WriteHistory writes an image to the history file.
func WriteHistory(image source.Image) error {
	// Write the image to the history file
	err := WriteList(This.cfg.HistoryFile, image)
	if err != nil {
		return err
	}

	// Trim the history file if it's too large
	err = TrimHistory()
	if err != nil {
		return err
	}

	return nil
}

// TrimHistory trims the history file to the configured size.
func TrimHistory() error {
	history, err := ReadList(This.cfg.HistoryFile)
	if err != nil {
		return err
	}

	if len(history) > This.cfg.HistorySize {
		history = history[len(history)-This.cfg.HistorySize:]
	}

	// Marshal the updated content back to json
	updatedBytes, err := json.MarshalIndent(history, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal the updated content: %w", err)
	}

	// Write the updated content back to the file
	err = os.WriteFile(This.cfg.HistoryFile, updatedBytes, 0o644)
	if err != nil {
		return fmt.Errorf("failed to write the updated content to the file: %w", err)
	}

	return nil
}

// cleanupTmpDir cleans up the tmp directory by removing old files.
func cleanupTmpDir() error {
	if !util.FileExists(This.cfg.CacheDir) {
		return nil
	}

	files, err := os.ReadDir(This.cfg.CacheDir)
	if err != nil {
		return err
	}

	if len(files) > This.cfg.CacheSize {
		// Sort by mod time
		util.SortFilesByMTime(files)

		// Get the paths of the current wallpapers
		currentWallpapers := make(map[string]bool)
		for _, display := range This.displays {
			currentWallpapers[display.Current.Path] = true
		}

		// Remove the oldest files that are not current wallpapers
		removedCount := 0
		for i := 0; i < len(files) && removedCount < len(files)-This.cfg.CacheSize; i++ {
			path := filepath.Join(This.cfg.CacheDir, files[i].Name())
			if !currentWallpapers[path] {
				err := os.Remove(path)
				if err != nil {
					return err
				}
				removedCount++
			}
		}
	}

	return nil
}

func ResetTicker() {
	if This.Ticker != nil {
		if This.interval == 0 {
			log.Debug("pausing wallpaper change", "interval", Interval())
			This.Ticker.Stop()
			return
		}

		log.Debug("updating change interval", "interval", Interval(), "next", NextTick())
		This.Ticker.Reset(Interval())

		return
	}

	if This.interval == 0 {
		log.Debug("wallpaper change is paused", "interval", Interval())
		return
	}

	log.Info("creating new ticker with interval",
		"interval", Interval(), "next", NextTick())
	This.Ticker = time.NewTicker(This.interval)
}

func NextTick() string {
	next := time.Now().Add(Interval())
	return next.Format("2006-01-02 15:04:05")
}
