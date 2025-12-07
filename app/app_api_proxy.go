package main

// Proxy methods for Kampus VPN
// This file contains Clash API proxy operations

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// GetProxiesWithDelay returns list of proxies with delay (ping)
func (a *App) GetProxiesWithDelay() map[string]interface{} {
	if !a.isRunning {
		return map[string]interface{}{
			"success": false,
			"error":   "VPN не запущен",
		}
	}

	client := &http.Client{Timeout: 5 * time.Second}

	// Get list of proxies
	resp, err := client.Get("http://127.0.0.1:9090/proxies")
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"error":   "Не удалось подключиться к API: " + err.Error(),
		}
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"error":   "Ошибка чтения ответа",
		}
	}

	var proxiesResp struct {
		Proxies map[string]struct {
			Name    string `json:"name"`
			Type    string `json:"type"`
			History []struct {
				Delay int `json:"delay"`
			} `json:"history"`
		} `json:"proxies"`
	}

	if err := json.Unmarshal(body, &proxiesResp); err != nil {
		return map[string]interface{}{
			"success": false,
			"error":   "Ошибка парсинга: " + err.Error(),
		}
	}

	// Form list of proxies with delays
	proxies := []map[string]interface{}{}
	for name, proxy := range proxiesResp.Proxies {
		// Skip service proxies
		if name == "DIRECT" || name == "REJECT" || name == "GLOBAL" ||
			proxy.Type == "Selector" || proxy.Type == "URLTest" || proxy.Type == "Fallback" {
			continue
		}

		delay := 0
		if len(proxy.History) > 0 {
			delay = proxy.History[len(proxy.History)-1].Delay
		}

		proxies = append(proxies, map[string]interface{}{
			"name":  name,
			"type":  proxy.Type,
			"delay": delay,
		})
	}

	return map[string]interface{}{
		"success": true,
		"proxies": proxies,
	}
}

// TestProxyDelay tests delay of a specific proxy
func (a *App) TestProxyDelay(proxyName string) map[string]interface{} {
	if !a.isRunning {
		return map[string]interface{}{
			"success": false,
			"error":   "VPN не запущен",
		}
	}

	client := &http.Client{Timeout: 10 * time.Second}

	// Test proxy delay
	url := fmt.Sprintf("http://127.0.0.1:9090/proxies/%s/delay?timeout=5000&url=http://www.gstatic.com/generate_204", proxyName)
	resp, err := client.Get(url)
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"delay":   0,
			"error":   err.Error(),
		}
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"delay":   0,
		}
	}

	var delayResp struct {
		Delay   int    `json:"delay"`
		Message string `json:"message"`
	}

	if err := json.Unmarshal(body, &delayResp); err != nil {
		return map[string]interface{}{
			"success": false,
			"delay":   0,
		}
	}

	if delayResp.Delay == 0 && delayResp.Message != "" {
		return map[string]interface{}{
			"success": false,
			"delay":   0,
			"error":   delayResp.Message,
		}
	}

	return map[string]interface{}{
		"success": true,
		"delay":   delayResp.Delay,
		"name":    proxyName,
	}
}

