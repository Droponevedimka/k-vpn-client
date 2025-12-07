// Kampus VPN - Servers Module
// Server selection and management

let servers = [];
let serverGroups = [];
let currentServer = null;
let currentServerGroup = null;

// Load servers
async function loadServers() {
    try {
        if (!API.GetServers) return;
        
        const result = await API.GetServers();
        servers = result || [];
        renderServers();
        
        // Get current server
        if (API.GetCurrentServer) {
            currentServer = await API.GetCurrentServer();
            highlightCurrentServer();
        }
    } catch (error) {
        console.error('Load servers error:', error);
    }
}

// Load server groups
async function loadServerGroups() {
    try {
        if (!API.GetServerGroups) return;
        
        const result = await API.GetServerGroups();
        serverGroups = result || [];
        renderServerGroups();
    } catch (error) {
        console.error('Load server groups error:', error);
    }
}

// Render servers list
function renderServers() {
    const container = document.getElementById('serversList');
    if (!container) return;
    
    if (servers.length === 0) {
        container.innerHTML = `<div class="empty-state">${t('noServers')}</div>`;
        return;
    }
    
    container.innerHTML = servers.map(server => `
        <div class="server-item ${server.selected ? 'selected' : ''}" 
             data-name="${escapeHtml(server.name)}"
             onclick="selectServer('${escapeHtml(server.name)}')">
            <div class="server-info">
                <span class="server-flag">${getCountryFlag(server.country)}</span>
                <span class="server-name">${escapeHtml(server.name)}</span>
            </div>
            <div class="server-ping ${getPingClass(server.ping)}">
                ${server.ping > 0 ? server.ping + ' ms' : '-'}
            </div>
        </div>
    `).join('');
}

// Render server groups
function renderServerGroups() {
    const container = document.getElementById('serverGroupsList');
    if (!container) return;
    
    if (serverGroups.length === 0) {
        container.innerHTML = '';
        return;
    }
    
    container.innerHTML = serverGroups.map(group => `
        <div class="server-group ${group.name === currentServerGroup ? 'active' : ''}"
             onclick="selectServerGroup('${escapeHtml(group.name)}')">
            <span class="group-name">${escapeHtml(group.name)}</span>
            <span class="group-count">${group.count || 0}</span>
        </div>
    `).join('');
}

// Select server
async function selectServer(serverName) {
    try {
        await API.SetServer(serverName);
        currentServer = serverName;
        showToast(t('serverSelected', { name: serverName }), 'success');
        highlightCurrentServer();
        
        // Reconnect if already connected
        if (vpnConnected) {
            showToast(t('reconnecting'), 'info');
            await API.Stop();
            await API.Start();
        }
    } catch (error) {
        console.error('Select server error:', error);
        showToast(error.toString(), 'error');
    }
}

// Select server group
async function selectServerGroup(groupName) {
    try {
        if (API.SetServerGroup) {
            await API.SetServerGroup(groupName);
        }
        currentServerGroup = groupName;
        renderServerGroups();
        loadServers(); // Reload servers for this group
    } catch (error) {
        console.error('Select server group error:', error);
        showToast(error.toString(), 'error');
    }
}

// Highlight current server
function highlightCurrentServer() {
    const items = document.querySelectorAll('.server-item');
    items.forEach(item => {
        const name = item.dataset.name;
        if (name === currentServer) {
            item.classList.add('selected');
        } else {
            item.classList.remove('selected');
        }
    });
}

// Refresh all pings
async function refreshAllPings() {
    try {
        showToast(t('refreshingPings'), 'info');
        
        if (API.RefreshAllPings) {
            await API.RefreshAllPings();
        }
        
        await loadServers();
        showToast(t('pingsRefreshed'), 'success');
    } catch (error) {
        console.error('Refresh pings error:', error);
        showToast(error.toString(), 'error');
    }
}

// Get single server ping
async function getServerPing(serverName) {
    try {
        if (!API.GetServerPing) return -1;
        return await API.GetServerPing(serverName);
    } catch {
        return -1;
    }
}

// Helper: Get country flag emoji
function getCountryFlag(countryCode) {
    if (!countryCode || countryCode.length !== 2) return '🌐';
    
    const codePoints = countryCode
        .toUpperCase()
        .split('')
        .map(char => 127397 + char.charCodeAt());
    
    return String.fromCodePoint(...codePoints);
}

// Helper: Get ping class for styling
function getPingClass(ping) {
    if (!ping || ping < 0) return 'ping-unknown';
    if (ping < 100) return 'ping-good';
    if (ping < 200) return 'ping-medium';
    return 'ping-bad';
}

// Toggle servers panel
function toggleServersPanel() {
    const panel = document.getElementById('serversPanel');
    if (!panel) return;
    
    if (panel.classList.contains('open')) {
        panel.classList.remove('open');
    } else {
        panel.classList.add('open');
        loadServers();
    }
}

// Close servers panel
function closeServersPanel() {
    const panel = document.getElementById('serversPanel');
    if (panel) {
        panel.classList.remove('open');
    }
}
