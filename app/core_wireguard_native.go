// Package main provides Native WireGuard Manager for KampusVPN.
// Manages WireGuard tunnels through Native OS integration:
// - Windows: WireGuard Service (wireguard.exe /installtunnelservice) with bundled binaries
// - macOS: wireguard-go with utun interface
// - Linux: wg-quick or wireguard-tools
// This separates WireGuard from sing-box for better stability and performance.
//
// Bundled binaries are stored in app/dependencies/wireguard-windows-v{version}/
// and copied to build/bin/ during wails build.
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

// Version variables - injected via ldflags at build time
var (
	// WireGuard version - bundled with the application
	WireGuardVersion = "0.5.3"
	
	// Wintun version (TUN driver for Windows)
	WintunVersion = "0.14.1"
)

const (
	// Tunnel naming
	TunnelPrefix = "kampus-wg-"
	
	// Timeouts
	TunnelStartTimeout = 10 * time.Second
	TunnelStopTimeout  = 5 * time.Second
)

// NativeWireGuardManager manages WireGuard tunnels via native OS integration
type NativeWireGuardManager struct {
	basePath      string                  // Application base path (where exe is)
	configDir     string                  // Directory for .conf files
	wireguardPath string                  // Path to wireguard executable
	wgPath        string                  // Path to wg tool (for status)
	wintunPath    string                  // Path to wintun.dll (Windows only)
	tunnels       map[string]*TunnelState // Active tunnels
	mu            sync.RWMutex
	logger        func(string)            // Logging function
}

// TunnelState tracks the state of a WireGuard tunnel
type TunnelState struct {
	Name       string    `json:"name"`
	ConfigID   int       `json:"config_id"`
	ConfigPath string    `json:"config_path"`
	StartedAt  time.Time `json:"started_at"`
	Active     bool      `json:"active"`
	PID        int       `json:"pid,omitempty"` // For Linux/macOS processes
}

// NewNativeWireGuardManager creates a new Native WireGuard Manager
// Expects bundled binaries in the same directory as the executable
func NewNativeWireGuardManager(basePath string, logger func(string)) *NativeWireGuardManager {
	m := &NativeWireGuardManager{
		basePath:  basePath,
		configDir: filepath.Join(basePath, "wireguard"),
		tunnels:   make(map[string]*TunnelState),
		logger:    logger,
	}
	
	// Set paths to bundled binaries (in same dir as executable)
	m.setPlatformPaths()
	
	return m
}

// setPlatformPaths sets executable paths based on current OS
// Binaries are bundled in bin/ subdirectory relative to the main executable
func (m *NativeWireGuardManager) setPlatformPaths() {
	binDir := filepath.Join(m.basePath, "bin")
	
	switch runtime.GOOS {
	case "windows":
		// Bundled binaries in bin/ subdirectory
		m.wireguardPath = filepath.Join(binDir, "wireguard.exe")
		m.wgPath = filepath.Join(binDir, "wg.exe")
		m.wintunPath = filepath.Join(binDir, "wintun.dll")
	case "darwin":
		m.wireguardPath = filepath.Join(binDir, "wireguard-go")
		m.wgPath = filepath.Join(binDir, "wg")
	case "linux":
		m.wireguardPath = filepath.Join(binDir, "wg-quick")
		m.wgPath = filepath.Join(binDir, "wg")
	}
}

// Init initializes the manager, creating directories
func (m *NativeWireGuardManager) Init() error {
	// Create wireguard config directory
	if err := os.MkdirAll(m.configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}
	
	// Check if WireGuard binaries exist
	if !m.IsInstalled() {
		m.log("WireGuard binaries not found - bundled binaries missing")
	} else {
		m.log(fmt.Sprintf("WireGuard v%s ready at: %s", WireGuardVersion, m.wireguardPath))
	}
	
	return nil
}

// log writes a log message
func (m *NativeWireGuardManager) log(msg string) {
	if m.logger != nil {
		m.logger(fmt.Sprintf("[WireGuard] %s", msg))
	}
}

