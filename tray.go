package browserchoose

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"

	"browserchoose/icon"
)

// trayApp ties the Fyne system-tray menu, the mux lock and the chooser window
// together. All UI work happens on Fyne's event loop.
type trayApp struct {
	app      fyne.App
	browsers *BrowserList
	cfgPath  string
	mux      *Mux
	desk     desktop.App
}

func newTrayApp(a fyne.App, browsers *BrowserList, cfgPath string) *trayApp {
	t := &trayApp{app: a, browsers: browsers, cfgPath: cfgPath}
	t.mux = NewMux(func() { fyne.Do(t.refreshMenu) })
	return t
}

func (t *trayApp) install() error {
	desk, ok := t.app.(desktop.App)
	if !ok {
		return fmt.Errorf("this Fyne driver has no system tray support")
	}
	t.desk = desk
	desk.SetSystemTrayIcon(fyne.NewStaticResource("browserchoose.ico", icon.Data))
	t.refreshMenu()

	// Fyne quits once its last window is destroyed. Keep one hidden window
	// alive so the tray keeps the daemon running while no chooser is open.
	keepAlive := t.app.NewWindow("BrowserChoose")
	keepAlive.SetCloseIntercept(keepAlive.Hide)
	keepAlive.Show()
	keepAlive.Hide()
	return nil
}

func (t *trayApp) refreshMenu() {
	if t.desk != nil {
		t.desk.SetSystemTrayMenu(t.buildMenu())
	}
}

func (t *trayApp) buildMenu() *fyne.Menu {
	mode := t.mux.Mode()
	status := "Mode: Choose (ask every time)"
	if mode != ChooseMode && mode < len(t.browsers.List) {
		status = fmt.Sprintf("Locked to %s (reverts to Choose after %d min idle)",
			t.browsers.List[mode].Name, int(LockTimeout.Minutes()))
	}

	items := []*fyne.MenuItem{
		fyne.NewMenuItem(status, nil),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem(menuMark(mode == ChooseMode)+"Choose", func() {
			t.mux.Unlock()
		}),
	}
	for i := range t.browsers.List {
		i := i
		items = append(items, fyne.NewMenuItem(menuMark(mode == i)+t.browsers.List[i].Name, func() {
			t.mux.Lock(i)
		}))
	}
	items = append(items,
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Open config", func() {
			if err := OpenConfig(t.cfgPath); err != nil {
				fyne.LogError("open config", err)
			}
		}),
		fyne.NewMenuItem("Quit", t.app.Quit),
	)
	return fyne.NewMenu("BrowserChoose", items...)
}

func menuMark(active bool) string {
	if active {
		return "✓ "
	}
	return ""
}

// handleURL routes one URL through the mux. It is safe to call from any
// goroutine.
func (t *trayApp) handleURL(url string) {
	if i := t.mux.Pick(); i != ChooseMode && i < len(t.browsers.List) {
		t.launch(i, url)
		return
	}
	fyne.Do(func() {
		NewChooser(t.app, url, t.browsers.List, func(i int) {
			t.launch(i, url)
		}).Show()
	})
}

func (t *trayApp) launch(i int, url string) {
	if err := t.browsers.List[i].Launch(url); err != nil {
		fyne.LogError("launch "+t.browsers.List[i].Name, err)
	}
}
