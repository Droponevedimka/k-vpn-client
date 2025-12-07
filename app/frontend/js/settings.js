// Kampus VPN - Settings Module
// Application settings management

let appSettings = {
    autoStart: false,
    autoConnect: false,
    minimizeToTray: true,
    showNotifications: true,
    language: 'ru',
    theme: 'dark',
    tunMode: false,
    systemProxy: true,
    logLevel: 'info',
    dnsMode: 'auto',
    customDns: '',
    clashApiPort: 9090
};

// Load settings from backend
async function loadSettings() {
    try {
        if (!API.GetSettings) return;
        
        const settings = await API.GetSettings();
        if (settings) {
            appSettings = { ...appSettings, ...settings };
            applySettingsToUI();
        }
    } catch (error) {
        console.error('Load settings error:', error);
    }
}

// Save settings to backend
async function saveSettings() {
    try {
        // Collect settings from UI
        collectSettingsFromUI();
        
        if (!API.SaveSettings) {
            showToast(t('settingsSaveNotAvailable'), 'error');
            return;
        }
        
        await API.SaveSettings(appSettings);
        showToast(t('settingsSaved'), 'success');
        closeSettings();
        
        // Apply language if changed
        if (typeof applyLanguage === 'function') {
            applyLanguage(appSettings.language);
        }
        
    } catch (error) {
        console.error('Save settings error:', error);
        showToast(t('settingsSaveFailed'), 'error');
    }
}

// Apply settings to UI elements
function applySettingsToUI() {
    // Auto start
    setToggle('autoStartToggle', appSettings.autoStart);
    
    // Auto connect
    setToggle('autoConnectToggle', appSettings.autoConnect);
    
    // Minimize to tray
    setToggle('minimizeToTrayToggle', appSettings.minimizeToTray);
    
    // Show notifications
    setToggle('showNotificationsToggle', appSettings.showNotifications);
    
    // TUN mode
    setToggle('tunModeToggle', appSettings.tunMode);
    
    // System proxy
    setToggle('systemProxyToggle', appSettings.systemProxy);
    
    // Language
    const langSelect = document.getElementById('languageSelect');
    if (langSelect) langSelect.value = appSettings.language;
    
    // Log level
    const logLevelSelect = document.getElementById('logLevelSelect');
    if (logLevelSelect) logLevelSelect.value = appSettings.logLevel;
    
    // DNS mode
    const dnsModeSelect = document.getElementById('dnsModeSelect');
    if (dnsModeSelect) dnsModeSelect.value = appSettings.dnsMode;
    
    // Custom DNS
    const customDnsInput = document.getElementById('customDnsInput');
    if (customDnsInput) customDnsInput.value = appSettings.customDns || '';
    
    // Clash API port
    const portInput = document.getElementById('clashApiPortInput');
    if (portInput) portInput.value = appSettings.clashApiPort || 9090;
}

// Collect settings from UI elements
function collectSettingsFromUI() {
    appSettings.autoStart = getToggle('autoStartToggle');
    appSettings.autoConnect = getToggle('autoConnectToggle');
    appSettings.minimizeToTray = getToggle('minimizeToTrayToggle');
    appSettings.showNotifications = getToggle('showNotificationsToggle');
    appSettings.tunMode = getToggle('tunModeToggle');
    appSettings.systemProxy = getToggle('systemProxyToggle');
    
    const langSelect = document.getElementById('languageSelect');
    if (langSelect) appSettings.language = langSelect.value;
    
    const logLevelSelect = document.getElementById('logLevelSelect');
    if (logLevelSelect) appSettings.logLevel = logLevelSelect.value;
    
    const dnsModeSelect = document.getElementById('dnsModeSelect');
    if (dnsModeSelect) appSettings.dnsMode = dnsModeSelect.value;
    
    const customDnsInput = document.getElementById('customDnsInput');
    if (customDnsInput) appSettings.customDns = customDnsInput.value.trim();
    
    const portInput = document.getElementById('clashApiPortInput');
    if (portInput) appSettings.clashApiPort = parseInt(portInput.value) || 9090;
}

// Load settings values when opening modal
function loadSettingsValues() {
    applySettingsToUI();
}

// Toggle helpers
function setToggle(elementId, value) {
    const toggle = document.getElementById(elementId);
    if (!toggle) return;
    
    if (value) {
        toggle.classList.add('active');
    } else {
        toggle.classList.remove('active');
    }
}

function getToggle(elementId) {
    const toggle = document.getElementById(elementId);
    return toggle ? toggle.classList.contains('active') : false;
}

function toggleSwitch(elementId) {
    const toggle = document.getElementById(elementId);
    if (!toggle) return;
    
    toggle.classList.toggle('active');
}

// Open logs folder
async function openLogsFolder() {
    try {
        if (!API.GetLogPath) return;
        
        const logPath = await API.GetLogPath();
        if (logPath && API.OpenFolder) {
            await API.OpenFolder(logPath);
        }
    } catch (error) {
        console.error('Open logs folder error:', error);
        showToast(t('openFolderFailed'), 'error');
    }
}

// Load logs into modal
async function loadLogs() {
    const container = document.getElementById('logsContent');
    if (!container) return;
    
    try {
        container.innerHTML = `<div class="loading">${t('loadingLogs')}</div>`;
        
        let logs = '';
        if (API.GetLogBuffer) {
            logs = await API.GetLogBuffer();
        } else if (API.GetRecentLogs) {
            logs = await API.GetRecentLogs(100);
        }
        
        if (logs) {
            container.innerHTML = `<pre>${escapeHtml(logs)}</pre>`;
            container.scrollTop = container.scrollHeight;
        } else {
            container.innerHTML = `<div class="empty-state">${t('noLogs')}</div>`;
        }
    } catch (error) {
        console.error('Load logs error:', error);
        container.innerHTML = `<div class="error-state">${t('loadLogsFailed')}</div>`;
    }
}

// Append log message (for real-time updates)
function appendToLogView(message) {
    const container = document.getElementById('logsContent');
    if (!container) return;
    
    const pre = container.querySelector('pre');
    if (pre) {
        pre.textContent += '\n' + message;
        container.scrollTop = container.scrollHeight;
    }
}

// Clear logs
async function clearLogs() {
    const container = document.getElementById('logsContent');
    if (container) {
        container.innerHTML = `<pre></pre>`;
    }
}

// Export settings
function exportSettings() {
    const dataStr = JSON.stringify(appSettings, null, 2);
    const blob = new Blob([dataStr], { type: 'application/json' });
    const url = URL.createObjectURL(blob);
    
    const a = document.createElement('a');
    a.href = url;
    a.download = 'kampusvpn-settings.json';
    a.click();
    
    URL.revokeObjectURL(url);
    showToast(t('settingsExported'), 'success');
}

// Import settings
function importSettings(file) {
    const reader = new FileReader();
    reader.onload = async (e) => {
        try {
            const imported = JSON.parse(e.target.result);
            appSettings = { ...appSettings, ...imported };
            applySettingsToUI();
            await saveSettings();
            showToast(t('settingsImported'), 'success');
        } catch (error) {
            console.error('Import settings error:', error);
            showToast(t('settingsImportFailed'), 'error');
        }
    };
    reader.readAsText(file);
}
