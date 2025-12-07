// Package main provides settings import/export functionality for KampusVPN.
package main

import (
	"encoding/json"
	"fmt"
	"time"
)

// ExportData represents all exportable application data.
type ExportData struct {
	Version          string              `json:"version"`           // App version that created export
	ExportedAt       time.Time           `json:"exported_at"`       // Export timestamp
	AppConfig        *AppConfig          `json:"app_config"`        // Application settings
	Profiles         []ConnectionProfile `json:"profiles"`          // Connection profiles
	WireGuardConfigs []UserWireGuardConfig `json:"wireguard_configs"` // WireGuard configurations
	TemplateContent  string              `json:"template_content"`  // Custom template.json content
}

// ExportSettings exports all application settings to JSON string.
func (a *App) ExportSettings() map[string]interface{} {
	a.waitForInit()

	export := ExportData{
		Version:    AppVersion,
		ExportedAt: time.Now(),
	}

	// Export app config
	if a.appConfig != nil {
		export.AppConfig = a.appConfig
	}

	// Export profiles
	if a.profileManager != nil {
		export.Profiles = a.profileManager.GetAll()
	}

	// Export WireGuard configs
	if a.configBuilder != nil {
		settings, err := a.configBuilder.LoadUserSettings()
		if err == nil && settings != nil {
			export.WireGuardConfigs = settings.WireGuardConfigs
		}
	}

	// Export template content
	if a.templatePath != "" {
		content, err := readFileContent(a.templatePath)
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

	var export ExportData
	if err := json.Unmarshal([]byte(jsonData), &export); err != nil {
		return map[string]interface{}{
			"success": false,
			"error":   fmt.Sprintf("Invalid settings format: %v", err),
		}
	}

	// Import app config (except active profile ID)
	if export.AppConfig != nil && a.appConfig != nil {
		activeProfileID := a.appConfig.ActiveProfileID
		*a.appConfig = *export.AppConfig
		a.appConfig.ActiveProfileID = activeProfileID // Preserve active profile

		if a.appConfigPath != "" {
			a.appConfig.Save(a.appConfigPath)
		}
	}

	// Import WireGuard configs
	if len(export.WireGuardConfigs) > 0 && a.configBuilder != nil {
		settings, err := a.configBuilder.LoadUserSettings()
		if err == nil {
			settings.WireGuardConfigs = export.WireGuardConfigs
			a.configBuilder.SaveUserSettings(settings)
		}
	}

	// Import template (if provided)
	if export.TemplateContent != "" && a.templatePath != "" {
		// Validate JSON before saving
		var jsonTest interface{}
		if err := json.Unmarshal([]byte(export.TemplateContent), &jsonTest); err == nil {
			writeFileContent(a.templatePath, export.TemplateContent)
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
