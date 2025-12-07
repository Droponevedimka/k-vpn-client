// Kampus VPN - VPN Control Module
// Connection management and status handling

let vpnConnected = false;
let connectingInProgress = false;
let connectionStartTime = null;
let connectionTimer = null;

// Toggle VPN connection
async function toggleVPN() {
    if (connectingInProgress) return;
    
    const powerBtn = document.getElementById('powerBtn');
    
    try {
        connectingInProgress = true;
        
        if (vpnConnected) {
            // Disconnect
            updateUIConnecting(false);
            await API.Stop();
            vpnConnected = false;
            updateUIDisconnected();
            showToast(t('disconnected'), 'info');
        } else {
            // Connect
            updateUIConnecting(true);
            await API.Start();
            vpnConnected = true;
            updateUIConnected();
            showToast(t('connected'), 'success');
        }
    } catch (error) {
        console.error('VPN toggle error:', error);
        showToast(error.toString(), 'error');
        updateUIDisconnected();
        vpnConnected = false;
    } finally {
        connectingInProgress = false;
    }
}

// Update UI for connected state
function updateUIConnected() {
    vpnConnected = true;
    const powerBtn = document.getElementById('powerBtn');
    const statusText = document.getElementById('statusText');
    const connectionTime = document.getElementById('connectionTime');
    
    if (powerBtn) {
        powerBtn.classList.add('active');
        powerBtn.classList.remove('connecting');
    }
    
    if (statusText) {
        statusText.textContent = t('connected');
        statusText.style.color = '#22c55e';
    }
    
    // Start connection timer
    connectionStartTime = Date.now();
    startConnectionTimer();
    
    // Start stats polling
    startStatsPolling();
}

// Update UI for disconnected state
function updateUIDisconnected() {
    vpnConnected = false;
    const powerBtn = document.getElementById('powerBtn');
    const statusText = document.getElementById('statusText');
    const connectionTime = document.getElementById('connectionTime');
    
    if (powerBtn) {
        powerBtn.classList.remove('active', 'connecting');
    }
    
    if (statusText) {
        statusText.textContent = t('disconnected');
        statusText.style.color = '#6b7280';
    }
    
    if (connectionTime) {
        connectionTime.textContent = '00:00:00';
    }
    
    // Stop timer and polling
    stopConnectionTimer();
    stopStatsPolling();
    
    // Reset speed history
    if (typeof speedHistory !== 'undefined') {
        speedHistory.upload.fill(0);
        speedHistory.download.fill(0);
        speedHistory.lastUpload = 0;
        speedHistory.lastDownload = 0;
    }
}

// Update UI for connecting state
function updateUIConnecting(connecting) {
    const powerBtn = document.getElementById('powerBtn');
    const statusText = document.getElementById('statusText');
    
    if (powerBtn) {
        if (connecting) {
            powerBtn.classList.add('connecting');
        } else {
            powerBtn.classList.remove('connecting');
        }
    }
    
    if (statusText) {
        statusText.textContent = connecting ? t('connecting') : t('disconnecting');
        statusText.style.color = '#f59e0b';
    }
}

// Connection timer
function startConnectionTimer() {
    stopConnectionTimer();
    connectionTimer = setInterval(() => {
        if (!connectionStartTime) return;
        
        const elapsed = Date.now() - connectionStartTime;
        const hours = Math.floor(elapsed / 3600000);
        const minutes = Math.floor((elapsed % 3600000) / 60000);
        const seconds = Math.floor((elapsed % 60000) / 1000);
        
        const timeStr = [
            hours.toString().padStart(2, '0'),
            minutes.toString().padStart(2, '0'),
            seconds.toString().padStart(2, '0')
        ].join(':');
        
        const connectionTime = document.getElementById('connectionTime');
        if (connectionTime) {
            connectionTime.textContent = timeStr;
        }
    }, 1000);
}

function stopConnectionTimer() {
    if (connectionTimer) {
        clearInterval(connectionTimer);
        connectionTimer = null;
    }
    connectionStartTime = null;
}

// Stats polling
let statsInterval = null;

function startStatsPolling() {
    stopStatsPolling();
    statsInterval = setInterval(updateStats, 1000);
    updateStats(); // Immediate first update
}

function stopStatsPolling() {
    if (statsInterval) {
        clearInterval(statsInterval);
        statsInterval = null;
    }
}

async function updateStats() {
    if (!vpnConnected || !API.GetClashStats) return;
    
    try {
        const stats = await API.GetClashStats();
        if (!stats) return;
        
        // Update speed displays
        const uploadSpeed = stats.uploadSpeed || 0;
        const downloadSpeed = stats.downloadSpeed || 0;
        
        // Update speed history for chart
        if (typeof speedHistory !== 'undefined') {
            speedHistory.upload.shift();
            speedHistory.upload.push(uploadSpeed);
            speedHistory.download.shift();
            speedHistory.download.push(downloadSpeed);
            speedHistory.lastUpload = uploadSpeed;
            speedHistory.lastDownload = downloadSpeed;
        }
        
        // Update UI elements
        updateSpeedDisplay('currentUploadSpeed', uploadSpeed);
        updateSpeedDisplay('currentDownloadSpeed', downloadSpeed);
        
        // Update total traffic
        const uploadEl = document.getElementById('statCurrentUpload');
        const downloadEl = document.getElementById('statCurrentDownload');
        if (uploadEl) uploadEl.textContent = formatBytes(stats.uploadTotal || 0);
        if (downloadEl) downloadEl.textContent = formatBytes(stats.downloadTotal || 0);
        
        // Redraw charts
        if (typeof drawCharts === 'function') {
            drawCharts();
        }
        
    } catch (error) {
        console.error('Stats update error:', error);
    }
}

function updateSpeedDisplay(elementId, bytesPerSec) {
    const el = document.getElementById(elementId);
    if (!el) return;
    el.textContent = formatSpeed(bytesPerSec);
}

function formatBytes(bytes) {
    if (bytes < 1024) return bytes + ' B';
    if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB';
    if (bytes < 1024 * 1024 * 1024) return (bytes / (1024 * 1024)).toFixed(1) + ' MB';
    return (bytes / (1024 * 1024 * 1024)).toFixed(2) + ' GB';
}

// Check initial VPN status
async function checkVPNStatus() {
    try {
        if (!API.GetStatus) return;
        
        const status = await API.GetStatus();
        if (status && status.connected) {
            vpnConnected = true;
            updateUIConnected();
            
            // Restore connection time if available
            if (status.connectedSince) {
                connectionStartTime = new Date(status.connectedSince).getTime();
            }
        } else {
            updateUIDisconnected();
        }
    } catch (error) {
        console.error('Status check error:', error);
        updateUIDisconnected();
    }
}
