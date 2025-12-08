# Kampus VPN Build Script
# This script builds the application and outputs to release/{version}/ folder

param(
    [switch]$Build,
    [switch]$Installer,
    [switch]$Portable,
    [switch]$All,
    [switch]$Clean,
    [string]$Version
)

$ErrorActionPreference = "Stop"
$ScriptRoot = $PSScriptRoot

# Read version from version.json
function Get-VersionInfo {
    $versionFile = Join-Path $ScriptRoot "version.json"
    if (-not (Test-Path $versionFile)) {
        Write-Host "[ERROR] version.json not found!" -ForegroundColor Red
        exit 1
    }
    return Get-Content $versionFile | ConvertFrom-Json
}

# Get latest version from release folder
function Get-LatestRelease {
    $releaseDir = Join-Path $ScriptRoot "release"
    if (-not (Test-Path $releaseDir)) {
        return $null
    }
    
    $versions = Get-ChildItem -Path $releaseDir -Directory | 
        Where-Object { $_.Name -match '^\d+\.\d+\.\d+$' } |
        Sort-Object { [version]$_.Name } -Descending
    
    if ($versions.Count -gt 0) {
        return $versions[0].Name
    }
    return $null
}

# Generate short random hash for build identification
function Get-BuildHash {
    $bytes = New-Object byte[] 4
    $rng = [System.Security.Cryptography.RandomNumberGenerator]::Create()
    $rng.GetBytes($bytes)
    return [BitConverter]::ToString($bytes).Replace("-", "").ToLower().Substring(0, 7)
}

$VersionInfo = Get-VersionInfo
$AppVersion = if ($Version) { $Version } else { $VersionInfo.version }
$SingBoxVersion = $VersionInfo.singbox.version
$WireGuardVersion = $VersionInfo.wireguard.version
$BuildTime = Get-Date -Format "yyyy-MM-dd HH:mm:ss"
$BuildHash = Get-BuildHash

# Build folder name includes hash for unique test environments
$BuildFolderName = "$AppVersion-$BuildHash"

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "   Kampus VPN Build System" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "Version:   $AppVersion" -ForegroundColor White
Write-Host "Build:     $BuildHash" -ForegroundColor Gray
Write-Host "sing-box:  $SingBoxVersion" -ForegroundColor White
Write-Host "WireGuard: $WireGuardVersion" -ForegroundColor White
Write-Host ""

# Paths
$AppDir = Join-Path $ScriptRoot "app"
$ReleaseDir = Join-Path $ScriptRoot "release"
$VersionDir = Join-Path $ReleaseDir $BuildFolderName
$DepsDir = Join-Path $ScriptRoot "dependencies"
$SingBoxDir = Join-Path $DepsDir "sing-box-v$SingBoxVersion"
$SingBoxExe = Join-Path $SingBoxDir "windows-amd64\sing-box-$SingBoxVersion-windows-amd64\sing-box.exe"
$InstallerDir = Join-Path $ScriptRoot "installer"

# Clean build
if ($Clean) {
    Write-Host "Cleaning..." -ForegroundColor Yellow
    if (Test-Path $VersionDir) {
        Remove-Item -Recurse -Force $VersionDir
    }
    Write-Host "[OK] Cleaned release/$BuildFolderName" -ForegroundColor Green
    if (-not $Build -and -not $All) {
        exit 0
    }
}

