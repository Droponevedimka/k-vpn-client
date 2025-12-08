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
	
	// Routing settings
	RoutingMode RoutingMode `json:"routing_mode"` // How traffic is routed: blocked_only, except_russia, all_traffic
	
	// Subscription settings
	AutoUpdateSub     bool      `json:"auto_update_sub"`
	SubUpdateInterval int       `json:"sub_update_interval"`
	LastSubUpdate     time.Time `json:"last_sub_update"`
	
	// Update tracking
	LastUpdateCheck string `json:"last_update_check"`
	
	// Active profile
	ActiveProfileID int `json:"active_profile_id"`
	
	// WireGuard settings
	WireGuardVersion string `json:"wireguard_version"` // Native WireGuard version (e.g., "0.5.3")
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
	}
	
	// Ensure default profile exists (ID=1, cannot be deleted)
	hasDefaultProfile := false
	for _, p := range s.data.Profiles {
		if p.ID == DefaultProfileID {
			hasDefaultProfile = true
			break
		}
	}
	if !hasDefaultProfile {
		// Insert default profile at the beginning
		s.data.Profiles = append([]ProfileData{s.createDefaultProfile()}, s.data.Profiles...)
	}
	
	// Ensure active profile ID is valid
	if s.data.App.ActiveProfileID <= 0 {
		s.data.App.ActiveProfileID = DefaultProfileID
	} else {
		// Check if active profile exists
		activeExists := false
		for _, p := range s.data.Profiles {
			if p.ID == s.data.App.ActiveProfileID {
				activeExists = true
				break
			}
		}
		if !activeExists {
			s.data.App.ActiveProfileID = DefaultProfileID
		}
	}
	
	return s.saveInternal()
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
			LogLevel:          LogLevelInfo, // Info by default
			Theme:             ThemeDark,
			Language:          LangRussian,
			RoutingMode:       DefaultRoutingMode, // blocked_only by default
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
// Always returns a valid profile ID (at least DefaultProfileID).
func (s *Storage) GetActiveProfileID() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	activeID := s.data.App.ActiveProfileID
	
	// If not set or invalid, return default
	if activeID <= 0 {
		return DefaultProfileID
	}
	
	// Verify the profile exists
	for _, p := range s.data.Profiles {
		if p.ID == activeID {
			return activeID
		}
	}
	
	// Profile doesn't exist, return default
	return DefaultProfileID
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
		return fmt.Errorf("нельзя удалить профиль по умолчанию")
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

