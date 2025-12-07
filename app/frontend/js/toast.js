// Kampus VPN - Toast Notifications Module

function showToast(type, message) {
    const container = document.getElementById('toastContainer');
    const toast = document.createElement('div');
    toast.className = `toast ${type}`;
    
    const icons = { success: '✅', error: '❌', warning: '⚠️', info: 'ℹ️' };
    toast.innerHTML = `
        <span class="icon">${icons[type] || 'ℹ️'}</span>
        <span class="message">${message}</span>
    `;
    
    container.appendChild(toast);
    
    setTimeout(() => {
        toast.classList.add('hide');
        setTimeout(() => toast.remove(), 300);
    }, 3000);
}

function showNotification(title, body) {
    // Check if notifications are enabled
    if (!appConfig?.notifications) return;
    
    // Use toast as fallback
    showToast('info', body);
}
