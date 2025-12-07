// Package main provides constants and configuration for KampusVPN application.
package main

import (
	"time"
)

// Application metadata
const (
	// AppName is the display name of the application.
	AppName = "KampusVPN"
	// AppVersion is the current version of the application.
	AppVersion = "1.2.0"
	// GitHubRepo is the GitHub repository path for updates.
	GitHubRepo = "Droponevedimka/k-vpn-client"
	// GitHubURL is the full GitHub URL.
	GitHubURL = "https://github.com/" + GitHubRepo
)

// File names used by the application
const (
	// ConfigFileName is the generated sing-box configuration file.
	ConfigFileName = "config.json"
	// TemplateFileName is the template for generating config.
	TemplateFileName = "template.json"
	// UserSettingsFileName stores user settings (subscription, wireguard configs).
	UserSettingsFileName = "user_settings.json"
	// AppConfigFileName stores application preferences.
	AppConfigFileName = "app_config.json"
	// TrafficStatsFileName stores traffic statistics.
	TrafficStatsFileName = "traffic_stats.json"
	// ProfilesFileName stores connection profiles.
	ProfilesFileName = "profiles.json"
	// LogFileName is the sing-box log file.
	LogFileName = "vpn.log"
	// CacheFileName is the sing-box cache database.
	CacheFileName = "cache.db"
	// SingboxExeName is the sing-box executable name.
	SingboxExeName = "sing-box.exe"
	// SingboxSubDir is the subdirectory containing sing-box.
	SingboxSubDir = "bin"
)

// HTTP client timeouts
const (
	// DefaultHTTPTimeout is the default timeout for HTTP requests.
	DefaultHTTPTimeout = 30 * time.Second
	// ShortHTTPTimeout is a shorter timeout for quick checks.
	ShortHTTPTimeout = 10 * time.Second
	// LongHTTPTimeout is a longer timeout for downloads.
	LongHTTPTimeout = 60 * time.Second
	// ClashAPITimeout is the timeout for Clash API requests.
	ClashAPITimeout = 5 * time.Second
)

// Clash API configuration
const (
	// ClashAPIHost is the host for Clash API.
	ClashAPIHost = "127.0.0.1"
	// ClashAPIPort is the port for Clash API.
	ClashAPIPort = 9090
	// ClashAPISecret is the secret for Clash API (empty = no auth).
	ClashAPISecret = ""
)

// Log configuration
const (
	// MaxLogSize is the maximum log file size before rotation.
	MaxLogSize = 10 * 1024 * 1024 // 10 MB
	// TruncateToSize is the size to truncate logs to when rotating.
	TruncateToSize = 5 * 1024 * 1024 // 5 MB
	// MaxLogBufferSize is the maximum number of log entries in UI buffer.
	MaxLogBufferSize = 1000
)

// LogLevel represents the logging level.
type LogLevel string

const (
	// LogLevelDebug enables all logging.
	LogLevelDebug LogLevel = "debug"
	// LogLevelInfo enables info and above.
	LogLevelInfo LogLevel = "info"
	// LogLevelWarn enables warnings and errors only.
	LogLevelWarn LogLevel = "warn"
	// LogLevelError enables only errors.
	LogLevelError LogLevel = "error"
	// LogLevelSilent disables all logging.
	LogLevelSilent LogLevel = "silent"
)

// Profile configuration
const (
	// DefaultProfileID is the ID of the default profile that cannot be deleted.
	DefaultProfileID = 1
	// DefaultProfileName is the default name for the first profile.
	DefaultProfileName = "Work"
	// MaxProfiles is the maximum number of profiles allowed.
	MaxProfiles = 10
)

// WireGuard configuration
const (
	// MaxWireGuardConfigs is the maximum number of WireGuard configs per profile.
	MaxWireGuardConfigs = 20
	// DefaultMTU is the default MTU for WireGuard.
	DefaultMTU = 1280
)

// UI configuration
const (
	// WindowWidth is the default window width.
	WindowWidth = 570
	// WindowHeight is the default window height.
	WindowHeight = 755
	// MinWindowWidth is the minimum window width.
	MinWindowWidth = 570
	// MinWindowHeight is the minimum window height.
	MinWindowHeight = 755
)

// Theme represents the UI theme.
type Theme string

const (
	// ThemeDark is the dark theme.
	ThemeDark Theme = "dark"
	// ThemeLight is the light theme.
	ThemeLight Theme = "light"
	// ThemeSystem follows system preference.
	ThemeSystem Theme = "system"
)

// Language represents the UI language.
type Language string

const (
	// LangRussian is Russian language.
	LangRussian Language = "ru"
	// LangEnglish is English language.
	LangEnglish Language = "en"
)
