# KampusVPN - Service Connectivity Test Script v2
# Tests connectivity to popular services in two phases:
# Phase 1: Direct access (should work WITHOUT VPN)
# Phase 2: Blocked services (should work only WITH VPN)

param(
    [switch]$Verbose,
    [switch]$Json,
    [int]$Timeout = 10,
    [switch]$Phase1Only,
    [switch]$Phase2Only
)

$ErrorActionPreference = "SilentlyContinue"
[Console]::OutputEncoding = [System.Text.Encoding]::UTF8

# ============================================
# PHASE 1: Direct Access Services
# These should work WITHOUT VPN (direct connection)
# ============================================
$directServices = @(
    # Russian services - must work directly
    @{ Name = "Yandex"; URL = "https://ya.ru"; Category = "Russian" }
    @{ Name = "Yandex Mail"; URL = "https://mail.yandex.ru"; Category = "Russian" }
    @{ Name = "VK"; URL = "https://vk.com"; Category = "Russian" }
    @{ Name = "Mail.ru"; URL = "https://mail.ru"; Category = "Russian" }
    @{ Name = "Sberbank"; URL = "https://www.sberbank.ru"; Category = "Russian" }
    @{ Name = "Tinkoff"; URL = "https://www.tinkoff.ru"; Category = "Russian" }
    @{ Name = "Gosuslugi"; URL = "https://www.gosuslugi.ru"; Category = "Russian" }
    @{ Name = "Ozon"; URL = "https://www.ozon.ru"; Category = "Russian" }
    @{ Name = "Wildberries"; URL = "https://www.wildberries.ru"; Category = "Russian" }
    @{ Name = "Avito"; URL = "https://www.avito.ru"; Category = "Russian" }
    @{ Name = "2GIS"; URL = "https://2gis.ru"; Category = "Russian" }
    @{ Name = "Habr"; URL = "https://habr.com"; Category = "Russian" }
    
    # International NOT blocked - must work directly
    @{ Name = "Google"; URL = "https://www.google.com"; Category = "International" }
    @{ Name = "Google Drive"; URL = "https://drive.google.com"; Category = "International" }
    @{ Name = "Gmail"; URL = "https://mail.google.com"; Category = "International" }
    @{ Name = "YouTube"; URL = "https://www.youtube.com"; Category = "International" }
    @{ Name = "GitHub"; URL = "https://github.com"; Category = "International" }
    @{ Name = "GitLab"; URL = "https://gitlab.com"; Category = "International" }
    @{ Name = "Stack Overflow"; URL = "https://stackoverflow.com"; Category = "International" }
    @{ Name = "Wikipedia"; URL = "https://www.wikipedia.org"; Category = "International" }
    @{ Name = "Reddit"; URL = "https://www.reddit.com"; Category = "International" }
    @{ Name = "Amazon"; URL = "https://www.amazon.com"; Category = "International" }
    @{ Name = "Microsoft"; URL = "https://www.microsoft.com"; Category = "International" }
    @{ Name = "Apple"; URL = "https://www.apple.com"; Category = "International" }
    @{ Name = "Telegram Web"; URL = "https://web.telegram.org"; Category = "International" }
    @{ Name = "WhatsApp Web"; URL = "https://web.whatsapp.com"; Category = "International" }
)

