# KampusVPN Tools

Утилиты для разработки и тестирования KampusVPN.

## check-services.ps1

Скрипт для проверки доступности популярных сервисов после подключения VPN.
Тестирование разделено на 2 фазы для проверки корректности маршрутизации.

### Использование

```powershell
# Полная проверка (обе фазы)
.\check-services.ps1

# С подробным выводом ошибок
.\check-services.ps1 -Verbose

# Только Phase 1 (прямой доступ)
.\check-services.ps1 -Phase1Only

# Только Phase 2 (заблокированные сервисы)
.\check-services.ps1 -Phase2Only

# С JSON выводом результатов
.\check-services.ps1 -Json

# С увеличенным таймаутом (по умолчанию 10 секунд)
.\check-services.ps1 -Timeout 15
```

### Фазы тестирования

**Phase 1: Direct Access (Прямой доступ)**
Сервисы, которые должны работать БЕЗ VPN (напрямую):
- Российские сервисы (Yandex, VK, Mail.ru, Sberbank, Tinkoff и др.)
- Международные незаблокированные (Google, GitHub, YouTube, Wikipedia и др.)

**Phase 2: Blocked Services (Заблокированные)**
Сервисы, которые должны работать ТОЛЬКО ЧЕРЕЗ VPN:
- Заблокированные РКН (Discord, LinkedIn, Instagram, Twitter, Facebook и др.)
- Гео-ограниченные (ChatGPT, Claude AI, Figma, Notion и др.)
- Игровые сервисы (Steam, Epic Games, Blizzard, Riot Games и др.)

### Категории сервисов

| Категория | Описание | Фаза |
|-----------|----------|------|
| **Russian** | Российские сервисы | Phase 1 |
| **International** | Международные незаблокированные | Phase 1 |
| **Blocked-RKN** | Заблокированы Роскомнадзором | Phase 2 |
| **Geo-blocked** | Ограничивают доступ из РФ | Phase 2 |
| **Gaming** | Игровые платформы и сервисы | Phase 2 |

### Пример вывода

```
================================================================
        KampusVPN Service Connectivity Test v2
================================================================

  Phase 1: Direct access (WITHOUT VPN bypass)
  Phase 2: Blocked services (WITH VPN bypass)

================================================================
 PHASE 1: Direct Access Services
================================================================

  Testing Yandex... OK
  Testing VK... OK
  Testing Google... OK
  Testing YouTube... OK
  Testing GitHub... OK

  Summary for Direct Access
  ----------------------------------------
  Russian: 12/12
  International: 14/14

  [OK] All direct services accessible

================================================================
 PHASE 2: Blocked/Restricted Services
================================================================

  Testing Discord... OK
  Testing LinkedIn... OK
  Testing ChatGPT... OK
  Testing Claude AI... OK
  Testing Steam... OK
  Testing Epic Games... OK

  Summary for Blocked Services
  ----------------------------------------
  Blocked-RKN: 17/17
  Geo-blocked: 11/11
  Gaming: 23/23

  [OK] All blocked services accessible via VPN

================================================================
                    FINAL RESULTS
================================================================

  Total: 77 passed, 0 failed

  [SUCCESS] VPN is working correctly!
     - Direct services: accessible without VPN bypass
     - Blocked services: accessible via VPN
```

### Тестируемые сервисы (77 шт.)

**Phase 1: Direct Access (26 сервисов)**

*Российские (12):*
- Yandex, Yandex Mail, VK, Mail.ru
- Sberbank, Tinkoff, Gosuslugi
- Ozon, Wildberries, Avito
- 2GIS, Habr

*Международные незаблокированные (14):*
- Google, Google Drive, Gmail, YouTube
- GitHub, GitLab, Stack Overflow, Wikipedia
- Reddit, Amazon, Microsoft, Apple
- Telegram Web, WhatsApp Web

**Phase 2: Blocked Services (51 сервис)**

*Заблокированные РКН (17):*
- Discord (discord.com, cdn.discordapp.com, status.discord.com)
- LinkedIn, Instagram, Twitter/X, Facebook
- Spotify, SoundCloud, Twitch, Medium
- Patreon, DeviantArt, Pinterest
- Dailymotion, Vimeo, Quora

*Гео-ограниченные (11):*
- ChatGPT, OpenAI, Claude AI, Anthropic
- Figma, Canva, Notion, Miro
- Slack, Grammarly, Zoom

*Игровые сервисы (23):*
- Steam, Steam Community, Epic Games
- EA, Ubisoft, Blizzard, Battle.net
- Riot Games, League of Legends, Valorant
- GOG, Xbox, PlayStation, Nintendo
- Rockstar Games, Roblox, Minecraft
- Fortnite, Apex Legends, Dota 2
- Counter-Strike, FACEIT, Overwolf
