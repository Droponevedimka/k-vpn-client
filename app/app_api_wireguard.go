package main

// WireGuard API methods for Kampus VPN
// This file contains WireGuard configuration management

import (
	"fmt"
)

// GetWireGuardList возвращает список WireGuard конфигов
func (a *App) GetWireGuardList() map[string]interface{} {
	a.waitForInit()
	
	if a.storage == nil {
		return map[string]interface{}{
			"success": false,
			"error":   "Storage не инициализирован",
			"configs": []WireGuardInfo{},
		}
	}

	settings, err := a.storage.GetUserSettings()
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"error":   err.Error(),
			"configs": []WireGuardInfo{},
		}
	}

	configs := make([]WireGuardInfo, 0, len(settings.WireGuardConfigs))
	for _, wg := range settings.WireGuardConfigs {
		configs = append(configs, wg.ToInfo())
	}

	return map[string]interface{}{
		"success": true,
		"configs": configs,
		"count":   len(configs),
	}
}

// ParseWireGuardConfigAPI парсит WireGuard конфиг и возвращает результат
func (a *App) ParseWireGuardConfigAPI(configText string) map[string]interface{} {
	wg, err := ParseWireGuardConfig(configText)
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		}
	}

	endpoint := wg.Endpoint
	if wg.EndpointPort > 0 {
		endpoint = fmt.Sprintf("%s:%d", wg.Endpoint, wg.EndpointPort)
	}

	return map[string]interface{}{
		"success":              true,
		"private_key":          wg.PrivateKey,
		"local_address":        wg.LocalAddress,
		"dns":                  wg.DNS,
		"mtu":                  wg.MTU,
		"public_key":           wg.PublicKey,
		"preshared_key":        wg.PresharedKey,
		"allowed_ips":          wg.AllowedIPs,
		"endpoint":             endpoint,
		"endpoint_port":        wg.EndpointPort,
		"persistent_keepalive": wg.PersistentKeepalive,
	}
}

// AddWireGuard добавляет новый WireGuard конфиг
func (a *App) AddWireGuard(tag string, name string, configText string) map[string]interface{} {
	a.waitForInit()
	
	// Проверяем что VPN выключен
	a.mu.Lock()
	if a.isRunning {
		a.mu.Unlock()
		return map[string]interface{}{
			"success": false,
			"error":   "Нельзя добавлять VPN пока соединение активно. Сначала отключите VPN.",
		}
	}
	a.mu.Unlock()

	if a.configBuilder == nil {
		return map[string]interface{}{
			"success": false,
			"error":   "ConfigBuilder не инициализирован",
		}
	}

	// Валидируем тег
	if err := ValidateTag(tag); err != nil {
		return map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		}
	}

	// Парсим конфиг
	wg, err := ParseWireGuardConfig(configText)
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"error":   fmt.Sprintf("Ошибка парсинга конфига: %v", err),
		}
	}

	wg.Tag = tag
	wg.Name = name
	if wg.Name == "" {
		wg.Name = tag
	}

	// Загружаем текущие настройки
	settings, err := a.storage.GetUserSettings()
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		}
	}

	// Проверяем лимит
	if len(settings.WireGuardConfigs) >= MaxWireGuardConfigs {
		return map[string]interface{}{
			"success": false,
			"error":   fmt.Sprintf("Достигнут лимит WireGuard конфигов (%d)", MaxWireGuardConfigs),
		}
	}

	// Проверяем уникальность тега
	for _, existing := range settings.WireGuardConfigs {
		if existing.Tag == tag {
			return map[string]interface{}{
				"success": false,
				"error":   fmt.Sprintf("Конфиг с тегом '%s' уже существует", tag),
			}
		}
	}

	// Добавляем конфиг
	settings.WireGuardConfigs = append(settings.WireGuardConfigs, *wg)

	// Перегенерируем конфиг
	if err := a.configBuilder.BuildConfigForProfile(a.storage.GetActiveProfileID(), settings.SubscriptionURL, settings.WireGuardConfigs); err != nil {
		return map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		}
	}

	return map[string]interface{}{
		"success": true,
		"count":   len(settings.WireGuardConfigs),
	}
}

