// Kampus VPN - QR Scanner Module
// QR code scanning for subscription URLs

let qrScanner = {
    stream: null,
    video: null,
    canvas: null,
    context: null,
    scanning: false,
    animationFrame: null
};

// Open QR Scanner modal
function openQrScanner() {
    openModal('qrScannerModal');
    startQrScanning();
}

// Close QR Scanner modal
function closeQrScanner() {
    stopQrScanning();
    closeModal('qrScannerModal');
}

// Start QR code scanning
async function startQrScanning() {
    qrScanner.video = document.getElementById('qrVideo');
    qrScanner.canvas = document.getElementById('qrCanvas');
    
    if (!qrScanner.video || !qrScanner.canvas) {
        showToast(t('qrScannerNotAvailable'), 'error');
        return;
    }
    
    qrScanner.context = qrScanner.canvas.getContext('2d');
    
    try {
        // Request camera access
        qrScanner.stream = await navigator.mediaDevices.getUserMedia({
            video: {
                facingMode: 'environment',
                width: { ideal: 640 },
                height: { ideal: 480 }
            }
        });
        
        qrScanner.video.srcObject = qrScanner.stream;
        qrScanner.video.play();
        
        qrScanner.scanning = true;
        scanQrFrame();
        
    } catch (error) {
        console.error('Camera access error:', error);
        
        if (error.name === 'NotAllowedError') {
            showToast(t('cameraPermissionDenied'), 'error');
        } else if (error.name === 'NotFoundError') {
            showToast(t('cameraNotFound'), 'error');
        } else {
            showToast(t('cameraError') + ': ' + error.message, 'error');
        }
        
        closeQrScanner();
    }
}

// Stop QR code scanning
function stopQrScanning() {
    qrScanner.scanning = false;
    
    if (qrScanner.animationFrame) {
        cancelAnimationFrame(qrScanner.animationFrame);
        qrScanner.animationFrame = null;
    }
    
    if (qrScanner.stream) {
        qrScanner.stream.getTracks().forEach(track => track.stop());
        qrScanner.stream = null;
    }
    
    if (qrScanner.video) {
        qrScanner.video.srcObject = null;
    }
}

// Scan single frame for QR code
function scanQrFrame() {
    if (!qrScanner.scanning) return;
    
    if (qrScanner.video.readyState === qrScanner.video.HAVE_ENOUGH_DATA) {
        qrScanner.canvas.width = qrScanner.video.videoWidth;
        qrScanner.canvas.height = qrScanner.video.videoHeight;
        
        qrScanner.context.drawImage(
            qrScanner.video,
            0, 0,
            qrScanner.canvas.width,
            qrScanner.canvas.height
        );
        
        const imageData = qrScanner.context.getImageData(
            0, 0,
            qrScanner.canvas.width,
            qrScanner.canvas.height
        );
        
        // Use jsQR library to decode
        if (typeof jsQR !== 'undefined') {
            const code = jsQR(imageData.data, imageData.width, imageData.height, {
                inversionAttempts: 'dontInvert'
            });
            
            if (code) {
                handleQrResult(code.data);
                return;
            }
        }
    }
    
    qrScanner.animationFrame = requestAnimationFrame(scanQrFrame);
}

// Handle QR code result
function handleQrResult(data) {
    console.log('QR code scanned:', data);
    
    stopQrScanning();
    closeModal('qrScannerModal');
    
    // Check if it's a valid subscription URL
    if (isValidSubscriptionUrl(data)) {
        // Fill in add profile form
        openAddProfile();
        
        const urlInput = document.getElementById('profileUrl');
        if (urlInput) {
            urlInput.value = data;
        }
        
        // Try to extract name from URL
        const nameInput = document.getElementById('profileName');
        if (nameInput) {
            nameInput.value = extractNameFromUrl(data);
        }
        
        showToast(t('qrCodeScanned'), 'success');
    } else {
        // Check if it's a WireGuard config
        if (data.includes('[Interface]') || data.includes('[Peer]')) {
            openImportWireGuard();
            toggleWireGuardImportMode('text');
            
            const textarea = document.getElementById('wireGuardConfigText');
            if (textarea) {
                textarea.value = data;
            }
            
            showToast(t('wireGuardConfigScanned'), 'success');
        } else {
            // Unknown format
            showToast(t('unknownQrFormat'), 'warning');
            console.log('Unknown QR data:', data);
        }
    }
}

// Validate subscription URL
function isValidSubscriptionUrl(url) {
    try {
        const parsed = new URL(url);
        // Accept http/https URLs
        return ['http:', 'https:'].includes(parsed.protocol);
    } catch {
        // Check for base64 encoded configs (vmess://, vless://, etc.)
        const protocols = ['vmess://', 'vless://', 'trojan://', 'ss://', 'ssr://', 'hysteria://', 'hysteria2://'];
        return protocols.some(p => url.toLowerCase().startsWith(p));
    }
}

// Manual QR code input (for pasting)
function handlePastedQrData() {
    const input = document.getElementById('qrManualInput');
    if (!input) return;
    
    const data = input.value.trim();
    if (data) {
        handleQrResult(data);
        input.value = '';
    }
}
