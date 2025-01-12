package chatlog

import (
	"fmt"
	"os"
	"strings"
	"sync"
)

const (
	filePermissions = 0664
)

type ChatLogger struct {
	mu      sync.RWMutex
	logFile *os.File
	logPath string
}

func NewChatLogger(logPath string) (*ChatLogger, error) {
	file, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, filePermissions)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	return &ChatLogger{
		logFile: file,
		logPath: logPath,
	}, nil
}

func (l *ChatLogger) SaveMessage(message string) error {
	if strings.TrimSpace(message) == "" {
		return fmt.Errorf("cannot save empty message")
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	if _, err := l.logFile.WriteString(message + "\n"); err != nil {
		return fmt.Errorf("failed to write message to log: %w", err)
	}

	return nil
}

func (l *ChatLogger) GetLastMessages(count int) ([]string, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	data, err := os.ReadFile(l.logPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read log file: %w", err)
	}

	lines := strings.Split(string(data), "\n")
	lines = filterEmptyLines(lines)

	if len(lines) > count {
		lines = lines[len(lines)-count:]
	}

	return lines, nil
}

func (l *ChatLogger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if err := l.logFile.Close(); err != nil {
		return fmt.Errorf("failed to close log file: %w", err)
	}

	return nil
}

func filterEmptyLines(lines []string) []string {
	var filtered []string
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			filtered = append(filtered, line)
		}
	}
	return filtered
}
