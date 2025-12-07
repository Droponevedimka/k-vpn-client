// Package main provides legacy type definitions for migration purposes.
// These types are used only to read old configuration files during migration.
package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"
)

// ConnectionProfile represents a VPN connection profile (legacy format).
// Used for migration from old profiles.json format.
type ConnectionProfile struct {
	ID              int       `json:"id"`
	Name            string    `json:"name"`
	CreatedAt       time.Time `json:"created_at"`
	SubscriptionURL string    `json:"subscription_url,omitempty"`
	ProxyCount      int       `json:"proxy_count,omitempty"`
	WireGuardTags   []string  `json:"wireguard_tags,omitempty"`
}

// AppConfig stores application preferences and settings (legacy format).
// Used for migration from old app_config.json format.
type AppConfig struct {
	AutoStart         bool      `json:"auto_start"`
	Notifications     bool      `json:"notifications"`
	CheckUpdates      bool      `json:"check_updates"`
	EnableLogging     bool      `json:"enable_logging"`
	LogLevel          LogLevel  `json:"log_level"`
	Theme             Theme     `json:"theme"`
	Language          Language  `json:"language"`
	AutoUpdateSub     bool      `json:"auto_update_sub"`
	SubUpdateInterval int       `json:"sub_update_interval"`
	LastSubUpdate     time.Time `json:"last_sub_update"`
	LastUpdateCheck   string    `json:"last_update_check"`
	ActiveProfileID   int       `json:"active_profile_id"`
}

// SubscriptionTestResult результат тестирования подписки
type SubscriptionTestResult struct {
	Success      bool        `json:"success"`
	Error        string      `json:"error,omitempty"`
	Count        int         `json:"count"`
	IsDirectLink bool        `json:"is_direct_link"`
	Proxies      []ProxyInfo `json:"proxies"`
}

// ProxyInfo информация о прокси для UI
type ProxyInfo struct {
	Type   string `json:"type"`
	Name   string `json:"name"`
	Server string `json:"server"`
	Port   int    `json:"port"`
}

// UserSettings хранит настройки пользователя для профиля
type UserSettings struct {
	SubscriptionURL  string                `json:"subscription_url"`
	LastUpdated      string                `json:"last_updated"`
	ProxyCount       int                   `json:"proxy_count"`
	WireGuardConfigs []UserWireGuardConfig `json:"wireguard_configs"`
}

// generateTag generates a unique tag for a proxy
func generateTag(p ProxyConfig, index int) string {
	name := p.Name
	if name == "" {
		name = p.Type
	}
	// Clean up name for tag
	tag := sanitizeTag(name)
	if tag == "" {
		tag = p.Type
	}
	return tag
}

// sanitizeTag removes invalid characters from tag
func sanitizeTag(s string) string {
	result := make([]byte, 0, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '-' || c == '_' {
			result = append(result, c)
		} else if c == ' ' {
			result = append(result, '-')
		}
	}
	return string(result)
}

// copyMap creates a deep copy of a map
func copyMap(m map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range m {
		if nested, ok := v.(map[string]interface{}); ok {
			result[k] = copyMap(nested)
		} else if arr, ok := v.([]interface{}); ok {
			newArr := make([]interface{}, len(arr))
			copy(newArr, arr)
			result[k] = newArr
		} else {
			result[k] = v
		}
	}
	return result
}

// SetAutoStart enables or disables system startup launch.
func SetAutoStart(enable bool) error {
	if runtime.GOOS != "windows" {
		return nil
	}

	exePath, err := os.Executable()
	if err != nil {
		return err
	}

	startupDir := filepath.Join(os.Getenv("APPDATA"), "Microsoft", "Windows", "Start Menu", "Programs", "Startup")
	shortcutPath := filepath.Join(startupDir, "KampusVPN.lnk")

	if enable {
		// Create shortcut using PowerShell
		psScript := `$WshShell = New-Object -ComObject WScript.Shell; $Shortcut = $WshShell.CreateShortcut('` + shortcutPath + `'); $Shortcut.TargetPath = '` + exePath + `'; $Shortcut.WorkingDirectory = '` + filepath.Dir(exePath) + `'; $Shortcut.Save()`
		cmd := exec.Command("powershell", "-Command", psScript)
		return cmd.Run()
	} else {
		// Remove shortcut
		return os.Remove(shortcutPath)
	}
}

// IsAutoStartEnabled checks if app is set to start with Windows.
func IsAutoStartEnabled() bool {
	if runtime.GOOS != "windows" {
		return false
	}

	startupDir := filepath.Join(os.Getenv("APPDATA"), "Microsoft", "Windows", "Start Menu", "Programs", "Startup")
	shortcutPath := filepath.Join(startupDir, "KampusVPN.lnk")

	_, err := os.Stat(shortcutPath)
	return err == nil
}
