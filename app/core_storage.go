// Package main provides unified storage management for KampusVPN.
// All profile data is stored in a single resources/settings.json file.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// ProfileData contains all data for a single profile.
type ProfileData struct {
	// Profile metadata
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	
	// Subscription settings (was user_settings.json)
	SubscriptionURL string                `json:"subscription_url,omitempty"`
	LastUpdated     string                `json:"last_updated,omitempty"`
	ProxyCount      int                   `json:"proxy_count,omitempty"`
	WireGuardConfigs []UserWireGuardConfig `json:"wireguard_configs,omitempty"`
	
	// Generated sing-box config (was config.json)
	SingboxConfig map[string]interface{} `json:"singbox_config,omitempty"`
}

// GlobalAppSettings contains global application settings (stored in settings.json).
type GlobalAppSettings struct {
	// General settings
	AutoStart     bool   `json:"auto_start"`
	Notifications bool   `json:"notifications"`
	CheckUpdates  bool   `json:"check_updates"`
	
	// Logging settings
	EnableLogging bool     `json:"enable_logging"`
	LogLevel      LogLevel `json:"log_level"`
	
	// Appearance
	Theme    Theme    `json:"theme"`
	Language Language `json:"language"`
	
	// Subscription settings
	AutoUpdateSub     bool      `json:"auto_update_sub"`
	SubUpdateInterval int       `json:"sub_update_interval"`
	LastSubUpdate     time.Time `json:"last_sub_update"`
	
	// Update tracking
	LastUpdateCheck string `json:"last_update_check"`
	
	// Active profile
	ActiveProfileID int `json:"active_profile_id"`
}

// SettingsFile represents the complete settings.json structure.
type SettingsFile struct {
	Version  int               `json:"version"`  // Schema version for migrations
	App      GlobalAppSettings `json:"app"`      // Global app settings
	Profiles []ProfileData     `json:"profiles"` // Array of profiles with their configs
}

// Storage manages the unified settings.json file.
type Storage struct {
	resourcesPath string       // Path to resources folder
	settingsPath  string       // Path to settings.json
	templatePath  string       // Path to template.json
	data          *SettingsFile
	mu            sync.RWMutex
}

const (
	SettingsVersion  = 1
	ResourcesFolder  = "resources"
	SettingsFileName = "settings.json"
)

// NewStorage creates a new storage manager.
func NewStorage(basePath string) *Storage {
	resourcesPath := filepath.Join(basePath, ResourcesFolder)
	
	s := &Storage{
		resourcesPath: resourcesPath,
		settingsPath:  filepath.Join(resourcesPath, SettingsFileName),
		templatePath:  filepath.Join(resourcesPath, TemplateFileName),
	}
	
	return s
}

// Init initializes storage, creating directories and files as needed.
func (s *Storage) Init() error {
	// Create resources directory
	if err := os.MkdirAll(s.resourcesPath, 0755); err != nil {
		return fmt.Errorf("failed to create resources directory: %w", err)
	}
	
	// Copy template.json to resources if not exists
	if !fileExists(s.templatePath) {
		if err := copyEmbeddedTemplate(s.templatePath); err != nil {
			return fmt.Errorf("failed to copy template.json: %w", err)
		}
	}
	
	// Load or create settings.json
	return s.Load()
}

// Load loads settings from file.
func (s *Storage) Load() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	data, err := os.ReadFile(s.settingsPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Create default settings
			s.data = s.createDefaultSettings()
			return s.saveInternal()
		}
		return fmt.Errorf("failed to read settings: %w", err)
	}
	
	var settings SettingsFile
	if err := json.Unmarshal(data, &settings); err != nil {
		// Backup corrupted file and create new
		backupPath := s.settingsPath + ".backup"
		os.Rename(s.settingsPath, backupPath)
		s.data = s.createDefaultSettings()
		return s.saveInternal()
	}
	
	s.data = &settings
	
	// Ensure at least one profile exists
	if len(s.data.Profiles) == 0 {
		s.data.Profiles = []ProfileData{s.createDefaultProfile()}
		return s.saveInternal()
	}
	
	return nil
}