// IsInstalled checks if WireGuard binaries are available
func (m *NativeWireGuardManager) IsInstalled() bool {
	// Check our bundled executable
	if fileExists(m.wireguardPath) {
		return true
	}
	
	// Check system-wide installation
	return m.checkSystemWireGuard()
}

// checkSystemWireGuard checks for system-wide WireGuard installation
func (m *NativeWireGuardManager) checkSystemWireGuard() bool {
	switch runtime.GOOS {
	case "windows":
		// Check common Windows paths
		paths := []string{
			`C:\Program Files\WireGuard\wireguard.exe`,
			`C:\Program Files (x86)\WireGuard\wireguard.exe`,
		}
		for _, p := range paths {
			if fileExists(p) {
				m.wireguardPath = p
				m.wgPath = filepath.Join(filepath.Dir(p), "wg.exe")
				return true
			}
		}
		
		// Check PATH
		if path, err := exec.LookPath("wireguard.exe"); err == nil {
			m.wireguardPath = path
			m.wgPath = filepath.Join(filepath.Dir(path), "wg.exe")
			return true
		}
		
	case "darwin":
		// macOS: check Homebrew installation
		brewPaths := []string{
			"/opt/homebrew/bin/wg",
			"/usr/local/bin/wg",
		}
		for _, p := range brewPaths {
			if fileExists(p) {
				m.wgPath = p
				// wireguard-go might be in same directory
				wgGo := filepath.Join(filepath.Dir(p), "wireguard-go")
				if fileExists(wgGo) {
					m.wireguardPath = wgGo
				}
				return true
			}
		}
		
		// Check PATH
		if path, err := exec.LookPath("wg"); err == nil {
			m.wgPath = path
			return true
		}
		
	case "linux":
		// Linux: check for wg-quick
		paths := []string{
			"/usr/bin/wg-quick",
			"/usr/local/bin/wg-quick",
		}
		for _, p := range paths {
			if fileExists(p) {
				m.wireguardPath = p
				m.wgPath = strings.TrimSuffix(p, "-quick")
				return true
			}
		}
		
		// Check PATH
		if path, err := exec.LookPath("wg-quick"); err == nil {
			m.wireguardPath = path
			if wgPath, err := exec.LookPath("wg"); err == nil {
				m.wgPath = wgPath
			}
			return true
		}
	}
	
	return false
}

// GetStatus returns current WireGuard status
func (m *NativeWireGuardManager) GetStatus() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	activeTunnels := make([]map[string]interface{}, 0)
	for _, t := range m.tunnels {
		if t.Active {
			activeTunnels = append(activeTunnels, map[string]interface{}{
				"name":       t.Name,
				"config_id":  t.ConfigID,
				"started_at": t.StartedAt.Format(time.RFC3339),
			})
		}
	}
	
	return map[string]interface{}{
		"installed":      m.IsInstalled(),
		"version":        WireGuardVersion,
		"wintun_version": WintunVersion,
		"platform":       runtime.GOOS,
		"arch":           runtime.GOARCH,
		"wireguard_path": m.wireguardPath,
		"active_tunnels": activeTunnels,
		"tunnel_count":   len(activeTunnels),
	}
}

