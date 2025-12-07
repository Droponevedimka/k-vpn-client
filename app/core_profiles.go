// Package main provides connection profile management for KampusVPN.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

// ConnectionProfile represents a VPN connection profile.
type ConnectionProfile struct {
	ID               int       `json:"id"`                // Unique profile ID
	Name             string    `json:"name"`              // Profile display name
	SubscriptionURL  string    `json:"subscription_url"`  // VPN subscription URL
	LastUpdated      time.Time `json:"last_updated"`      // Last subscription update
	ProxyCount       int       `json:"proxy_count"`       // Number of proxies from subscription
	WireGuardConfigs []string  `json:"wireguard_configs"` // WireGuard config tags for this profile
	CreatedAt        time.Time `json:"created_at"`        // Profile creation time
}

// ProfileManager manages connection profiles.
type ProfileManager struct {
	profiles   []ConnectionProfile
	configPath string
	mu         sync.RWMutex
}

// NewProfileManager creates a new profile manager.
func NewProfileManager(configPath string) *ProfileManager {
	pm := &ProfileManager{
		configPath: configPath,
		profiles:   make([]ConnectionProfile, 0),
	}
	pm.Load()
	return pm
}

// Load loads profiles from file.
func (pm *ProfileManager) Load() error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	data, err := os.ReadFile(pm.configPath)
	if err != nil {
		// Create default profile if file doesn't exist
		pm.profiles = []ConnectionProfile{
			{
				ID:        DefaultProfileID,
				Name:      DefaultProfileName,
				CreatedAt: time.Now(),
			},
		}
		return pm.saveInternal()
	}

	if err := json.Unmarshal(data, &pm.profiles); err != nil {
		// Reset to default on parse error
		pm.profiles = []ConnectionProfile{
			{
				ID:        DefaultProfileID,
				Name:      DefaultProfileName,
				CreatedAt: time.Now(),
			},
		}
		return pm.saveInternal()
	}

	// Ensure default profile exists
	if len(pm.profiles) == 0 || pm.profiles[0].ID != DefaultProfileID {
		defaultProfile := ConnectionProfile{
			ID:        DefaultProfileID,
			Name:      DefaultProfileName,
			CreatedAt: time.Now(),
		}
		pm.profiles = append([]ConnectionProfile{defaultProfile}, pm.profiles...)
	}

	return nil
}

// saveInternal saves profiles without locking (caller must hold lock).
func (pm *ProfileManager) saveInternal() error {
	data, err := json.MarshalIndent(pm.profiles, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal profiles: %w", err)
	}
	return os.WriteFile(pm.configPath, data, 0644)
}

// Save saves profiles to file.
func (pm *ProfileManager) Save() error {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	return pm.saveInternal()
}

// GetAll returns all profiles.
func (pm *ProfileManager) GetAll() []ConnectionProfile {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	result := make([]ConnectionProfile, len(pm.profiles))
	copy(result, pm.profiles)
	return result
}

// GetByID returns a profile by ID.
func (pm *ProfileManager) GetByID(id int) (*ConnectionProfile, error) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	for i := range pm.profiles {
		if pm.profiles[i].ID == id {
			profile := pm.profiles[i]
			return &profile, nil
		}
	}
	return nil, fmt.Errorf("profile with ID %d not found", id)
}

// Create creates a new profile.
func (pm *ProfileManager) Create(name string) (*ConnectionProfile, error) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if len(pm.profiles) >= MaxProfiles {
		return nil, fmt.Errorf("maximum number of profiles (%d) reached", MaxProfiles)
	}

	// Find next available ID
	maxID := 0
	for _, p := range pm.profiles {
		if p.ID > maxID {
			maxID = p.ID
		}
	}

	profile := ConnectionProfile{
		ID:        maxID + 1,
		Name:      name,
		CreatedAt: time.Now(),
	}

	pm.profiles = append(pm.profiles, profile)
	if err := pm.saveInternal(); err != nil {
		return nil, err
	}

	return &profile, nil
}

// Update updates a profile.
func (pm *ProfileManager) Update(id int, name string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	for i := range pm.profiles {
		if pm.profiles[i].ID == id {
			pm.profiles[i].Name = name
			return pm.saveInternal()
		}
	}
	return fmt.Errorf("profile with ID %d not found", id)
}