// createDefaultSettings creates default settings structure.
func (s *Storage) createDefaultSettings() *SettingsFile {
	return &SettingsFile{
		Version: SettingsVersion,
		App: GlobalAppSettings{
			AutoStart:         false,
			Notifications:     true,
			CheckUpdates:      true,
			EnableLogging:     true,
			LogLevel:          LogLevelWarn,
			Theme:             ThemeDark,
			Language:          LangRussian,
			AutoUpdateSub:     false,
			SubUpdateInterval: 24,
			ActiveProfileID:   DefaultProfileID,
		},
		Profiles: []ProfileData{s.createDefaultProfile()},
	}
}

// createDefaultProfile creates a default profile.
func (s *Storage) createDefaultProfile() ProfileData {
	return ProfileData{
		ID:        DefaultProfileID,
		Name:      DefaultProfileName,
		CreatedAt: time.Now(),
	}
}

// saveInternal saves settings without locking.
func (s *Storage) saveInternal() error {
	data, err := json.MarshalIndent(s.data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal settings: %w", err)
	}
	return os.WriteFile(s.settingsPath, data, 0644)
}

// Save saves settings to file.
func (s *Storage) Save() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.saveInternal()
}

// GetTemplatePath returns path to template.json.
func (s *Storage) GetTemplatePath() string {
	return s.templatePath
}

// GetResourcesPath returns path to resources folder.
func (s *Storage) GetResourcesPath() string {
	return s.resourcesPath
}

// --- App Settings ---

// GetAppSettings returns a copy of app settings.
func (s *Storage) GetAppSettings() GlobalAppSettings {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.data.App
}

// UpdateAppSettings updates app settings.
func (s *Storage) UpdateAppSettings(settings GlobalAppSettings) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data.App = settings
	return s.saveInternal()
}

// GetActiveProfileID returns the active profile ID.
func (s *Storage) GetActiveProfileID() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.data.App.ActiveProfileID
}

// SetActiveProfileID sets the active profile ID.
func (s *Storage) SetActiveProfileID(id int) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data.App.ActiveProfileID = id
	return s.saveInternal()
}

// --- Profile Management ---

// GetAllProfiles returns all profiles.
func (s *Storage) GetAllProfiles() []ProfileData {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	result := make([]ProfileData, len(s.data.Profiles))
	copy(result, s.data.Profiles)
	return result
}

// GetProfile returns a profile by ID.
func (s *Storage) GetProfile(id int) (*ProfileData, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	for i := range s.data.Profiles {
		if s.data.Profiles[i].ID == id {
			profile := s.data.Profiles[i]
			return &profile, nil
		}
	}
	return nil, fmt.Errorf("profile with ID %d not found", id)
}

// GetActiveProfile returns the currently active profile.
func (s *Storage) GetActiveProfile() (*ProfileData, error) {
	return s.GetProfile(s.GetActiveProfileID())
}

// CreateProfile creates a new profile.
func (s *Storage) CreateProfile(name string) (*ProfileData, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	if len(s.data.Profiles) >= MaxProfiles {
		return nil, fmt.Errorf("maximum number of profiles (%d) reached", MaxProfiles)
	}
	
	// Find next available ID
	maxID := 0
	for _, p := range s.data.Profiles {
		if p.ID > maxID {
			maxID = p.ID
		}
	}
	
	profile := ProfileData{
		ID:        maxID + 1,
		Name:      name,
		CreatedAt: time.Now(),
	}
	
	s.data.Profiles = append(s.data.Profiles, profile)
	if err := s.saveInternal(); err != nil {
		return nil, err
	}
	
	return &profile, nil
}

// UpdateProfile updates a profile's metadata.
func (s *Storage) UpdateProfile(id int, name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	for i := range s.data.Profiles {
		if s.data.Profiles[i].ID == id {
			s.data.Profiles[i].Name = name
			return s.saveInternal()
		}
	}
	return fmt.Errorf("profile with ID %d not found", id)
}

