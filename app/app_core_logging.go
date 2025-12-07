package main

// Logging methods for Kampus VPN
// This file contains all logging-related operations

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

// setupLogPath sets up the log file path
func (a *App) setupLogPath() {
	var logDir string

	switch runtime.GOOS {
	case "windows":
		// %LOCALAPPDATA%\KampusVPN\logs
		logDir = filepath.Join(os.Getenv("LOCALAPPDATA"), "KampusVPN", "logs")
	case "darwin":
		// ~/Library/Logs/KampusVPN
		home, _ := os.UserHomeDir()
		logDir = filepath.Join(home, "Library", "Logs", "KampusVPN")
	default:
		// ~/.local/share/kampusvpn/logs
		home, _ := os.UserHomeDir()
		logDir = filepath.Join(home, ".local", "share", "kampusvpn", "logs")
	}

	os.MkdirAll(logDir, 0755)
	a.logPath = filepath.Join(logDir, "vpn.log")
}

// openLogFile opens log file with rotation
func (a *App) openLogFile() error {
	// Check existing file size and rotate if needed
	if err := a.rotateLogIfNeeded(); err != nil {
		// Not critical, continue
	}

	var err error
	a.logFile, err = os.OpenFile(a.logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	// Write session separator
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	a.logFile.WriteString(fmt.Sprintf("\n=== VPN Session Started: %s ===\n", timestamp))

	return nil
}

// rotateLogIfNeeded checks log size and truncates if needed
func (a *App) rotateLogIfNeeded() error {
	info, err := os.Stat(a.logPath)
	if err != nil {
		return nil // File doesn't exist - ok
	}

	if info.Size() < MaxLogSize {
		return nil // Size is ok
	}

	// Read last TruncateToSize bytes
	file, err := os.Open(a.logPath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Seek to position (size - TruncateToSize)
	offset := info.Size() - TruncateToSize
	if offset < 0 {
		offset = 0
	}
	file.Seek(offset, 0)

	// Skip first incomplete line
	reader := bufio.NewReader(file)
	reader.ReadString('\n')

	// Read remainder
	remainingData, err := io.ReadAll(reader)
	if err != nil {
		return err
	}

	// Rewrite file
	file.Close()
	err = os.WriteFile(a.logPath, remainingData, 0644)
	if err != nil {
		return err
	}

	// Add rotation marker
	marker := fmt.Sprintf("=== Log rotated at %s (old logs truncated) ===\n",
		time.Now().Format("2006-01-02 15:04:05"))
	f, _ := os.OpenFile(a.logPath, os.O_APPEND|os.O_WRONLY, 0644)
	if f != nil {
		f.WriteString(marker)
		f.Close()
	}

	return nil
}

// closeLogFile closes log file
func (a *App) closeLogFile() {
	if a.logFile != nil {
		timestamp := time.Now().Format("2006-01-02 15:04:05")
		a.logFile.WriteString(fmt.Sprintf("=== VPN Session Ended: %s ===\n", timestamp))
		a.logFile.Close()
		a.logFile = nil
	}
}

// writeLog writes to log file
func (a *App) writeLog(message string) {
	if a.logFile != nil {
		timestamp := time.Now().Format("15:04:05")
		a.logFile.WriteString(fmt.Sprintf("[%s] %s\n", timestamp, message))
	}
}

// AddToLogBuffer adds message to log buffer for UI
func (a *App) AddToLogBuffer(message string) {
	a.logBufferMu.Lock()
	defer a.logBufferMu.Unlock()

	// Limit buffer size
	if len(a.logBuffer) >= MaxLogBufferSize {
		a.logBuffer = a.logBuffer[100:] // Remove first 100 entries
	}

	timestamp := time.Now().Format("15:04:05")
	a.logBuffer = append(a.logBuffer, fmt.Sprintf("[%s] %s", timestamp, message))
}

// GetLogs returns logs from buffer (API for frontend)
func (a *App) GetLogs(lastN int) map[string]interface{} {
	a.logBufferMu.RLock()
	defer a.logBufferMu.RUnlock()

	if lastN <= 0 || lastN > len(a.logBuffer) {
		lastN = len(a.logBuffer)
	}

	// Return last N entries
	startIdx := len(a.logBuffer) - lastN
	if startIdx < 0 {
		startIdx = 0
	}

	logs := make([]string, lastN)
	copy(logs, a.logBuffer[startIdx:])

	return map[string]interface{}{
		"success": true,
		"logs":    logs,
		"total":   len(a.logBuffer),
	}
}

// ClearLogs clears log buffer
func (a *App) ClearLogs() map[string]interface{} {
	a.logBufferMu.Lock()
	defer a.logBufferMu.Unlock()

	a.logBuffer = make([]string, 0, MaxLogBufferSize)

	return map[string]interface{}{
		"success": true,
		"message": "Логи очищены",
	}
}
