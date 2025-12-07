package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

// App is the main application struct that holds all state and dependencies.
type App struct {
	ctx             context.Context
	cmd             *exec.Cmd
	isRunning       bool
	hasError        bool
	stoppedManually bool // Manual stop flag
	initialized     bool // Initialization complete flag
	windowVisible   bool // Window visibility flag for ping optimization
	mu              sync.Mutex
	configPath      string
	templatePath    string
	singboxPath     string
	logPath         string
	appConfigPath   string
	logFile         *os.File
	configBuilder   *ConfigBuilder
	appConfig       *AppConfig
	trafficStats    *TrafficStats
	profileManager  *ProfileManager
	logBuffer       []string // Log buffer for UI
	logBufferMu     sync.RWMutex
}

// NewApp creates a new App application struct.
func NewApp() *App {
	return &App{
		logBuffer:     make([]string, 0, MaxLogBufferSize),
		windowVisible: true,
	}
}

// startup is called when the app starts.
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	
	// Perform heavy initialization in goroutine to not block UI
	go func() {
		a.setupLogPath()
		a.findPaths()
		
		// Initialize app settings FIRST (to get active profile ID)
		a.initAppConfig()
		
		// Initialize profile manager
		a.initProfileManager()
		
		// Initialize config builder (uses active profile from app config)
		a.initConfigBuilder()
		
		// Initialize traffic stats
		a.initTrafficStats()
		
		a.mu.Lock()
		a.initialized = true
		a.mu.Unlock()
	}()
}

// waitForInit waits for initialization to complete (max 5 sec)
func (a *App) waitForInit() bool {
	for i := 0; i < 50; i++ {
		a.mu.Lock()
		if a.initialized {
			a.mu.Unlock()
			return true
		}
		a.mu.Unlock()
		time.Sleep(100 * time.Millisecond)
	}
	return false
}

// shutdown is called when the app is closing
func (a *App) shutdown(ctx context.Context) {
	a.Stop()
	a.closeLogFile()
	
	// Save traffic stats
	if a.trafficStats != nil {
		a.trafficStats.Save()
	}
	
	// Save settings
	if a.appConfig != nil {
		a.saveAppConfig()
	}
}

// initConfigBuilder initializes ConfigBuilder
func (a *App) initConfigBuilder() {
	if a.configPath != "" {
		basePath := filepath.Dir(a.configPath)
		a.configBuilder = NewConfigBuilder(basePath)
		a.templatePath = filepath.Join(basePath, "template.json")
		
		// Set active profile from app config (if already loaded)
		if a.appConfig != nil && a.appConfig.ActiveProfileID != 0 {
			a.configBuilder.SetActiveProfile(a.appConfig.ActiveProfileID)
		}
		
		// Auto-regenerate config.json if there's a subscription but no config
		a.regenerateConfigIfNeeded()
	}
}

// regenerateConfigIfNeeded regenerates config.json if it's missing but subscription exists
func (a *App) regenerateConfigIfNeeded() {
	if a.configBuilder == nil {
		return
	}
	
	// Check if config.json exists
	if fileExists(a.configPath) {
		return
	}
	
	// Load settings
	settings, err := a.configBuilder.LoadUserSettings()
	if err != nil || settings.SubscriptionURL == "" {
		return
	}
	
	// Regenerate config
	a.configBuilder.BuildConfig(settings.SubscriptionURL)
}

// findPaths finds paths to sing-box and config.json
func (a *App) findPaths() {
	// Get executable directory
	exePath, err := os.Executable()
	if err != nil {
		return
	}
	exePath, _ = filepath.EvalSymlinks(exePath)
	exeDir := filepath.Dir(exePath)

	// Set path to config.json next to exe (even if file doesn't exist yet)
	a.configPath = filepath.Join(exeDir, "config.json")
	
	// Copy template.json if it doesn't exist
	templatePath := filepath.Join(exeDir, "template.json")
	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		if err := copyEmbeddedTemplate(templatePath); err != nil {
			a.writeLog(fmt.Sprintf("Failed to copy template.json: %v", err))
		} else {
			a.writeLog("Created template.json from embedded resource")
		}
	}

	// Determine sing-box binary name
	singboxName := "sing-box"
	if runtime.GOOS == "windows" {
		singboxName = "sing-box.exe"
	}

	// 1. Look in bin/ folder next to exe (portable distribution)
	singboxPath := filepath.Join(exeDir, "bin", singboxName)
	if _, err := os.Stat(singboxPath); err == nil {
		a.singboxPath = singboxPath
		a.writeLog(fmt.Sprintf("Using bundled sing-box: %s", singboxPath))
		return
	}

	// 2. Look next to exe
	singboxPath = filepath.Join(exeDir, singboxName)
	if _, err := os.Stat(singboxPath); err == nil {
		a.singboxPath = singboxPath
		a.writeLog(fmt.Sprintf("Using sing-box: %s", singboxPath))
		return
	}

	// 3. Platform-specific fallbacks
	if runtime.GOOS == "windows" {
		// In Program Files
		singboxPath = "C:\\Program Files\\sing-box\\sing-box.exe"
		if _, err := os.Stat(singboxPath); err == nil {
			a.singboxPath = singboxPath
			return
		}
	} else {
		// In PATH
		if path, err := exec.LookPath("sing-box"); err == nil {
			a.singboxPath = path
			return
		}
		// In /usr/local/bin
		singboxPath = "/usr/local/bin/sing-box"
		if _, err := os.Stat(singboxPath); err == nil {
			a.singboxPath = singboxPath
		}
	}
}

// fileExists checks if a file exists
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