// DeleteProfile deletes a profile.
func (s *Storage) DeleteProfile(id int) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	if id == DefaultProfileID {
		return fmt.Errorf("cannot delete default profile")
	}
	
	for i := range s.data.Profiles {
		if s.data.Profiles[i].ID == id {
			s.data.Profiles = append(s.data.Profiles[:i], s.data.Profiles[i+1:]...)
			
			// Switch to default profile if deleted profile was active
			if s.data.App.ActiveProfileID == id {
				s.data.App.ActiveProfileID = DefaultProfileID
			}
			
			return s.saveInternal()
		}
	}
	return fmt.Errorf("profile with ID %d not found", id)
}

// --- Profile Settings (Subscription, WireGuard) ---

// UpdateProfileSubscription updates a profile's subscription settings.
func (s *Storage) UpdateProfileSubscription(id int, subscriptionURL string, proxyCount int, wireGuardConfigs []UserWireGuardConfig) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	for i := range s.data.Profiles {
		if s.data.Profiles[i].ID == id {
			s.data.Profiles[i].SubscriptionURL = subscriptionURL
			s.data.Profiles[i].ProxyCount = proxyCount
			s.data.Profiles[i].WireGuardConfigs = wireGuardConfigs
			s.data.Profiles[i].LastUpdated = time.Now().Format("2006-01-02 15:04:05")
			return s.saveInternal()
		}
	}
	return fmt.Errorf("profile with ID %d not found", id)
}

// UpdateProfileWireGuard updates only WireGuard configs for a profile.
func (s *Storage) UpdateProfileWireGuard(id int, wireGuardConfigs []UserWireGuardConfig) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	for i := range s.data.Profiles {
		if s.data.Profiles[i].ID == id {
			s.data.Profiles[i].WireGuardConfigs = wireGuardConfigs
			return s.saveInternal()
		}
	}
	return fmt.Errorf("profile with ID %d not found", id)
}

// --- Sing-box Config ---

// UpdateProfileConfig updates the generated sing-box config for a profile.
func (s *Storage) UpdateProfileConfig(id int, config map[string]interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	for i := range s.data.Profiles {
		if s.data.Profiles[i].ID == id {
			s.data.Profiles[i].SingboxConfig = config
			return s.saveInternal()
		}
	}
	return fmt.Errorf("profile with ID %d not found", id)
}

// GetProfileConfig returns the sing-box config for a profile.
func (s *Storage) GetProfileConfig(id int) (map[string]interface{}, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	for i := range s.data.Profiles {
		if s.data.Profiles[i].ID == id {
			return s.data.Profiles[i].SingboxConfig, nil
		}
	}
	return nil, fmt.Errorf("profile with ID %d not found", id)
}

// WriteActiveConfigToFile writes the active profile's config to a temporary file for sing-box.
// This is needed because sing-box requires a file path.
func (s *Storage) WriteActiveConfigToFile() (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	activeID := s.data.App.ActiveProfileID
	
	for i := range s.data.Profiles {
		if s.data.Profiles[i].ID == activeID {
			config := s.data.Profiles[i].SingboxConfig
			if config == nil || len(config) == 0 {
				return "", fmt.Errorf("no config for profile %d", activeID)
			}
			
			// Write to temp config file
			configPath := filepath.Join(s.resourcesPath, "active_config.json")
			data, err := json.MarshalIndent(config, "", "  ")
			if err != nil {
				return "", fmt.Errorf("failed to marshal config: %w", err)
			}
			
			if err := os.WriteFile(configPath, data, 0644); err != nil {
				return "", fmt.Errorf("failed to write config: %w", err)
			}
			
			return configPath, nil
		}
	}
	
	return "", fmt.Errorf("active profile %d not found", activeID)
}

// --- Migration from old format ---

// ConfigBuilderForStorage provides config building functionality for Storage.
type ConfigBuilderForStorage struct {
	storage *Storage
	fetcher *SubscriptionFetcher
}

// NewConfigBuilderForStorage creates a config builder that works with Storage.
func NewConfigBuilderForStorage(storage *Storage) *ConfigBuilderForStorage {
	return &ConfigBuilderForStorage{
		storage: storage,
		fetcher: NewSubscriptionFetcher(),
	}
}

