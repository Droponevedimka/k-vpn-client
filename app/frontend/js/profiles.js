// Kampus VPN - Profiles Module
// Subscription profile management

let profiles = [];
let activeProfileId = null;

// Load all profiles
async function loadProfiles() {
    try {
        if (!API.GetProfiles) return;
        
        const result = await API.GetProfiles();
        if (result && result.success && result.profiles) {
            profiles = result.profiles;
        } else {
            profiles = [];
        }
        renderProfiles();
        
        // Update active profile indicator from activeProfile field or isActive flag
        const activeId = result?.activeProfile;
        const current = profiles.find(p => p.isActive || p.id === activeId);
        if (current) {
            activeProfileId = current.id;
            updateActiveProfileDisplay();
        }
    } catch (error) {
        console.error('Load profiles error:', error);
        showToast(t('errorLoadingProfiles'), 'error');
    }
}

// Render profiles list
function renderProfiles() {
    const container = document.getElementById('profilesList');
    if (!container) return;
    
    if (profiles.length === 0) {
        container.innerHTML = `<div class="empty-state">${t('noProfiles')}</div>`;
        return;
    }
    
    container.innerHTML = profiles.map(profile => `
        <div class="profile-item ${profile.isActive ? 'active' : ''}" data-id="${profile.id}">
            <div class="profile-info">
                <div class="profile-name">${escapeHtml(profile.name)}</div>
                <div class="profile-url">${escapeHtml(truncateUrl(profile.subscription || ''))}</div>
            </div>
            <div class="profile-actions">
                ${profile.isActive ? `<span class="profile-badge">${t('active')}</span>` : ''}
                <button class="btn-icon" onclick="selectProfile('${profile.id}')" title="${t('select')}">
                    <svg viewBox="0 0 24 24" width="16" height="16" fill="currentColor">
                        <path d="M9 16.17L4.83 12l-1.42 1.41L9 19 21 7l-1.41-1.41z"/>
                    </svg>
                </button>
                <button class="btn-icon" onclick="updateProfile('${profile.id}')" title="${t('update')}">
                    <svg viewBox="0 0 24 24" width="16" height="16" fill="currentColor">
                        <path d="M17.65 6.35A7.958 7.958 0 0012 4c-4.42 0-7.99 3.58-7.99 8s3.57 8 7.99 8c3.73 0 6.84-2.55 7.73-6h-2.08A5.99 5.99 0 0112 18c-3.31 0-6-2.69-6-6s2.69-6 6-6c1.66 0 3.14.69 4.22 1.78L13 11h7V4l-2.35 2.35z"/>
                    </svg>
                </button>
                <button class="btn-icon btn-danger" onclick="deleteProfile('${profile.id}')" title="${t('delete')}">
                    <svg viewBox="0 0 24 24" width="16" height="16" fill="currentColor">
                        <path d="M6 19c0 1.1.9 2 2 2h8c1.1 0 2-.9 2-2V7H6v12zM19 4h-3.5l-1-1h-5l-1 1H5v2h14V4z"/>
                    </svg>
                </button>
            </div>
        </div>
    `).join('');
}

// Add new profile
async function addProfile() {
    const urlInput = document.getElementById('profileUrl');
    const nameInput = document.getElementById('profileName');
    
    const url = urlInput?.value?.trim();
    const name = nameInput?.value?.trim() || extractNameFromUrl(url);
    
    if (!url) {
        showToast(t('enterProfileUrl'), 'error');
        return;
    }
    
    try {
        await API.AddProfile(url, name);
        showToast(t('profileAdded'), 'success');
        closeAddProfile();
        loadProfiles();
    } catch (error) {
        console.error('Add profile error:', error);
        showToast(error.toString(), 'error');
    }
}

// Select profile
async function selectProfile(profileId) {
    try {
        await API.SelectProfile(profileId);
        activeProfileId = profileId;
        showToast(t('profileSelected'), 'success');
        loadProfiles();
    } catch (error) {
        console.error('Select profile error:', error);
        showToast(error.toString(), 'error');
    }
}

// Update profile (refresh subscription)
async function updateProfile(profileId) {
    try {
        showToast(t('updatingProfile'), 'info');
        await API.UpdateProfile(profileId);
        showToast(t('profileUpdated'), 'success');
        loadProfiles();
    } catch (error) {
        console.error('Update profile error:', error);
        showToast(error.toString(), 'error');
    }
}

// Delete profile
async function deleteProfile(profileId) {
    if (!confirm(t('confirmDeleteProfile'))) return;
    
    try {
        await API.DeleteProfile(profileId);
        showToast(t('profileDeleted'), 'success');
        loadProfiles();
    } catch (error) {
        console.error('Delete profile error:', error);
        showToast(error.toString(), 'error');
    }
}

// Check for profile updates
async function checkProfileUpdates() {
    try {
        if (!API.CheckProfileUpdates) return;
        
        const updates = await API.CheckProfileUpdates();
        if (updates && updates.length > 0) {
            showToast(t('profileUpdatesAvailable', { count: updates.length }), 'info');
        }
    } catch (error) {
        console.error('Check profile updates error:', error);
    }
}

// Helper functions
function extractNameFromUrl(url) {
    try {
        const urlObj = new URL(url);
        return urlObj.hostname;
    } catch {
        return 'Profile';
    }
}

function truncateUrl(url, maxLength = 40) {
    if (!url || url.length <= maxLength) return url;
    return url.substring(0, maxLength) + '...';
}

function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}
