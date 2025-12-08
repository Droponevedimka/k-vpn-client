package main

// App settings methods for Kampus VPN (App methods extension)
// This file contains app configuration API methods

import (
	"fmt"
	"os"
	"time"
	
	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// GetAppConfig возвращает текущие настройки приложения (API для фронтенда)
func (a *App) GetAppConfig() map[string]interface{} {
	a.waitForInit()
	
	if a.storage == nil {
		return map[string]interface{}{
			"success": false,
			"error":   "Хранилище не инициализировано",
		}
	}
	
	settings := a.storage.GetAppSettings()
	
	return map[string]interface{}{
		"success":           true,
		"autoStart":         settings.AutoStart,
		"enableLogging":     settings.EnableLogging,
		"checkUpdates":      settings.CheckUpdates,
		"notifications":     settings.Notifications,
		"theme":             settings.Theme,
		"language":          settings.Language,
		"logLevel":          settings.LogLevel,
		"autoUpdateSub":     settings.AutoUpdateSub,
		"subUpdateInterval": settings.SubUpdateInterval,
		"lastSubUpdate":     settings.LastSubUpdate.Format(time.RFC3339),
		"wireGuardVersion":  settings.WireGuardVersion,
		"appVersion":        Version,
		"appName":           AppName,
		"singboxVersion":    SingBoxVersion,
		"buildHash":         BuildHash,
		"buildTime":         BuildTime,
	}
}

// SaveAppConfig сохраняет настройки приложения (API для фронтенда)
func (a *App) SaveAppConfig(autoStart, enableLogging, checkUpdates, notifications, autoUpdateSub bool, theme, language, logLevel string, subUpdateInterval int) map[string]interface{} {
	a.waitForInit()
	
	if a.storage == nil {
		return map[string]interface{}{
			"success": false,
			"error":   "Хранилище не инициализировано",
		}
	}
	
	settings := a.storage.GetAppSettings()
	
	// Обновляем настройки
	settings.AutoStart = autoStart
	settings.EnableLogging = enableLogging
	settings.CheckUpdates = checkUpdates
	settings.Notifications = notifications
	settings.AutoUpdateSub = autoUpdateSub
	settings.Theme = Theme(theme)
	settings.Language = Language(language)
	settings.SubUpdateInterval = subUpdateInterval
	
	// Обновляем уровень логирования
	if logLevel != "" {
		settings.LogLevel = LogLevel(logLevel)
	}
	
	// Сохраняем в storage
	if err := a.storage.UpdateAppSettings(settings); err != nil {
		return map[string]interface{}{
			"success": false,
			"error":   fmt.Sprintf("Ошибка сохранения настроек: %v", err),
		}
	}
	
	// Применяем автозапуск
	if err := SetAutoStart(autoStart); err != nil {
		return map[string]interface{}{
			"success": false,
			"error":   fmt.Sprintf("Ошибка настройки автозапуска: %v", err),
		}
	}
	
	return map[string]interface{}{
		"success": true,
		"message": "Настройки сохранены",
	}
}

// GetWireGuardVersion returns current WireGuard version (bundled with app)
func (a *App) GetWireGuardVersion() map[string]interface{} {
	installed := false
	wireguardPath := ""
	
	if a.nativeWG != nil {
		installed = a.nativeWG.IsInstalled()
		wireguardPath = a.nativeWG.wireguardPath
	}
	
	return map[string]interface{}{
		"success":       true,
		"version":       WireGuardVersion,
		"wintunVersion": WintunVersion,
		"installed":     installed,
		"wireguardPath": wireguardPath,
	}
}

// GetAutoStartStatus проверяет статус автозапуска
func (a *App) GetAutoStartStatus() map[string]interface{} {
	return map[string]interface{}{
		"success":   true,
		"autoStart": IsAutoStartEnabled(),
	}
}

// ============================================================================
// Import/Export API methods
// ============================================================================

// ExportProfilesToFile opens save dialog and exports all profiles to JSON file.
func (a *App) ExportProfilesToFile() map[string]interface{} {
	a.waitForInit()
	
	// Get export data first
	exportResult := a.ExportAllProfiles()
	if !exportResult["success"].(bool) {
		return exportResult
	}
	
	jsonData := exportResult["data"].(string)
	
	// Open save dialog
	filename, err := wailsRuntime.SaveFileDialog(a.ctx, wailsRuntime.SaveDialogOptions{
		Title:           "Экспорт профилей",
		DefaultFilename: fmt.Sprintf("kampus-vpn-profiles-%s.json", time.Now().Format("2006-01-02")),
		Filters: []wailsRuntime.FileFilter{
			{
				DisplayName: "JSON файлы (*.json)",
				Pattern:     "*.json",
			},
		},
	})
	
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"error":   fmt.Sprintf("Ошибка диалога сохранения: %v", err),
		}
	}
	
	if filename == "" {
		// User cancelled
		return map[string]interface{}{
			"success": false,
			"error":   "Отменено пользователем",
		}
	}
	
	// Write to file
	if err := os.WriteFile(filename, []byte(jsonData), 0644); err != nil {
		return map[string]interface{}{
			"success": false,
			"error":   fmt.Sprintf("Ошибка записи файла: %v", err),
		}
	}
	
	profilesCount := exportResult["profiles_count"].(int)
	
	a.writeLog(fmt.Sprintf("Exported %d profiles to %s", profilesCount, filename))
	a.AddToLogBuffer(fmt.Sprintf("Экспортировано %d профилей", profilesCount))
	
	return map[string]interface{}{
		"success":        true,
		"message":        fmt.Sprintf("Экспортировано %d профилей", profilesCount),
		"filename":       filename,
		"profiles_count": profilesCount,
	}
}

// ImportProfilesFromFile opens file dialog and imports profiles from JSON file.
func (a *App) ImportProfilesFromFile() map[string]interface{} {
	a.waitForInit()
	
	// Check VPN is not running
	a.mu.Lock()
	if a.isRunning {
		a.mu.Unlock()
		return map[string]interface{}{
			"success": false,
			"error":   "Нельзя импортировать пока VPN активен. Сначала отключите VPN.",
		}
	}
	a.mu.Unlock()
	
	// Open file dialog
	filename, err := wailsRuntime.OpenFileDialog(a.ctx, wailsRuntime.OpenDialogOptions{
		Title: "Импорт профилей",
		Filters: []wailsRuntime.FileFilter{
			{
				DisplayName: "JSON файлы (*.json)",
				Pattern:     "*.json",
			},
		},
	})
	
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"error":   fmt.Sprintf("Ошибка диалога открытия: %v", err),
		}
	}
	
	if filename == "" {
		// User cancelled
		return map[string]interface{}{
			"success": false,
			"error":   "Отменено пользователем",
		}
	}
	
	// Read file
	data, err := os.ReadFile(filename)
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"error":   fmt.Sprintf("Ошибка чтения файла: %v", err),
		}
	}
	
	// Validate first
	validationResult := a.ValidateImportData(string(data))
	if !validationResult["success"].(bool) {
		return validationResult
	}
	
	// Return validation info for user confirmation
	validationResult["filename"] = filename
	validationResult["file_data"] = string(data)
	validationResult["needs_confirmation"] = true
	
	return validationResult
}

// ConfirmImportProfiles confirms and executes import after user approval.
func (a *App) ConfirmImportProfiles(jsonData string) map[string]interface{} {
	return a.ImportAllProfiles(jsonData)
}