// Delete deletes a profile by ID.
func (pm *ProfileManager) Delete(id int) error {
	if id == DefaultProfileID {
		return fmt.Errorf("cannot delete the default profile")
	}

	pm.mu.Lock()
	defer pm.mu.Unlock()

	for i := range pm.profiles {
		if pm.profiles[i].ID == id {
			pm.profiles = append(pm.profiles[:i], pm.profiles[i+1:]...)
			return pm.saveInternal()
		}
	}
	return fmt.Errorf("profile with ID %d not found", id)
}

// SetSubscription sets subscription data for a profile.
func (pm *ProfileManager) SetSubscription(id int, url string, proxyCount int) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	for i := range pm.profiles {
		if pm.profiles[i].ID == id {
			pm.profiles[i].SubscriptionURL = url
			pm.profiles[i].ProxyCount = proxyCount
			pm.profiles[i].LastUpdated = time.Now()
			return pm.saveInternal()
		}
	}
	return fmt.Errorf("profile with ID %d not found", id)
}

// ClearSubscription clears subscription from a profile.
func (pm *ProfileManager) ClearSubscription(id int) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	for i := range pm.profiles {
		if pm.profiles[i].ID == id {
			pm.profiles[i].SubscriptionURL = ""
			pm.profiles[i].ProxyCount = 0
			pm.profiles[i].LastUpdated = time.Time{}
			return pm.saveInternal()
		}
	}
	return fmt.Errorf("profile with ID %d not found", id)
}

// AddWireGuard adds a WireGuard config tag to a profile.
func (pm *ProfileManager) AddWireGuard(id int, tag string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	for i := range pm.profiles {
		if pm.profiles[i].ID == id {
			// Check if already exists
			for _, t := range pm.profiles[i].WireGuardConfigs {
				if t == tag {
					return nil // Already exists
				}
			}
			pm.profiles[i].WireGuardConfigs = append(pm.profiles[i].WireGuardConfigs, tag)
			return pm.saveInternal()
		}
	}
	return fmt.Errorf("profile with ID %d not found", id)
}

// RemoveWireGuard removes a WireGuard config tag from a profile.
func (pm *ProfileManager) RemoveWireGuard(id int, tag string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	for i := range pm.profiles {
		if pm.profiles[i].ID == id {
			for j, t := range pm.profiles[i].WireGuardConfigs {
				if t == tag {
					pm.profiles[i].WireGuardConfigs = append(
						pm.profiles[i].WireGuardConfigs[:j],
						pm.profiles[i].WireGuardConfigs[j+1:]...,
					)
					return pm.saveInternal()
				}
			}
			return nil // Tag not found, not an error
		}
	}
	return fmt.Errorf("profile with ID %d not found", id)
}

// ExportProfile exports a single profile as JSON.
func (pm *ProfileManager) ExportProfile(id int) (string, error) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	for _, p := range pm.profiles {
		if p.ID == id {
			data, err := json.MarshalIndent(p, "", "  ")
			if err != nil {
				return "", err
			}
			return string(data), nil
		}
	}
	return "", fmt.Errorf("profile with ID %d not found", id)
}

// ImportProfile imports a profile from JSON.
func (pm *ProfileManager) ImportProfile(jsonData string) (*ConnectionProfile, error) {
	var profile ConnectionProfile
	if err := json.Unmarshal([]byte(jsonData), &profile); err != nil {
		return nil, fmt.Errorf("invalid profile JSON: %w", err)
	}

	pm.mu.Lock()
	defer pm.mu.Unlock()

	if len(pm.profiles) >= MaxProfiles {
		return nil, fmt.Errorf("maximum number of profiles (%d) reached", MaxProfiles)
	}

	// Find next available ID
	maxID := 0
	for _, p := range pm.profiles {
		if p.ID > maxID {
			maxID = p.ID
		}
	}

	profile.ID = maxID + 1
	profile.CreatedAt = time.Now()

	pm.profiles = append(pm.profiles, profile)
	if err := pm.saveInternal(); err != nil {
		return nil, err
	}

	return &profile, nil
}