# ============================================
# PHASE 2: Blocked/Restricted Services  
# These should work only WITH VPN
# ============================================
$blockedServices = @(
    # === BLOCKED BY RKN (Roskomnadzor) ===
    @{ Name = "Discord"; URL = "https://discord.com"; Category = "Blocked-RKN" }
    @{ Name = "Discord CDN"; URL = "https://cdn.discordapp.com"; Category = "Blocked-RKN" }
    @{ Name = "Discord Status"; URL = "https://status.discord.com"; Category = "Blocked-RKN" }
    @{ Name = "LinkedIn"; URL = "https://www.linkedin.com"; Category = "Blocked-RKN" }
    @{ Name = "Instagram"; URL = "https://www.instagram.com"; Category = "Blocked-RKN" }
    @{ Name = "Twitter/X"; URL = "https://twitter.com"; Category = "Blocked-RKN" }
    @{ Name = "Facebook"; URL = "https://www.facebook.com"; Category = "Blocked-RKN" }
    @{ Name = "Spotify"; URL = "https://www.spotify.com"; Category = "Blocked-RKN" }
    @{ Name = "SoundCloud"; URL = "https://soundcloud.com"; Category = "Blocked-RKN" }
    @{ Name = "Medium"; URL = "https://medium.com"; Category = "Blocked-RKN" }
    @{ Name = "Twitch"; URL = "https://www.twitch.tv"; Category = "Blocked-RKN" }
    @{ Name = "Patreon"; URL = "https://www.patreon.com"; Category = "Blocked-RKN" }
    @{ Name = "DeviantArt"; URL = "https://www.deviantart.com"; Category = "Blocked-RKN" }
    @{ Name = "Pinterest"; URL = "https://www.pinterest.com"; Category = "Blocked-RKN" }
    @{ Name = "Dailymotion"; URL = "https://www.dailymotion.com"; Category = "Blocked-RKN" }
    @{ Name = "Vimeo"; URL = "https://vimeo.com"; Category = "Blocked-RKN" }
    @{ Name = "Quora"; URL = "https://www.quora.com"; Category = "Blocked-RKN" }
    
    # === GEO-RESTRICTED (block access from Russia) ===
    @{ Name = "ChatGPT"; URL = "https://chat.openai.com"; Category = "Geo-blocked" }
    @{ Name = "OpenAI"; URL = "https://openai.com"; Category = "Geo-blocked" }
    @{ Name = "Claude AI"; URL = "https://claude.ai"; Category = "Geo-blocked" }
    @{ Name = "Anthropic"; URL = "https://www.anthropic.com"; Category = "Geo-blocked" }
    @{ Name = "Figma"; URL = "https://www.figma.com"; Category = "Geo-blocked" }
    @{ Name = "Canva"; URL = "https://www.canva.com"; Category = "Geo-blocked" }
    @{ Name = "Notion"; URL = "https://www.notion.so"; Category = "Geo-blocked" }
    @{ Name = "Miro"; URL = "https://miro.com"; Category = "Geo-blocked" }
    @{ Name = "Slack"; URL = "https://slack.com"; Category = "Geo-blocked" }
    @{ Name = "Grammarly"; URL = "https://www.grammarly.com"; Category = "Geo-blocked" }
    @{ Name = "Zoom"; URL = "https://zoom.us"; Category = "Geo-blocked" }
    
    # === GAMING SERVICES ===
    @{ Name = "Steam"; URL = "https://store.steampowered.com"; Category = "Gaming" }
    @{ Name = "Steam Community"; URL = "https://steamcommunity.com"; Category = "Gaming" }
    @{ Name = "Epic Games"; URL = "https://www.epicgames.com"; Category = "Gaming" }
    @{ Name = "EA"; URL = "https://www.ea.com"; Category = "Gaming" }
    @{ Name = "Ubisoft"; URL = "https://www.ubisoft.com"; Category = "Gaming" }
    @{ Name = "Blizzard"; URL = "https://www.blizzard.com"; Category = "Gaming" }
    @{ Name = "Battle.net"; URL = "https://battle.net"; Category = "Gaming" }
    @{ Name = "Riot Games"; URL = "https://www.riotgames.com"; Category = "Gaming" }
    @{ Name = "League of Legends"; URL = "https://www.leagueoflegends.com"; Category = "Gaming" }
    @{ Name = "Valorant"; URL = "https://playvalorant.com"; Category = "Gaming" }
    @{ Name = "GOG"; URL = "https://www.gog.com"; Category = "Gaming" }
    @{ Name = "Xbox"; URL = "https://www.xbox.com"; Category = "Gaming" }
    @{ Name = "PlayStation"; URL = "https://www.playstation.com"; Category = "Gaming" }
    @{ Name = "Nintendo"; URL = "https://www.nintendo.com"; Category = "Gaming" }
    @{ Name = "Rockstar Games"; URL = "https://www.rockstargames.com"; Category = "Gaming" }
    @{ Name = "Roblox"; URL = "https://www.roblox.com"; Category = "Gaming" }
    @{ Name = "Minecraft"; URL = "https://www.minecraft.net"; Category = "Gaming" }
    @{ Name = "Fortnite"; URL = "https://www.fortnite.com"; Category = "Gaming" }
    @{ Name = "Apex Legends"; URL = "https://www.ea.com/games/apex-legends"; Category = "Gaming" }
    @{ Name = "Dota 2"; URL = "https://www.dota2.com"; Category = "Gaming" }
    @{ Name = "Counter-Strike"; URL = "https://www.counter-strike.net"; Category = "Gaming" }
    @{ Name = "FACEIT"; URL = "https://www.faceit.com"; Category = "Gaming" }
    @{ Name = "Overwolf"; URL = "https://www.overwolf.com"; Category = "Gaming" }
)

