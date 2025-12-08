// Kampus VPN - i18n (Internationalization) Module

const i18n = {
    ru: {
        // Status
        connected: 'Подключено',
        disconnected: 'Отключено',
        connecting: 'Подключение...',
        disconnecting: 'Отключение...',
        // Badges
        addVpn: '⚡ Добавить VPN',
        proxies: 'прокси',
        workNetworks: '🏢 Рабочие сети',
        // Profile
        profile: 'Профиль',
        profiles: 'Профили',
        selectProfile: 'Выберите или создайте профиль подключения',
        createProfile: '➕ Создать профиль',
        newProfile: '➕ Новый профиль',
        editProfile: '✏️ Редактировать профиль',
        profileName: 'Название профиля',
        default: 'По умолчанию',
        noSubscription: 'Нет подписки',
        cantDeleteDefault: 'Нельзя удалить профиль по умолчанию',
        deleteProfile: 'Удаление профиля',
        deleteProfileConfirm: 'Удалить профиль "{name}"? Все настройки профиля будут потеряны.',
        profileCreated: 'Профиль создан',
        profileUpdated: 'Профиль обновлён',
        profileDeleted: 'Профиль удалён',
        profileActivated: 'Профиль активирован',
        disconnectFirst: 'Отключите VPN перед сменой профиля',
        enterProfileName: 'Введите название профиля',
        // Settings
        settings: 'Настройки',
        general: 'Общие',
        autoStart: 'Автозапуск',
        autoStartDesc: 'Запускать при входе в систему',
        notifications: 'Уведомления',
        notificationsDesc: 'Показывать уведомления о подключении',
        logging: 'Логирование sing-box',
        loggingDesc: 'Записывать логи в файл',
        subscription: 'Подписка',
        autoUpdate: 'Авто-обновление',
        autoUpdateDesc: 'Обновлять подписку автоматически',
        updates: 'Обновления',
        checkUpdates: 'Проверять обновления',
        checkUpdatesDesc: 'Уведомлять о новых версиях',
        appearance: 'Внешний вид',
        theme: 'Тема',
        themeDesc: 'Оформление приложения',
        themeDark: 'Тёмная',
        themeLight: 'Светлая',
        themeSystem: 'Системная',
        language: 'Язык',
        languageDesc: 'Язык интерфейса',
        configuration: 'Конфигурация',
        templateEditor: 'Редактор шаблона',
        // Actions
        cancel: 'Отмена',
        close: 'Закрыть',
        save: 'Сохранить',
        create: 'Создать',
        delete: 'Удалить',
        edit: 'Редактировать',
        folder: 'Папка',
        settingsSaved: 'Настройки сохранены',
        // Errors
        error: 'Ошибка',
        warning: 'Предупреждение',
        // VPN
        vpnConnected: 'VPN подключен',
        vpnDisconnected: 'VPN отключен',
        disconnectVpnFirst: 'Сначала отключите VPN',
        secureConnection: 'Безопасное подключение',
        // WireGuard
        wireGuardInstallTitle: 'WireGuard не установлен',
        wireGuardInstallDesc: 'Для работы рабочих сетей необходим WireGuard',
        wireGuardInstallBtn: 'Установить WireGuard',
        wireGuardInstalling: 'Установка...',
        wireGuardInstalled: 'WireGuard установлен',
        wireGuardInstallError: 'Ошибка установки WireGuard',
        tunnelRunning: 'Работает',
        tunnelStopped: 'Остановлен',
        startTunnel: 'Запустить',
        stopTunnel: 'Остановить',
        startAllTunnels: '▶ Запустить все',
        stopAllTunnels: '◼ Остановить все',
        tunnelStarted: 'Туннель запущен',
        tunnelStopped: 'Туннель остановлен',
        tunnelError: 'Ошибка туннеля',
        tunnelActive: 'Активен',
        noWireGuardConfigs: 'Нет конфигураций WireGuard',
        addWireGuardConfig: 'Добавьте конфигурацию для рабочей сети',
        wireGuardVersion: 'Версия WireGuard',
        wireGuardVersionDesc: 'Версия Native WireGuard для рабочих сетей',
        wireGuardVersionChanged: 'Версия WireGuard изменена',
        internalDomains: 'Внутренние домены',
        // Import/Export
        exportProfiles: 'Экспорт профилей',
        importProfiles: 'Импорт профилей',
        profilesExported: 'Экспортировано {count} профилей',
        profilesImported: 'Импортировано {count} профилей',
        exportFailed: 'Ошибка экспорта',
        importFailed: 'Ошибка импорта',
        exportNotAvailable: 'Экспорт недоступен',
        importNotAvailable: 'Импорт недоступен',
        importConfirmMessage: 'Вы уверены что хотите импортировать профили?',
        profilesFound: 'Найдено профилей',
        wireGuardConfigs: 'WireGuard конфигов',
        hasTemplate: 'Включает шаблон',
        yes: 'Да',
        no: 'Нет',
        importWarning: '⚠️ ВНИМАНИЕ: Все текущие профили будут заменены!',
    },
    en: {
        // Status
        connected: 'Connected',
        disconnected: 'Disconnected',
        connecting: 'Connecting...',
        disconnecting: 'Disconnecting...',
        // Badges
        addVpn: '⚡ Add VPN',
        proxies: 'proxies',
        workNetworks: '🏢 Work networks',
        // Profile
        profile: 'Profile',
        profiles: 'Profiles',
        selectProfile: 'Select or create a connection profile',
        createProfile: '➕ Create profile',
        newProfile: '➕ New profile',
        editProfile: '✏️ Edit profile',
        profileName: 'Profile name',
        default: 'Default',
        noSubscription: 'No subscription',
        cantDeleteDefault: 'Cannot delete the default profile',
        deleteProfile: 'Delete profile',
        deleteProfileConfirm: 'Delete profile "{name}"? All profile settings will be lost.',
        profileCreated: 'Profile created',
        profileUpdated: 'Profile updated',
        profileDeleted: 'Profile deleted',
        profileActivated: 'Profile activated',
        disconnectFirst: 'Disconnect VPN before switching profile',
        enterProfileName: 'Enter profile name',
        // Settings
        settings: 'Settings',
        general: 'General',
        autoStart: 'Auto-start',
        autoStartDesc: 'Launch at system startup',
        notifications: 'Notifications',
        notificationsDesc: 'Show connection notifications',
        logging: 'sing-box logging',
        loggingDesc: 'Write logs to file',
        subscription: 'Subscription',
        autoUpdate: 'Auto-update',
        autoUpdateDesc: 'Update subscription automatically',
        updates: 'Updates',
        checkUpdates: 'Check for updates',
        checkUpdatesDesc: 'Notify about new versions',
        appearance: 'Appearance',
        theme: 'Theme',
        themeDesc: 'App theme',
        themeDark: 'Dark',
        themeLight: 'Light',
        themeSystem: 'System',
        language: 'Language',
        languageDesc: 'Interface language',
        configuration: 'Configuration',
        templateEditor: 'Template editor',
        // Actions
        cancel: 'Cancel',
        close: 'Close',
        save: 'Save',
        create: 'Create',
        delete: 'Delete',
        edit: 'Edit',
        folder: 'Folder',
        settingsSaved: 'Settings saved',
        // Errors
        error: 'Error',
        warning: 'Warning',
        // VPN
        vpnConnected: 'VPN connected',
        vpnDisconnected: 'VPN disconnected',
        disconnectVpnFirst: 'Disconnect VPN first',
        secureConnection: 'Secure connection',
        // WireGuard
        wireGuardInstallTitle: 'WireGuard not installed',
        wireGuardInstallDesc: 'WireGuard is required for work networks',
        wireGuardInstallBtn: 'Install WireGuard',
        wireGuardInstalling: 'Installing...',
        wireGuardInstalled: 'WireGuard installed',
        wireGuardInstallError: 'WireGuard installation error',
        tunnelRunning: 'Running',
        tunnelStopped: 'Stopped',
        startTunnel: 'Start',
        stopTunnel: 'Stop',
        startAllTunnels: '▶ Start all',
        stopAllTunnels: '◼ Stop all',
        tunnelStarted: 'Tunnel started',
        tunnelStopped: 'Tunnel stopped',
        tunnelError: 'Tunnel error',
        tunnelActive: 'Active',
        noWireGuardConfigs: 'No WireGuard configurations',
        addWireGuardConfig: 'Add a work network configuration',
        wireGuardVersion: 'WireGuard version',
        wireGuardVersionDesc: 'Native WireGuard version for work networks',
        wireGuardVersionChanged: 'WireGuard version changed',
        internalDomains: 'Internal domains',
        // Import/Export
        exportProfiles: 'Export profiles',
        importProfiles: 'Import profiles',
        profilesExported: 'Exported {count} profiles',
        profilesImported: 'Imported {count} profiles',
        exportFailed: 'Export failed',
        importFailed: 'Import failed',
        exportNotAvailable: 'Export not available',
        importNotAvailable: 'Import not available',
        importConfirmMessage: 'Are you sure you want to import profiles?',
        profilesFound: 'Profiles found',
        wireGuardConfigs: 'WireGuard configs',
        hasTemplate: 'Includes template',
        yes: 'Yes',
        no: 'No',
        importWarning: '⚠️ WARNING: All current profiles will be replaced!',
    }
};

let currentLang = 'ru';

// Get translation
function t(key, params = {}) {
    let text = i18n[currentLang][key] || i18n['ru'][key] || key;
    // Replace placeholders like {name}
    Object.keys(params).forEach(k => {
        text = text.replace(`{${k}}`, params[k]);
    });
    return text;
}

// Apply language to UI
function applyLanguage(lang) {
    currentLang = lang;
    document.documentElement.setAttribute('lang', lang);
    updateUITexts();
}

// Update all UI texts
function updateUITexts() {
    // Status will update on next status update
    // Subtitle
    const subtitle = document.getElementById('activeProfileName');
    if (subtitle && currentProfiles && currentProfiles.length > 0) {
        updateActiveProfileDisplay();
    } else if (subtitle) {
        subtitle.textContent = t('secureConnection');
    }
}