// TestSubscription tests a subscription URL and returns available proxies.
func (b *ConfigBuilderForStorage) TestSubscription(subscriptionURL string) (*SubscriptionTestResult, error) {
	result := &SubscriptionTestResult{
		Success: false,
		Proxies: []ProxyInfo{},
	}
	
	isDirectLink := isDirectProxyLink(subscriptionURL)
	
	var proxies []ProxyConfig
	var err error
	
	if isDirectLink {
		proxy, err := b.fetcher.ParseSingleLink(subscriptionURL)
		if err != nil {
			result.Error = fmt.Sprintf("Ошибка парсинга ссылки: %v", err)
			return result, nil
		}
		proxies = []ProxyConfig{proxy}
	} else {
		proxies, err = b.fetcher.FetchAndParse(subscriptionURL)
		if err != nil {
			result.Error = fmt.Sprintf("Ошибка загрузки подписки: %v", err)
			return result, nil
		}
	}
	
	if len(proxies) == 0 {
		result.Error = "Подписка не содержит доступных прокси"
		return result, nil
	}
	
	result.Success = true
	result.Count = len(proxies)
	result.IsDirectLink = isDirectLink
	
	for _, p := range proxies {
		result.Proxies = append(result.Proxies, ProxyInfo{
			Type:   p.Type,
			Name:   p.Name,
			Server: p.Server,
			Port:   p.ServerPort,
		})
	}
	
	return result, nil
}

// BuildConfig builds sing-box config for the active profile.
func (b *ConfigBuilderForStorage) BuildConfig(subscriptionURL string) error {
	profile, err := b.storage.GetActiveProfile()
	if err != nil || profile == nil {
		return fmt.Errorf("no active profile")
	}
	
	return b.BuildConfigForProfile(profile.ID, subscriptionURL, profile.WireGuardConfigs)
}

// BuildConfigForProfile builds sing-box config for a specific profile.
func (b *ConfigBuilderForStorage) BuildConfigForProfile(profileID int, subscriptionURL string, wireGuardConfigs []UserWireGuardConfig) error {
	// Load template
	templateData, err := os.ReadFile(b.storage.templatePath)
	if err != nil {
		return fmt.Errorf("не удалось загрузить template.json: %w", err)
	}
	
	var template map[string]interface{}
	if err := json.Unmarshal(templateData, &template); err != nil {
		return fmt.Errorf("ошибка парсинга template.json: %w", err)
	}
	
	// Get proxies from subscription
	var proxies []ProxyConfig
	
	if subscriptionURL != "" {
		isDirectLink := isDirectProxyLink(subscriptionURL)
		
		if isDirectLink {
			proxy, err := b.fetcher.ParseSingleLink(subscriptionURL)
			if err != nil {
				return fmt.Errorf("ошибка парсинга ссылки: %w", err)
			}
			proxy.Tag = generateTag(proxy, 0)
			proxies = []ProxyConfig{proxy}
		} else {
			proxies, err = b.fetcher.FetchAndParse(subscriptionURL)
			if err != nil {
				return fmt.Errorf("ошибка загрузки подписки: %w", err)
			}
			for i := range proxies {
				proxies[i].Tag = generateTag(proxies[i], i)
			}
		}
	}
	
	// Generate outbounds
	outbounds := b.generateOutbounds(template, proxies)
	template["outbounds"] = outbounds
	
	// Add WireGuard endpoints
	b.addWireGuardEndpoints(template, wireGuardConfigs)
	
	// Add DNS servers for WireGuard
	b.addWireGuardDNS(template, wireGuardConfigs)
	
	// Update route rules for WireGuard
	b.updateRouteRulesForWireGuard(template, wireGuardConfigs)
	
	// Add experimental section
	b.addExperimentalAPI(template)
	
	// Remove template fields
	delete(template, "outbounds_template")
	delete(template, "_comment_outbounds")
	
	// Update profile in storage
	if err := b.storage.UpdateProfileSubscription(profileID, subscriptionURL, len(proxies), wireGuardConfigs); err != nil {
		return err
	}
	
	if err := b.storage.UpdateProfileConfig(profileID, template); err != nil {
		return err
	}
	
	return nil
}

