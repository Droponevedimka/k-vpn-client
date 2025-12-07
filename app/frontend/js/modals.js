// Kampus VPN - Modals Module
// Modal management and helpers

// Modal state
let currentModal = null;

// Open modal by ID
function openModal(modalId) {
    const modal = document.getElementById(modalId);
    if (!modal) return;
    
    closeAllModals();
    modal.style.display = 'flex';
    currentModal = modalId;
    document.body.style.overflow = 'hidden';
    
    // Add escape key listener
    document.addEventListener('keydown', handleEscapeKey);
}

// Close specific modal
function closeModal(modalId) {
    const modal = document.getElementById(modalId);
    if (!modal) return;
    
    modal.style.display = 'none';
    if (currentModal === modalId) {
        currentModal = null;
        document.body.style.overflow = '';
    }
    
    document.removeEventListener('keydown', handleEscapeKey);
}

// Close all modals
function closeAllModals() {
    const modals = document.querySelectorAll('.modal-overlay');
    modals.forEach(modal => {
        modal.style.display = 'none';
    });
    currentModal = null;
    document.body.style.overflow = '';
    document.removeEventListener('keydown', handleEscapeKey);
}

// Handle escape key
function handleEscapeKey(e) {
    if (e.key === 'Escape' && currentModal) {
        closeModal(currentModal);
    }
}

// Modal overlay click handler
function setupModalOverlayClicks() {
    document.querySelectorAll('.modal-overlay').forEach(modal => {
        modal.addEventListener('click', (e) => {
            if (e.target === modal) {
                closeModal(modal.id);
            }
        });
    });
}

// Specific modal openers
function openSettings() {
    openModal('settingsModal');
    loadSettingsValues();
}

function closeSettings() {
    closeModal('settingsModal');
}

function openAddProfile() {
    openModal('addProfileModal');
    document.getElementById('profileUrl').value = '';
    document.getElementById('profileName').value = '';
}

function closeAddProfile() {
    closeModal('addProfileModal');
}

function openLogsModal() {
    openModal('logsModal');
    loadLogs();
}

function closeLogsModal() {
    closeModal('logsModal');
}

function openImportWireGuard() {
    openModal('importWireGuardModal');
}

function closeImportWireGuard() {
    closeModal('importWireGuardModal');
}

function openUpdateModal(updateInfo) {
    const modal = document.getElementById('updateModal');
    if (!modal || !updateInfo) return;
    
    // Fill update info
    const versionEl = document.getElementById('updateVersion');
    const notesEl = document.getElementById('updateNotes');
    const downloadBtn = document.getElementById('startUpdateBtn');
    
    if (versionEl) versionEl.textContent = updateInfo.latestVersion;
    if (notesEl) notesEl.innerHTML = updateInfo.releaseNotes || t('noReleaseNotes');
    if (downloadBtn) downloadBtn.dataset.downloadUrl = updateInfo.downloadURL;
    
    // Reset progress
    const progressBar = document.getElementById('updateProgressBar');
    const progressText = document.getElementById('updateProgressText');
    if (progressBar) progressBar.style.width = '0%';
    if (progressText) progressText.textContent = '';
    
    openModal('updateModal');
}

function closeUpdateModal() {
    closeModal('updateModal');
}

// Proxy/Server modals
function openProxyList(groupName) {
    window.currentProxyGroup = groupName;
    openModal('proxyListModal');
    loadProxiesForGroup(groupName);
}

function closeProxyList() {
    closeModal('proxyListModal');
    window.currentProxyGroup = null;
}

// Confirm dialog
function showConfirmDialog(message, onConfirm, onCancel) {
    if (confirm(message)) {
        if (onConfirm) onConfirm();
    } else {
        if (onCancel) onCancel();
    }
}

// Initialize modals
function initModals() {
    setupModalOverlayClicks();
}
