package logger

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
)

type Logger struct {
	*log.Logger
}

const TraceLevel = log.DebugLevel - 1

func (l *Logger) Trace(msg string, args ...any) {
	l.Log(TraceLevel, msg, args...)
}

func Setup(level, file string) (*Logger, *os.File, error) {
	var err error
	logH := os.Stderr
	if file != "" {
		logH, err = os.OpenFile(file, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to open log file: %w", err)
		}
	}

	if level == "" {
		logH, err = os.Open(os.DevNull)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to open log file: %w", err)
		}
	}

	l := new(Logger)
	logger := log.New(logH)

	styles := log.DefaultStyles()
	styles.Levels[TraceLevel] = lipgloss.NewStyle().
		SetString("TRACE").
		Bold(true).
		MaxWidth(4).
		Foreground(lipgloss.Color("61"))
	styles.Caller = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	styles.Timestamp = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	styles.Key = lipgloss.NewStyle().Foreground(lipgloss.Color("246"))
	styles.Value = lipgloss.NewStyle().Foreground(lipgloss.Color("250"))

	logger.SetStyles(styles)
	logger.SetOutput(logH)
	logger.SetReportTimestamp(true)

	logLevel, err := parseLevel(level)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse log level: %w", err)
	}
	logger.SetLevel(logLevel)

	if logLevel == log.DebugLevel || logLevel == TraceLevel {
		logger.SetReportCaller(true)
	}

	l.Logger = logger
	log.SetDefault(logger)

	return l, logH, nil
}

func parseLevel(level string) (log.Level, error) {
	if level == strings.ToLower("trace") {
		return TraceLevel, nil
	}

	logLevel, err := log.ParseLevel(level)
	if err != nil {
		return log.FatalLevel, fmt.Errorf("could not parse log level: %w", err)
	}

	return logLevel, nil
}
