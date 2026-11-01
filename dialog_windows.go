//go:build windows

package browserchoose

import "golang.org/x/sys/windows"

// showError displays a modal error box.
func showError(title, msg string) {
	t, _ := windows.UTF16PtrFromString(title)
	m, _ := windows.UTF16PtrFromString(msg)
	windows.MessageBox(0, m, t, windows.MB_ICONERROR)
}