// generateOutbounds generates outbounds list.
func (b *ConfigBuilderForStorage) generateOutbounds(template map[string]interface{}, proxies []ProxyConfig) []interface{} {
	outbounds := []interface{}{}
	proxyTags := []string{}
	
	for _, p := range proxies {
		outbounds = append(outbounds, p.ToSingboxOutbound())
		proxyTags = append(proxyTags, p.Tag)
	}
	
	outboundsTemplate, ok := template["outbounds_template"].(map[string]interface{})
	if !ok {
		outboundsTemplate = map[string]interface{}{}
	}
	
	if len(proxyTags) > 0 {
		if urltest, ok := outboundsTemplate["urltest"].(map[string]interface{}); ok {
			urltest = copyMap(urltest)
			urltest["outbounds"] = proxyTags
			outbounds = append(outbounds, urltest)
		} else {
			outbounds = append(outbounds, map[string]interface{}{
				"type":      "urltest",
				"tag":       "auto-select",
				"outbounds": proxyTags,
				"url":       "https://www.gstatic.com/generate_204",
				"interval":  "3m",
				"tolerance": 50,
			})
		}
		
		selectorOutbounds := append([]string{"auto-select"}, proxyTags...)
		selectorOutbounds = append(selectorOutbounds, "direct")
		
		if selector, ok := outboundsTemplate["selector"].(map[string]interface{}); ok {
			selector = copyMap(selector)
			selector["outbounds"] = selectorOutbounds
			outbounds = append(outbounds, selector)
		} else {
			outbounds = append(outbounds, map[string]interface{}{
				"type":      "selector",
				"tag":       "proxy",
				"outbounds": selectorOutbounds,
				"default":   "auto-select",
			})
		}
	} else {
		outbounds = append(outbounds, map[string]interface{}{
			"type":      "selector",
			"tag":       "proxy",
			"outbounds": []string{"direct"},
			"default":   "direct",
		})
	}
	
	if direct, ok := outboundsTemplate["direct"].(map[string]interface{}); ok {
		outbounds = append(outbounds, copyMap(direct))
	} else {
		outbounds = append(outbounds, map[string]interface{}{
			"type": "direct",
			"tag":  "direct",
		})
	}
	
	if block, ok := outboundsTemplate["block"].(map[string]interface{}); ok {
		outbounds = append(outbounds, copyMap(block))
	} else {
		outbounds = append(outbounds, map[string]interface{}{
			"type": "block",
			"tag":  "block",
		})
	}
	
	if dnsOut, ok := outboundsTemplate["dns-out"].(map[string]interface{}); ok {
		outbounds = append(outbounds, copyMap(dnsOut))
	} else {
		outbounds = append(outbounds, map[string]interface{}{
			"type": "dns",
			"tag":  "dns-out",
		})
	}
	
	return outbounds
}

// addWireGuardEndpoints adds WireGuard to endpoints section.
func (b *ConfigBuilderForStorage) addWireGuardEndpoints(template map[string]interface{}, wireGuardConfigs []UserWireGuardConfig) {
	if len(wireGuardConfigs) == 0 {
		return
	}
	
	endpoints, _ := template["endpoints"].([]interface{})
	if endpoints == nil {
		endpoints = []interface{}{}
	}
	
	for _, wg := range wireGuardConfigs {
		// Use the existing ToSingboxEndpoint method
		endpoint := wg.ToSingboxEndpoint()
		endpoints = append(endpoints, endpoint)
	}
	
	template["endpoints"] = endpoints
}