// ReplaceAllProfiles replaces ALL profiles with imported ones.
// This is used for full import - all existing profiles are removed and replaced.
func (s *Storage) ReplaceAllProfiles(profiles []ProfileData) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	if len(profiles) == 0 {
		return fmt.Errorf("cannot import empty profiles list")
	}
	
	// Ensure at least one profile has ID=1 (default profile)
	hasDefault := false
	for _, p := range profiles {
		if p.ID == DefaultProfileID {
			hasDefault = true
			break
		}
	}
	
	if !hasDefault {
		// Set first profile as default
		profiles[0].ID = DefaultProfileID
	}
	
	// Replace all profiles
	s.data.Profiles = profiles
	
	// Validate active profile ID
	activeExists := false
	for _, p := range profiles {
		if p.ID == s.data.App.ActiveProfileID {
			activeExists = true
			break
		}
	}
	
	if !activeExists {
		// Set to default profile
		s.data.App.ActiveProfileID = DefaultProfileID
	}
	
	return s.saveInternal()
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
			
			// WireGuard is now managed by Native WireGuard Manager
			// Remove old WireGuard outbounds from config if present
			s.removeWireGuardFromConfig(config)
			
			// Clean up deprecated/problematic fields
			// Remove endpoints (WireGuard is managed separately)
			delete(config, "endpoints")
			
			// Remove log output to make sing-box write to stdout
			if logSection, ok := config["log"].(map[string]interface{}); ok {
				delete(logSection, "output")
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

// removeWireGuardFromConfig removes WireGuard outbounds and related DNS/route rules
// WireGuard is now managed by Native WireGuard Manager
func (s *Storage) removeWireGuardFromConfig(config map[string]interface{}) {
	// Remove WireGuard outbounds
	if outbounds, ok := config["outbounds"].([]interface{}); ok {
		filtered := []interface{}{}
		for _, ob := range outbounds {
			if obMap, ok := ob.(map[string]interface{}); ok {
				if obType, _ := obMap["type"].(string); obType != "wireguard" {
					filtered = append(filtered, ob)
				}
			}
		}
		config["outbounds"] = filtered
	}
	
	// Remove dns-wg-* servers and rules
	if dns, ok := config["dns"].(map[string]interface{}); ok {
		// Remove WireGuard DNS servers
		if servers, ok := dns["servers"].([]interface{}); ok {
			filtered := []interface{}{}
			for _, srv := range servers {
				if srvMap, ok := srv.(map[string]interface{}); ok {
					if tag, _ := srvMap["tag"].(string); !strings.HasPrefix(tag, "dns-wg-") {
						filtered = append(filtered, srv)
					}
				}
			}
			dns["servers"] = filtered
		}
		
		// Remove WireGuard DNS rules (those with dns-wg-* server)
		if rules, ok := dns["rules"].([]interface{}); ok {
			filtered := []interface{}{}
			for _, rule := range rules {
				if ruleMap, ok := rule.(map[string]interface{}); ok {
					if server, _ := ruleMap["server"].(string); !strings.HasPrefix(server, "dns-wg-") {
						filtered = append(filtered, rule)
					}
				}
			}
			dns["rules"] = filtered
		}
	}
	
	// Remove WireGuard route rules (those routing to wg-* outbounds)
	if route, ok := config["route"].(map[string]interface{}); ok {
		if rules, ok := route["rules"].([]interface{}); ok {
			filtered := []interface{}{}
			for _, rule := range rules {
				if ruleMap, ok := rule.(map[string]interface{}); ok {
					if outbound, _ := ruleMap["outbound"].(string); !strings.HasPrefix(outbound, "wg-") {
						filtered = append(filtered, rule)
					}
				}
			}
			route["rules"] = filtered
		}
	}
}

// migrateWireGuardDNS ensures DNS rules for .local domains exist
func (s *Storage) migrateWireGuardDNS(config map[string]interface{}, wireGuardConfigs []UserWireGuardConfig) {
	if len(wireGuardConfigs) == 0 {
		return
	}
	
	dns, ok := config["dns"].(map[string]interface{})
	if !ok {
		return
	}
	
	servers, ok := dns["servers"].([]interface{})
	if !ok {
		servers = []interface{}{}
	}
	
	rules, _ := dns["rules"].([]interface{})
	if rules == nil {
		rules = []interface{}{}
	}
	
	for _, wg := range wireGuardConfigs {
		if wg.DNS == "" {
			continue
		}
		
		dnsTag := wg.Tag + "-dns"
		
		// Check if DNS server exists
		serverExists := false
		for _, srv := range servers {
			if srvMap, ok := srv.(map[string]interface{}); ok {
				if tag, ok := srvMap["tag"].(string); ok && tag == dnsTag {
					serverExists = true
					break
				}
			}
		}
		
		// Add DNS server if not exists
		if !serverExists {
			servers = append(servers, map[string]interface{}{
				"type":   "udp",
				"tag":    dnsTag,
				"server": wg.DNS,
				"detour": wg.Tag,
			})
		}
		
		// Check if .local DNS rule exists
		localRuleExists := false
		for _, rule := range rules {
			if ruleMap, ok := rule.(map[string]interface{}); ok {
				if server, ok := ruleMap["server"].(string); ok && server == dnsTag {
					localRuleExists = true
					break
				}
			}
		}
		
		// Add .local DNS rule at the beginning if not exists
		if !localRuleExists {
			localRule := map[string]interface{}{
				"domain_suffix": []string{".local", "." + wg.Tag + ".local"},
				"action":        "route",
				"server":        dnsTag,
			}
			rules = append([]interface{}{localRule}, rules...)
		}
	}
	
	dns["servers"] = servers
	dns["rules"] = rules
}

// migrateWireGuardRouteRules ensures route rules for WireGuard AllowedIPs exist
// Порядок: WireGuard IP rules → ip_is_private → geosite-ru → proxy
func (s *Storage) migrateWireGuardRouteRules(config map[string]interface{}, wireGuardConfigs []UserWireGuardConfig) {
	if len(wireGuardConfigs) == 0 {
		return
	}
	
	route, ok := config["route"].(map[string]interface{})
	if !ok {
		return
	}
	
	rules, _ := route["rules"].([]interface{})
	if rules == nil {
		rules = []interface{}{}
	}
	
	// Находим позицию после hijack-dns (перед ip_is_private)
	insertIdx := 0
	for i, rule := range rules {
		if ruleMap, ok := rule.(map[string]interface{}); ok {
			action, _ := ruleMap["action"].(string)
			if action == "hijack-dns" {
				insertIdx = i + 1
				break
			}
			if action == "sniff" {
				insertIdx = i + 1
			}
		}
	}
	
	// Проверяем и добавляем IP rules для каждого WireGuard
	for _, wg := range wireGuardConfigs {
		if len(wg.AllowedIPs) == 0 {
			continue
		}
		
		// Проверяем существует ли уже правило для этого WireGuard
		ruleExists := false
		for _, rule := range rules {
			if ruleMap, ok := rule.(map[string]interface{}); ok {
				if outbound, ok := ruleMap["outbound"].(string); ok && outbound == wg.Tag {
					if _, hasIP := ruleMap["ip_cidr"]; hasIP {
						ruleExists = true
						break
					}
				}
			}
		}
		
		// Добавляем правило если не существует
		if !ruleExists {
			ipRule := map[string]interface{}{
				"ip_cidr":  wg.AllowedIPs,
				"outbound": wg.Tag,
			}
			// Вставляем в позицию insertIdx
			newRules := make([]interface{}, 0, len(rules)+1)
			newRules = append(newRules, rules[:insertIdx]...)
			newRules = append(newRules, ipRule)
			newRules = append(newRules, rules[insertIdx:]...)
			rules = newRules
			insertIdx++ // Сдвигаем позицию для следующего WireGuard
		}
	}
	
	route["rules"] = rules
}

// --- Migration from old format ---

// ConfigBuilderForStorage provides config building functionality for Storage.
type ConfigBuilderForStorage struct {
	storage       *Storage
	fetcher       *SubscriptionFetcher
	routingMode   RoutingMode
	filterManager *FilterManager
}

// NewConfigBuilderForStorage creates a config builder that works with Storage.
func NewConfigBuilderForStorage(storage *Storage) *ConfigBuilderForStorage {
	// Filter manager path: go up from resources to parent, then bin/filters
	basePath := filepath.Dir(storage.resourcesPath)
	
	return &ConfigBuilderForStorage{
		storage:       storage,
		fetcher:       NewSubscriptionFetcher(),
		routingMode:   DefaultRoutingMode,
		filterManager: NewFilterManager(basePath),
	}
}

// SetRoutingMode sets the routing mode for config generation
func (b *ConfigBuilderForStorage) SetRoutingMode(mode RoutingMode) {
	b.routingMode = mode
}

// GetRoutingMode returns current routing mode
func (b *ConfigBuilderForStorage) GetRoutingMode() RoutingMode {
	return b.routingMode
}

// GetFilterManager returns the filter manager
func (b *ConfigBuilderForStorage) GetFilterManager() *FilterManager {
	return b.filterManager
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

	// Filter unsupported transports (e.g., xhttp which is Xray-only)
	filterResult := FilterUnsupportedTransports(proxies)
	proxies = filterResult.Supported
	
	if len(proxies) == 0 {
		if filterResult.AllFiltered {
			result.Error = filterResult.Message
		} else {
			result.Error = "Подписка не содержит доступных прокси"
		}
		return result, nil
	}
	
	result.Success = true
	result.Count = len(proxies)
	result.IsDirectLink = isDirectLink

	// Add warning about filtered proxies
	if len(filterResult.Filtered) > 0 {
		result.Warning = filterResult.Message
		result.FilteredCount = len(filterResult.Filtered)
	}
	
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
	fmt.Printf("[BuildConfigForProfile] Called with profileID=%d, %d WireGuard configs\n", profileID, len(wireGuardConfigs))
	for i, wg := range wireGuardConfigs {
		fmt.Printf("[BuildConfigForProfile] WireGuard[%d]: tag=%s, dns=%s, allowedIPs=%v\n", i, wg.Tag, wg.DNS, wg.AllowedIPs)
	}
	
	// Load template
	templateData, err := os.ReadFile(b.storage.templatePath)
	if err != nil {
		return fmt.Errorf("не удалось загрузить template.json: %w", err)
	}
	
	var template map[string]interface{}
	if err := json.Unmarshal(templateData, &template); err != nil {
		return fmt.Errorf("ошибка парсинга template.json: %w", err)
	}
	
	// Disable strict_route when WireGuard is used to allow system routes to work
	fmt.Printf("[BuildConfigForProfile] Configuring TUN for WireGuard compatibility...\n")
	b.disableStrictRouteForWireGuard(template, wireGuardConfigs)
	
	// Add DNS servers and rules for WireGuard networks
	// (WireGuard works natively, DNS queries go through direct and WireGuard interface handles routing)
	fmt.Printf("[BuildConfigForProfile] Adding WireGuard DNS rules for %d configs...\n", len(wireGuardConfigs))
	b.addWireGuardDNSNew(template, wireGuardConfigs)
	
	// Update route rules for WireGuard AllowedIPs
	fmt.Printf("[BuildConfigForProfile] Adding WireGuard route rules...\n")
	b.updateRouteRulesForWireGuardNew(template, wireGuardConfigs)
	
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

		// Filter unsupported transports (e.g., xhttp which is Xray-only)
		filterResult := FilterUnsupportedTransports(proxies)
		if filterResult.AllFiltered {
			return fmt.Errorf("%s", filterResult.Message)
		}
		if len(filterResult.Filtered) > 0 {
			fmt.Printf("[BuildConfigForProfile] Warning: %s\n", filterResult.Message)
		}
		proxies = filterResult.Supported
	}
	
	// Generate outbounds
	outbounds := b.generateOutbounds(template, proxies)
	template["outbounds"] = outbounds
	
	// WireGuard is now managed by Native WireGuard Manager
	// Remove any existing WireGuard from config
	delete(template, "endpoints")
	
	// Apply routing mode (blocked_only, except_russia, all_traffic)
	b.applyRoutingMode(template)
	
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
	
	// block и dns-out удалены - в sing-box 1.11+ используются rule actions
	// action: "reject" вместо outbound: "block"
	// action: "hijack-dns" вместо outbound: "dns-out"
	
	return outbounds
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
		// New sing-box 1.12+ DNS server format
		server := map[string]interface{}{
			"type":   "udp",
			"tag":    serverTag,
			"server": wg.DNS,
			"detour": wg.Tag,
		}
		servers = append(servers, server)
	}
	
	dns["servers"] = servers
}

