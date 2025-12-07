package main

// Subscription management methods for Kampus VPN
// This file contains subscription-related API methods

import (
	"encoding/json"
	"fmt"
	"time"
)

// GetSettingsPath returns the path to settings file
func (a *App) GetSettingsPath() string {
	return a.configBuilder.GetSettingsPath()
}

// GetSettings returns current settings
func (a *App) GetSettings() map[string]interface{} {
	settingsPath := a.GetSettingsPath()
	settings, err := LoadSettings(settingsPath)
	if err != nil {
		// Return default settings
		settings = DefaultSettings()
	}

	// Convert to map for frontend
	data, _ := json.Marshal(settings)
	var result map[string]interface{}
	json.Unmarshal(data, &result)

	return result
}

// SaveSettingsFromUI saves settings from frontend
func (a *App) SaveSettingsFromUI(settingsJSON string) map[string]interface{} {
	var settings AppSettings
	if err := json.Unmarshal([]byte(settingsJSON), &settings); err != nil {
		return map[string]interface{}{
			"success": false,
			"error":   fmt.Sprintf("Invalid settings JSON: %v", err),
		}
	}

	settingsPath := a.GetSettingsPath()
	if err := SaveSettings(&settings, settingsPath); err != nil {
		return map[string]interface{}{
			"success": false,
			"error":   fmt.Sprintf("Failed to save settings: %v", err),
		}
	}

	return map[string]interface{}{
		"success": true,
	}
}

// AddSubscription adds a new subscription URL
func (a *App) AddSubscription(url string, name string) map[string]interface{} {
	settingsPath := a.GetSettingsPath()
	settings, err := LoadSettings(settingsPath)
	if err != nil {
		settings = DefaultSettings()
	}

	// Check for duplicates
	for _, sub := range settings.Subscriptions {
		if sub.URL == url {
			return map[string]interface{}{
				"success": false,
				"error":   "Subscription already exists",
			}
		}
	}

	settings.Subscriptions = append(settings.Subscriptions, SubscriptionConfig{
		URL:            url,
		Enabled:        true,
		Name:           name,
		UpdateInterval: "24h",
	})

	if err := SaveSettings(settings, settingsPath); err != nil {
		return map[string]interface{}{
			"success": false,
			"error":   fmt.Sprintf("Failed to save settings: %v", err),
		}
	}

	return map[string]interface{}{
		"success": true,
	}
}

// RemoveSubscription removes a subscription by URL
func (a *App) RemoveSubscription(url string) map[string]interface{} {
	settingsPath := a.GetSettingsPath()
	settings, err := LoadSettings(settingsPath)
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"error":   "No settings found",
		}
	}

	newSubs := []SubscriptionConfig{}
	for _, sub := range settings.Subscriptions {
		if sub.URL != url {
			newSubs = append(newSubs, sub)
		}
	}
	settings.Subscriptions = newSubs

	if err := SaveSettings(settings, settingsPath); err != nil {
		return map[string]interface{}{
			"success": false,
			"error":   fmt.Sprintf("Failed to save settings: %v", err),
		}
	}

	return map[string]interface{}{
		"success": true,
	}
}

// TestSubscription tests a subscription URL and returns available proxies
func (a *App) TestSubscription(url string) map[string]interface{} {
	fetcher := NewSubscriptionFetcher()
	proxies, err := fetcher.FetchAndParse(url)
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"error":   err.Error(),
			"count":   0,
		}
	}

	// Convert proxies to simple format for frontend
	proxyList := []map[string]interface{}{}
	for _, p := range proxies {
		proxyList = append(proxyList, map[string]interface{}{
			"type":   p.Type,
			"name":   p.Name,
			"server": p.Server,
			"port":   p.ServerPort,
		})
	}

	return map[string]interface{}{
		"success": true,
		"count":   len(proxies),
		"proxies": proxyList,
	}
}

// GenerateAndSaveConfig generates config from settings and saves it
func (a *App) GenerateAndSaveConfig() map[string]interface{} {
	if a.configBuilder == nil {
		return map[string]interface{}{
			"success": false,
			"error":   "ConfigBuilder не инициализирован",
		}
	}

	settings, err := a.configBuilder.LoadUserSettings()
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"error":   fmt.Sprintf("Failed to load settings: %v", err),
		}
	}

	if err := a.configBuilder.BuildConfigFull(settings.SubscriptionURL, settings.WireGuardConfigs); err != nil {
		return map[string]interface{}{
			"success": false,
			"error":   fmt.Sprintf("Failed to generate config: %v", err),
		}
	}

	return map[string]interface{}{
		"success": true,
		"path":    a.configPath,
	}
}

// UpdateSubscriptions fetches all subscriptions and regenerates config
func (a *App) UpdateSubscriptions() map[string]interface{} {
	// Stop VPN if running
	wasRunning := a.isRunning
	if wasRunning {
		a.Stop()
	}

	// Generate new config
	result := a.GenerateAndSaveConfig()
	if !result["success"].(bool) {
		return result
	}

	// Restart VPN if it was running
	if wasRunning {
		a.Start()
	}

	return map[string]interface{}{
		"success":    true,
		"wasRunning": wasRunning,
	}
}