// GenerateConfFile generates a WireGuard .conf file from config
func (m *NativeWireGuardManager) GenerateConfFile(config *WireGuardConfig) string {
	var sb strings.Builder
	
	sb.WriteString("[Interface]\n")
	sb.WriteString(fmt.Sprintf("PrivateKey = %s\n", config.PrivateKey))
	
	// Address - can be multiple
	if len(config.Address) > 0 {
		sb.WriteString(fmt.Sprintf("Address = %s\n", strings.Join(config.Address, ", ")))
	}
	
	// DNS
	if config.DNS != "" {
		sb.WriteString(fmt.Sprintf("DNS = %s\n", config.DNS))
	}
	
	// MTU
	if config.MTU > 0 {
		sb.WriteString(fmt.Sprintf("MTU = %d\n", config.MTU))
	}
	
	// Peers
	for _, peer := range config.Peers {
		sb.WriteString("\n[Peer]\n")
		sb.WriteString(fmt.Sprintf("PublicKey = %s\n", peer.PublicKey))
		
		if peer.PresharedKey != "" {
			sb.WriteString(fmt.Sprintf("PresharedKey = %s\n", peer.PresharedKey))
		}
		
		// Endpoint
		if peer.Endpoint != "" && peer.Port > 0 {
			sb.WriteString(fmt.Sprintf("Endpoint = %s:%d\n", peer.Endpoint, peer.Port))
		}
		
		// AllowedIPs
		if len(peer.AllowedIPs) > 0 {
			sb.WriteString(fmt.Sprintf("AllowedIPs = %s\n", strings.Join(peer.AllowedIPs, ", ")))
		}
		
		// PersistentKeepalive
		if peer.PersistentKeepalive > 0 {
			sb.WriteString(fmt.Sprintf("PersistentKeepalive = %d\n", peer.PersistentKeepalive))
		}
	}
	
	return sb.String()
}

// WriteConfigFile writes config to .conf file and returns path
func (m *NativeWireGuardManager) WriteConfigFile(name string, config *WireGuardConfig) (string, error) {
	confPath := filepath.Join(m.configDir, name+".conf")
	content := m.GenerateConfFile(config)
	
	// Write with restricted permissions (contains private key)
	if err := os.WriteFile(confPath, []byte(content), 0600); err != nil {
		return "", fmt.Errorf("failed to write config file: %w", err)
	}
	
	m.log(fmt.Sprintf("Config written to: %s", confPath))
	return confPath, nil
}

// StartTunnel starts a WireGuard tunnel
func (m *NativeWireGuardManager) StartTunnel(configID int, config *WireGuardConfig) error {
	if !m.IsInstalled() {
		return fmt.Errorf("WireGuard is not installed")
	}
	
	m.mu.Lock()
	defer m.mu.Unlock()
	
	// Generate tunnel name
	name := fmt.Sprintf("%s%d", TunnelPrefix, configID)
	
	// Check if already running
	if state, exists := m.tunnels[name]; exists && state.Active {
		m.log(fmt.Sprintf("Tunnel %s already running", name))
		return nil
	}
	
	// Write config file
	confPath, err := m.WriteConfigFile(name, config)
	if err != nil {
		return err
	}
	
	m.log(fmt.Sprintf("Starting tunnel: %s", name))
	
	// Start tunnel using wireguard.exe /installtunnelservice
	cmd := exec.Command(m.wireguardPath, "/installtunnelservice", confPath)
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		m.log(fmt.Sprintf("Failed to start tunnel: %v, output: %s", err, string(output)))
		return fmt.Errorf("failed to start tunnel: %w", err)
	}
	
	// Track tunnel state
	m.tunnels[name] = &TunnelState{
		Name:       name,
		ConfigID:   configID,
		ConfigPath: confPath,
		StartedAt:  time.Now(),
		Active:     true,
	}
	
	m.log(fmt.Sprintf("Tunnel %s started successfully", name))
	return nil
}

// StopTunnel stops a WireGuard tunnel
func (m *NativeWireGuardManager) StopTunnel(configID int) error {
	if !m.IsInstalled() {
		return fmt.Errorf("WireGuard is not installed")
	}
	
	m.mu.Lock()
	defer m.mu.Unlock()
	
	name := fmt.Sprintf("%s%d", TunnelPrefix, configID)
	
	state, exists := m.tunnels[name]
	if !exists || !state.Active {
		m.log(fmt.Sprintf("Tunnel %s not running", name))
		return nil
	}
	
	m.log(fmt.Sprintf("Stopping tunnel: %s", name))
	
	// Stop tunnel using wireguard.exe /uninstalltunnelservice
	cmd := exec.Command(m.wireguardPath, "/uninstalltunnelservice", name)
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		m.log(fmt.Sprintf("Failed to stop tunnel: %v, output: %s", err, string(output)))
		// Continue anyway to clean up state
	}
	
	// Update state
	state.Active = false
	
	m.log(fmt.Sprintf("Tunnel %s stopped", name))
	return nil
}

