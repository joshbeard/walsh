package cli

import (
	"fmt"
	"os"
	"strings"
	"unicode/utf8"

	"github.com/joshbeard/walsh/internal/config"
	"github.com/joshbeard/walsh/internal/session"
	"github.com/joshbeard/walsh/internal/util"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh/terminal"
)

func Setup(cmd *cobra.Command, args []string) (string, *session.Session, error) {
	// Load config
	cfg, err := config.Load("")
	if err != nil {
		return "", nil, fmt.Errorf("error loading config: %w", err)
	}

	// Create session
	sess := session.NewSession(cfg)

	displays := sess.Displays()
	display, _ := cmd.Flags().GetString("display")
	if len(args) > 0 && display == "" {
		// If the argument is a digit, assume it's a display. If it's a display
		// name, use it. Otherwise, assume it's a source.
		matchName := false
		for _, d := range displays {
			if d.Name == args[0] {
				matchName = true
				break
			}
		}

		if matchName || util.IsNumber(args[0]) {
			display = args[0]
		}
	}

	return display, sess, nil
}

// GetTerminalWidth returns the current terminal width.
func GetTerminalWidth() (int, error) {
	width, _, err := terminal.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		return 80, err // default width if error occurs
	}
	if width > 80 {
		width = 80
	}
	return width, nil
}

// WrapText wraps the input text to fit within a given width, preserving newlines.
func WrapText(text string, width int) []string {
	var lines []string
	for _, paragraph := range strings.Split(text, "\n") {
		if utf8.RuneCountInString(paragraph) == 0 {
			lines = append(lines, "")
			continue
		}
		var currentLine string
		for _, word := range strings.Fields(paragraph) {
			if utf8.RuneCountInString(currentLine)+utf8.RuneCountInString(word)+1 > width {
				lines = append(lines, currentLine)
				currentLine = word
			} else {
				if len(currentLine) > 0 {
					currentLine += " "
				}
				currentLine += word
			}
		}
		if len(currentLine) > 0 {
			lines = append(lines, currentLine)
		}
	}
	return lines
}

// Banner creates a banner with the given text.
func Banner(text string) (string, error) {
	width, err := GetTerminalWidth()
	if err != nil {
		return "", err
	}
	textWidth := width - 4 // for padding and borders
	wrappedText := WrapText(text, textWidth)
	borderTop := "┌" + strings.Repeat("─", width-2) + "┐"
	borderBottom := "└" + strings.Repeat("─", width-2) + "┘"
	var result strings.Builder

	result.WriteString(borderTop + "\n")
	for _, line := range wrappedText {
		padding := strings.Repeat(" ", textWidth-utf8.RuneCountInString(line))
		result.WriteString(fmt.Sprintf("│ %s%s │\n", line, padding))
	}
	result.WriteString(borderBottom)

	return result.String(), nil
}
