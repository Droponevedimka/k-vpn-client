package main

// App settings methods for Kampus VPN (App methods extension)
// This file contains app configuration API methods

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

// initAppConfig инициализирует настройки приложения
func (a *App) initAppConfig() {
	a.appConfigPath = a.getAppConfigPath()
	a.appConfig = LoadAppConfig(a.appConfigPath)
}

// initProfileManager инициализирует менеджер профилей
func (a *App) initProfileManager() {
	basePath := a.getConfigDir()
	profilesPath := filepath.Join(basePath, "profiles.json")
	a.profileManager = NewProfileManager(profilesPath)
}

// getConfigDir возвращает директорию конфигурации приложения
func (a *App) getConfigDir() string {
	var configDir string
	
	switch runtime.GOOS {
	case "windows":
		configDir = filepath.Join(os.Getenv("LOCALAPPDATA"), "KampusVPN")
	case "darwin":
		home, _ := os.UserHomeDir()
		configDir = filepath.Join(home, "Library", "Application Support", "KampusVPN")
	default:
		home, _ := os.UserHomeDir()
		configDir = filepath.Join(home, ".config", "kampusvpn")
	}
	
	os.MkdirAll(configDir, 0755)
	return configDir
}

// getAppConfigPath возвращает путь к файлу настроек
func (a *App) getAppConfigPath() string {
	return filepath.Join(a.getConfigDir(), "app_config.json")
}

// saveAppConfig сохраняет настройки приложения
func (a *App) saveAppConfig() error {
	if a.appConfig == nil {
		return nil
	}
	configPath := a.getAppConfigPath()
	return a.appConfig.Save(configPath)
}

// GetAppConfig возвращает текущие настройки приложения (API для фронтенда)
func (a *App) GetAppConfig() map[string]interface{} {
	a.waitForInit()
	
	if a.appConfig == nil {
		return map[string]interface{}{
			"success": false,
			"error":   "Настройки не загружены",
		}
	}
	
	return map[string]interface{}{
		"success":           true,
		"autoStart":         a.appConfig.AutoStart,
		"enableLogging":     a.appConfig.EnableLogging,
		"checkUpdates":      a.appConfig.CheckUpdates,
		"notifications":     a.appConfig.Notifications,
		"theme":             a.appConfig.Theme,
		"language":          a.appConfig.Language,
		"autoUpdateSub":     a.appConfig.AutoUpdateSub,
		"subUpdateInterval": a.appConfig.SubUpdateInterval,
		"lastSubUpdate":     a.appConfig.LastSubUpdate.Format(time.RFC3339),
		"appVersion":        AppVersion,
		"appName":           AppName,
	}
}

// SaveAppConfig сохраняет настройки приложения (API для фронтенда)
func (a *App) SaveAppConfig(autoStart, enableLogging, checkUpdates, notifications, autoUpdateSub bool, theme, language string, subUpdateInterval int) map[string]interface{} {
	a.waitForInit()
	
	if a.appConfig == nil {
		return map[string]interface{}{
			"success": false,
			"error":   "Настройки не загружены",
		}
	}
	
	// Обновляем настройки
	a.appConfig.AutoStart = autoStart
	a.appConfig.EnableLogging = enableLogging
	a.appConfig.CheckUpdates = checkUpdates
	a.appConfig.Notifications = notifications
	a.appConfig.AutoUpdateSub = autoUpdateSub
	a.appConfig.Theme = Theme(theme)
	a.appConfig.Language = Language(language)
	a.appConfig.SubUpdateInterval = subUpdateInterval
	
	// Применяем автозапуск
	if err := a.appConfig.SetAutoStart(autoStart); err != nil {
		return map[string]interface{}{
			"success": false,
			"error":   fmt.Sprintf("Ошибка настройки автозапуска: %v", err),
		}
	}
	
	// Сохраняем в файл
	if err := a.saveAppConfig(); err != nil {
		return map[string]interface{}{
			"success": false,
			"error":   fmt.Sprintf("Ошибка сохранения настроек: %v", err),
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
