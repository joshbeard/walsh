package source

import (
	"bufio"
	"errors"
	"fmt"
	"math/rand"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/log"
	"github.com/joshbeard/walsh/internal/util"
)

type SourceProvider interface {
	GetImages(srcs []string) ([]Image, error)
	Random(images []Image) (Image, error)
}

type Images struct {
	Sources []string
	service SourceProvider
}

type Image struct {
	Source string
	Path   string
	ShaSum string
}

type SourceType int

const (
	SourceDirectory SourceType = iota
	SourceList
	SourceSSH
)

var sourcePrefixes = map[SourceType]string{
	SourceDirectory: "dir://",
	SourceList:      "list://",
	SourceSSH:       "ssh://",
}

func (st SourceType) String() string {
	return sourcePrefixes[st]
}

func NewImages(sources []string, service SourceProvider) *Images {
	return &Images{
		Sources: sources,
		service: service,
	}
}

// GetImages retrieves a list of images from a specified list (text file),
// directory, or remote source and returns a slice of paths.
func GetImages(srcs []string) ([]Image, error) {
	var err error
	var results, images []Image

	for _, src := range srcs {
		log.Debugf("getting images from source '%s'", src)
		switch {
		case util.IsFilePath(src):
			log.Debugf("source '%s' is a file", src)
			results, err = getDirImages(src)
		case strings.HasPrefix(src, SourceDirectory.String()):
			results, err = getDirImages(src)
		case strings.HasPrefix(src, SourceList.String()):
			results, err = getListImages(src)
		case strings.HasPrefix(src, SourceSSH.String()):
			results, err = getSSHImages(src)
		default:
			return nil, fmt.Errorf("invalid source format: %s", src)
		}

		if err != nil {
			log.Errorf("failed to get images from source '%s': %s", src, err)
			continue
		}

		log.Debugf("found %d images in source '%s'", len(results), src)
		images = append(images, results...)
	}

	if len(images) == 0 {
		return nil, errors.New("no images were found")
	}

	return images, nil
}

// Random selects a random image from a list of images.
func Random(images []Image, tmpDir string) (Image, error) {
	var err error

	if len(images) == 0 {
		return Image{}, errors.New("no images available")
	}

	seed := time.Now().UnixNano()
	rng := rand.New(rand.NewSource(seed))
	randomIndex := rng.Intn(len(images))

	image := images[randomIndex]
	if strings.HasPrefix(image.Source, SourceSSH.String()) {
		// Create tmpDir if it doesn't exist.
		if !util.FileExists(tmpDir) {
			err = os.Mkdir(tmpDir, 0700)
			if err != nil {
				return Image{}, fmt.Errorf("failed to create temporary directory: %w", err)
			}
		}

		dest := filepath.Join(tmpDir, filepath.Base(image.Source))
		image, err = downloadSSHImage(image, dest)
		if err != nil {
			return Image{}, fmt.Errorf("failed to download SSH image: %w", err)
		}
	}

	return image, nil
}

// isImageFile checks if a file has a valid image extension.
func isImageFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".jpg", ".jpeg", ".png", ".gif", ".bmp", ".tiff":
		return true
	default:
		return false
	}
}

// getDirImages retrieves a list of images from a local directory and
// returns a slice of paths.
// The source string should be in the format:
// dir:///path/to/images
func getDirImages(src string) ([]Image, error) {
	// Remove the "dir://" prefix from the source string to get the directory path.
	dirPath := strings.TrimPrefix(src, SourceDirectory.String())

	// Open the directory.
	dir, err := os.Open(dirPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open directory: %w", err)
	}
	defer dir.Close()

	// Read the directory contents.
	entries, err := dir.Readdir(-1)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	// Filter the entries to include only image files.
	var images []Image
	for _, entry := range entries {
		if !entry.IsDir() {
			if isImageFile(entry.Name()) {
				checksum, err := util.Sha256(filepath.Join(dirPath, entry.Name()))
				if err != nil {
					return nil, fmt.Errorf("failed to calculate checksum: %w", err)
				}

				images = append(images, Image{
					Source: SourceDirectory.String(),
					Path:   filepath.Join(dirPath, entry.Name()),
					ShaSum: checksum,
				})
			}
		}
	}

	return images, nil
}

