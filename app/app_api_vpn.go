package main

// VPN Control methods for Kampus VPN
// This file contains VPN start/stop/toggle operations

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
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

	// Get log level from settings and update config file
	logLevel := "info" // default - info
	if a.storage != nil {
		settings := a.storage.GetAppSettings()
		if settings.LogLevel != "" {
			logLevel = string(settings.LogLevel)
		}
	}
	
	// Update log level in config file
	if err := a.updateConfigLogLevel(configPath, logLevel); err != nil {
		a.writeLog(fmt.Sprintf("Warning: could not update log level in config: %v", err))
	}

	a.writeLog(fmt.Sprintf("Starting sing-box: %s", a.singboxPath))
	a.writeLog(fmt.Sprintf("Config: %s", configPath))
	a.writeLog(fmt.Sprintf("Log level: %s", logLevel))

	// Start sing-box with config for current profile
	a.cmd = exec.Command(a.singboxPath, "run", "-c", configPath)

	// WireGuard is now handled by Native WireGuard Manager, not sing-box
	// No need for ENABLE_DEPRECATED_WIREGUARD_OUTBOUND

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

	// Start Native WireGuard tunnels (internal/corporate VPNs)
	if a.nativeWG != nil && a.nativeWG.IsInstalled() {
		a.startNativeWireGuardTunnels()
	}

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
	a.writeLog(fmt.Sprintf("[%s] Log reader started", prefix))
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

		// Check for critical errors only (not DNS resolution failures)
		lineLower := strings.ToLower(line)
		isCriticalError := (strings.Contains(lineLower, "error") || strings.Contains(lineLower, "fatal")) &&
			// Ignore DNS resolution errors for local/internal domains
			!strings.Contains(lineLower, "dns: exchange failed") &&
			!strings.Contains(lineLower, "context deadline exceeded") &&
			// Ignore connection refused (normal when server is down)
			!strings.Contains(lineLower, "connection refused") &&
			// Ignore i/o timeout (normal network fluctuation)
			!strings.Contains(lineLower, "i/o timeout")
		
		if isCriticalError {
			a.mu.Lock()
			a.hasError = true
			a.mu.Unlock()
			UpdateTrayIcon("error")
		}
	}
	if err := scanner.Err(); err != nil {
		a.writeLog(fmt.Sprintf("[%s] Log reader error: %v", prefix, err))
	} else {
		a.writeLog(fmt.Sprintf("[%s] Log reader finished", prefix))
	}
}

// Stop stops VPN
func (a *App) Stop() map[string]interface{} {
	a.mu.Lock()
	defer a.mu.Unlock()

	if !a.isRunning || a.cmd == nil || a.cmd.Process == nil {
		a.isRunning = false
		a.stoppedManually = false
		// Also stop Native WireGuard tunnels
		a.stopNativeWireGuardTunnels()
		UpdateTrayIcon("disconnected")
		return map[string]interface{}{
			"success": true,
		}
	}

	a.writeLog("Stopping VPN...")

	// Stop Native WireGuard tunnels first
	a.stopNativeWireGuardTunnels()

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

// updateConfigLogLevel updates the log level in the config file
func (a *App) updateConfigLogLevel(configPath, logLevel string) error {
	// Read config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read config: %w", err)
	}

	// Parse JSON
	var config map[string]interface{}
	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to parse config: %w", err)
	}

	// Update log level
	if logSection, ok := config["log"].(map[string]interface{}); ok {
		logSection["level"] = logLevel
	} else {
		// Create log section if not exists
		config["log"] = map[string]interface{}{
			"level": logLevel,
		}
	}

	// Write back
	newData, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, newData, 0644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

// startNativeWireGuardTunnels starts all configured Native WireGuard tunnels
func (a *App) startNativeWireGuardTunnels() {
	a.writeLog("[WireGuard] startNativeWireGuardTunnels called")
	
	if a.nativeWG == nil {
		a.writeLog("[WireGuard] nativeWG is nil, skipping")
		return
	}
	
	if a.storage == nil {
		a.writeLog("[WireGuard] storage is nil, skipping")
		return
	}
	
	settings, err := a.storage.GetUserSettings()
	if err != nil {
		a.writeLog(fmt.Sprintf("[WireGuard] Error getting user settings: %v", err))
		return
	}
	
	a.writeLog(fmt.Sprintf("[WireGuard] Found %d WireGuard config(s)", len(settings.WireGuardConfigs)))
	
	if len(settings.WireGuardConfigs) == 0 {
		a.writeLog("[WireGuard] No WireGuard configs found, skipping")
		return
	}
	
	a.writeLog(fmt.Sprintf("Starting %d Native WireGuard tunnel(s)...", len(settings.WireGuardConfigs)))
	
	started := 0
	for i, wg := range settings.WireGuardConfigs {
		a.writeLog(fmt.Sprintf("[WireGuard] Processing config %d: tag=%s, name=%s, endpoint=%s, allowedIPs=%v", 
			i, wg.Tag, wg.Name, wg.Endpoint, wg.AllowedIPs))
		
		nativeConfig := wg.ToWireGuardConfig()
		a.writeLog(fmt.Sprintf("[WireGuard] Native config: Address=%v, DNS=%s, Peers=%d", 
			nativeConfig.Address, nativeConfig.DNS, len(nativeConfig.Peers)))
		
		if err := a.nativeWG.StartTunnel(i, nativeConfig); err != nil {
			a.writeLog(fmt.Sprintf("[WireGuard] Failed to start %s: %v", wg.Tag, err))
			a.AddToLogBuffer(fmt.Sprintf("WireGuard %s: ошибка запуска", wg.Name))
		} else {
			started++
			a.AddToLogBuffer(fmt.Sprintf("WireGuard %s: подключен", wg.Name))
		}
	}
	
	if started > 0 {
		a.writeLog(fmt.Sprintf("[WireGuard] Started %d/%d tunnels", started, len(settings.WireGuardConfigs)))
	}
}

// stopNativeWireGuardTunnels stops all Native WireGuard tunnels
func (a *App) stopNativeWireGuardTunnels() {
	if a.nativeWG == nil {
		return
	}
	
	a.writeLog("Stopping Native WireGuard tunnels...")
	a.nativeWG.StopAllTunnels()
	a.writeLog("Native WireGuard tunnels stopped")
}