// TestAllProxiesDelay tests delay of all proxies in parallel
func (a *App) TestAllProxiesDelay() map[string]interface{} {
	if !a.isRunning {
		return map[string]interface{}{
			"success": false,
			"error":   "VPN не запущен",
		}
	}

	client := &http.Client{Timeout: 5 * time.Second}

	// Get list of proxies from selector proxy
	resp, err := client.Get("http://127.0.0.1:9090/proxies/proxy")
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		}
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"error":   "Ошибка чтения",
		}
	}

	var selectorInfo struct {
		All []string `json:"all"`
		Now string   `json:"now"`
	}

	if err := json.Unmarshal(body, &selectorInfo); err != nil {
		return map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		}
	}

	// Filter service proxies
	filteredProxies := []string{}
	for _, name := range selectorInfo.All {
		// Skip service elements
		if name == "direct" || name == "block" || name == "dns-out" || name == "auto-select" {
			continue
		}
		filteredProxies = append(filteredProxies, name)
	}

	// Get WireGuard configs from settings
	settings, _ := a.storage.GetUserSettings()
	wireGuardTags := []string{}
	wireGuardNames := map[string]string{} // tag -> name
	if settings != nil && len(settings.WireGuardConfigs) > 0 {
		for _, wg := range settings.WireGuardConfigs {
			wireGuardTags = append(wireGuardTags, wg.Tag)
			wireGuardNames[wg.Tag] = wg.Name
		}
	}

	totalCount := len(filteredProxies) + len(wireGuardTags)
	if totalCount == 0 {
		return map[string]interface{}{
			"success":      true,
			"proxies":      []map[string]interface{}{},
			"currentProxy": selectorInfo.Now,
			"count":        0,
		}
	}

	// Test delay for each proxy in parallel
	type proxyResult struct {
		Name       string
		Delay      int
		Type       string
		IsInternal bool
	}

	results := make(chan proxyResult, totalCount)

	// Test external proxies
	for _, proxyName := range filteredProxies {
		go func(name string) {
			delay := 0
			proxyType := ""

			// Get proxy info
			infoResp, err := client.Get(fmt.Sprintf("http://127.0.0.1:9090/proxies/%s", name))
			if err == nil {
				defer infoResp.Body.Close()
				infoBody, _ := io.ReadAll(infoResp.Body)
				var info struct {
					Type    string `json:"type"`
					History []struct {
						Delay int `json:"delay"`
					} `json:"history"`
				}
				if json.Unmarshal(infoBody, &info) == nil {
					proxyType = info.Type
					if len(info.History) > 0 {
						delay = info.History[len(info.History)-1].Delay
					}
				}
			}

			// If no history, test delay
			if delay == 0 {
				delayResp, err := client.Get(fmt.Sprintf("http://127.0.0.1:9090/proxies/%s/delay?timeout=3000&url=http://www.gstatic.com/generate_204", name))
				if err == nil {
					defer delayResp.Body.Close()
					delayBody, _ := io.ReadAll(delayResp.Body)
					var d struct {
						Delay int `json:"delay"`
					}
					if json.Unmarshal(delayBody, &d) == nil {
						delay = d.Delay
					}
				}
			}

			results <- proxyResult{Name: name, Delay: delay, Type: proxyType, IsInternal: false}
		}(proxyName)
	}

	// Test WireGuard servers
	for _, wgTag := range wireGuardTags {
		go func(tag string) {
			delay := -1 // -1 means "active but ping not measured"
			displayName := wireGuardNames[tag]
			if displayName == "" {
				displayName = tag
			}

			// Check that WireGuard endpoint is accessible in Clash API
			infoResp, err := client.Get(fmt.Sprintf("http://127.0.0.1:9090/proxies/%s", tag))
			if err == nil {
				defer infoResp.Body.Close()
				infoBody, _ := io.ReadAll(infoResp.Body)
				var info struct {
					Type string `json:"type"`
				}
				if json.Unmarshal(infoBody, &info) == nil && info.Type == "WireGuard" {
					delay = -1 // WireGuard is active
				}
			}

			results <- proxyResult{Name: displayName + " (внутр.)", Delay: delay, Type: "WireGuard", IsInternal: true}
		}(wgTag)
	}

	// Collect results
	proxies := []map[string]interface{}{}
	timeout := time.After(10 * time.Second)

	for i := 0; i < totalCount; i++ {
		select {
		case result := <-results:
			proxies = append(proxies, map[string]interface{}{
				"name":       result.Name,
				"delay":      result.Delay,
				"type":       result.Type,
				"isInternal": result.IsInternal,
			})
		case <-timeout:
			break
		}
	}

	return map[string]interface{}{
		"success":      true,
		"proxies":      proxies,
		"currentProxy": selectorInfo.Now,
		"count":        len(proxies),
	}
}

// GetCurrentProxy returns current active proxy and its delay
func (a *App) GetCurrentProxy() map[string]interface{} {
	if !a.isRunning {
		return map[string]interface{}{
			"success": false,
		}
	}

	client := &http.Client{Timeout: 5 * time.Second}

	// Get info about proxy selector
	resp, err := client.Get("http://127.0.0.1:9090/proxies/proxy")
	if err != nil {
		return map[string]interface{}{
			"success": false,
		}
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return map[string]interface{}{
			"success": false,
		}
	}

	var proxyInfo struct {
		Name string `json:"name"`
		Now  string `json:"now"`
		Type string `json:"type"`
	}

	if err := json.Unmarshal(body, &proxyInfo); err != nil {
		return map[string]interface{}{
			"success": false,
		}
	}

	currentProxy := proxyInfo.Now
	if currentProxy == "" {
		currentProxy = proxyInfo.Name
	}

	// Get delay for current proxy
	delay := 0
	if currentProxy != "" {
		delayResp, err := client.Get(fmt.Sprintf("http://127.0.0.1:9090/proxies/%s/delay?timeout=3000&url=http://www.gstatic.com/generate_204", currentProxy))
		if err == nil {
			defer delayResp.Body.Close()
			delayBody, _ := io.ReadAll(delayResp.Body)
			var delayInfo struct {
				Delay int `json:"delay"`
			}
			if json.Unmarshal(delayBody, &delayInfo) == nil {
				delay = delayInfo.Delay
			}
		}
	}

	return map[string]interface{}{
		"success": true,
		"name":    currentProxy,
		"type":    proxyInfo.Type,
		"delay":   delay,
	}
}
