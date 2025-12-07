package main

// Profile management methods for Kampus VPN
// This file contains profile CRUD operations

import (
	"time"
)

// GetProfiles возвращает список всех профилей (API для фронтенда)
func (a *App) GetProfiles() map[string]interface{} {
	a.waitForInit()
	
	if a.profileManager == nil {
		return map[string]interface{}{
			"success": false,
			"error":   "Менеджер профилей не инициализирован",
		}
	}
	
	profiles := a.profileManager.GetAll()
	activeID := DefaultProfileID
	if a.appConfig != nil {
		activeID = a.appConfig.ActiveProfileID
	}
	
	var profilesData []map[string]interface{}
	for _, p := range profiles {
		profilesData = append(profilesData, map[string]interface{}{
			"id":           p.ID,
			"name":         p.Name,
			"subscription": p.SubscriptionURL,
			"wireguards":   p.WireGuardConfigs,
			"isActive":     p.ID == activeID,
			"createdAt":    p.CreatedAt.Format(time.RFC3339),
			"proxyCount":   p.ProxyCount,
		})
	}
	
	return map[string]interface{}{
		"success":       true,
		"profiles":      profilesData,
		"activeProfile": activeID,
	}
}

// GetActiveProfile возвращает активный профиль (API для фронтенда)
func (a *App) GetActiveProfile() map[string]interface{} {
	a.waitForInit()
	
	if a.profileManager == nil {
		return map[string]interface{}{
			"success": false,
			"error":   "Менеджер профилей не инициализирован",
		}
	}
	
	activeID := DefaultProfileID
	if a.appConfig != nil {
		activeID = a.appConfig.ActiveProfileID
	}
	
	profile, err := a.profileManager.GetByID(activeID)
	if err != nil {
		// Fallback to default profile
		profile, _ = a.profileManager.GetByID(DefaultProfileID)
	}
	
	return map[string]interface{}{
		"success": true,
		"profile": map[string]interface{}{
			"id":           profile.ID,
			"name":         profile.Name,
			"subscription": profile.SubscriptionURL,
			"wireguards":   profile.WireGuardConfigs,
			"isActive":     true,
			"createdAt":    profile.CreatedAt.Format(time.RFC3339),
			"proxyCount":   profile.ProxyCount,
		},
	}
}

// SetActiveProfile устанавливает активный профиль (API для фронтенда)
func (a *App) SetActiveProfile(id int) map[string]interface{} {
	a.waitForInit()
	
	// Check if VPN is running - don't allow profile change while connected
	a.mu.Lock()
	if a.isRunning {
		a.mu.Unlock()
		return map[string]interface{}{
			"success": false,
			"error":   "Отключите VPN перед сменой профиля",
		}
	}
	a.mu.Unlock()
	
	if a.profileManager == nil {
		return map[string]interface{}{
			"success": false,
			"error":   "Менеджер профилей не инициализирован",
		}
	}
	
	// Verify profile exists
	if _, err := a.profileManager.GetByID(id); err != nil {
		return map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		}
	}
	
	// Switch ConfigBuilder to new profile's settings
	if a.configBuilder != nil {
		a.configBuilder.SetActiveProfile(id)
	}
	
	// Save active profile to config
	if a.appConfig != nil {
		a.appConfig.ActiveProfileID = id
		a.saveAppConfig()
	}
	
	return map[string]interface{}{
		"success": true,
		"message": "Профиль активирован",
	}
}

// CreateProfile создает новый профиль (API для фронтенда)
func (a *App) CreateProfile(name string) map[string]interface{} {
	a.waitForInit()
	
	if a.profileManager == nil {
		return map[string]interface{}{
			"success": false,
			"error":   "Менеджер профилей не инициализирован",
		}
	}
	
	profile, err := a.profileManager.Create(name)
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		}
	}
	
	return map[string]interface{}{
		"success": true,
		"profile": map[string]interface{}{
			"id":           profile.ID,
			"name":         profile.Name,
			"subscription": profile.SubscriptionURL,
			"wireguards":   profile.WireGuardConfigs,
			"isActive":     false,
			"createdAt":    profile.CreatedAt.Format(time.RFC3339),
			"proxyCount":   profile.ProxyCount,
		},
	}
}

// UpdateProfile обновляет профиль (API для фронтенда)
func (a *App) UpdateProfile(id int, name string) map[string]interface{} {
	a.waitForInit()
	
	if a.profileManager == nil {
		return map[string]interface{}{
			"success": false,
			"error":   "Менеджер профилей не инициализирован",
		}
	}
	
	if err := a.profileManager.Update(id, name); err != nil {
		return map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		}
	}
	
	return map[string]interface{}{
		"success": true,
		"message": "Профиль обновлен",
	}
}

// DeleteProfile удаляет профиль (API для фронтенда)
func (a *App) DeleteProfile(id int) map[string]interface{} {
	a.waitForInit()
	
	if a.profileManager == nil {
		return map[string]interface{}{
			"success": false,
			"error":   "Менеджер профилей не инициализирован",
		}
	}
	
	if err := a.profileManager.Delete(id); err != nil {
		return map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		}
	}
	
	// If deleted profile was active, switch to default
	if a.appConfig != nil && a.appConfig.ActiveProfileID == id {
		a.appConfig.ActiveProfileID = DefaultProfileID
		a.saveAppConfig()
	}
	
	return map[string]interface{}{
		"success": true,
		"message": "Профиль удален",
	}
}