// disableStrictRouteForWireGuard disables strict_route in TUN when WireGuard is used.
// This allows system routes (WireGuard interface) to work alongside sing-box TUN.
func (b *ConfigBuilderForStorage) disableStrictRouteForWireGuard(template map[string]interface{}, wireGuardConfigs []UserWireGuardConfig) {
	if len(wireGuardConfigs) == 0 {
		return
	}
	
	inbounds, ok := template["inbounds"].([]interface{})
	if !ok {
		return
	}
	
	for i, inbound := range inbounds {
		if inboundMap, ok := inbound.(map[string]interface{}); ok {
			if inboundMap["type"] == "tun" {
				// Disable strict_route to allow WireGuard routes to work
				inboundMap["strict_route"] = false
				inbounds[i] = inboundMap
				fmt.Printf("[disableStrictRouteForWireGuard] Disabled strict_route for TUN\n")
				break
			}
		}
	}
	
	template["inbounds"] = inbounds
}

// addWireGuardDNSNew adds DNS servers for WireGuard networks (native WireGuard mode).
// DNS queries go through "direct" - the WireGuard interface handles routing.
func (b *ConfigBuilderForStorage) addWireGuardDNSNew(template map[string]interface{}, wireGuardConfigs []UserWireGuardConfig) {
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
	
	dnsRules, _ := dns["rules"].([]interface{})
	if dnsRules == nil {
		dnsRules = []interface{}{}
	}
	
	for _, wg := range wireGuardConfigs {
		if wg.DNS == "" {
			continue
		}
		
		dnsTag := fmt.Sprintf("dns-%s", wg.Tag)
		
		// Add DNS server - no special binding needed
		// Traffic to DNS server IP will be excluded from TUN and go through WireGuard
		server := map[string]interface{}{
			"type":        "udp",
			"tag":         dnsTag,
			"server":      wg.DNS,
			"server_port": 53,
		}
		servers = append(servers, server)
		
		// Build domain suffixes for DNS rule
		domainSuffixes := []string{}
		if wg.Endpoint != "" {
			parts := strings.Split(wg.Endpoint, ".")
			if len(parts) >= 2 {
				baseDomain := "." + strings.Join(parts[len(parts)-2:], ".")
				domainSuffixes = append(domainSuffixes, baseDomain)
			}
		}
		domainSuffixes = append(domainSuffixes, ".local", fmt.Sprintf(".%s.local", wg.Tag))
		
		// Add DNS rule at the beginning
		dnsRule := map[string]interface{}{
			"domain_suffix": domainSuffixes,
			"action":        "route",
			"server":        dnsTag,
		}
		dnsRules = append([]interface{}{dnsRule}, dnsRules...)
		
		fmt.Printf("[addWireGuardDNSNew] Added DNS server %s (%s) for domains: %v\n", dnsTag, wg.DNS, domainSuffixes)
	}
	
	dns["servers"] = servers
	dns["rules"] = dnsRules
}

