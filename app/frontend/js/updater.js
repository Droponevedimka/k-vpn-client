// Kampus VPN - Updater Module
// Auto-update functionality

let updateState = {
    checking: false,
    downloading: false,
    downloadProgress: 0,
    latestVersion: null,
    currentVersion: null
};

// Check for updates
async function checkForUpdates(showNoUpdateToast = true) {
    if (updateState.checking) return;
    
    try {
        updateState.checking = true;
        showToast(t('checkingForUpdates'), 'info');
        
        // Get current version
        if (API.GetVersion) {
            updateState.currentVersion = await API.GetVersion();
        }
        
        // Check for updates
        if (!API.CheckUpdates) {
            showToast(t('updateCheckNotAvailable'), 'error');
            return;
        }
        
        const updateInfo = await API.CheckUpdates();
        
        if (updateInfo && updateInfo.success && updateInfo.hasUpdate) {
            updateState.latestVersion = updateInfo.latestVersion;
            showUpdateAvailable(updateInfo);
        } else if (showNoUpdateToast) {
            showToast(t('noUpdatesAvailable'), 'info');
        }
        
    } catch (error) {
        console.error('Check updates error:', error);
        showToast(t('updateCheckFailed'), 'error');
    } finally {
        updateState.checking = false;
    }
}

// Show update available notification
function showUpdateAvailable(updateInfo) {
    // Show notification badge
    const badge = document.getElementById('updateBadge');
    if (badge) {
        badge.style.display = 'flex';
        badge.textContent = updateInfo.latestVersion;
    }
    
    // Show toast with action
    showToast(
        t('updateAvailable', { version: updateInfo.latestVersion }),
        'info',
        10000,
        [{ text: t('viewUpdate'), action: () => openUpdateModal(updateInfo) }]
    );
}

// Start update download and installation
async function startUpdate() {
    const downloadBtn = document.getElementById('startUpdateBtn');
    if (!downloadBtn || updateState.downloading) return;
    
    const downloadUrl = downloadBtn.dataset.downloadUrl;
    if (!downloadUrl) {
        showToast(t('noDownloadUrl'), 'error');
        return;
    }
    
    try {
        updateState.downloading = true;
        downloadBtn.disabled = true;
        downloadBtn.textContent = t('downloading');
        
        // Reset progress
        updateDownloadProgress(0, t('preparingDownload'));
        
        // Start download
        await API.DownloadAndInstallUpdate(downloadUrl);
        
        // If we get here without restart, show success
        updateDownloadProgress(100, t('updateComplete'));
        showToast(t('updateInstalled'), 'success');
        
    } catch (error) {
        console.error('Update download error:', error);
        showToast(t('updateFailed') + ': ' + error.toString(), 'error');
        
        downloadBtn.disabled = false;
        downloadBtn.textContent = t('retry');
        updateDownloadProgress(0, t('downloadFailed'));
    } finally {
        updateState.downloading = false;
    }
}

// Handle update progress events from backend
function handleUpdateProgress(data) {
    if (!data) return;
    
    const progress = data.progress || 0;
    const status = data.status || '';
    const speed = data.speed || 0;
    const downloaded = data.downloaded || 0;
    const total = data.total || 0;
    
    let statusText = status;
    if (speed > 0) {
        statusText += ` (${formatSpeed(speed)})`;
    }
    if (downloaded > 0 && total > 0) {
        statusText += ` - ${formatFileSize(downloaded)} / ${formatFileSize(total)}`;
    }
    
    updateDownloadProgress(progress, statusText);
}

// Update progress bar UI
function updateDownloadProgress(percent, statusText) {
    const progressBar = document.getElementById('updateProgressBar');
    const progressText = document.getElementById('updateProgressText');
    const progressPercent = document.getElementById('updateProgressPercent');
    
    if (progressBar) {
        progressBar.style.width = percent + '%';
    }
    
    if (progressText) {
        progressText.textContent = statusText || '';
    }
    
    if (progressPercent) {
        progressPercent.textContent = Math.round(percent) + '%';
    }
    
    updateState.downloadProgress = percent;
}

// Cancel update (if possible)
function cancelUpdate() {
    // Currently not implemented - close modal instead
    closeUpdateModal();
    updateState.downloading = false;
    
    const downloadBtn = document.getElementById('startUpdateBtn');
    if (downloadBtn) {
        downloadBtn.disabled = false;
        downloadBtn.textContent = t('downloadAndInstall');
    }
}

// Get changelog/release notes
async function getChangelog(version) {
    try {
        // This would need backend implementation
        // For now, return empty
        return '';
    } catch {
        return '';
    }
}

// Check updates on startup (delayed)
function scheduleUpdateCheck() {
    // Check after 5 seconds
    setTimeout(() => {
        checkForUpdates(false); // Don't show toast if no updates
    }, 5000);
}

// Format helpers (if not available from charts.js)
if (typeof formatSpeed === 'undefined') {
    function formatSpeed(bytesPerSec) {
        if (bytesPerSec < 1024) return bytesPerSec.toFixed(0) + ' B/s';
        if (bytesPerSec < 1024 * 1024) return (bytesPerSec / 1024).toFixed(1) + ' KB/s';
        return (bytesPerSec / (1024 * 1024)).toFixed(2) + ' MB/s';
    }
}

if (typeof formatFileSize === 'undefined') {
    function formatFileSize(bytes) {
        if (bytes < 1024) return bytes + ' B';
        if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB';
        return (bytes / (1024 * 1024)).toFixed(1) + ' MB';
    }
}
