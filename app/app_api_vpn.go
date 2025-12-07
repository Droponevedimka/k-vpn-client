package main

// VPN Control methods for Kampus VPN
// This file contains VPN start/stop/toggle operations

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
	"runtime"
	"strings"
	"syscall"

	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// getActiveConfigPath writes active config to file and returns the path.
// This is needed because sing-box requires a file path, but we store configs in settings.json.
func (a *App) getActiveConfigPath() (string, error) {
	if a.storage == nil {
		return "", fmt.Errorf("storage not initialized")
	}
	return a.storage.WriteActiveConfigToFile()
}

// GetStatus returns current VPN status
func (a *App) GetStatus() map[string]interface{} {
	// Wait for initialization if not completed
	a.waitForInit()

	a.mu.Lock()
	defer a.mu.Unlock()

	configPath, _ := a.getActiveConfigPath()
	hasConfig := configPath != "" && fileExists(configPath)
	
	return map[string]interface{}{
		"running":       a.isRunning,
		"hasError":      a.hasError,
		"configPath":    configPath,
		"singboxPath":   a.singboxPath,
		"configExists":  hasConfig,
		"singboxExists": a.singboxPath != "" && fileExists(a.singboxPath),
		"logPath":       a.logPath,
	}
}

// Start starts VPN
func (a *App) Start() map[string]interface{} {
	// Wait for initialization
	a.waitForInit()

	a.mu.Lock()
	defer a.mu.Unlock()

	if a.isRunning {
		return map[string]interface{}{
			"success": false,
			"error":   "VPN уже запущен",
		}
	}

	if a.singboxPath == "" || !fileExists(a.singboxPath) {
		a.hasError = true
		UpdateTrayIcon("error")
		return map[string]interface{}{
			"success": false,
			"error":   "sing-box не найден. Установите sing-box.",
		}
	}

	configPath, err := a.getActiveConfigPath()
	if err != nil || configPath == "" {
		a.hasError = true
		UpdateTrayIcon("error")
		return map[string]interface{}{
			"success": false,
			"error":   "Конфиг не найден. Добавьте подписку для текущего профиля.",
		}
	}

	// Open log file
	if err := a.openLogFile(); err != nil {
		a.writeLog(fmt.Sprintf("Warning: could not open log file: %v", err))
	}

	a.writeLog(fmt.Sprintf("Starting sing-box: %s", a.singboxPath))
	a.writeLog(fmt.Sprintf("Config: %s", configPath))

	// Start sing-box with config for current profile
	a.cmd = exec.Command(a.singboxPath, "run", "-c", configPath)

	// Get stdout and stderr for logging
	stdout, _ := a.cmd.StdoutPipe()
	stderr, _ := a.cmd.StderrPipe()

	// Hide console window on Windows
	if runtime.GOOS == "windows" {
		a.cmd.SysProcAttr = &syscall.SysProcAttr{
			HideWindow:    true,
			CreationFlags: 0x08000000, // CREATE_NO_WINDOW
		}
	}

	// Set working directory to resources folder
	if a.storage != nil {
		a.cmd.Dir = a.storage.GetResourcesPath()
	} else {
		a.cmd.Dir = a.basePath
	}

	if err := a.cmd.Start(); err != nil {
		a.hasError = true
		UpdateTrayIcon("error")
		a.writeLog(fmt.Sprintf("ERROR: Failed to start: %v", err))
		return map[string]interface{}{
			"success": false,
			"error":   fmt.Sprintf("Ошибка запуска: %v", err),
		}
	}

	a.isRunning = true
	a.hasError = false
	UpdateTrayIcon("connected")
	a.writeLog("VPN started successfully")
	a.AddToLogBuffer("VPN запущен")

	// Start tracking traffic statistics
	if a.trafficStats != nil {
		a.trafficStats.StartSession()
	}

	// Log output in goroutines
	go a.logOutput(stdout, "OUT")
	go a.logOutput(stderr, "ERR")

	// Monitor process in goroutine
	go func() {
		err := a.cmd.Wait()
		a.mu.Lock()
		wasStoppedManually := a.stoppedManually
		a.isRunning = false
		a.stoppedManually = false

		// End traffic session
		if a.trafficStats != nil {
			a.trafficStats.EndSession()
			a.trafficStats.Save()
		}

		if wasStoppedManually {
			// Manual stop - not an error
			a.writeLog("VPN stopped by user")
			a.AddToLogBuffer("VPN остановлен пользователем")
			UpdateTrayIcon("disconnected")
		} else if err != nil {
			a.hasError = true
			a.writeLog(fmt.Sprintf("VPN process exited with error: %v", err))
			a.AddToLogBuffer(fmt.Sprintf("VPN завершился с ошибкой: %v", err))
			UpdateTrayIcon("error")
		} else {
			a.writeLog("VPN process exited normally")
			a.AddToLogBuffer("VPN завершил работу")
			UpdateTrayIcon("disconnected")
		}
		a.closeLogFile()
		a.mu.Unlock()
		// Notify frontend about status change
		wailsRuntime.EventsEmit(a.ctx, "vpn-status-changed", false)
	}()

	return map[string]interface{}{
		"success": true,
	}
}

// logOutput reads and logs process output
func (a *App) logOutput(reader io.Reader, prefix string) {
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := scanner.Text()

		// Check logging setting from storage
		loggingEnabled := true
		if a.storage != nil {
			settings := a.storage.GetAppSettings()
			loggingEnabled = settings.EnableLogging
		}

		if loggingEnabled {
			a.writeLog(fmt.Sprintf("[%s] %s", prefix, line))
		}

		// Add to log buffer for UI (always)
		a.AddToLogBuffer(fmt.Sprintf("[%s] %s", prefix, line))

		// Check for errors
		if strings.Contains(strings.ToLower(line), "error") ||
			strings.Contains(strings.ToLower(line), "fatal") {
			a.mu.Lock()
			a.hasError = true
			a.mu.Unlock()
			UpdateTrayIcon("error")
		}
	}
}

// Stop stops VPN
func (a *App) Stop() map[string]interface{} {
	a.mu.Lock()
	defer a.mu.Unlock()

	if !a.isRunning || a.cmd == nil || a.cmd.Process == nil {
		a.isRunning = false
		a.stoppedManually = false
		UpdateTrayIcon("disconnected")
		return map[string]interface{}{
			"success": true,
		}
	}

	a.writeLog("Stopping VPN...")

	// Set manual stop flag BEFORE terminating process
	a.stoppedManually = true

	// Terminate process
	if runtime.GOOS == "windows" {
		// On Windows use taskkill for proper termination
		// Hide console window
		killCmd := exec.Command("taskkill", "/F", "/T", "/PID", fmt.Sprintf("%d", a.cmd.Process.Pid))
		killCmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
		killCmd.Run()
	} else {
		// On Unix send SIGTERM
		a.cmd.Process.Signal(syscall.SIGTERM)
	}

	a.hasError = false
	// DO NOT set isRunning = false here, goroutine will do it
	// DO NOT call UpdateTrayIcon here, goroutine will do it

	return map[string]interface{}{
		"success": true,
	}
}

// Toggle toggles VPN state
func (a *App) Toggle() map[string]interface{} {
	if a.isRunning {
		return a.Stop()
	}
	return a.Start()
}

// CanModifyVPN checks if VPN settings can be modified
func (a *App) CanModifyVPN() map[string]interface{} {
	a.mu.Lock()
	defer a.mu.Unlock()

	return map[string]interface{}{
		"canModify": !a.isRunning,
		"message":   "Сначала отключите VPN для изменения настроек",
	}
}