function Test-ServiceURL {
    param($URL, $TimeoutSec)
    
    $stopwatch = [System.Diagnostics.Stopwatch]::StartNew()
    
    try {
        $response = Invoke-WebRequest -Uri $URL -TimeoutSec $TimeoutSec -UseBasicParsing -MaximumRedirection 10 -ErrorAction Stop
        $stopwatch.Stop()
        
        return @{
            Success = $true
            StatusCode = $response.StatusCode
            Time = $stopwatch.ElapsedMilliseconds
        }
    } catch {
        $stopwatch.Stop()
        $statusCode = 0
        
        if ($_.Exception.Response) {
            $statusCode = [int]$_.Exception.Response.StatusCode
        }
        
        # Consider 2xx, 3xx, 4xx as "reachable" (server responded)
        $isReachable = ($statusCode -ge 200 -and $statusCode -lt 500) -or 
                       ($statusCode -eq 0 -and $_.Exception.Message -match "redirect")
        
        return @{
            Success = $isReachable
            StatusCode = $statusCode
            Error = $_.Exception.Message
            Time = $stopwatch.ElapsedMilliseconds
        }
    }
}

function Test-Services {
    param($Services, $PhaseName)
    
    $results = [System.Collections.ArrayList]@()
    $passedCount = 0
    $failedCount = 0
    
    foreach ($service in $Services) {
        $svcName = $service.Name
        $svcCategory = $service.Category
        
        Write-Host -NoNewline "  Testing $svcName... "
        
        $result = Test-ServiceURL -URL $service.URL -TimeoutSec $Timeout
        
        if ($result.Success) {
            Write-Host "OK" -ForegroundColor Green
            $passedCount++
        } else {
            Write-Host "FAIL" -ForegroundColor Red
            $failedCount++
            if ($Verbose -and $result.Error) {
                Write-Host "    Error: $($result.Error)" -ForegroundColor DarkGray
            }
        }
        
        [void]$results.Add(@{
            Name = $svcName
            Category = $svcCategory
            Success = $result.Success
            StatusCode = $result.StatusCode
            Error = $result.Error
            Time = $result.Time
        })
    }
    
    return @{
        Results = $results
        Passed = $passedCount
        Failed = $failedCount
    }
}

function Show-Summary {
    param($Results, $Title)
    
    Write-Host ""
    Write-Host "  Summary for $Title" -ForegroundColor White
    Write-Host "  ----------------------------------------" -ForegroundColor DarkGray
    
    $categories = $Results | Group-Object Category
    foreach ($cat in $categories) {
        $catPassed = ($cat.Group | Where-Object { $_.Success }).Count
        $catTotal = $cat.Group.Count
        $catColor = if ($catPassed -eq $catTotal) { "Green" } elseif ($catPassed -gt 0) { "Yellow" } else { "Red" }
        
        Write-Host "  $($cat.Name): $catPassed/$catTotal" -ForegroundColor $catColor
    }
}

# ============================================
# MAIN EXECUTION
# ============================================

Write-Host ""
Write-Host "================================================================" -ForegroundColor Cyan
Write-Host "        KampusVPN Service Connectivity Test v2" -ForegroundColor Cyan
Write-Host "================================================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "  Phase 1: Direct access (WITHOUT VPN bypass)" -ForegroundColor Cyan
Write-Host "  Phase 2: Blocked services (WITH VPN bypass)" -ForegroundColor Cyan
Write-Host ""

$allResults = [System.Collections.ArrayList]@()
$totalPassed = 0
$totalFailed = 0

