package util

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"syscall"
	"time"

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

// IsFilePath checks if the given path is a valid file path.
// It supports both absolute and relative paths and replaces environment variables.
func IsFilePath(path string) bool {
	// Replace environment variables in the path
	path = os.ExpandEnv(path)

	// Get the absolute path
	absPath, err := filepath.Abs(path)
	if err != nil {
		return false
	}

	log.Debugf("Checking if path is a file: %s", absPath)

	_, err = os.Stat(absPath)
	return !os.IsNotExist(err)
}

func MkDir(path string) error {
	if !FileExists(path) {
		err := os.MkdirAll(path, 0o755)
		if err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}
	}

	return nil
}

// RunCmd executes the given command and returns its output as a string.
func RunCmd(cmd string) (string, error) {
	// Assume everything we run is low priority.
	const priority = 0

	// Create an exec.Command object to represent the command.
	command := exec.Command("sh", "-c", cmd)

	// Use a buffer to capture the standard output and error.
	var out bytes.Buffer
	var stderr bytes.Buffer
	command.Stdout = &out
	command.Stderr = &stderr

	// Run the command.
	log.Debug("running command", "cmd", cmd)
	if err := command.Start(); err != nil {
		return "", fmt.Errorf("failed to start command: %w", err)
	}

	// Set the process priority to low.
	if err := syscall.Setpriority(syscall.PRIO_PROCESS, command.Process.Pid, priority); err != nil {
		return "", fmt.Errorf("failed to set process priority: %w", err)
	}

	// Wait for the command to finish.
	if err := command.Wait(); err != nil {
		return "", fmt.Errorf("failed to wait for command: %w", err)
	}

	// Log stderr
	if stderr.Len() > 0 {
		log.Warnf("stderr: %s", stderr.String())
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

func Retry(maxRetries int, retryInterval time.Duration, operation func() error) error {
	var err error
	for i := 0; i < maxRetries; i++ {
		err = operation()
		if err == nil {
			return nil
		}
		log.Errorf("Error encountered: %s. Retrying in %v...", err, retryInterval)
		time.Sleep(retryInterval)
	}
	return err
}