// addWireGuardDNS adds DNS servers for WireGuard networks.
func (b *ConfigBuilderForStorage) addWireGuardDNS(template map[string]interface{}, wireGuardConfigs []UserWireGuardConfig) {
	if len(wireGuardConfigs) == 0 {
		return
	}
	
	dns, ok := template["dns"].(map[string]interface{})
	if !ok {
		return
	}
	
	servers, _ := dns["servers"].([]interface{})
	if servers == nil {
		servers = []interface{}{}
	}
	
	for _, wg := range wireGuardConfigs {
		if wg.DNS == "" {
			continue
		}
		
		serverTag := fmt.Sprintf("%s-dns", wg.Tag)
		server := map[string]interface{}{
			"tag":      serverTag,
			"address":  wg.DNS,
			"detour":   wg.Tag,
			"strategy": "ipv4_only",
		}
		servers = append(servers, server)
	}
	
	dns["servers"] = servers
}

// updateRouteRulesForWireGuard updates route rules for WireGuard.
func (b *ConfigBuilderForStorage) updateRouteRulesForWireGuard(template map[string]interface{}, wireGuardConfigs []UserWireGuardConfig) {
	if len(wireGuardConfigs) == 0 {
		return
	}
	
	route, ok := template["route"].(map[string]interface{})
	if !ok {
		route = map[string]interface{}{}
		template["route"] = route
	}
	
	rules, _ := route["rules"].([]interface{})
	if rules == nil {
		rules = []interface{}{}
	}
	
	// Use existing GenerateRouteRulesForWireGuard function
	newRules := GenerateRouteRulesForWireGuard(wireGuardConfigs)
	
	// Convert to []interface{}
	newRulesInterface := make([]interface{}, len(newRules))
	for i, r := range newRules {
		newRulesInterface[i] = r
	}
	
	// Prepend new rules to existing ones
	newRulesInterface = append(newRulesInterface, rules...)
	route["rules"] = newRulesInterface
}

// addExperimentalAPI adds experimental section for traffic stats.
func (b *ConfigBuilderForStorage) addExperimentalAPI(template map[string]interface{}) {
	experimental, ok := template["experimental"].(map[string]interface{})
	if !ok {
		experimental = map[string]interface{}{}
		template["experimental"] = experimental
	}
	
	clashAPI, ok := experimental["clash_api"].(map[string]interface{})
	if !ok {
		experimental["clash_api"] = map[string]interface{}{
			"external_controller": "127.0.0.1:9090",
		}
	} else {
		if _, exists := clashAPI["external_controller"]; !exists {
			clashAPI["external_controller"] = "127.0.0.1:9090"
		}
	}
}

// isDirectProxyLink checks if URL is a direct proxy link.
func isDirectProxyLink(url string) bool {
	if len(url) < 5 {
		return false
	}
	return strings.HasPrefix(url, "vless://") ||
		strings.HasPrefix(url, "trojan://") ||
		strings.HasPrefix(url, "ss://") ||
		strings.HasPrefix(url, "vmess://")
}

// GetUserSettings returns user settings for active profile (compatibility method).
func (s *Storage) GetUserSettings() (*UserSettings, error) {
	profile, err := s.GetActiveProfile()
	if err != nil || profile == nil {
		return &UserSettings{}, nil
	}
	
	return &UserSettings{
		SubscriptionURL:  profile.SubscriptionURL,
		LastUpdated:      profile.LastUpdated,
		ProxyCount:       profile.ProxyCount,
		WireGuardConfigs: profile.WireGuardConfigs,
	}, nil
}

// GetConfigPath returns path to active config file (written on demand).
func (s *Storage) GetConfigPath() (string, error) {
	return s.WriteActiveConfigToFile()
}

