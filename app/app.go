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
	basePath        string // Base path (exe directory)
	singboxPath     string
	logPath         string
	logFile         *os.File
	storage         *Storage                  // Unified storage for all settings
	configBuilder   *ConfigBuilderForStorage  // Config builder for storage
	trafficStats    *TrafficStats
	nativeWG        *NativeWireGuardManager   // Native WireGuard tunnel manager
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
		
		// Initialize unified storage (replaces appConfig, profileManager, configBuilder)
		a.initStorage()
		
		// Initialize Native WireGuard Manager
		a.initNativeWireGuard()
		
		// Initialize traffic stats
		a.initTrafficStats()
		
		a.mu.Lock()
		a.initialized = true
		a.mu.Unlock()
		
		// Set initial tray icon to disconnected (grey)
		UpdateTrayIcon("disconnected")
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
	// Stop sing-box
	a.Stop()
	
	// Stop all Native WireGuard tunnels
	if a.nativeWG != nil {
		a.writeLog("Stopping all Native WireGuard tunnels...")
		a.nativeWG.StopAllTunnels()
	}
	
	a.closeLogFile()
	
	// Save traffic stats
	if a.trafficStats != nil {
		a.trafficStats.Save()
	}
	
	// Storage auto-saves on every change, no need to save here
}

// initStorage initializes the unified storage
func (a *App) initStorage() {
	if a.basePath == "" {
		return
	}
	
	a.storage = NewStorage(a.basePath)
	if err := a.storage.Init(); err != nil {
		a.writeLog(fmt.Sprintf("Failed to init storage: %v", err))
		return
	}
	
	// Create config builder for storage
	a.configBuilder = NewConfigBuilderForStorage(a.storage)
	
	// Migrate from old format if needed
	if err := a.storage.MigrateFromOldFormat(a.basePath); err != nil {
		a.writeLog(fmt.Sprintf("Migration error: %v", err))
	}
	
	a.writeLog("Storage initialized: " + a.storage.GetResourcesPath())
}

// initNativeWireGuard initializes the Native WireGuard Manager
func (a *App) initNativeWireGuard() {
	if a.basePath == "" {
		return
	}
	
	// Create native WireGuard manager - uses bundled binaries
	a.nativeWG = NewNativeWireGuardManager(a.basePath, a.writeLog)
	
	if err := a.nativeWG.Init(); err != nil {
		a.writeLog(fmt.Sprintf("Failed to init Native WireGuard: %v", err))
		return
	}
	
	if a.nativeWG.IsInstalled() {
		a.writeLog(fmt.Sprintf("Native WireGuard v%s available: %s", WireGuardVersion, a.nativeWG.wireguardPath))
	} else {
		a.writeLog(fmt.Sprintf("Native WireGuard v%s - bundled binaries not found", WireGuardVersion))
	}
}
// findPaths finds paths to sing-box and base directory
func (a *App) findPaths() {
	// Get executable directory
	exePath, err := os.Executable()
	if err != nil {
		return
	}
	exePath, _ = filepath.EvalSymlinks(exePath)
	exeDir := filepath.Dir(exePath)

	// Set base path
	a.basePath = exeDir

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
