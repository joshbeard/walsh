package session

import (
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
	Index   int          `json:"index"`
	Name    string       `json:"name"`
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
func NewSession(cfg *config.Config) (*Session, error) {
	sessType, err := detect()
	if err != nil {
		return nil, err
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
		log.Warnf("Unknown session type: %d", sessType)
		return nil, errors.New("unknown session type")
	}

	display, err := svc.GetDisplays()
	if err != nil {
		log.Errorf("Error getting displays: %s", err)
		return nil, err
	}

	return &Session{
		svc:      svc,
		sessType: sessType,
		displays: display,
		cfg:      cfg,
	}, nil
}

// Config returns the session's config.
func (s Session) Config() *config.Config {
	return s.cfg
}

// getImages gets images from the sources and filters them based on the
// blacklist and history files.
func (s Session) getImages(sources []string) ([]source.Image, error) {
	log.Debugf("Getting images from sources")
	images, err := source.GetImages(sources)
	if err != nil {
		log.Errorf("Error getting images: %s", err)
		return nil, err
	}

	log.Debugf("Filtering blacklisted images")
	blacklist, err := s.ReadList(s.cfg.BlacklistFile)
	if err != nil {
		log.Errorf("Error reading blacklist: %s", err)
		return nil, err
	}
	images = source.FilterImages(images, blacklist)

	history, err := s.ReadList(s.cfg.HistoryFile)
	if err != nil {
		log.Errorf("Error reading history: %s", err)
		return nil, err
	}
	// if the number of images is fewer than the history size, don't filter
	if len(images) > s.cfg.HistorySize {
		log.Debugf("Filtering images in history")
		images = source.FilterImages(images, history)
	}

	return images, nil
}

// SetWallpaper sets the wallpaper for the session.
func (s *Session) SetWallpaper(sources []string, displayStr string) error {
	var err error
	if len(sources) == 0 {
		sources = s.cfg.Sources
	}

	images, err := s.getImages(sources)
	if err != nil {
		return err
	}

	var display Display
	displays := s.displays

	if displayStr != "" {
		_, display, err = s.GetDisplay(displayStr)
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
				image, err = source.Random(images, s.cfg.CacheDir)
			} else {
				mu.Unlock()
				errChan <- errors.New("no images available")
				return
			}
			mu.Unlock()

			if err != nil {
				log.Errorf("Error selecting random image for display %s: %s", d.Name, err)
				time.Sleep(1 * time.Second)
				continue
			}
			log.Debugf("Number of images: %d", len(images))

			err = s.svc.SetWallpaper(image.Path, d)
			if err != nil {
				log.Errorf("Error setting wallpaper for display %s: %s. Will retry", d.Name, err)
				time.Sleep(1 * time.Second)
				continue
			}

			mu.Lock()

			// Don't re-use the same image for multiple displays, unless there aren't
			// enough images to go around.
			if len(displays) <= len(images) {
				images = source.RemoveImage(images, image)
			}

			err = s.WriteCurrent(d, image)
			if err != nil {
				log.Errorf("Error saving to history for display %s: %s", d.Name, err)
				errChan <- err
				mu.Unlock()
				return
			}

			err = s.WriteHistory(image)
			if err != nil {
				log.Errorf("Error writing to history for display %s: %s", d.Name, err)
				errChan <- err
				mu.Unlock()
				return
			}

			mu.Unlock()

			log.Infof("Set wallpaper for display %s: %s", d.Name, image.Path)
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

	err = s.cleanupTmpDir()
	if err != nil {
		log.Errorf("Error cleaning up tmp dir: %s", err)
	}

	return nil
}

// GetDisplay gets a display by index or name.
func (s Session) GetDisplay(display string) (int, Display, error) {
	// If it's a number, assume it's an index. Otherwise, look up by name.
	if util.IsNumber(display) {
		i, err := strconv.Atoi(display)
		if err != nil {
			return -1, Display{}, err
		}

		if i >= len(s.displays) {
			return -1, Display{}, errors.New("display index out of range")
		}

		return i, s.displays[i], nil
	}

	for i, d := range s.displays {
		if d.Name == display {
			return i, d, nil
		}
	}

	return -1, Display{}, errors.New("display not found")
}

// Displays returns the displays for the session.
func (s Session) Displays() []Display {
	return s.displays
}

// Display returns the display for the session by index or name.
func (c CurrentWallpaper) Display(display string) (Display, error) {
	for _, d := range c.Displays {
		// If the display arg is a number, match the index. otherwise, match the name.
		if strconv.Itoa(d.Index) == display || d.Name == display {
			return d, nil
		}
	}

	return Display{}, errors.New("current wallpaper not found for display")
}

// detect the current session type based on the environment and/or
// commands.
func detect() (SessionType, error) {
	// Hyprland - XDG_CURRENT_DESKTOP=Hyprland
	// X11 - XDG_SESSION_TYPE=x11
	xdgCurrentDesktop := os.Getenv("XDG_CURRENT_DESKTOP")
	xdgSessionType := os.Getenv("XDG_SESSION_TYPE")
	swaySocket := os.Getenv("SWAYSOCK")
	isMac := isMacOS()

	switch {
	case isMac:
		log.Debugf("Detected macOS session")
		return SessionTypeMacOS, nil
	case xdgCurrentDesktop == "Hyprland":
		log.Debugf("Detected Hyprland desktop")
		return SessionTypeHyprland, nil
	case xdgSessionType == "wayland" && swaySocket != "":
		log.Debugf("Detected Sway session")
		return SessionTypeSway, nil
	case xdgSessionType == "wayland":
		log.Debugf("Detected Wayland session")
		return SessionTypeWayland, nil
	case xdgSessionType == "x11":
		log.Debugf("Detected X11 session")
		return SessionTypeX11Unknown, nil
	default:
		return SessionTypeUnknown, errors.New("unknown session type")
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
func (s Session) GetCurrentWallpaper(display string) (string, error) {
	_, d, err := s.GetDisplay(display)
	if err != nil {
		return "", err
	}

	currentFile, err := s.ReadCurrent()
	if err != nil {
		log.Fatal(err)
	}

	currentDisplay, err := currentFile.Display(display)
	if err != nil {
		log.Fatal(err)
	}

	return s.svc.GetCurrentWallpaper(d, currentDisplay)
}

// View opens an image in the default image viewer.
func (s Session) View(image string) error {
	log.Debugf("Viewing %s", image)
	cmd, err := s.getViewCommand(image)
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

func (s Session) getViewCommand(image string) (string, error) {
	if s.cfg.ViewCommand != "" {
		return parseViewCmd(s.cfg.ViewCommand, image), nil
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
func (s Session) WriteCurrent(display Display, path source.Image) error {
	var err error
	// Update the display's current path
	display.Current = path

	// Ensure the file exists before reading
	if !util.FileExists(s.cfg.CurrentFile) {
		// Create the file if it doesn't exist
		err = os.WriteFile(s.cfg.CurrentFile, []byte("{}"), 0o644)
		if err != nil {
			return fmt.Errorf("failed to create the file: %w", err)
		}
	}

	// Read the existing file content
	fileBytes, err := os.ReadFile(s.cfg.CurrentFile)
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
		if d.Index == display.Index {
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
	err = os.WriteFile(s.cfg.CurrentFile, updatedBytes, 0o644)
	if err != nil {
		return fmt.Errorf("failed to write the updated content to the file: %w", err)
	}

	return nil
}

// ReadCurrent reads the current wallpaper data for the session.
func (s Session) ReadCurrent() (CurrentWallpaper, error) {
	if !util.FileExists(s.cfg.CurrentFile) {
		return CurrentWallpaper{}, nil
	}

	// Read the file
	fileBytes, err := os.ReadFile(s.cfg.CurrentFile)
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

// ReadList reads a list of images from a file.
func (s Session) ReadList(file string) ([]source.Image, error) {
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
func (s Session) WriteList(file string, image source.Image) error {
	list, err := s.ReadList(file)
	if err != nil {
		return err
	}

	// Check if it's already in the list
	if source.ImageInList(image, list) {
		log.Warnf("Image already in list: %s", image.Path)
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

	log.Debugf("Added image %s to list: %s", image.Path, file)

	return nil
}

// WriteHistory writes an image to the history file.
func (s Session) WriteHistory(image source.Image) error {
	// Write the image to the history file
	err := s.WriteList(s.cfg.HistoryFile, image)
	if err != nil {
		return err
	}

	// Trim the history file if it's too large
	err = s.TrimHistory()
	if err != nil {
		return err
	}

	return nil
}

// TrimHistory trims the history file to the configured size.
func (s Session) TrimHistory() error {
	history, err := s.ReadList(s.cfg.HistoryFile)
	if err != nil {
		return err
	}

	if len(history) > s.cfg.HistorySize {
		history = history[len(history)-s.cfg.HistorySize:]
	}

	// Marshal the updated content back to json
	updatedBytes, err := json.MarshalIndent(history, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal the updated content: %w", err)
	}

	// Write the updated content back to the file
	err = os.WriteFile(s.cfg.HistoryFile, updatedBytes, 0o644)
	if err != nil {
		return fmt.Errorf("failed to write the updated content to the file: %w", err)
	}

	return nil
}

// cleanupTmpDir cleans up the tmp directory by removing old files.
func (s Session) cleanupTmpDir() error {
	if !util.FileExists(s.cfg.CacheDir) {
		return nil
	}

	files, err := os.ReadDir(s.cfg.CacheDir)
	if err != nil {
		return err
	}

	if len(files) > s.cfg.CacheSize {
		// Sort by mod time
		util.SortFilesByMTime(files)

		// Get the paths of the current wallpapers
		currentWallpapers := make(map[string]bool)
		for _, display := range s.displays {
			currentWallpapers[display.Current.Path] = true
		}

		// Remove the oldest files that are not current wallpapers
		removedCount := 0
		for i := 0; i < len(files) && removedCount < len(files)-s.cfg.CacheSize; i++ {
			path := filepath.Join(s.cfg.CacheDir, files[i].Name())
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
