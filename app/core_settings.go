package main

import (
	"encoding/json"
	"os"
)

// SubscriptionConfig represents a single subscription configuration.
type SubscriptionConfig struct {
	URL            string `json:"url"`
	Enabled        bool   `json:"enabled"`
	Name           string `json:"name"`
	UpdateInterval string `json:"update_interval"`
}

// AppSettings represents the application settings stored in settings.json.
// This handles subscription management and other connection-related settings.
type AppSettings struct {
	Subscriptions []SubscriptionConfig `json:"subscriptions"`
	DirectProxies []string             `json:"direct_proxies"` // Direct proxy links (vless://, vmess://, etc.)
}

// DefaultSettings returns default application settings.
func DefaultSettings() *AppSettings {
	return &AppSettings{
		Subscriptions: []SubscriptionConfig{},
		DirectProxies: []string{},
	}
}

// LoadSettings loads settings from the given path.
// Returns default settings if file doesn't exist or can't be parsed.
func LoadSettings(path string) (*AppSettings, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return DefaultSettings(), nil
		}
		return nil, err
	}

	var settings AppSettings
	if err := json.Unmarshal(data, &settings); err != nil {
		return DefaultSettings(), nil
	}

	return &settings, nil
}

// SaveSettings saves settings to the given path.
func SaveSettings(settings *AppSettings, path string) error {
	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
