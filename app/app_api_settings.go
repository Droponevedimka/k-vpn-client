package main

// App settings methods for Kampus VPN (App methods extension)
// This file contains app configuration API methods

import (
	"fmt"
	"time"
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
