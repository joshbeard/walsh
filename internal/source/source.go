package source

import (
	"bufio"
	"errors"
	"fmt"
	"math/rand"
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
		switch {
		case strings.HasPrefix(src, "/"):
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
		if _, err := os.Stat(tmpDir); os.IsNotExist(err) {
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
			// Check if the file has a valid image extension.
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
				//filepath.Join(dirPath, entry.Name()))
			}
		}
	}

	return images, nil
}

// getListImages reads a list of image paths from a text file and returns
// a slice of paths. Basically just returns each line as a path.
func getListImages(src string) ([]Image, error) {
	// Remove the "list://" prefix from the source string to get the file path.
	if !strings.HasPrefix(src, SourceList.String()) {
		return nil, fmt.Errorf("invalid source format")
	}
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
	if !strings.HasPrefix(src, SourceSSH.String()) {
		return nil, fmt.Errorf("invalid source format")
	}

	source := strings.TrimPrefix(src, SourceSSH.String())
	addrParts := strings.Split(source, ":")
	if len(addrParts) != 2 {
		return nil, fmt.Errorf("invalid source format")
	}
	addr := addrParts[0]
	path := addrParts[1]

	// Test connectivity
	ok, err := isSSHAlive(addr)
	if err != nil {
		return nil, err
	}

	if !ok {
		return nil, fmt.Errorf("SSH connection to %s not successful", addr)
	}

	// Construct the SSH command.
	sshCmd := fmt.Sprintf("ssh %s ls %s", addr, path)

	// Run the SSH command.
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

func downloadSSHImage(src Image, dest string) (Image, error) {
	if !strings.HasPrefix(src.Source, SourceSSH.String()) {
		return Image{}, fmt.Errorf("invalid source format")
	}

	source := strings.TrimPrefix(src.Source, SourceSSH.String())
	addrParts := strings.Split(source, ":")
	if len(addrParts) != 2 {
		return Image{}, fmt.Errorf("invalid source format")
	}
	addr := addrParts[0]
	path := addrParts[1]

	// Construct the SSH command.
	sshCmd := fmt.Sprintf("scp %s:\"%s\" \"%s\"", addr, path, dest)

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

func ImageInList(image Image, list []Image) bool {
	for _, i := range list {
		if i.ShaSum == image.ShaSum {
			return true
		}
	}

	return false
}