// StopAllTunnels stops all managed tunnels
func (m *NativeWireGuardManager) StopAllTunnels() {
	m.mu.RLock()
	tunnelIDs := make([]int, 0)
	for _, state := range m.tunnels {
		if state.Active {
			tunnelIDs = append(tunnelIDs, state.ConfigID)
		}
	}
	m.mu.RUnlock()
	
	for _, id := range tunnelIDs {
		if err := m.StopTunnel(id); err != nil {
			m.log(fmt.Sprintf("Error stopping tunnel %d: %v", id, err))
		}
	}
}

// GetActiveTunnels returns list of active tunnels
func (m *NativeWireGuardManager) GetActiveTunnels() []TunnelState {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	var active []TunnelState
	for _, state := range m.tunnels {
		if state.Active {
			active = append(active, *state)
		}
	}
	return active
}

// IsTunnelActive checks if a specific tunnel is active
func (m *NativeWireGuardManager) IsTunnelActive(configID int) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	name := fmt.Sprintf("%s%d", TunnelPrefix, configID)
	if state, exists := m.tunnels[name]; exists {
		return state.Active
	}
	return false
}

// GetTunnelStats gets statistics for a tunnel (requires wg.exe)
func (m *NativeWireGuardManager) GetTunnelStats(configID int) (map[string]interface{}, error) {
	if !fileExists(m.wgPath) {
		return nil, fmt.Errorf("wg.exe not found")
	}
	
	name := fmt.Sprintf("%s%d", TunnelPrefix, configID)
	
	cmd := exec.Command(m.wgPath, "show", name)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get tunnel stats: %w", err)
	}
	
	// Parse wg show output
	stats := m.parseWgShowOutput(string(output))
	return stats, nil
}

// parseWgShowOutput parses the output of `wg show` command
func (m *NativeWireGuardManager) parseWgShowOutput(output string) map[string]interface{} {
	stats := make(map[string]interface{})
	
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "transfer:") {
			// Parse transfer stats
			parts := strings.Split(line, ",")
			for _, part := range parts {
				part = strings.TrimSpace(part)
				if strings.Contains(part, "received") {
					stats["received"] = strings.TrimSuffix(strings.TrimPrefix(part, "transfer: "), " received")
				} else if strings.Contains(part, "sent") {
					stats["sent"] = strings.TrimSuffix(part, " sent")
				}
			}
		} else if strings.HasPrefix(line, "latest handshake:") {
			stats["last_handshake"] = strings.TrimPrefix(line, "latest handshake: ")
		}
	}
	
	return stats
}

// CleanupConfigs removes all .conf files for stopped tunnels
func (m *NativeWireGuardManager) CleanupConfigs() error {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	files, err := os.ReadDir(m.configDir)
	if err != nil {
		return err
	}
	
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".conf") {
			name := strings.TrimSuffix(file.Name(), ".conf")
			
			// Check if tunnel is active
			if state, exists := m.tunnels[name]; !exists || !state.Active {
				confPath := filepath.Join(m.configDir, file.Name())
				if err := os.Remove(confPath); err != nil {
					m.log(fmt.Sprintf("Failed to remove config: %s", confPath))
				}
			}
		}
	}
	
	return nil
}

// GetWireGuardVersion returns the bundled WireGuard version
func (m *NativeWireGuardManager) GetWireGuardVersion() string {
	return WireGuardVersion
}

// copyFileNative copies a file from src to dst (for WireGuard native manager)
func copyFileNative(src, dst string) error {
	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()
	
	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destination.Close()
	
	_, err = io.Copy(destination, source)
	return err
}

// checksumFile calculates SHA256 checksum of a file
func checksumFile(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()
	
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	
	return hex.EncodeToString(hash.Sum(nil)), nil
}
