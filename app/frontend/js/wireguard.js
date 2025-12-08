// Kampus VPN - WireGuard Module
// WireGuard configuration management with Native WireGuard support

let wireGuardConfigs = [];
let selectedWireGuardConfig = null;
let nativeWireGuardStatus = {
    installed: false,
    tunnels: []
};

// Load WireGuard configs and native status
async function loadWireGuardConfigs() {
    try {
        if (!API.GetWireGuardConfigs) return;
        
        const result = await API.GetWireGuardConfigs();
        wireGuardConfigs = result || [];
        
        // Load Native WireGuard status
        await loadNativeWireGuardStatus();
        
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

// Load Native WireGuard status
async function loadNativeWireGuardStatus() {
    try {
        if (!API.GetNativeWireGuardStatus) return;
        
        const status = await API.GetNativeWireGuardStatus();
        if (status && status.success) {
            nativeWireGuardStatus.installed = status.installed;
            nativeWireGuardStatus.tunnels = status.active_tunnels || [];
        }
        
        // Also get active tunnels
        if (API.GetNativeWireGuardTunnels) {
            const tunnels = await API.GetNativeWireGuardTunnels();
            if (tunnels && tunnels.success) {
                nativeWireGuardStatus.tunnels = tunnels.tunnels || [];
            }
        }
    } catch (error) {
        console.error('Load Native WireGuard status error:', error);
    }
}

// Check if a config has active tunnel
function isNativeTunnelActive(tag) {
    return nativeWireGuardStatus.tunnels.some(t => t.tag === tag && t.active);
}

// Render WireGuard configs list with native tunnel controls
function renderWireGuardConfigs() {
    const container = document.getElementById('wireGuardConfigsList');
    if (!container) return;
    
    if (wireGuardConfigs.length === 0) {
        container.innerHTML = `
            <div class="empty-state">${t('noWireGuardConfigs')}</div>
            ${!nativeWireGuardStatus.installed ? renderWireGuardInstallBanner() : ''}
        `;
        return;
    }
    
    let html = '';
    
    // Show install banner if WireGuard not installed
    if (!nativeWireGuardStatus.installed) {
        html += renderWireGuardInstallBanner();
    }
    
    // Render configs
    html += wireGuardConfigs.map(config => {
        const isActive = isNativeTunnelActive(config.tag);
        const domains = config.internal_domains || [];
        const domainsText = domains.length > 0 ? domains.join(', ') : '';
        return `
        <div class="wireguard-item ${isActive ? 'active-tunnel' : ''}" 
             data-tag="${escapeHtml(config.tag)}">
            <div class="wireguard-info">
                <svg viewBox="0 0 24 24" width="20" height="20" fill="currentColor">
                    <path d="M12 1L3 5v6c0 5.55 3.84 10.74 9 12 5.16-1.26 9-6.45 9-12V5l-9-4zm0 10.99h7c-.53 4.12-3.28 7.79-7 8.94V12H5V6.3l7-3.11v8.8z"/>
                </svg>
                <div class="wireguard-details">
                    <span class="wireguard-name">${escapeHtml(config.name || config.tag)}</span>
                    <span class="wireguard-endpoint">${escapeHtml(config.endpoint || '')}</span>
                    ${domainsText ? `<span class="wireguard-domains" title="${t('internalDomains')}: ${escapeHtml(domainsText)}">${escapeHtml(domainsText)}</span>` : ''}
                </div>
            </div>
            <div class="wireguard-actions">
                ${isActive ? `<span class="tunnel-badge active">${t('tunnelActive')}</span>` : ''}
                
                <!-- Native tunnel toggle button -->
                ${nativeWireGuardStatus.installed ? `
                    <button class="btn-icon ${isActive ? 'btn-success' : 'btn-primary'}" 
                            onclick="event.stopPropagation(); toggleNativeTunnel('${escapeHtml(config.tag)}')" 
                            title="${isActive ? t('stopTunnel') : t('startTunnel')}">
                        ${isActive ? 
                            '<svg viewBox="0 0 24 24" width="16" height="16" fill="currentColor"><path d="M6 6h12v12H6z"/></svg>' :
                            '<svg viewBox="0 0 24 24" width="16" height="16" fill="currentColor"><path d="M8 5v14l11-7z"/></svg>'
                        }
                    </button>
                ` : ''}
                
                <button class="btn-icon btn-danger" 
                        onclick="event.stopPropagation(); deleteWireGuardConfig('${escapeHtml(config.tag)}')" 
                        title="${t('delete')}">
                    <svg viewBox="0 0 24 24" width="16" height="16" fill="currentColor">
                        <path d="M6 19c0 1.1.9 2 2 2h8c1.1 0 2-.9 2-2V7H6v12zM19 4h-3.5l-1-1h-5l-1 1H5v2h14V4z"/>
                    </svg>
                </button>
            </div>
        </div>
    `}).join('');
    
    // Add "Start All" / "Stop All" buttons if there are configs
    if (wireGuardConfigs.length > 0 && nativeWireGuardStatus.installed) {
        const hasActiveTunnels = nativeWireGuardStatus.tunnels.length > 0;
        html += `
            <div class="wireguard-bulk-actions">
                <button class="btn btn-primary btn-sm" onclick="startAllNativeTunnels()">
                    <svg viewBox="0 0 24 24" width="14" height="14" fill="currentColor"><path d="M8 5v14l11-7z"/></svg>
                    ${t('startAllTunnels')}
                </button>
                ${hasActiveTunnels ? `
                    <button class="btn btn-secondary btn-sm" onclick="stopAllNativeTunnels()">
                        <svg viewBox="0 0 24 24" width="14" height="14" fill="currentColor"><path d="M6 6h12v12H6z"/></svg>
                        ${t('stopAllTunnels')}
                    </button>
                ` : ''}
            </div>
        `;
    }
    
    container.innerHTML = html;
}

// Render WireGuard install banner
function renderWireGuardInstallBanner() {
    return `
        <div class="wireguard-install-banner">
            <div class="banner-icon">
                <svg viewBox="0 0 24 24" width="32" height="32" fill="currentColor">
                    <path d="M12 1L3 5v6c0 5.55 3.84 10.74 9 12 5.16-1.26 9-6.45 9-12V5l-9-4z"/>
                </svg>
            </div>
            <div class="banner-content">
                <h4>${t('wireGuardNotInstalled')}</h4>
                <p>${t('wireGuardInstallDescription')}</p>
                <button class="btn btn-primary btn-sm" onclick="downloadWireGuard()">
                    ${t('downloadWireGuard')}
                </button>
            </div>
        </div>
    `;
}

// Toggle native WireGuard tunnel
async function toggleNativeTunnel(tag) {
    try {
        const isActive = isNativeTunnelActive(tag);
        
        if (isActive) {
            // Stop tunnel
            const result = await API.StopNativeWireGuard(tag);
            if (result && result.success) {
                showToast(t('tunnelStopped'), 'info');
            } else {
                showToast(result?.error || t('tunnelStopError'), 'error');
            }
        } else {
            // Start tunnel
            const result = await API.StartNativeWireGuard(tag);
            if (result && result.success) {
                showToast(t('tunnelStarted'), 'success');
            } else if (result?.install_required) {
                showToast(t('wireGuardNotInstalled'), 'warning');
                return;
            } else {
                showToast(result?.error || t('tunnelStartError'), 'error');
            }
        }
        
        // Reload status
        await loadWireGuardConfigs();
    } catch (error) {
        console.error('Toggle native tunnel error:', error);
        showToast(error.toString(), 'error');
    }
}

// Start all native tunnels
async function startAllNativeTunnels() {
    try {
        const result = await API.StartAllNativeWireGuard();
        if (result && result.success) {
            showToast(t('allTunnelsStarted').replace('{count}', result.started), 'success');
        } else if (result?.install_required) {
            showToast(t('wireGuardNotInstalled'), 'warning');
        } else {
            showToast(result?.errors?.join(', ') || t('tunnelStartError'), 'error');
        }
        await loadWireGuardConfigs();
    } catch (error) {
        console.error('Start all tunnels error:', error);
        showToast(error.toString(), 'error');
    }
}

// Stop all native tunnels
async function stopAllNativeTunnels() {
    try {
        const result = await API.StopAllNativeWireGuard();
        if (result && result.success) {
            showToast(t('allTunnelsStopped'), 'info');
        } else {
            showToast(result?.error || t('tunnelStopError'), 'error');
        }
        await loadWireGuardConfigs();
    } catch (error) {
        console.error('Stop all tunnels error:', error);
        showToast(error.toString(), 'error');
    }
}

// Download WireGuard
async function downloadWireGuard() {
    try {
        showToast(t('downloadingWireGuard'), 'info');
        
        const result = await API.DownloadWireGuard();
        
        if (result && result.success) {
            showToast(t('wireGuardInstalled'), 'success');
            await loadWireGuardConfigs();
        } else if (result?.install_manual) {
            // Need manual installation
            showToast(t('wireGuardInstallManual'), 'warning');
            if (result.installer_path && API.OpenFolder) {
                // Open folder with installer
                const folder = result.installer_path.substring(0, result.installer_path.lastIndexOf('\\'));
                await API.OpenFolder(folder);
            }
        } else {
            showToast(result?.error || t('wireGuardDownloadError'), 'error');
        }
    } catch (error) {
        console.error('Download WireGuard error:', error);
        showToast(error.toString(), 'error');
    }
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

// Select WireGuard config (legacy - for sing-box integration)
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
async function deleteWireGuardConfig(tag) {
    if (!confirm(t('confirmDeleteWireGuardConfig'))) return;
    
    try {
        // Stop tunnel first if active
        if (isNativeTunnelActive(tag)) {
            await API.StopNativeWireGuard(tag);
        }
        
        await API.DeleteWireGuardConfig(tag);
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