// getListImages reads a list of image paths from a text file and returns
// a slice of paths. Basically just returns each line as a path.
func getListImages(src string) ([]Image, error) {
	// Remove the "list://" prefix from the source string to get the file path.
	listPath := strings.TrimPrefix(src, SourceList.String())

	// Open the list file.
	file, err := os.Open(listPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open list file: %w", err)
	}
	defer file.Close()

	// Read the list file line by line.
	var images []Image
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		path := scanner.Text()
		checksum, err := util.Sha256(path)
		if err != nil {
			return nil, fmt.Errorf("failed to calculate checksum: %w", err)
		}

		images = append(images, Image{
			Source: SourceList.String(),
			Path:   path,
			ShaSum: checksum,
		})
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read list file: %w", err)
	}

	return images, nil
}

// getSSHImages retrieves a list of images from a remote SSH server and
// returns a slice of paths.
// The source string should be in the format:
// ssh://user@host:/path/to/images
func getSSHImages(src string) ([]Image, error) {
	// Parse the SSH URI.
	sshURI, err := ParseSSHURI(src)
	if err != nil {
		return nil, fmt.Errorf("failed to parse SSH URI: %w", err)
	}

	addr := sshURI.Address

	// Test connectivity
	ok, err := isSSHAlive(addr)
	if err != nil {
		return nil, err
	}

	if !ok {
		return nil, fmt.Errorf("SSH connection to %s not successful", addr)
	}

	// Construct the SSH command.
	sshCmd := fmt.Sprintf("ssh %s ls %s", addr, sshURI.Path)

	// Run the SSH command.
	log.Debugf("running SSH command: %s", sshCmd)
	output, err := util.RunCmd(sshCmd)
	if err != nil {
		return nil, fmt.Errorf("failed to run SSH command: %w", err)
	}

	// Split the output into lines (each line is a path to an image).
	images := strings.Split(strings.TrimSpace(output), "\n")

	// Filter the paths to include only image files.
	var validImages []Image
	for _, img := range images {
		if isImageFile(img) {
			validImages = append(validImages, Image{
				Source: src + "/" + img,
			})
		}
	}

	log.Debugf("found %d images in source '%s'", len(validImages), src)

	return validImages, nil
}

// isSSHAlive checks if we can connect to an SSH source.
func isSSHAlive(addr string) (bool, error) {
	cmd := fmt.Sprintf("ssh -v -o BatchMode=yes -o ConnectTimeout=5 '%s' "+
		"echo 'SSH connection test' 2>&1", addr)

	log.Debugf("testing ssh connection with command: %s", cmd)
	_, err := util.RunCmd(cmd)
	if err != nil {
		return false, fmt.Errorf("could not connect to SSH host %s: %w", addr, err)
	}

	return true, nil
}

type SSHURI struct {
	User    string
	Server  string
	Path    string
	Address string
}

func ParseSSHURI(uri string) (*SSHURI, error) {
	// Ensure the URI starts with "ssh://"
	if !strings.HasPrefix(uri, "ssh://") {
		return nil, fmt.Errorf("invalid SSH URI: %s", uri)
	}

	// Parse the URI using net/url
	parsedURI, err := url.Parse(uri)
	if err != nil {
		return nil, fmt.Errorf("failed to parse SSH URI: %w", err)
	}

	// Initialize the SSHURI struct
	sshURI := &SSHURI{
		Server:  parsedURI.Hostname(),
		Path:    parsedURI.Path,
		Address: parsedURI.Hostname(),
	}

	// Extract user info if present
	if parsedURI.User != nil {
		sshURI.User = parsedURI.User.Username()
		sshURI.Address = fmt.Sprintf("%s@%s", sshURI.User, sshURI.Server)
	}

	// Ensure the path is not empty
	if sshURI.Path == "" {
		return nil, fmt.Errorf("path is missing in the SSH URI: %s", uri)
	}

	return sshURI, nil
}