// UpdateWireGuard обновляет существующий WireGuard конфиг
func (a *App) UpdateWireGuard(oldTag string, tag string, name string, configText string) map[string]interface{} {
	a.waitForInit()
	
	// Проверяем что VPN выключен
	a.mu.Lock()
	if a.isRunning {
		a.mu.Unlock()
		return map[string]interface{}{
			"success": false,
			"error":   "Нельзя редактировать VPN пока соединение активно. Сначала отключите VPN.",
		}
	}
	a.mu.Unlock()

	if a.configBuilder == nil {
		return map[string]interface{}{
			"success": false,
			"error":   "ConfigBuilder не инициализирован",
		}
	}

	// Валидируем новый тег
	if err := ValidateTag(tag); err != nil {
		return map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		}
	}

	// Парсим конфиг
	wg, err := ParseWireGuardConfig(configText)
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"error":   fmt.Sprintf("Ошибка парсинга конфига: %v", err),
		}
	}

	wg.Tag = tag
	wg.Name = name
	if wg.Name == "" {
		wg.Name = tag
	}

	// Загружаем текущие настройки
	settings, err := a.storage.GetUserSettings()
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		}
	}

	// Находим и обновляем конфиг
	found := false
	for i, existing := range settings.WireGuardConfigs {
		if existing.Tag == oldTag {
			// Проверяем уникальность нового тега (если изменился)
			if tag != oldTag {
				for _, other := range settings.WireGuardConfigs {
					if other.Tag == tag {
						return map[string]interface{}{
							"success": false,
							"error":   fmt.Sprintf("Конфиг с тегом '%s' уже существует", tag),
						}
					}
				}
			}
			settings.WireGuardConfigs[i] = *wg
			found = true
			break
		}
	}

	if !found {
		return map[string]interface{}{
			"success": false,
			"error":   fmt.Sprintf("Конфиг с тегом '%s' не найден", oldTag),
		}
	}

	// Перегенерируем конфиг
	if err := a.configBuilder.BuildConfigForProfile(a.storage.GetActiveProfileID(), settings.SubscriptionURL, settings.WireGuardConfigs); err != nil {
		return map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		}
	}

	return map[string]interface{}{
		"success": true,
	}
}

// DeleteWireGuard удаляет WireGuard конфиг
func (a *App) DeleteWireGuard(tag string) map[string]interface{} {
	a.waitForInit()
	
	// Проверяем что VPN выключен
	a.mu.Lock()
	if a.isRunning {
		a.mu.Unlock()
		return map[string]interface{}{
			"success": false,
			"error":   "Нельзя удалять VPN пока соединение активно. Сначала отключите VPN.",
		}
	}
	a.mu.Unlock()

	if a.storage == nil {
		return map[string]interface{}{
			"success": false,
			"error":   "Storage не инициализирован",
		}
	}

	// Загружаем текущие настройки
	settings, err := a.storage.GetUserSettings()
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		}
	}

	// Находим и удаляем конфиг
	newConfigs := make([]UserWireGuardConfig, 0, len(settings.WireGuardConfigs)-1)
	found := false
	for _, existing := range settings.WireGuardConfigs {
		if existing.Tag == tag {
			found = true
			continue
		}
		newConfigs = append(newConfigs, existing)
	}

	if !found {
		return map[string]interface{}{
			"success": false,
			"error":   fmt.Sprintf("Конфиг с тегом '%s' не найден", tag),
		}
	}

	settings.WireGuardConfigs = newConfigs

	// Перегенерируем конфиг
	if err := a.configBuilder.BuildConfigForProfile(a.storage.GetActiveProfileID(), settings.SubscriptionURL, settings.WireGuardConfigs); err != nil {
		return map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		}
	}

	return map[string]interface{}{
		"success": true,
		"count":   len(settings.WireGuardConfigs),
	}
}

// GetWireGuardConfig возвращает полный конфиг WireGuard для редактирования
func (a *App) GetWireGuardConfig(tag string) map[string]interface{} {
	a.waitForInit()
	
	if a.storage == nil {
		return map[string]interface{}{
			"success": false,
			"error":   "Storage не инициализирован",
		}
	}

	settings, err := a.storage.GetUserSettings()
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		}
	}

	for _, wg := range settings.WireGuardConfigs {
		if wg.Tag == tag {
			endpoint := wg.Endpoint
			if wg.EndpointPort > 0 {
				endpoint = fmt.Sprintf("%s:%d", wg.Endpoint, wg.EndpointPort)
			}
			
			return map[string]interface{}{
				"success":              true,
				"tag":                  wg.Tag,
				"name":                 wg.Name,
				"private_key":          wg.PrivateKey,
				"local_address":        wg.LocalAddress,
				"dns":                  wg.DNS,
				"mtu":                  wg.MTU,
				"public_key":           wg.PublicKey,
				"preshared_key":        wg.PresharedKey,
				"allowed_ips":          wg.AllowedIPs,
				"endpoint":             endpoint,
				"persistent_keepalive": wg.PersistentKeepalive,
			}
		}
	}

	return map[string]interface{}{
		"success": false,
		"error":   fmt.Sprintf("Конфиг с тегом '%s' не найден", tag),
	}
}