# ============================================
# PHASE 1: Direct Access
# ============================================
if (-not $Phase2Only) {
    Write-Host "================================================================" -ForegroundColor Yellow
    Write-Host " PHASE 1: Direct Access Services" -ForegroundColor Yellow
    Write-Host " These should work WITHOUT VPN (direct connection)" -ForegroundColor DarkYellow
    Write-Host "================================================================" -ForegroundColor Yellow
    Write-Host ""
    
    $phase1 = Test-Services -Services $directServices -PhaseName "Direct Access"
    
    foreach ($r in $phase1.Results) {
        [void]$allResults.Add($r)
    }
    $totalPassed += $phase1.Passed
    $totalFailed += $phase1.Failed
    
    Show-Summary -Results $phase1.Results -Title "Direct Access"
    
    Write-Host ""
    if ($phase1.Failed -eq 0) {
        Write-Host "  [OK] All direct services accessible" -ForegroundColor Green
    } else {
        Write-Host "  [WARN] Some direct services failed: $($phase1.Failed)" -ForegroundColor Yellow
    }
    Write-Host ""
}

# ============================================
# PHASE 2: Blocked/Restricted Services
# ============================================
if (-not $Phase1Only) {
    Write-Host "================================================================" -ForegroundColor Magenta
    Write-Host " PHASE 2: Blocked/Restricted Services" -ForegroundColor Magenta
    Write-Host " These should work only WITH VPN enabled" -ForegroundColor DarkMagenta
    Write-Host "================================================================" -ForegroundColor Magenta
    Write-Host ""
    
    $phase2 = Test-Services -Services $blockedServices -PhaseName "Blocked Services"
    
    foreach ($r in $phase2.Results) {
        [void]$allResults.Add($r)
    }
    $totalPassed += $phase2.Passed
    $totalFailed += $phase2.Failed
    
    Show-Summary -Results $phase2.Results -Title "Blocked Services"
    
    Write-Host ""
    if ($phase2.Failed -eq 0) {
        Write-Host "  [OK] All blocked services accessible via VPN" -ForegroundColor Green
    } else {
        Write-Host "  [FAIL] Some blocked services NOT accessible: $($phase2.Failed)" -ForegroundColor Red
    }
    Write-Host ""
}

# ============================================
# FINAL SUMMARY
# ============================================
Write-Host "================================================================" -ForegroundColor Cyan
Write-Host "                    FINAL RESULTS" -ForegroundColor Cyan
Write-Host "================================================================" -ForegroundColor Cyan
Write-Host ""

Write-Host "  Total: $totalPassed passed, $totalFailed failed"
Write-Host ""

if ($totalFailed -eq 0) {
    Write-Host "  [SUCCESS] VPN is working correctly!" -ForegroundColor Green
    Write-Host "     - Direct services: accessible without VPN bypass" -ForegroundColor DarkGray
    Write-Host "     - Blocked services: accessible via VPN" -ForegroundColor DarkGray
} elseif ($totalFailed -le 3) {
    Write-Host "  [WARN] VPN is mostly working" -ForegroundColor Yellow
    Write-Host "     Some services may have temporary issues" -ForegroundColor DarkGray
} else {
    Write-Host "  [ERROR] VPN may have issues" -ForegroundColor Red
    Write-Host "     Check your connection and VPN settings" -ForegroundColor DarkGray
}

Write-Host ""

# Failed services list
$failedServices = $allResults | Where-Object { -not $_.Success }
if ($failedServices.Count -gt 0) {
    Write-Host "  Failed services:" -ForegroundColor Red
    foreach ($svc in $failedServices) {
        Write-Host "    - $($svc.Name) [$($svc.Category)]" -ForegroundColor DarkRed
        if ($Verbose -and $svc.Error) {
            Write-Host "      $($svc.Error)" -ForegroundColor DarkGray
        }
    }
    Write-Host ""
}

# Output JSON if requested
if ($Json) {
    Write-Host "JSON Output:" -ForegroundColor Cyan
    $allResults | ConvertTo-Json -Depth 3
}

Write-Host ""

# Exit code
if ($totalFailed -gt 0) {
    exit 1
} else {
    exit 0
}