// updateRouteRulesForWireGuardNew updates route rules for WireGuard (native mode).
// Traffic goes through "direct" - the WireGuard interface handles routing based on AllowedIPs.
func (b *ConfigBuilderForStorage) updateRouteRulesForWireGuardNew(template map[string]interface{}, wireGuardConfigs []UserWireGuardConfig) {
	if len(wireGuardConfigs) == 0 {
		return
	}
	
	route, ok := template["route"].(map[string]interface{})
	if !ok {
		return
	}
	
	rules, ok := route["rules"].([]interface{})
	if !ok {
		rules = []interface{}{}
	}
	
	// Collect all AllowedIPs from WireGuard configs
	allWireGuardCIDRs := []string{}
	for _, wg := range wireGuardConfigs {
		allWireGuardCIDRs = append(allWireGuardCIDRs, wg.AllowedIPs...)
	}
	
	if len(allWireGuardCIDRs) == 0 {
		return
	}
	
	// Find position after hijack-dns
	insertIdx := 0
	for i, rule := range rules {
		if ruleMap, ok := rule.(map[string]interface{}); ok {
			action, _ := ruleMap["action"].(string)
			if action == "hijack-dns" {
				insertIdx = i + 1
				break
			}
			if action == "sniff" {
				insertIdx = i + 1
			}
		}
	}
	
	// Create route rule for WireGuard networks
	wgRule := map[string]interface{}{
		"ip_cidr":  allWireGuardCIDRs,
		"outbound": "direct",
	}
	
	// Insert rule after hijack-dns
	finalRules := make([]interface{}, 0, len(rules)+1)
	finalRules = append(finalRules, rules[:insertIdx]...)
	finalRules = append(finalRules, wgRule)
	finalRules = append(finalRules, rules[insertIdx:]...)
	
	route["rules"] = finalRules
	
	fmt.Printf("[updateRouteRulesForWireGuardNew] Added route rule for CIDRs: %v at position %d\n", allWireGuardCIDRs, insertIdx)
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

// applyRoutingMode applies routing rules based on the selected routing mode.
func (b *ConfigBuilderForStorage) applyRoutingMode(template map[string]interface{}) {
	route, ok := template["route"].(map[string]interface{})
	if !ok {
		route = map[string]interface{}{}
		template["route"] = route
	}

	// Clean up DNS rules that reference remote rule_sets (geosite-*)
	b.cleanupDNSRuleSets(template)

	switch b.routingMode {
	case RoutingModeBlockedOnly:
		// Only blocked sites through VPN - use Re:filter + community rule-sets
		b.applyBlockedOnlyMode(route)
		
	case RoutingModeExceptRussia:
		// All except Russia through VPN - use built-in RU domain list
		b.applyExceptRussiaMode(route)
		
	case RoutingModeAllTraffic:
		// All traffic through VPN - remove direct rules for Russia
		b.applyAllTrafficMode(route)
		
	default:
		// Unknown mode, use blocked_only as safest default
		fmt.Printf("[applyRoutingMode] Unknown mode %s, using blocked_only\n", b.routingMode)
		b.applyBlockedOnlyMode(route)
	}
}

// cleanupDNSRuleSets removes DNS rules that reference remote rule_sets (geosite-*).
// These are not available in blocked_only and all_traffic modes.
func (b *ConfigBuilderForStorage) cleanupDNSRuleSets(template map[string]interface{}) {
	dns, ok := template["dns"].(map[string]interface{})
	if !ok {
		return
	}

	rules, ok := dns["rules"].([]interface{})
	if !ok {
		return
	}

	// Filter out rules that use rule_set with geosite-*
	newRules := make([]interface{}, 0, len(rules))
	for _, rule := range rules {
		ruleMap, ok := rule.(map[string]interface{})
		if !ok {
			newRules = append(newRules, rule)
			continue
		}

		// Check if this rule uses rule_set
		if ruleSet, hasRuleSet := ruleMap["rule_set"]; hasRuleSet {
			// Skip rules with geosite-* rule_sets
			if ruleSetArr, ok := ruleSet.([]interface{}); ok {
				hasGeosite := false
				for _, rs := range ruleSetArr {
					if rsStr, ok := rs.(string); ok {
						if strings.HasPrefix(rsStr, "geosite-") || strings.HasPrefix(rsStr, "geoip-") {
							hasGeosite = true
							break
						}
					}
				}
				if hasGeosite {
					fmt.Printf("[cleanupDNSRuleSets] Removed DNS rule with remote rule_set: %v\n", ruleSet)
					continue
				}
			}
		}

		newRules = append(newRules, rule)
	}

	dns["rules"] = newRules
}

// applyBlockedOnlyMode configures routing for blocked sites only.
func (b *ConfigBuilderForStorage) applyBlockedOnlyMode(route map[string]interface{}) {
	fmt.Printf("[applyRoutingMode] Using blocked_only mode with local filters\n")

	// Get local filter rule_sets
	filterRuleSets := b.filterManager.GetRuleSetConfigs()
	if len(filterRuleSets) == 0 {
		fmt.Printf("[applyRoutingMode] WARNING: No filter files found, falling back to except_russia\n")
		return
	}

	// Build new rule_set array with only local filters
	newRuleSets := make([]interface{}, 0, len(filterRuleSets))
	for _, rs := range filterRuleSets {
		newRuleSets = append(newRuleSets, rs)
	}
	route["rule_set"] = newRuleSets

	// Build new rules for blocked_only mode
	newRules := []interface{}{
		// 1. Sniff for protocol detection
		map[string]interface{}{
			"action": "sniff",
		},
		// 2. Local domains direct
		map[string]interface{}{
			"domain_suffix": []string{".local", ".internal", ".corp", ".lan", ".home", ".intranet", ".private"},
			"action":        "route",
			"outbound":      "direct",
		},
		// 3. Hijack DNS
		map[string]interface{}{
			"protocol": "dns",
			"action":   "hijack-dns",
		},
		// 4. Private IPs direct
		map[string]interface{}{
			"ip_is_private": true,
			"action":        "route",
			"outbound":      "direct",
		},
	}

	// 5. Add rules for blocked domains/IPs through proxy
	newRules = append(newRules, map[string]interface{}{
		"rule_set": []string{"refilter-domains"},
		"action":   "route",
		"outbound": "proxy",
	})
	
	newRules = append(newRules, map[string]interface{}{
		"rule_set": []string{"refilter-ips"},
		"action":   "route",
		"outbound": "proxy",
	})
	
	newRules = append(newRules, map[string]interface{}{
		"rule_set": []string{"community-domains"},
		"action":   "route",
		"outbound": "proxy",
	})
	
	newRules = append(newRules, map[string]interface{}{
		"rule_set": []string{"community-ips"},
		"action":   "route",
		"outbound": "proxy",
	})
	
	newRules = append(newRules, map[string]interface{}{
		"rule_set": []string{"discord-ips"},
		"action":   "route",
		"outbound": "proxy",
	})

	route["rules"] = newRules
	route["final"] = "direct"
	
	fmt.Printf("[applyRoutingMode] Applied blocked_only: %d rule_sets, %d rules, final=direct\n", 
		len(newRuleSets), len(newRules))
}

// applyAllTrafficMode configures routing for all traffic through VPN.
func (b *ConfigBuilderForStorage) applyAllTrafficMode(route map[string]interface{}) {
	fmt.Printf("[applyRoutingMode] Using all_traffic mode\n")

	// Remove rule_sets (not needed for all traffic mode)
	route["rule_set"] = []interface{}{}

	// Minimal rules
	newRules := []interface{}{
		map[string]interface{}{
			"action": "sniff",
		},
		map[string]interface{}{
			"domain_suffix": []string{".local", ".internal", ".corp", ".lan", ".home", ".intranet", ".private"},
			"action":        "route",
			"outbound":      "direct",
		},
		map[string]interface{}{
			"protocol": "dns",
			"action":   "hijack-dns",
		},
		map[string]interface{}{
			"ip_is_private": true,
			"action":        "route",
			"outbound":      "direct",
		},
	}

	route["rules"] = newRules
	route["final"] = "proxy"
	
	fmt.Printf("[applyRoutingMode] Applied all_traffic: minimal rules, final=proxy\n")
}

// applyExceptRussiaMode configures routing for all traffic except Russia through VPN.
// Uses built-in domain list instead of remote geosite to avoid download issues.
func (b *ConfigBuilderForStorage) applyExceptRussiaMode(route map[string]interface{}) {
	fmt.Printf("[applyRoutingMode] Using except_russia mode with built-in domain list\n")

	// No remote rule_sets needed - we use built-in domain suffixes
	route["rule_set"] = []interface{}{}

	// Russian domain suffixes for direct routing
	ruDomainSuffixes := []string{
		// Top-level domains
		".ru", ".su", ".рф",
		// Yandex
		".yandex.com", ".yandex.net", ".yandex.ru", ".ya.ru", ".yandex.by", ".yandex.kz",
		// VK / Mail.ru
		".vk.com", ".vkontakte.ru", ".vk.me", ".userapi.com",
		".mail.ru", ".mailru.com", ".mycdn.me", ".imgsmail.ru",
		".ok.ru", ".odnoklassniki.ru",
		// Banks
		".sberbank.ru", ".sber.ru", ".tinkoff.ru", ".tinkoff.com", ".vtb.ru", ".alfabank.ru",
		".raiffeisen.ru", ".gazprombank.ru", ".open.ru", ".rosbank.ru",
		// Government
		".gosuslugi.ru", ".mos.ru", ".nalog.ru", ".government.ru", ".kremlin.ru",
		".duma.gov.ru", ".cbr.ru", ".pfrf.ru", ".fss.ru",
		// News
		".ria.ru", ".rbc.ru", ".interfax.ru", ".tass.ru", ".kommersant.ru",
		".lenta.ru", ".gazeta.ru", ".kp.ru", ".mk.ru", ".iz.ru", ".rt.com",
		// E-commerce
		".ozon.ru", ".wildberries.ru", ".lamoda.ru", ".dns-shop.ru", ".mvideo.ru",
		".eldorado.ru", ".citilink.ru", ".avito.ru", ".youla.ru",
		// Retail
		".perekrestok.ru", ".magnit.ru", ".5ka.ru", ".dixy.ru", ".lenta.com",
		".sbermarket.ru", ".delivery-club.ru",
		// Transport
		".rzd.ru", ".aeroflot.ru", ".s7.ru", ".utair.ru", ".pobeda.aero",
		".pochta.ru", ".cdek.ru", ".boxberry.ru", ".dpd.ru",
		// Telecom
		".mts.ru", ".megafon.ru", ".beeline.ru", ".tele2.ru",
		".rostelecom.ru", ".rt.ru",
		// Media
		".vgtrk.ru", ".1tv.ru", ".ntv.ru", ".ren.tv", ".ctc.ru",
		".rutube.ru", ".ivi.ru", ".okko.tv", ".more.tv", ".kinopoisk.ru",
		".dzen.ru", ".zen.yandex.ru",
		// Maps / Navigation
		".2gis.ru", ".2gis.com",
		// Other popular
		".sports.ru", ".championat.com", ".sport-express.ru",
		".hh.ru", ".superjob.ru", ".rabota.ru",
		".cian.ru", ".domclick.ru", ".avito.ru",
		".pikabu.ru", ".habr.com", ".vc.ru", ".dtf.ru",
	}

	// Russian domain keywords for additional matching
	ruDomainKeywords := []string{
		"yandex", "sber", "tinkoff", "gosuslugi", "rutube",
		"vkontakte", "mailru", "rambler", "wildberries", "ozon",
	}

	newRules := []interface{}{
		// 1. Sniff for protocol detection
		map[string]interface{}{
			"action": "sniff",
		},
		// 2. Local domains direct
		map[string]interface{}{
			"domain_suffix": []string{".local", ".internal", ".corp", ".lan", ".home", ".intranet", ".private"},
			"action":        "route",
			"outbound":      "direct",
		},
		// 3. Hijack DNS
		map[string]interface{}{
			"protocol": "dns",
			"action":   "hijack-dns",
		},
		// 4. Private IPs direct
		map[string]interface{}{
			"ip_is_private": true,
			"action":        "route",
			"outbound":      "direct",
		},
		// 5. Russian domains direct
		map[string]interface{}{
			"domain_suffix": ruDomainSuffixes,
			"action":        "route",
			"outbound":      "direct",
		},
		// 6. Russian domain keywords direct
		map[string]interface{}{
			"domain_keyword": ruDomainKeywords,
			"action":         "route",
			"outbound":       "direct",
		},
	}

	route["rules"] = newRules
	route["final"] = "proxy"

	fmt.Printf("[applyRoutingMode] Applied except_russia: %d domain suffixes, %d keywords, final=proxy\n",
		len(ruDomainSuffixes), len(ruDomainKeywords))
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
			var oldConfig AppConfigLegacy
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
