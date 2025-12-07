// Kampus VPN - Main Application
// Entry point and initialization

// Application state
let appInitialized = false;

// Initialize application
async function initApp() {
    if (appInitialized) return;
    
    console.log('Initializing Kampus VPN...');
    
    try {
        // Wait for Wails API
        await initializeAPI();
        console.log('API initialized');
        
        // Setup event listeners
        setupEventListeners();
        
        // Initialize modals
        if (typeof initModals === 'function') {
            initModals();
        }
        
        // Load language
        await loadLanguage();
        
        // Load settings
        await loadSettings();
        
        // Check VPN status
        await checkVPNStatus();
        
        // Load profiles
        await loadProfiles();
        
        // Load servers
        await loadServers();
        
        // Load WireGuard configs
        if (typeof loadWireGuardConfigs === 'function') {
            await loadWireGuardConfigs();
        }
        
        // Initialize charts
        if (typeof drawCharts === 'function') {
            drawCharts();
        }
        
        // Schedule update check
        if (typeof scheduleUpdateCheck === 'function') {
            scheduleUpdateCheck();
        }
        
        // Get version info
        await loadVersionInfo();
        
        appInitialized = true;
        console.log('Kampus VPN initialized successfully');
        
    } catch (error) {
        console.error('App initialization error:', error);
        showToast('Initialization error: ' + error.message, 'error');
    }
}

// Load language settings
async function loadLanguage() {
    try {
        let lang = 'ru'; // Default
        
        if (API.GetLanguage) {
            lang = await API.GetLanguage() || 'ru';
        }
        
        if (typeof applyLanguage === 'function') {
            applyLanguage(lang);
        }
    } catch (error) {
        console.error('Load language error:', error);
    }
}

// Load version info
async function loadVersionInfo() {
    try {
        const versionEl = document.getElementById('appVersion');
        const singboxVersionEl = document.getElementById('singboxVersion');
        
        if (versionEl && API.GetVersion) {
            const version = await API.GetVersion();
            versionEl.textContent = version || '-';
        }
        
        if (singboxVersionEl && API.GetSingBoxVersion) {
            const singboxVersion = await API.GetSingBoxVersion();
            singboxVersionEl.textContent = singboxVersion || '-';
        }
    } catch (error) {
        console.error('Load version info error:', error);
    }
}

// Window drag functionality
function initWindowDrag() {
    const header = document.querySelector('.header');
    if (!header) return;
    
    header.addEventListener('mousedown', (e) => {
        if (e.target.closest('button')) return;
        if (typeof runtime !== 'undefined' && runtime.WindowStartDrag) {
            runtime.WindowStartDrag();
        }
    });
}

// Window controls
function minimizeWindow() {
    if (typeof runtime !== 'undefined' && runtime.WindowMinimize) {
        runtime.WindowMinimize();
    }
}

function closeWindow() {
    if (typeof runtime !== 'undefined' && runtime.Quit) {
        runtime.Quit();
    }
}

// Keyboard shortcuts
function setupKeyboardShortcuts() {
    document.addEventListener('keydown', (e) => {
        // Ctrl+Shift+C - Toggle VPN
        if (e.ctrlKey && e.shiftKey && e.key === 'C') {
            e.preventDefault();
            toggleVPN();
        }
        
        // Ctrl+, - Open settings
        if (e.ctrlKey && e.key === ',') {
            e.preventDefault();
            openSettings();
        }
        
        // Ctrl+L - Open logs
        if (e.ctrlKey && e.key === 'l') {
            e.preventDefault();
            openLogsModal();
        }
        
        // F5 - Refresh servers
        if (e.key === 'F5') {
            e.preventDefault();
            loadServers();
        }
    });
}

// Context menu prevention
function setupContextMenu() {
    document.addEventListener('contextmenu', (e) => {
        // Allow context menu only on input fields
        if (!e.target.closest('input, textarea')) {
            e.preventDefault();
        }
    });
}

// Handle visibility change (tab/window focus)
function setupVisibilityHandler() {
    document.addEventListener('visibilitychange', () => {
        if (!document.hidden && vpnConnected) {
            // Refresh stats when window becomes visible
            if (typeof updateStats === 'function') {
                updateStats();
            }
        }
    });
}

// Error boundary for uncaught errors
function setupErrorHandler() {
    window.onerror = (message, source, lineno, colno, error) => {
        console.error('Uncaught error:', { message, source, lineno, colno, error });
        return false;
    };
    
    window.onunhandledrejection = (event) => {
        console.error('Unhandled promise rejection:', event.reason);
    };
}

// DOM Ready handler
function onDOMReady(callback) {
    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', callback);
    } else {
        callback();
    }
}

// Initialize when DOM is ready
onDOMReady(() => {
    // Setup error handling first
    setupErrorHandler();
    
    // Initialize drag
    initWindowDrag();
    
    // Setup keyboard shortcuts
    setupKeyboardShortcuts();
    
    // Setup context menu
    setupContextMenu();
    
    // Setup visibility handler
    setupVisibilityHandler();
    
    // Initialize app (async)
    initApp();
});

// Export for external access if needed
window.KampusVPN = {
    toggleVPN,
    openSettings,
    openAddProfile,
    openLogsModal,
    checkForUpdates,
    refreshAllPings,
    openQrScanner
};
