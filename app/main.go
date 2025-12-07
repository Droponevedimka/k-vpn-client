package main

import (
	"embed"
	"log"
	"os"
	"syscall"
	"unsafe"

	"github.com/energye/systray"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
)

//go:embed all:frontend/dist
var assets embed.FS

//go:embed assets/icons/icon_grey.ico
var iconGrey []byte

//go:embed assets/icons/icon_green.ico
var iconGreen []byte

//go:embed assets/icons/icon_red.ico
var iconRed []byte

//go:embed config/template.json
var embeddedTemplate []byte

var appInstance *App
var systrayReady = make(chan struct{})

// Windows API для single instance и смены иконки
var (
	kernel32        = syscall.NewLazyDLL("kernel32.dll")
	user32          = syscall.NewLazyDLL("user32.dll")
	createMutex     = kernel32.NewProc("CreateMutexW")
	findWindow      = user32.NewProc("FindWindowW")
	showWindow      = user32.NewProc("ShowWindow")
	setForeground   = user32.NewProc("SetForegroundWindow")
	sendMessage     = user32.NewProc("SendMessageW")
	createIconFromResourceEx = user32.NewProc("CreateIconFromResourceEx")
	destroyIcon     = user32.NewProc("DestroyIcon")
	lookupIconIdFromDirectoryEx = user32.NewProc("LookupIconIdFromDirectoryEx")
)

const (
	SW_RESTORE     = 9
	WM_SETICON     = 0x0080
	ICON_SMALL     = 0
	ICON_BIG       = 1
	LR_DEFAULTCOLOR = 0x00000000
)

// copyEmbeddedTemplate копирует встроенный template.json в указанный путь
func copyEmbeddedTemplate(destPath string) error {
	return os.WriteFile(destPath, embeddedTemplate, 0644)
}

func main() {
	// Проверяем single instance
	mutexName, _ := syscall.UTF16PtrFromString("KampusVPN_SingleInstance")
	handle, _, _ := createMutex.Call(0, 0, uintptr(unsafe.Pointer(mutexName)))

	if syscall.GetLastError() == syscall.ERROR_ALREADY_EXISTS {
		// Приложение уже запущено - показываем существующее окно
		windowName, _ := syscall.UTF16PtrFromString("Kampus VPN")
		hwnd, _, _ := findWindow.Call(0, uintptr(unsafe.Pointer(windowName)))
		if hwnd != 0 {
			showWindow.Call(hwnd, SW_RESTORE)
			setForeground.Call(hwnd)
		}
		os.Exit(0)
	}
	defer syscall.CloseHandle(syscall.Handle(handle))

	appInstance = NewApp()

	// Запускаем systray в отдельной горутине (более надёжно на Windows)
	go func() {
		systray.Run(onSystrayReady, onSystrayExit)
	}()

	// Небольшая задержка для инициализации systray
	<-systrayReady

	// Запускаем Wails в main goroutine (более стабильно для GUI)
	runWails()
}

func runWails() {
	err := wails.Run(&options.App{
		Title:     "Kampus VPN",
		Width:     570,
		Height:    755,
		MinWidth:  570,
		MinHeight: 755,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        appInstance.startup,
		OnShutdown:       appInstance.shutdown,
		Bind: []interface{}{
			appInstance,
		},
		Windows: &windows.Options{
			WebviewIsTransparent: false,
			WindowIsTranslucent:  false,
			DisableWindowIcon:    false,
		},
		Frameless: false,
		// При закрытии окна - скрывать в трей
		HideWindowOnClose: true,
	})

	if err != nil {
		log.Fatal(err)
	}
}

func onSystrayReady() {
	systray.SetIcon(iconGrey)
	systray.SetTitle("Kampus VPN")
	systray.SetTooltip("Kampus VPN - Отключено")

	// Левый клик - открыть приложение
	systray.SetOnClick(func(menu systray.IMenu) {
		if appInstance != nil {
			appInstance.ShowWindow()
		}
	})

	// Двойной клик - тоже открыть
	systray.SetOnDClick(func(menu systray.IMenu) {
		if appInstance != nil {
			appInstance.ShowWindow()
		}
	})

	// Правый клик - показать меню
	systray.SetOnRClick(func(menu systray.IMenu) {
		menu.ShowMenu()
	})

	// Пункты меню (показываются по правому клику)
	mShow := systray.AddMenuItem("Открыть", "Показать окно")
	systray.AddSeparator()
	mLogs := systray.AddMenuItem("Логи", "Открыть файл логов")
	mAbout := systray.AddMenuItem("О программе", "Информация о программе")
	systray.AddSeparator()
	mQuit := systray.AddMenuItem("Выход", "Закрыть приложение")

	// Сигнализируем что systray готов
	close(systrayReady)

	// Обработка кликов по пунктам меню
	mShow.Click(func() {
		if appInstance != nil {
			appInstance.ShowWindow()
		}
	})

	mLogs.Click(func() {
		if appInstance != nil {
			appInstance.OpenLogs()
		}
	})

	mAbout.Click(func() {
		if appInstance != nil {
			appInstance.ShowAbout()
		}
	})

	mQuit.Click(func() {
		if appInstance != nil {
			appInstance.QuitApp()
		}
		systray.Quit()
	})
}

