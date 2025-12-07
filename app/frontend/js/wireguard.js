// Kampus VPN - WireGuard Module
// WireGuard configuration management

let wireGuardConfigs = [];
let selectedWireGuardConfig = null;

// Load WireGuard configs
async function loadWireGuardConfigs() {
    try {
        if (!API.GetWireGuardConfigs) return;
        
        const result = await API.GetWireGuardConfigs();
        wireGuardConfigs = result || [];
        renderWireGuardConfigs();
        
        // Get selected config
        if (API.GetSelectedWireGuardConfig) {
            selectedWireGuardConfig = await API.GetSelectedWireGuardConfig();
            highlightSelectedWireGuardConfig();
        }
    } catch (error) {
        console.error('Load WireGuard configs error:', error);
    }
}

// Render WireGuard configs list
function renderWireGuardConfigs() {
    const container = document.getElementById('wireGuardConfigsList');
    if (!container) return;
    
    if (wireGuardConfigs.length === 0) {
        container.innerHTML = `<div class="empty-state">${t('noWireGuardConfigs')}</div>`;
        return;
    }
    
    container.innerHTML = wireGuardConfigs.map(config => `
        <div class="wireguard-item ${config.selected ? 'selected' : ''}" 
             data-name="${escapeHtml(config.name)}"
             onclick="selectWireGuardConfig('${escapeHtml(config.name)}')">
            <div class="wireguard-info">
                <svg viewBox="0 0 24 24" width="20" height="20" fill="currentColor">
                    <path d="M12 1L3 5v6c0 5.55 3.84 10.74 9 12 5.16-1.26 9-6.45 9-12V5l-9-4zm0 10.99h7c-.53 4.12-3.28 7.79-7 8.94V12H5V6.3l7-3.11v8.8z"/>
                </svg>
                <span class="wireguard-name">${escapeHtml(config.name)}</span>
            </div>
            <div class="wireguard-actions">
                ${config.selected ? `<span class="config-badge">${t('active')}</span>` : ''}
                <button class="btn-icon btn-danger" onclick="event.stopPropagation(); deleteWireGuardConfig('${escapeHtml(config.name)}')" title="${t('delete')}">
                    <svg viewBox="0 0 24 24" width="16" height="16" fill="currentColor">
                        <path d="M6 19c0 1.1.9 2 2 2h8c1.1 0 2-.9 2-2V7H6v12zM19 4h-3.5l-1-1h-5l-1 1H5v2h14V4z"/>
                    </svg>
                </button>
            </div>
        </div>
    `).join('');
}

// Import WireGuard config
async function importWireGuardConfig() {
    try {
        // Open file dialog
        if (!API.OpenFileDialog) {
            showToast(t('featureNotAvailable'), 'error');
            return;
        }
        
        const filePath = await API.OpenFileDialog();
        if (!filePath) return; // User cancelled
        
        await API.ImportWireGuardConfig(filePath);
        showToast(t('wireGuardConfigImported'), 'success');
        closeImportWireGuard();
        loadWireGuardConfigs();
    } catch (error) {
        console.error('Import WireGuard config error:', error);
        showToast(error.toString(), 'error');
    }
}

// Import WireGuard from text
async function importWireGuardFromText() {
    const textarea = document.getElementById('wireGuardConfigText');
    const nameInput = document.getElementById('wireGuardConfigName');
    
    const configText = textarea?.value?.trim();
    const configName = nameInput?.value?.trim() || 'WireGuard Config';
    
    if (!configText) {
        showToast(t('enterWireGuardConfig'), 'error');
        return;
    }
    
    try {
        // Parse and validate config
        if (!configText.includes('[Interface]')) {
            showToast(t('invalidWireGuardConfig'), 'error');
            return;
        }
        
        await API.ImportWireGuardConfig(configText, configName, true); // true = from text
        showToast(t('wireGuardConfigImported'), 'success');
        closeImportWireGuard();
        loadWireGuardConfigs();
        
        // Clear form
        if (textarea) textarea.value = '';
        if (nameInput) nameInput.value = '';
    } catch (error) {
        console.error('Import WireGuard config error:', error);
        showToast(error.toString(), 'error');
    }
}

// Select WireGuard config
async function selectWireGuardConfig(configName) {
    try {
        await API.SelectWireGuardConfig(configName);
        selectedWireGuardConfig = configName;
        showToast(t('wireGuardConfigSelected'), 'success');
        highlightSelectedWireGuardConfig();
        
        // Reconnect if already connected
        if (typeof vpnConnected !== 'undefined' && vpnConnected) {
            showToast(t('reconnecting'), 'info');
            await API.Stop();
            await API.Start();
        }
    } catch (error) {
        console.error('Select WireGuard config error:', error);
        showToast(error.toString(), 'error');
    }
}

// Delete WireGuard config
async function deleteWireGuardConfig(configName) {
    if (!confirm(t('confirmDeleteWireGuardConfig'))) return;
    
    try {
        await API.DeleteWireGuardConfig(configName);
        showToast(t('wireGuardConfigDeleted'), 'success');
        loadWireGuardConfigs();
    } catch (error) {
        console.error('Delete WireGuard config error:', error);
        showToast(error.toString(), 'error');
    }
}

// Highlight selected config
function highlightSelectedWireGuardConfig() {
    const items = document.querySelectorAll('.wireguard-item');
    items.forEach(item => {
        const name = item.dataset.name;
        if (name === selectedWireGuardConfig) {
            item.classList.add('selected');
        } else {
            item.classList.remove('selected');
        }
    });
}

// Toggle import mode (file/text)
function toggleWireGuardImportMode(mode) {
    const fileSection = document.getElementById('wireGuardFileImport');
    const textSection = document.getElementById('wireGuardTextImport');
    const fileTab = document.getElementById('wireGuardFileTab');
    const textTab = document.getElementById('wireGuardTextTab');
    
    if (mode === 'file') {
        if (fileSection) fileSection.style.display = 'block';
        if (textSection) textSection.style.display = 'none';
        if (fileTab) fileTab.classList.add('active');
        if (textTab) textTab.classList.remove('active');
    } else {
        if (fileSection) fileSection.style.display = 'none';
        if (textSection) textSection.style.display = 'block';
        if (fileTab) fileTab.classList.remove('active');
        if (textTab) textTab.classList.add('active');
    }
}
