package util

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sort"
	"strconv"

	"github.com/charmbracelet/log"
)

func Contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func IsNumber(s string) bool {
	_, err := strconv.Atoi(s)
	return err == nil
}

// OpenFile opens a file and returns its contents as a byte slice.
func OpenFile(filename string) ([]byte, error) {
	// #nosec G304
	f, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	// Read file into byte slice
	data := make([]byte, 0)
	buf := make([]byte, 1024)
	for {
		n, err := f.Read(buf)
		if err != nil {
			break
		}
		data = append(data, buf[:n]...)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return data, nil
}

// WriteFile writes a byte slice to a file.
func WriteFile(filename string, data []byte) error {
	if data == nil {
		return fmt.Errorf("data is nil")
	}

	if filename == "" {
		return fmt.Errorf("filename is empty")
	}

	// #nosec G304
	f, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer f.Close()

	_, err = f.Write(data)
	if err != nil {
		return fmt.Errorf("failed to write to file: %w", err)
	}

	return nil
}

// FileExists returns true if the file exists at the given path.
func FileExists(filename string) bool {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return false
	}

	return true
}

func MkDir(path string) error {
	if !FileExists(path) {
		err := os.MkdirAll(path, 0755)
		if err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}
	}

	return nil
}

// RunCmd executes the given command and returns its output as a string.
func RunCmd(cmd string) (string, error) {
	// Create an exec.Command object to represent the command.
	command := exec.Command("sh", "-c", cmd)

	// Use a buffer to capture the standard output and error.
	var out bytes.Buffer
	var stderr bytes.Buffer
	command.Stdout = &out
	command.Stderr = &stderr

	// Run the command.
	log.Debugf("Running command: %s", cmd)
	err := command.Run()
	if err != nil {
		return "", err
	}

	// Log stderr
	if stderr.Len() > 0 {
		fmt.Println("stderr:", stderr.String())
	}

	// Return the output as a string.
	return out.String(), nil
}

func Sha256(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	hash := sha256.New()

	_, err = io.Copy(hash, file)
	if err != nil {
		return "", fmt.Errorf("failed to copy file content to hash: %w", err)
	}

	sum := hash.Sum(nil)

	return hex.EncodeToString(sum), nil
}

func SortFilesByMTime(files []os.DirEntry) {
	sort.Slice(files, func(i, j int) bool {
		infoI, err := files[i].Info()
		if err != nil {
			return false
		}
		infoJ, err := files[j].Info()
		if err != nil {
			return false
		}
		return infoI.ModTime().Before(infoJ.ModTime())
	})
}