// AddDirectProxy adds a direct proxy link
func (a *App) AddDirectProxy(link string) map[string]interface{} {
	// Validate link
	fetcher := NewSubscriptionFetcher()
	proxy, err := fetcher.ParseSingleLink(link)
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"error":   fmt.Sprintf("Invalid proxy link: %v", err),
		}
	}

	settingsPath := a.GetSettingsPath()
	settings, err := LoadSettings(settingsPath)
	if err != nil {
		settings = DefaultSettings()
	}

	// Check for duplicates
	for _, p := range settings.DirectProxies {
		if p == link {
			return map[string]interface{}{
				"success": false,
				"error":   "Proxy already exists",
			}
		}
	}

	settings.DirectProxies = append(settings.DirectProxies, link)

	if err := SaveSettings(settings, settingsPath); err != nil {
		return map[string]interface{}{
			"success": false,
			"error":   fmt.Sprintf("Failed to save settings: %v", err),
		}
	}

	return map[string]interface{}{
		"success": true,
		"proxy": map[string]interface{}{
			"type":   proxy.Type,
			"name":   proxy.Name,
			"server": proxy.Server,
			"port":   proxy.ServerPort,
		},
	}
}

// ==================== Subscription Management (New API) ====================

// GetCurrentSubscription возвращает текущую подписку пользователя
func (a *App) GetCurrentSubscription() map[string]interface{} {
	// Ждём инициализации
	a.waitForInit()
	
	if a.configBuilder == nil {
		return map[string]interface{}{
			"hasSubscription": false,
			"error":           "ConfigBuilder не инициализирован",
		}
	}

	settings, err := a.configBuilder.LoadUserSettings()
	if err != nil {
		return map[string]interface{}{
			"hasSubscription": false,
			"error":           err.Error(),
		}
	}

	if settings.SubscriptionURL == "" {
		return map[string]interface{}{
			"hasSubscription": false,
		}
	}

	return map[string]interface{}{
		"hasSubscription": true,
		"url":             settings.SubscriptionURL,
		"lastUpdated":     settings.LastUpdated,
		"proxyCount":      settings.ProxyCount,
	}
}

// TestVPNConnection тестирует подписку или прямую ссылку
func (a *App) TestVPNConnection(url string) map[string]interface{} {
	// Ждём инициализации
	a.waitForInit()
	
	if a.configBuilder == nil {
		return map[string]interface{}{
			"success": false,
			"error":   "ConfigBuilder не инициализирован",
		}
	}

	result, err := a.configBuilder.TestSubscription(url)
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		}
	}

	return map[string]interface{}{
		"success":      result.Success,
		"error":        result.Error,
		"count":        result.Count,
		"isDirectLink": result.IsDirectLink,
		"proxies":      result.Proxies,
	}
}

// SetVPNSubscription устанавливает подписку и генерирует конфиг
func (a *App) SetVPNSubscription(url string) map[string]interface{} {
	// Ждём инициализации
	a.waitForInit()
	
	if a.configBuilder == nil {
		return map[string]interface{}{
			"success": false,
			"error":   "ConfigBuilder не инициализирован",
		}
	}

	// Останавливаем VPN если запущен
	wasRunning := a.isRunning
	if wasRunning {
		a.Stop()
	}

	// Генерируем новый конфиг
	if err := a.configBuilder.BuildConfig(url); err != nil {
		return map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		}
	}

	// Перезапускаем VPN если был запущен
	if wasRunning {
		go func() {
			// Небольшая задержка чтобы конфиг сохранился
			time.Sleep(500 * time.Millisecond)
			a.Start()
		}()
	}

	// Загружаем обновлённые настройки
	settings, _ := a.configBuilder.LoadUserSettings()

	return map[string]interface{}{
		"success":    true,
		"proxyCount": settings.ProxyCount,
	}
}

// RemoveVPNSubscription удаляет подписку и генерирует конфиг без прокси
func (a *App) RemoveVPNSubscription() map[string]interface{} {
	// Ждём инициализации
	a.waitForInit()
	
	if a.configBuilder == nil {
		return map[string]interface{}{
			"success": false,
			"error":   "ConfigBuilder не инициализирован",
		}
	}

	// Останавливаем VPN
	wasRunning := a.isRunning
	if wasRunning {
		a.Stop()
	}

	// Генерируем конфиг без подписки
	if err := a.configBuilder.BuildConfig(""); err != nil {
		return map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		}
	}

	return map[string]interface{}{
		"success":    true,
		"wasRunning": wasRunning,
	}
}

// RefreshVPNSubscription обновляет текущую подписку
func (a *App) RefreshVPNSubscription() map[string]interface{} {
	if a.configBuilder == nil {
		return map[string]interface{}{
			"success": false,
			"error":   "ConfigBuilder не инициализирован",
		}
	}

	settings, err := a.configBuilder.LoadUserSettings()
	if err != nil || settings.SubscriptionURL == "" {
		return map[string]interface{}{
			"success": false,
			"error":   "Нет сохранённой подписки",
		}
	}

	return a.SetVPNSubscription(settings.SubscriptionURL)
}
