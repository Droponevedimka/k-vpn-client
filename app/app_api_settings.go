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

// ============================================================================
// Routing Mode API methods
// ============================================================================

// GetRoutingMode returns current routing mode
func (a *App) GetRoutingMode() map[string]interface{} {
	a.waitForInit()
	
	if a.storage == nil {
		return map[string]interface{}{
			"success": false,
			"error":   "Хранилище не инициализировано",
		}
	}
	
	settings := a.storage.GetAppSettings()
	mode := settings.RoutingMode
	
	// Default to blocked_only if empty
	if mode == "" {
		mode = DefaultRoutingMode
	}
	
	// Get mode descriptions for UI
	modeDescriptions := map[string]string{
		string(RoutingModeBlockedOnly):   "Только заблокированные",
		string(RoutingModeExceptRussia):  "Всё кроме России",
		string(RoutingModeAllTraffic):    "Весь трафик",
	}
	
	return map[string]interface{}{
		"success":     true,
		"mode":        string(mode),
		"description": modeDescriptions[string(mode)],
		"modes": []map[string]string{
			{"value": string(RoutingModeBlockedOnly), "label": "Только заблокированные", "description": "Через VPN идут только заблокированные сайты (РКН + сервисы, блокирующие РФ). Минимальная нагрузка на VPN."},
			{"value": string(RoutingModeExceptRussia), "label": "Всё кроме России", "description": "Весь зарубежный трафик через VPN, российские сайты напрямую."},
			{"value": string(RoutingModeAllTraffic), "label": "Весь трафик", "description": "Весь трафик через VPN. Максимальная приватность, высокая нагрузка."},
		},
	}
}

// SetRoutingMode sets routing mode and rebuilds config
func (a *App) SetRoutingMode(mode string) map[string]interface{} {
	a.waitForInit()
	
	if a.storage == nil {
		return map[string]interface{}{
			"success": false,
			"error":   "Хранилище не инициализировано",
		}
	}
	
	// Validate mode
	routingMode := RoutingMode(mode)
	switch routingMode {
	case RoutingModeBlockedOnly, RoutingModeExceptRussia, RoutingModeAllTraffic:
		// Valid mode
	default:
		return map[string]interface{}{
			"success": false,
			"error":   fmt.Sprintf("Неизвестный режим маршрутизации: %s", mode),
		}
	}
	
	// Check if VPN is running
	a.mu.Lock()
	isRunning := a.isRunning
	a.mu.Unlock()
	
	if isRunning {
		return map[string]interface{}{
			"success": false,
			"error":   "Нельзя изменить режим пока VPN активен. Сначала отключите VPN.",
		}
	}
	
	// Update settings
	settings := a.storage.GetAppSettings()
	settings.RoutingMode = routingMode
	
	if err := a.storage.UpdateAppSettings(settings); err != nil {
		return map[string]interface{}{
			"success": false,
			"error":   fmt.Sprintf("Ошибка сохранения настроек: %v", err),
		}
	}
	
	// Update config builder
	if a.configBuilder != nil {
		a.configBuilder.SetRoutingMode(routingMode)
	}
	
	// Rebuild config for active profile
	if err := a.RebuildActiveProfileConfig(); err != nil {
		return map[string]interface{}{
			"success": false,
			"error":   fmt.Sprintf("Ошибка перестройки конфига: %v", err),
		}
	}
	
	a.writeLog(fmt.Sprintf("Routing mode changed to: %s", mode))
	
	return map[string]interface{}{
		"success": true,
		"message": "Режим маршрутизации изменён",
		"mode":    mode,
	}
}

// ============================================================================
// Filters API methods
// ============================================================================

// GetFiltersInfo returns information about bundled filters
func (a *App) GetFiltersInfo() map[string]interface{} {
	a.waitForInit()
	
	// Create filter manager pointing to bin/filters
	filterManager := NewFilterManager(a.basePath)
	
	info, err := filterManager.GetInfo()
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"error":   fmt.Sprintf("Ошибка получения информации о фильтрах: %v", err),
		}
	}
	
	files := filterManager.GetFilterFiles()
	
	return map[string]interface{}{
		"success":        true,
		"version":        info.Version,
		"updated_at":     info.UpdatedAt,
		"days_old":       info.DaysOld,
		"max_age_days":   info.MaxAgeDays,
		"is_outdated":    info.IsOutdated,
		"filter_count":   info.FilterCount,
		"total_size_kb":  info.TotalSizeKB,
		"update_message": info.UpdateMessage,
		"can_update":     info.CanUpdate,
		"files":          files,
	}
}

// UpdateFilters downloads latest Re:filter rule-sets
func (a *App) UpdateFilters() map[string]interface{} {
	a.waitForInit()
	
	// Check if VPN is running
	a.mu.Lock()
	isRunning := a.isRunning
	a.mu.Unlock()
	
	if isRunning {
		return map[string]interface{}{
			"success": false,
			"error":   "Нельзя обновить фильтры пока VPN активен. Сначала отключите VPN.",
		}
	}
	
	// Create filter manager
	filterManager := NewFilterManager(a.basePath)
	
	a.writeLog("Updating Re:filter rule-sets...")
	a.AddToLogBuffer("Обновление фильтров...")
	
	updated, err := filterManager.UpdateRefilters()
	if err != nil {
		a.AddToLogBuffer(fmt.Sprintf("Ошибка обновления: %v", err))
		return map[string]interface{}{
			"success": false,
			"error":   fmt.Sprintf("Ошибка обновления фильтров: %v", err),
		}
	}
	
	if updated == 0 {
		return map[string]interface{}{
			"success": false,
			"error":   "Не удалось обновить ни один фильтр",
		}
	}
	
	// Rebuild config if in blocked_only mode
	settings := a.storage.GetAppSettings()
	if settings.RoutingMode == RoutingModeBlockedOnly {
		if err := a.RebuildActiveProfileConfig(); err != nil {
			a.writeLog(fmt.Sprintf("Warning: Failed to rebuild config after filter update: %v", err))
		}
	}
	
	a.writeLog(fmt.Sprintf("Updated %d filter files", updated))
	a.AddToLogBuffer(fmt.Sprintf("Обновлено %d файлов фильтров", updated))
	
	// Return fresh info
	info, _ := filterManager.GetInfo()
	
	return map[string]interface{}{
		"success":      true,
		"message":      fmt.Sprintf("Обновлено %d файлов фильтров", updated),
		"updated":      updated,
		"version":      info.Version,
		"updated_at":   info.UpdatedAt,
		"is_outdated":  info.IsOutdated,
	}
}

// RebuildActiveProfileConfig rebuilds config for active profile
func (a *App) RebuildActiveProfileConfig() error {
	if a.storage == nil {
		return fmt.Errorf("storage not initialized")
	}
	
	profile, err := a.storage.GetActiveProfile()
	if err != nil || profile == nil {
		return fmt.Errorf("no active profile: %v", err)
	}
	
	// Get routing mode from settings
	settings := a.storage.GetAppSettings()
	if a.configBuilder != nil {
		a.configBuilder.SetRoutingMode(settings.RoutingMode)
	}
	
	// Rebuild using config builder
	return a.configBuilder.BuildConfig(profile.SubscriptionURL)
}
