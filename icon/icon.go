// Package icon embeds the application icon. Regenerate WebIcon.ico from
// WebIcon.png with `go run ./icon/makeico` after changing the artwork.
package icon

import (
	_ "embed"

	"fyne.io/fyne/v2"
)

//go:embed WebIcon.png
var pngData []byte

//go:embed WebIcon.ico
var icoData []byte

// AppIcon is the window and application icon for Fyne windows.
var AppIcon = fyne.NewStaticResource("WebIcon.png", pngData)

// TrayIcon is the systray icon. The Windows tray loads icons from ICO data.
var TrayIcon = fyne.NewStaticResource("WebIcon.ico", icoData)
