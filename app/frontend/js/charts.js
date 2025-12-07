// Kampus VPN - Charts Module

// Speed history for chart (last 30 data points)
const speedHistory = {
    upload: new Array(30).fill(0),
    download: new Array(30).fill(0),
    lastUpload: 0,
    lastDownload: 0
};

function drawCharts() {
    drawSpeedChart();
    drawTrafficChart();
}

function drawSpeedChart() {
    const canvas = document.getElementById('speedChart');
    if (!canvas) return;
    
    const ctx = canvas.getContext('2d');
    const width = canvas.width;
    const height = canvas.height;
    
    // Clear
    ctx.clearRect(0, 0, width, height);
    
    // Find max value for scaling
    const maxSpeed = Math.max(...speedHistory.upload, ...speedHistory.download, 1024); // min 1KB
    
    // Draw grid
    ctx.strokeStyle = 'rgba(255,255,255,0.1)';
    ctx.lineWidth = 1;
    for (let i = 0; i <= 4; i++) {
        const y = (height / 4) * i;
        ctx.beginPath();
        ctx.moveTo(0, y);
        ctx.lineTo(width, y);
        ctx.stroke();
    }
    
    // Draw download line (blue)
    ctx.strokeStyle = '#3b82f6';
    ctx.lineWidth = 2;
    ctx.beginPath();
    speedHistory.download.forEach((val, i) => {
        const x = (width / (speedHistory.download.length - 1)) * i;
        const y = height - (val / maxSpeed) * (height - 10);
        if (i === 0) ctx.moveTo(x, y);
        else ctx.lineTo(x, y);
    });
    ctx.stroke();
    
    // Draw upload line (green)
    ctx.strokeStyle = '#22c55e';
    ctx.beginPath();
    speedHistory.upload.forEach((val, i) => {
        const x = (width / (speedHistory.upload.length - 1)) * i;
        const y = height - (val / maxSpeed) * (height - 10);
        if (i === 0) ctx.moveTo(x, y);
        else ctx.lineTo(x, y);
    });
    ctx.stroke();
    
    // Legend
    ctx.font = '10px sans-serif';
    ctx.fillStyle = '#22c55e';
    ctx.fillText('↑ ' + formatSpeed(speedHistory.upload[speedHistory.upload.length - 1]), 5, 12);
    ctx.fillStyle = '#3b82f6';
    ctx.fillText('↓ ' + formatSpeed(speedHistory.download[speedHistory.download.length - 1]), 80, 12);
}

function drawTrafficChart() {
    const canvas = document.getElementById('trafficChart');
    if (!canvas) return;
    
    const ctx = canvas.getContext('2d');
    const width = canvas.width;
    const height = canvas.height;
    
    // Get current stats
    const uploadEl = document.getElementById('statCurrentUpload');
    const downloadEl = document.getElementById('statCurrentDownload');
    const uploadText = uploadEl ? uploadEl.textContent : '0 B';
    const downloadText = downloadEl ? downloadEl.textContent : '0 B';
    
    // Parse values (approximate)
    const upload = parseTrafficValue(uploadText);
    const download = parseTrafficValue(downloadText);
    const total = upload + download || 1;
    
    // Clear
    ctx.clearRect(0, 0, width, height);
    
    // Draw bar chart
    const barHeight = 25;
    const gap = 10;
    
    // Upload bar
    ctx.fillStyle = 'rgba(34,197,94,0.2)';
    ctx.fillRect(0, 0, width, barHeight);
    ctx.fillStyle = '#22c55e';
    ctx.fillRect(0, 0, (upload / total) * width, barHeight);
    ctx.fillStyle = '#fff';
    ctx.font = '11px sans-serif';
    ctx.fillText('↑ ' + uploadText, 8, barHeight - 8);
    
    // Download bar
    ctx.fillStyle = 'rgba(59,130,246,0.2)';
    ctx.fillRect(0, barHeight + gap, width, barHeight);
    ctx.fillStyle = '#3b82f6';
    ctx.fillRect(0, barHeight + gap, (download / total) * width, barHeight);
    ctx.fillStyle = '#fff';
    ctx.fillText('↓ ' + downloadText, 8, barHeight + gap + barHeight - 8);
}

function formatSpeed(bytesPerSec) {
    if (bytesPerSec < 1024) return bytesPerSec.toFixed(0) + ' B/s';
    if (bytesPerSec < 1024 * 1024) return (bytesPerSec / 1024).toFixed(1) + ' KB/s';
    return (bytesPerSec / (1024 * 1024)).toFixed(2) + ' MB/s';
}

function parseTrafficValue(str) {
    const num = parseFloat(str) || 0;
    if (str.includes('GB')) return num * 1024 * 1024 * 1024;
    if (str.includes('MB')) return num * 1024 * 1024;
    if (str.includes('KB')) return num * 1024;
    return num;
}

function formatFileSize(bytes) {
    if (bytes < 1024) return bytes + ' B';
    if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB';
    return (bytes / (1024 * 1024)).toFixed(1) + ' MB';
}
