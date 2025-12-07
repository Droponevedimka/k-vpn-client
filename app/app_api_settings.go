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
		"autoUpdateSub":     settings.AutoUpdateSub,
		"subUpdateInterval": settings.SubUpdateInterval,
		"lastSubUpdate":     settings.LastSubUpdate.Format(time.RFC3339),
		"appVersion":        AppVersion,
		"appName":           AppName,
	}
}

// SaveAppConfig сохраняет настройки приложения (API для фронтенда)
func (a *App) SaveAppConfig(autoStart, enableLogging, checkUpdates, notifications, autoUpdateSub bool, theme, language string, subUpdateInterval int) map[string]interface{} {
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

// GetAutoStartStatus проверяет статус автозапуска
func (a *App) GetAutoStartStatus() map[string]interface{} {
	return map[string]interface{}{
		"success":   true,
		"autoStart": IsAutoStartEnabled(),
	}
}
