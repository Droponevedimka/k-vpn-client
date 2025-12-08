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

// ============================================================================
// Import/Export Profiles
// ============================================================================

// Export all profiles to JSON file
async function exportProfiles() {
    try {
        if (!API.ExportProfilesToFile) {
            showToast(t('exportNotAvailable'), 'error');
            return;
        }
        
        const result = await API.ExportProfilesToFile();
        
        if (result.success) {
            showToast(t('profilesExported', { count: result.profiles_count }), 'success');
        } else {
            if (result.error !== 'Отменено пользователем') {
                showToast(result.error, 'error');
            }
        }
    } catch (error) {
        console.error('Export profiles error:', error);
        showToast(t('exportFailed'), 'error');
    }
}

// Import profiles from JSON file
async function importProfiles() {
    try {
        if (!API.ImportProfilesFromFile) {
            showToast(t('importNotAvailable'), 'error');
            return;
        }
        
        const result = await API.ImportProfilesFromFile();
        
        if (!result.success) {
            if (result.error !== 'Отменено пользователем') {
                showToast(result.error, 'error');
            }
            return;
        }
        
        // Show confirmation dialog
        if (result.needs_confirmation) {
            showImportConfirmDialog(result);
        }
    } catch (error) {
        console.error('Import profiles error:', error);
        showToast(t('importFailed'), 'error');
    }
}

// Show import confirmation dialog
function showImportConfirmDialog(validationResult) {
    const profileNames = validationResult.profile_names || [];
    const profilesList = profileNames.map(name => `• ${name}`).join('\n');
    
    const message = `${t('importConfirmMessage')}

${t('profilesFound')}: ${validationResult.profiles_count}
${t('wireGuardConfigs')}: ${validationResult.wireguard_count}
${t('hasTemplate')}: ${validationResult.has_template ? t('yes') : t('no')}

${t('profiles')}:
${profilesList}

${t('importWarning')}`;
    
    if (confirm(message)) {
        confirmImport(validationResult.file_data);
    }
}

// Confirm and execute import
async function confirmImport(jsonData) {
    try {
        if (!API.ConfirmImportProfiles) {
            showToast(t('importNotAvailable'), 'error');
            return;
        }
        
        const result = await API.ConfirmImportProfiles(jsonData);
        
        if (result.success) {
            showToast(t('profilesImported', { count: result.profiles_count }), 'success');
            // Reload profiles UI
            if (typeof loadProfiles === 'function') {
                await loadProfiles();
            }
            // Reload WireGuard configs
            if (typeof loadWireGuardConfigs === 'function') {
                await loadWireGuardConfigs();
            }
        } else {
            showToast(result.error, 'error');
        }
    } catch (error) {
        console.error('Confirm import error:', error);
        showToast(t('importFailed'), 'error');
    }
}

// Legacy functions for backward compatibility
function exportSettings() {
    exportProfiles();
}

function importSettings(file) {
    importProfiles();
}
