// Package main provides settings import/export functionality for KampusVPN.
package main

import (
	"encoding/json"
	"fmt"
	"time"
)

// ExportData represents all exportable application data.
type ExportData struct {
	Version          string                `json:"version"`           // App version that created export
	ExportedAt       time.Time             `json:"exported_at"`       // Export timestamp
	AppSettings      GlobalAppSettings     `json:"app_settings"`      // Application settings
	Profiles         []ProfileData         `json:"profiles"`          // Connection profiles with configs
	WireGuardConfigs []UserWireGuardConfig `json:"wireguard_configs"` // WireGuard configurations (for active profile)
	TemplateContent  string                `json:"template_content"`  // Custom template.json content
}

// ExportSettings exports all application settings to JSON string.
func (a *App) ExportSettings() map[string]interface{} {
	a.waitForInit()

	if a.storage == nil {
		return map[string]interface{}{
			"success": false,
			"error":   "Storage not initialized",
		}
	}

	export := ExportData{
		Version:    Version,
		ExportedAt: time.Now(),
	}

	// Export app settings
	export.AppSettings = a.storage.GetAppSettings()

	// Export profiles
	export.Profiles = a.storage.GetAllProfiles()

	// Export WireGuard configs from active profile
	settings, err := a.storage.GetUserSettings()
	if err == nil && settings != nil {
		export.WireGuardConfigs = settings.WireGuardConfigs
	}

	// Export template content
	templatePath := a.storage.GetTemplatePath()
	if templatePath != "" {
		content, err := readFileContent(templatePath)
		if err == nil {
			export.TemplateContent = content
		}
	}

	data, err := json.MarshalIndent(export, "", "  ")
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"error":   fmt.Sprintf("Failed to export settings: %v", err),
		}
	}

	return map[string]interface{}{
		"success": true,
		"data":    string(data),
	}
}

// ImportSettings imports application settings from JSON string.
func (a *App) ImportSettings(jsonData string) map[string]interface{} {
	a.waitForInit()

	if a.storage == nil {
		return map[string]interface{}{
			"success": false,
			"error":   "Storage not initialized",
		}
	}

	var export ExportData
	if err := json.Unmarshal([]byte(jsonData), &export); err != nil {
		return map[string]interface{}{
			"success": false,
			"error":   fmt.Sprintf("Invalid settings format: %v", err),
		}
	}

	// Import app settings (preserve active profile ID)
	currentSettings := a.storage.GetAppSettings()
	activeProfileID := currentSettings.ActiveProfileID
	export.AppSettings.ActiveProfileID = activeProfileID
	a.storage.UpdateAppSettings(export.AppSettings)

	// Import WireGuard configs to active profile
	if len(export.WireGuardConfigs) > 0 {
		settings, err := a.storage.GetUserSettings()
		if err == nil {
			a.storage.UpdateProfileWireGuard(activeProfileID, export.WireGuardConfigs)
			// Rebuild config with new WireGuard settings
			if a.configBuilder != nil {
				a.configBuilder.BuildConfigForProfile(activeProfileID, settings.SubscriptionURL, export.WireGuardConfigs)
			}
		}
	}

	// Import template (if provided)
	templatePath := a.storage.GetTemplatePath()
	if export.TemplateContent != "" && templatePath != "" {
		// Validate JSON before saving
		var jsonTest interface{}
		if err := json.Unmarshal([]byte(export.TemplateContent), &jsonTest); err == nil {
			writeFileContent(templatePath, export.TemplateContent)
		}
	}

	a.writeLog("Settings imported successfully")
	a.AddToLogBuffer("Настройки импортированы")

	return map[string]interface{}{
		"success": true,
		"message": "Settings imported successfully",
	}
}

// readFileContent reads file content as string.
func readFileContent(path string) (string, error) {
	data, err := readFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// writeFileContent writes string content to file.
func writeFileContent(path, content string) error {
	return writeFile(path, []byte(content))
}