# Build application
function Build-Application {
    Write-Host ""
    Write-Host "Building application..." -ForegroundColor Yellow
    
    # Check wails
    $wails = Get-Command wails -ErrorAction SilentlyContinue
    if (-not $wails) {
        Write-Host "[ERROR] Wails not found! Install it with: go install github.com/wailsapp/wails/v2/cmd/wails@latest" -ForegroundColor Red
        exit 1
    }
    
    # Clean old builds with same version (different hashes)
    if (Test-Path $ReleaseDir) {
        $oldBuilds = Get-ChildItem -Path $ReleaseDir -Directory | Where-Object { $_.Name -match "^$AppVersion-" }
        foreach ($oldBuild in $oldBuilds) {
            Write-Host "[CLEAN] Removing old build: $($oldBuild.Name)" -ForegroundColor Yellow
            Remove-Item -Path $oldBuild.FullName -Recurse -Force
        }
        # Also remove old ZIP files with hashes
        $oldZips = Get-ChildItem -Path $ReleaseDir -File -Filter "KampusVPN-$AppVersion-*.zip"
        foreach ($oldZip in $oldZips) {
            Write-Host "[CLEAN] Removing old ZIP: $($oldZip.Name)" -ForegroundColor Yellow
            Remove-Item -Path $oldZip.FullName -Force
        }
    }
    
    # Create release directory
    if (-not (Test-Path $VersionDir)) {
        New-Item -ItemType Directory -Path $VersionDir | Out-Null
    }
    
    # Create resources directory
    $resourcesDir = Join-Path $VersionDir "resources"
    if (-not (Test-Path $resourcesDir)) {
        New-Item -ItemType Directory -Path $resourcesDir | Out-Null
    }
    
    # Update wails.json version
    $wailsJson = Join-Path $AppDir "wails.json"
    $wailsConfig = Get-Content $wailsJson | ConvertFrom-Json
    $wailsConfig.info.productVersion = $AppVersion
    $wailsConfig | ConvertTo-Json -Depth 10 | Set-Content $wailsJson
    
    # Build with ldflags (include build hash for dev identification)
    $ldflags = "-X 'main.Version=$AppVersion' -X 'main.BuildTime=$BuildTime' -X 'main.BuildHash=$BuildHash' -X 'main.SingBoxVersion=$SingBoxVersion' -X 'main.WireGuardVersion=$WireGuardVersion' -s -w -H=windowsgui"
    
    # Temp build directory
    $buildBin = Join-Path $AppDir "build\bin"
    
    Push-Location $AppDir
    try {
        Write-Host "Building version $AppVersion (hash: $BuildHash)..." -ForegroundColor Gray
        & wails build -ldflags $ldflags -clean
        
        if ($LASTEXITCODE -ne 0) {
            Write-Host "[ERROR] Build failed!" -ForegroundColor Red
            exit 1
        }
        
        # Move built exe to release folder
        $appExe = Join-Path $buildBin "KampusVPN.exe"
        if (Test-Path $appExe) {
            Move-Item $appExe $VersionDir -Force
            Write-Host "[OK] Built KampusVPN.exe" -ForegroundColor Green
        } else {
            Write-Host "[ERROR] KampusVPN.exe not found after build!" -ForegroundColor Red
            exit 1
        }
        
        # Clean build directory
        if (Test-Path $buildBin) {
            Remove-Item -Path $buildBin -Recurse -Force -ErrorAction SilentlyContinue
        }
        
    } finally {
        Pop-Location
    }
    
    # Create bin directory for sing-box
    $binDir = Join-Path $VersionDir "bin"
    if (-not (Test-Path $binDir)) {
        New-Item -ItemType Directory -Path $binDir | Out-Null
    }
    
    # Copy sing-box.exe to bin/ folder
    $singBoxDest = Join-Path $binDir "sing-box.exe"
    if (Test-Path $SingBoxExe) {
        Copy-Item $SingBoxExe $singBoxDest -Force
        Write-Host "[OK] Copied bin/sing-box.exe (v$SingBoxVersion)" -ForegroundColor Green
    } else {
        Write-Host "[WARNING] sing-box.exe not found at: $SingBoxExe" -ForegroundColor Yellow
        Write-Host "          Run download-singbox.ps1 to download it" -ForegroundColor Yellow
    }
    
    # Copy WireGuard dependencies to bin/ folder
    $WireGuardDir = Join-Path $DepsDir "wireguard-windows-v$WireGuardVersion"
    $WireGuardFiles = @("wireguard.exe", "wg.exe", "wintun.dll")
    
    foreach ($file in $WireGuardFiles) {
        $src = Join-Path $WireGuardDir $file
        $dst = Join-Path $binDir $file
        if (Test-Path $src) {
            Copy-Item $src $dst -Force
            Write-Host "[OK] Copied bin/$file" -ForegroundColor Green
        } else {
            Write-Host "[WARNING] $file not found at: $src" -ForegroundColor Yellow
        }
    }
    
    # Copy template.json
    $templateSrc = Join-Path $AppDir "config\template.json"
    if (Test-Path $templateSrc) {
        Copy-Item $templateSrc $resourcesDir -Force
        Write-Host "[OK] Copied template.json" -ForegroundColor Green
    }
    
    # Copy filters (rule-sets for routing)
    $filtersDir = Join-Path $DepsDir "filters"
    $filtersDest = Join-Path $binDir "filters"
    if (Test-Path $filtersDir) {
        if (-not (Test-Path $filtersDest)) {
            New-Item -ItemType Directory -Path $filtersDest | Out-Null
        }
        # Copy all .srs files and version.json
        Get-ChildItem -Path $filtersDir -Filter "*.srs" | ForEach-Object {
            Copy-Item $_.FullName $filtersDest -Force
        }
        $filtersVersion = Join-Path $filtersDir "version.json"
        if (Test-Path $filtersVersion) {
            Copy-Item $filtersVersion $filtersDest -Force
        }
        $filterCount = (Get-ChildItem -Path $filtersDest -Filter "*.srs").Count
        Write-Host "[OK] Copied bin/filters/ ($filterCount rule-sets)" -ForegroundColor Green
    } else {
        Write-Host "[WARNING] Filters not found at: $filtersDir" -ForegroundColor Yellow
    }
    
    # Create portable ZIP automatically after build (clean version name for distribution)
    $zipFile = Join-Path $ReleaseDir "KampusVPN-$AppVersion.zip"
    if (Test-Path $zipFile) {
        Remove-Item $zipFile -Force
    }
    Compress-Archive -Path "$VersionDir\*" -DestinationPath $zipFile -CompressionLevel Optimal
    $zipSize = (Get-Item $zipFile).Length / 1MB
    Write-Host "[OK] Created KampusVPN-$AppVersion.zip ($([math]::Round($zipSize, 2)) MB)" -ForegroundColor Green
    
    Write-Host ""
    Write-Host "[SUCCESS] Build completed: release/$BuildFolderName/" -ForegroundColor Green
    
    # Show files
    Write-Host ""
    Write-Host "Output files:" -ForegroundColor Cyan
    Get-ChildItem $VersionDir -Recurse | ForEach-Object {
        $size = if ($_.PSIsContainer) { "" } else { " ({0:N2} MB)" -f ($_.Length / 1MB) }
        $relativePath = $_.FullName.Replace($VersionDir, "").TrimStart("\")
        Write-Host "  $relativePath$size" -ForegroundColor White
    }
    Write-Host "  KampusVPN-$AppVersion.zip ($([math]::Round($zipSize, 2)) MB)" -ForegroundColor White
}

# Create portable ZIP (standalone, for manual use)
function Create-Portable {
    Write-Host ""
    Write-Host "Creating portable ZIP..." -ForegroundColor Yellow
    
    $sourceDir = $VersionDir
    if (-not (Test-Path $sourceDir)) {
        # Try to find latest version
        $latestVer = Get-LatestRelease
        if ($latestVer) {
            $sourceDir = Join-Path $ReleaseDir $latestVer
            Write-Host "Using latest release: $latestVer" -ForegroundColor Gray
        } else {
            Write-Host "[ERROR] No built version found. Run with -Build first." -ForegroundColor Red
            exit 1
        }
    }
    
    $appExe = Join-Path $sourceDir "KampusVPN.exe"
    if (-not (Test-Path $appExe)) {
        Write-Host "[ERROR] KampusVPN.exe not found in $sourceDir" -ForegroundColor Red
        exit 1
    }
    
    $zipFile = Join-Path $ReleaseDir "KampusVPN-$AppVersion.zip"
    
    if (Test-Path $zipFile) {
        Remove-Item $zipFile
    }
    
    # Create ZIP from version folder
    Compress-Archive -Path "$sourceDir\*" -DestinationPath $zipFile -CompressionLevel Optimal
    
    $fileSize = (Get-Item $zipFile).Length / 1MB
    Write-Host "[OK] Created: KampusVPN-$AppVersion.zip ($([math]::Round($fileSize, 2)) MB)" -ForegroundColor Green
}

# Create installer
function Create-Installer {
    Write-Host ""
    Write-Host "Creating installer..." -ForegroundColor Yellow
    
    $sourceDir = $VersionDir
    if (-not (Test-Path $sourceDir)) {
        $latestVer = Get-LatestRelease
        if ($latestVer) {
            $sourceDir = Join-Path $ReleaseDir $latestVer
            $AppVersion = $latestVer
            Write-Host "Using latest release: $latestVer" -ForegroundColor Gray
        } else {
            Write-Host "[ERROR] No built version found. Run with -Build first." -ForegroundColor Red
            exit 1
        }
    }
    
    # Check for NSIS
    $nsisPath = $null
    $nsisLocations = @(
        "C:\Program Files (x86)\NSIS\makensis.exe",
        "C:\Program Files\NSIS\makensis.exe"
    )
    
    foreach ($path in $nsisLocations) {
        if (Test-Path $path) {
            $nsisPath = $path
            break
        }
    }
    
    if (-not $nsisPath) {
        $nsisCmd = Get-Command makensis -ErrorAction SilentlyContinue
        if ($nsisCmd) {
            $nsisPath = $nsisCmd.Source
        }
    }
    
    if (-not $nsisPath) {
        Write-Host "[WARNING] NSIS not found. Skipping installer creation." -ForegroundColor Yellow
        Write-Host "          Install NSIS from: https://nsis.sourceforge.io/Download" -ForegroundColor Yellow
        Write-Host "          Or run: winget install NSIS.NSIS" -ForegroundColor Yellow
        return
    }
    
    Write-Host "[OK] NSIS found: $nsisPath" -ForegroundColor Green
    
    # Generate NSIS script with current version and paths
    $nsiScript = @"
; Kampus VPN NSIS Installer (Auto-generated)
!define PRODUCT_VERSION "$AppVersion"
!define SOURCE_DIR "$sourceDir"
!define OUTPUT_DIR "$ReleaseDir"

!include "MUI2.nsh"

Name "Kampus VPN `${PRODUCT_VERSION}"
OutFile "`${OUTPUT_DIR}\KampusVPN-`${PRODUCT_VERSION}-setup.exe"
InstallDir "`$PROGRAMFILES64\KampusVPN"
RequestExecutionLevel admin

!define MUI_ICON "$InstallerDir\assets\icon.ico"
!define MUI_UNICON "$InstallerDir\assets\icon.ico"

!insertmacro MUI_PAGE_WELCOME
!insertmacro MUI_PAGE_LICENSE "$InstallerDir\assets\license.txt"
!insertmacro MUI_PAGE_DIRECTORY
!insertmacro MUI_PAGE_INSTFILES
!define MUI_FINISHPAGE_RUN "`$INSTDIR\KampusVPN.exe"
!insertmacro MUI_PAGE_FINISH

!insertmacro MUI_UNPAGE_CONFIRM
!insertmacro MUI_UNPAGE_INSTFILES

!insertmacro MUI_LANGUAGE "English"
!insertmacro MUI_LANGUAGE "Russian"

Section "Install"
    SetOutPath "`$INSTDIR"
    
    File "`${SOURCE_DIR}\KampusVPN.exe"
    
    CreateDirectory "`$INSTDIR\bin"
    SetOutPath "`$INSTDIR\bin"
    File "`${SOURCE_DIR}\bin\sing-box.exe"
    
    CreateDirectory "`$INSTDIR\resources"
    SetOutPath "`$INSTDIR\resources"
    File /nonfatal "`${SOURCE_DIR}\resources\template.json"
    
    SetOutPath "`$INSTDIR"
    
    CreateDirectory "`$SMPROGRAMS\Kampus VPN"
    CreateShortCut "`$SMPROGRAMS\Kampus VPN\Kampus VPN.lnk" "`$INSTDIR\KampusVPN.exe"
    CreateShortCut "`$SMPROGRAMS\Kampus VPN\Uninstall.lnk" "`$INSTDIR\uninst.exe"
    CreateShortCut "`$DESKTOP\Kampus VPN.lnk" "`$INSTDIR\KampusVPN.exe"
    
    WriteUninstaller "`$INSTDIR\uninst.exe"
    
    WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\KampusVPN" "DisplayName" "Kampus VPN"
    WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\KampusVPN" "UninstallString" "`$INSTDIR\uninst.exe"
    WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\KampusVPN" "DisplayVersion" "`${PRODUCT_VERSION}"
    WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\KampusVPN" "Publisher" "K-AMPUS"
    WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\KampusVPN" "DisplayIcon" "`$INSTDIR\KampusVPN.exe"
SectionEnd

Section "Uninstall"
    nsExec::ExecToLog 'taskkill /F /IM KampusVPN.exe'
    nsExec::ExecToLog 'taskkill /F /IM sing-box.exe'
    
    Delete "`$INSTDIR\KampusVPN.exe"
    Delete "`$INSTDIR\bin\sing-box.exe"
    RMDir "`$INSTDIR\bin"
    Delete "`$INSTDIR\uninst.exe"
    Delete "`$INSTDIR\resources\*.*"
    RMDir "`$INSTDIR\resources"
    RMDir "`$INSTDIR"
    
    Delete "`$SMPROGRAMS\Kampus VPN\*.lnk"
    RMDir "`$SMPROGRAMS\Kampus VPN"
    Delete "`$DESKTOP\Kampus VPN.lnk"
    
    DeleteRegKey HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\KampusVPN"
SectionEnd
"@
    
    $tempNsi = Join-Path $env:TEMP "KampusVPN_installer.nsi"
    $nsiScript | Out-File -FilePath $tempNsi -Encoding UTF8
    
    # Build installer
    & $nsisPath $tempNsi
    
    if ($LASTEXITCODE -eq 0) {
        $installerFile = Join-Path $ReleaseDir "KampusVPN-$AppVersion-setup.exe"
        if (Test-Path $installerFile) {
            $fileSize = (Get-Item $installerFile).Length / 1MB
            Write-Host "[OK] Created: KampusVPN-$AppVersion-setup.exe ($([math]::Round($fileSize, 2)) MB)" -ForegroundColor Green
        }
    } else {
        Write-Host "[ERROR] NSIS build failed!" -ForegroundColor Red
    }
    
    Remove-Item $tempNsi -ErrorAction SilentlyContinue
}

# Main execution
if ($All -or (-not $Build -and -not $Installer -and -not $Portable)) {
    Build-Application
    Create-Portable
    Create-Installer
} else {
    if ($Build) { Build-Application }
    if ($Portable) { Create-Portable }
    if ($Installer) { Create-Installer }
}

Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "   Done!" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Cyan
