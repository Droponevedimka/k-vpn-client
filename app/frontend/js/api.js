// Kampus VPN - API Module
// Wails API wrapper and backend communication

// API object will be populated when Wails is ready
let API = {};

// Initialize API bindings
function initializeAPI() {
    return new Promise((resolve) => {
        if (typeof go !== 'undefined' && go.main && go.main.App) {
            API = {
                // VPN Control
                Start: go.main.App.Start,
                Stop: go.main.App.Stop,
                ToggleVPN: go.main.App.ToggleVPN,
                GetStatus: go.main.App.GetStatus,
                GetClashStats: go.main.App.GetClashStats,
                GetConnectionInfo: go.main.App.GetConnectionInfo,
                
                // Profiles
                GetProfiles: go.main.App.GetProfiles,
                AddProfile: go.main.App.AddProfile,
                DeleteProfile: go.main.App.DeleteProfile,
                SelectProfile: go.main.App.SelectProfile,
                SetActiveProfile: go.main.App.SetActiveProfile,
                UpdateProfile: go.main.App.UpdateProfile,
                CheckProfileUpdates: go.main.App.CheckProfileUpdates,
                
                // Servers
                GetServers: go.main.App.GetServers,
                SetServer: go.main.App.SetServer,
                GetCurrentServer: go.main.App.GetCurrentServer,
                GetServerGroups: go.main.App.GetServerGroups,
                SetServerGroup: go.main.App.SetServerGroup,
                GetServerPing: go.main.App.GetServerPing,
                RefreshAllPings: go.main.App.RefreshAllPings,
                
                // WireGuard
                GetWireGuardConfigs: go.main.App.GetWireGuardConfigs,
                ImportWireGuardConfig: go.main.App.ImportWireGuardConfig,
                DeleteWireGuardConfig: go.main.App.DeleteWireGuardConfig,
                SelectWireGuardConfig: go.main.App.SelectWireGuardConfig,
                GetSelectedWireGuardConfig: go.main.App.GetSelectedWireGuardConfig,
                OpenFileDialog: go.main.App.OpenFileDialog,
                
                // Logs
                GetLogPath: go.main.App.GetLogPath,
                GetRecentLogs: go.main.App.GetRecentLogs,
                GetLogBuffer: go.main.App.GetLogBuffer,
                
                // Updates
                CheckUpdates: go.main.App.CheckUpdates,
                DownloadAndInstallUpdate: go.main.App.DownloadAndInstallUpdate,
                
                // Settings
                GetSettings: go.main.App.GetSettings,
                SaveSettings: go.main.App.SaveSettings,
                GetLanguage: go.main.App.GetLanguage,
                SetLanguage: go.main.App.SetLanguage,
                
                // System
                GetVersion: go.main.App.GetVersion,
                GetSingBoxVersion: go.main.App.GetSingBoxVersion,
                OpenFolder: go.main.App.OpenFolder,
                GetSystemProxyStatus: go.main.App.GetSystemProxyStatus,
                GetTUNStatus: go.main.App.GetTUNStatus,
                Restart: go.main.App.Restart,
                
                // Proxies (Clash API)
                GetProxies: go.main.App.GetProxies,
                SelectProxy: go.main.App.SelectProxy,
            };
            resolve(true);
        } else {
            // Wait for Wails to load
            setTimeout(() => initializeAPI().then(resolve), 100);
        }
    });
}

// Helper for safe API calls
async function safeAPICall(method, ...args) {
    try {
        if (!API[method]) {
            console.error(`API method ${method} not available`);
            return null;
        }
        return await API[method](...args);
    } catch (error) {
        console.error(`API call ${method} failed:`, error);
        throw error;
    }
}

// Event system for Wails events
const eventHandlers = {};

function onEvent(eventName, handler) {
    if (!eventHandlers[eventName]) {
        eventHandlers[eventName] = [];
    }
    eventHandlers[eventName].push(handler);
    
    // Register with Wails runtime
    if (typeof runtime !== 'undefined') {
        runtime.EventsOn(eventName, handler);
    }
}

function offEvent(eventName, handler) {
    if (eventHandlers[eventName]) {
        eventHandlers[eventName] = eventHandlers[eventName].filter(h => h !== handler);
    }
}

// Register Wails event listeners
function setupEventListeners() {
    if (typeof runtime === 'undefined') {
        setTimeout(setupEventListeners, 100);
        return;
    }
    
    // VPN status events
    runtime.EventsOn("vpn-connected", () => {
        updateUIConnected();
    });
    
    runtime.EventsOn("vpn-disconnected", () => {
        updateUIDisconnected();
    });
    
    runtime.EventsOn("vpn-error", (error) => {
        showToast(error, 'error');
        updateUIDisconnected();
    });
    
    // Update progress events
    runtime.EventsOn("update-progress", (data) => {
        handleUpdateProgress(data);
    });
    
    // Profile events
    runtime.EventsOn("profile-updated", () => {
        loadProfiles();
    });
    
    // Log events
    runtime.EventsOn("log", (message) => {
        appendToLogView(message);
    });
}