// MigrateFromOldFormat migrates data from old file structure to new settings.json.
// Only migrates if settings.json didn't exist before (was just created).
func (s *Storage) MigrateFromOldFormat(basePath string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	// Skip migration if we already have profiles with data (settings.json existed)
	if len(s.data.Profiles) > 0 && s.data.Profiles[0].SubscriptionURL != "" {
		return nil // Already have data, skip migration
	}
	
	migrated := false
	
	// Try to migrate old profiles.json
	oldProfilesPath := filepath.Join(basePath, "profiles.json")
	if fileExists(oldProfilesPath) {
		data, err := os.ReadFile(oldProfilesPath)
		if err == nil {
			var oldProfiles []ConnectionProfile
			if json.Unmarshal(data, &oldProfiles) == nil {
				for _, oldP := range oldProfiles {
					// Check if profile already exists
					exists := false
					for _, p := range s.data.Profiles {
						if p.ID == oldP.ID {
							exists = true
							break
						}
					}
					
					if !exists {
						s.data.Profiles = append(s.data.Profiles, ProfileData{
							ID:        oldP.ID,
							Name:      oldP.Name,
							CreatedAt: oldP.CreatedAt,
						})
					}
				}
				migrated = true
			}
		}
	}
	
	// Try to migrate old user_settings files
	for i := range s.data.Profiles {
		profileID := s.data.Profiles[i].ID
		
		var settingsPath string
		if profileID == DefaultProfileID {
			settingsPath = filepath.Join(basePath, "user_settings.json")
		} else {
			settingsPath = filepath.Join(basePath, fmt.Sprintf("user_settings_%d.json", profileID))
		}
		
		if fileExists(settingsPath) {
			data, err := os.ReadFile(settingsPath)
			if err == nil {
				var oldSettings UserSettings
				if json.Unmarshal(data, &oldSettings) == nil {
					s.data.Profiles[i].SubscriptionURL = oldSettings.SubscriptionURL
					s.data.Profiles[i].LastUpdated = oldSettings.LastUpdated
					s.data.Profiles[i].ProxyCount = oldSettings.ProxyCount
					s.data.Profiles[i].WireGuardConfigs = oldSettings.WireGuardConfigs
					migrated = true
				}
			}
		}
		
		// Try to migrate old config files
		var configPath string
		if profileID == DefaultProfileID {
			configPath = filepath.Join(basePath, "config.json")
		} else {
			configPath = filepath.Join(basePath, fmt.Sprintf("config_%d.json", profileID))
		}
		
		if fileExists(configPath) {
			data, err := os.ReadFile(configPath)
			if err == nil {
				var oldConfig map[string]interface{}
				if json.Unmarshal(data, &oldConfig) == nil {
					s.data.Profiles[i].SingboxConfig = oldConfig
					migrated = true
				}
			}
		}
	}
	
	// Migrate old app_config.json
	oldAppConfigPath := filepath.Join(os.Getenv("LOCALAPPDATA"), "KampusVPN", "app_config.json")
	if fileExists(oldAppConfigPath) {
		data, err := os.ReadFile(oldAppConfigPath)
		if err == nil {
			var oldConfig AppConfig
			if json.Unmarshal(data, &oldConfig) == nil {
				s.data.App.AutoStart = oldConfig.AutoStart
				s.data.App.Notifications = oldConfig.Notifications
				s.data.App.CheckUpdates = oldConfig.CheckUpdates
				s.data.App.EnableLogging = oldConfig.EnableLogging
				s.data.App.LogLevel = oldConfig.LogLevel
				s.data.App.Theme = oldConfig.Theme
				s.data.App.Language = oldConfig.Language
				s.data.App.AutoUpdateSub = oldConfig.AutoUpdateSub
				s.data.App.SubUpdateInterval = oldConfig.SubUpdateInterval
				s.data.App.LastSubUpdate = oldConfig.LastSubUpdate
				s.data.App.LastUpdateCheck = oldConfig.LastUpdateCheck
				s.data.App.ActiveProfileID = oldConfig.ActiveProfileID
				migrated = true
				// Remove old file after migration
				os.Remove(oldAppConfigPath)
			}
		}
	}
	
	if migrated {
		// Remove old files after successful migration
		os.Remove(filepath.Join(basePath, "profiles.json"))
		os.Remove(filepath.Join(basePath, "user_settings.json"))
		os.Remove(filepath.Join(basePath, "config.json"))
		// Remove profile-specific old files
		for i := 2; i <= 10; i++ {
			os.Remove(filepath.Join(basePath, fmt.Sprintf("user_settings_%d.json", i)))
			os.Remove(filepath.Join(basePath, fmt.Sprintf("config_%d.json", i)))
		}
		return s.saveInternal()
	}
	
	return nil
}
