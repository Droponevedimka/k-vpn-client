package main

// UI and Window methods for Kampus VPN
// This file contains window management and UI operations

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// Quit closes the application (called from UI)
func (a *App) Quit() {
	a.Stop()
	wailsRuntime.Quit(a.ctx)
}

// QuitApp closes the application (alias)
func (a *App) QuitApp() {
	a.Stop()
	if a.ctx != nil {
		wailsRuntime.Quit(a.ctx)
	}
	os.Exit(0)
}

// ShowWindow shows the application window
func (a *App) ShowWindow() {
	if a.ctx != nil {
		wailsRuntime.WindowShow(a.ctx)
		a.SetWindowVisible(true)
	}
}

// ShowAbout shows about dialog
func (a *App) ShowAbout() {
	if a.ctx != nil {
		version := a.GetVersion()
		wailsRuntime.MessageDialog(a.ctx, wailsRuntime.MessageDialogOptions{
			Type:    wailsRuntime.InfoDialog,
			Title:   "О программе Kampus VPN",
			Message: fmt.Sprintf("Версия: %s\nБесплатный VPN клиент на базе sing-box", version),
		})
	}
}

// HideWindow hides the application window
func (a *App) HideWindow() {
	if a.ctx != nil {
		wailsRuntime.WindowHide(a.ctx)
		a.SetWindowVisible(false)
	}
}

// OpenConfigFolder opens the config folder in file explorer
func (a *App) OpenConfigFolder() {
	var configDir string
	switch runtime.GOOS {
	case "windows":
		configDir = filepath.Dir(a.configPath)
		if configDir == "" {
			configDir = filepath.Join(os.Getenv("LOCALAPPDATA"), "KampusVPN")
		}
	case "darwin":
		home, _ := os.UserHomeDir()
		configDir = filepath.Join(home, "Library", "Application Support", "KampusVPN")
	default:
		home, _ := os.UserHomeDir()
		configDir = filepath.Join(home, ".config", "kampusvpn")
	}

	openFolder(configDir)
}

// OpenLogs opens the logs folder in file explorer
func (a *App) OpenLogs() {
	var logDir string
	switch runtime.GOOS {
	case "windows":
		logDir = filepath.Join(os.Getenv("LOCALAPPDATA"), "KampusVPN", "logs")
	case "darwin":
		home, _ := os.UserHomeDir()
		logDir = filepath.Join(home, "Library", "Logs", "KampusVPN")
	default:
		home, _ := os.UserHomeDir()
		logDir = filepath.Join(home, ".local", "share", "kampusvpn", "logs")
	}

	// Create logs folder if it doesn't exist
	os.MkdirAll(logDir, 0755)

	openFolder(logDir)
}

// openFolder opens a folder in the system file manager
func openFolder(path string) {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("explorer", path)
	case "darwin":
		cmd = exec.Command("open", path)
	default:
		cmd = exec.Command("xdg-open", path)
	}

	cmd.Start()
}

// GetVersion returns application version
func (a *App) GetVersion() string {
	return AppVersion
}

// GetSingBoxInfo returns sing-box information
func (a *App) GetSingBoxInfo() map[string]interface{} {
	result := map[string]interface{}{
		"found":   false,
		"path":    "",
		"version": "",
	}

	if a.singboxPath != "" && fileExists(a.singboxPath) {
		result["found"] = true
		result["path"] = a.singboxPath
	}

	return result
}

// SetWindowVisible sets window visibility flag (for ping optimization)
func (a *App) SetWindowVisible(visible bool) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.windowVisible = visible
}

// IsWindowVisible returns window visibility flag
func (a *App) IsWindowVisible() bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.windowVisible
}