func onSystrayExit() {
	// Cleanup при выходе из systray
}

// UpdateTrayIcon обновляет иконку в трее и в окне приложения
func UpdateTrayIcon(status string) {
	var iconData []byte
	var tooltip string
	
	switch status {
	case "connected":
		iconData = iconGreen
		tooltip = "Kampus VPN - Подключено"
	case "error":
		iconData = iconRed
		tooltip = "Kampus VPN - Ошибка"
	default:
		iconData = iconGrey
		tooltip = "Kampus VPN - Отключено"
	}
	
	// Обновляем иконку в трее
	systray.SetIcon(iconData)
	systray.SetTooltip(tooltip)
	
	// Обновляем иконку окна
	go setWindowIcon(iconData)
}

// setWindowIcon устанавливает иконку окна через Windows API
func setWindowIcon(iconData []byte) {
	if len(iconData) == 0 {
		return
	}
	
	// Находим окно по заголовку
	windowName, _ := syscall.UTF16PtrFromString("Kampus VPN")
	hwnd, _, _ := findWindow.Call(0, uintptr(unsafe.Pointer(windowName)))
	if hwnd == 0 {
		return
	}
	
	// Создаем иконку из данных .ico файла
	// ICO файл содержит директорию иконок, нужно найти нужный размер
	hIcon := createIconFromICO(iconData, 32, 32) // Большая иконка
	hIconSmall := createIconFromICO(iconData, 16, 16) // Маленькая иконка
	
	if hIcon != 0 {
		sendMessage.Call(hwnd, WM_SETICON, ICON_BIG, hIcon)
	}
	if hIconSmall != 0 {
		sendMessage.Call(hwnd, WM_SETICON, ICON_SMALL, hIconSmall)
	}
}

// createIconFromICO создает HICON из данных .ico файла
func createIconFromICO(icoData []byte, width, height int) uintptr {
	if len(icoData) < 6 {
		return 0
	}
	
	// Структура ICO файла:
	// ICONDIR (6 bytes): reserved(2), type(2), count(2)
	// ICONDIRENTRY (16 bytes each): width, height, colorCount, reserved, planes(2), bitCount(2), bytesInRes(4), imageOffset(4)
	
	// Проверяем заголовок ICO
	if icoData[0] != 0 || icoData[1] != 0 || icoData[2] != 1 || icoData[3] != 0 {
		return 0 // Не ICO файл
	}
	
	count := int(icoData[4]) | int(icoData[5])<<8
	if count == 0 {
		return 0
	}
	
	// Ищем подходящий размер иконки
	bestIdx := 0
	bestSize := 0
	
	for i := 0; i < count; i++ {
		entryOffset := 6 + i*16
		if entryOffset+16 > len(icoData) {
			break
		}
		
		w := int(icoData[entryOffset])
		h := int(icoData[entryOffset+1])
		if w == 0 {
			w = 256
		}
		if h == 0 {
			h = 256
		}
		
		// Ищем ближайший размер к запрошенному
		size := w
		if w == width && h == height {
			bestIdx = i
			break
		}
		if size > bestSize && size <= width*2 {
			bestSize = size
			bestIdx = i
		}
	}
	
	// Получаем данные выбранной иконки
	entryOffset := 6 + bestIdx*16
	if entryOffset+16 > len(icoData) {
		return 0
	}
	
	bytesInRes := int(icoData[entryOffset+8]) | int(icoData[entryOffset+9])<<8 | 
		int(icoData[entryOffset+10])<<16 | int(icoData[entryOffset+11])<<24
	imageOffset := int(icoData[entryOffset+12]) | int(icoData[entryOffset+13])<<8 | 
		int(icoData[entryOffset+14])<<16 | int(icoData[entryOffset+15])<<24
	
	if imageOffset+bytesInRes > len(icoData) {
		return 0
	}
	
	// Создаем иконку из ресурса
	imageData := icoData[imageOffset : imageOffset+bytesInRes]
	
	hIcon, _, _ := createIconFromResourceEx.Call(
		uintptr(unsafe.Pointer(&imageData[0])),
		uintptr(bytesInRes),
		1, // TRUE = icon
		0x00030000, // Version
		uintptr(width),
		uintptr(height),
		LR_DEFAULTCOLOR,
	)
	
	return hIcon
}