// BuildSCPCommand builds an scp command string from the SSHURI struct.
func BuildSCPCommand(source string, destination string) (string, error) {
	var src, dst string

	sourceIsSSH := strings.HasPrefix(source, "ssh://")

	// Parse the SSH URI if sourceIsSSH is true
	if sourceIsSSH {
		parsedURI, err := ParseSSHURI(source)
		if err != nil {
			return "", err
		}
		if parsedURI.User != "" {
			src = fmt.Sprintf("%s@%s:\"%s\"", parsedURI.User, parsedURI.Server, parsedURI.Path)
		} else {
			src = fmt.Sprintf("%s:\"%s\"", parsedURI.Server, parsedURI.Path)
		}
		dst = escapePath(destination)
	} else {
		// Parse the SSH URI if the destination is an SSH URI
		parsedURI, err := ParseSSHURI(destination)
		if err != nil {
			return "", err
		}
		if parsedURI.User != "" {
			dst = fmt.Sprintf("%s@%s:\"%s\"", parsedURI.User, parsedURI.Server, parsedURI.Path)
		} else {
			dst = fmt.Sprintf("%s:\"%s\"", parsedURI.Server, parsedURI.Path)
		}
		src = escapePath(source)
	}

	return fmt.Sprintf("scp %s %s", src, dst), nil
}

func escapePath(path string) string {
	return strings.ReplaceAll(path, " ", "\\ ")
}

func downloadSSHImage(src Image, dest string) (Image, error) {
	sshCmd, err := BuildSCPCommand(src.Source, dest)
	if err != nil {
		return Image{}, fmt.Errorf("failed to build SCP command: %w", err)
	}

	// Run the SSH command.
	result, err := util.RunCmd(sshCmd)
	if err != nil {
		return Image{}, fmt.Errorf("failed to run SSH command: %w output: %s", err, result)
	}

	checksum, err := util.Sha256(dest)
	if err != nil {
		return Image{}, fmt.Errorf("failed to calculate checksum: %w", err)
	}

	src.Path = dest
	src.ShaSum = checksum

	return src, nil
}

func UploadSSHImage(src Image, dest string) error {
	sshCmd, err := BuildSCPCommand(src.Path, dest)
	if err != nil {
		return fmt.Errorf("failed to build SCP command: %w", err)
	}

	// Run the SSH command.
	result, err := util.RunCmd(sshCmd)
	if err != nil {
		return fmt.Errorf("failed to run SSH command: %w output: %s", err, result)
	}

	// Remove the source file
	log.Debugf("Removing source file: %s", src.Path)
	if err := os.Remove(src.Path); err != nil {
		return fmt.Errorf("failed to remove source file: %w", err)
	}

	return nil
}

func ImageInList(image Image, list []Image) bool {
	for _, i := range list {
		if i.ShaSum == image.ShaSum {
			return true
		}
	}

	return false
}

// FilterImages filters out images that are in a list.
func FilterImages(images []Image, list []Image) []Image {
	var filtered []Image
	for _, i := range images {
		if !ImageInList(i, list) {
			filtered = append(filtered, i)
		}
	}

	return filtered
}

func GetMatches(images []Image, list []Image) []Image {
	var matches []Image
	for _, i := range images {
		if ImageInList(i, list) {
			matches = append(matches, i)
		}
	}

	return matches
}

func RemoveImage(list []Image, image Image) []Image {
	var newList []Image
	for _, i := range list {
		if i.ShaSum != image.ShaSum {
			newList = append(newList, i)
		}
	}

	return newList
}
