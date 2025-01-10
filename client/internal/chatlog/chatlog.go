package chatlog

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
)

const FilePermissions = 0664

var (
	ErrOpenLogFile = errors.New("failed to open log file")
	ErrWriteLog    = errors.New("failed to write to log file")
	ErrReadLog     = errors.New("failed to read log file")
)

type ChatLogger struct {
	mu      sync.Mutex
	logPath string
}

func NewChatLogger(logPath string) *ChatLogger {
	return &ChatLogger{logPath: logPath}
}

func (l *ChatLogger) SaveMessage(message string) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	file, err := os.OpenFile(l.logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, FilePermissions)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrOpenLogFile, err)
	}
	defer file.Close()

	_, err = file.WriteString(message + "\n")
	if err != nil {
		return fmt.Errorf("%w: %w", ErrWriteLog, err)
	}

	return nil
}

func (l *ChatLogger) GetLastMessages(count int) ([]string, error) {
	data, err := os.ReadFile(l.logPath)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrReadLog, err)
	}

	lines := strings.Split(string(data), "\n")
	if len(lines) > count {
		lines = lines[len(lines)-count:]
	}

	return lines, nil
}
