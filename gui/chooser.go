// Package gui renders the pick-a-browser window.
//
// It is deliberately kept separate from the systray so the window can be built,
// run and debugged on its own (see the guimain command). Wiring the two into a
// single process is a later step.
package gui

import (
	"fmt"
	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"browserchoose"
)

// Chooser is the window that asks which browser should handle a URL.
type Chooser struct {
	App      fyne.App
	URL      string
	Browsers []browserchoose.BrowserItem
	OnPick   func(index int)

	window fyne.Window
}

// New builds a chooser window for url. It is shown with Show.
func New(a fyne.App, url string, browsers []browserchoose.BrowserItem, onPick func(index int)) *Chooser {
	c := &Chooser{App: a, URL: url, Browsers: browsers, OnPick: onPick}
	c.window = a.NewWindow("Open URL")
	c.window.SetContent(c.build())
	c.window.Resize(fyne.NewSize(420, 200))
	return c
}

// Show displays the chooser window without blocking.
func (c *Chooser) Show() {
	c.window.Show()
}

// RunChooser is the standalone entry point: it creates its own Fyne app, shows
// the chooser, and blocks until the window is closed. Picking a browser launches
// it with the URL.
func RunChooser(url string, browsers []browserchoose.BrowserItem) {
	a := app.New()
	c := New(a, url, browsers, func(i int) {
		if err := browsers[i].Launch(url); err != nil {
			fmt.Fprintf(os.Stderr, "launch %s: %v\n", browsers[i].Name, err)
		}
		a.Quit()
	})
	c.Show()
	a.Run()
}

func (c *Chooser) build() fyne.CanvasObject {
	title := widget.NewLabelWithStyle("Which browser should open this URL?", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	urlEntry := widget.NewEntry()
	urlEntry.SetText(c.URL)
	urlEntry.Disable()

	buttons := make([]fyne.CanvasObject, 0, len(c.Browsers))
	for i := range c.Browsers {
		buttons = append(buttons, c.browserButton(i))
	}

	body := container.NewVBox(
		title,
		urlEntry,
		container.NewHBox(buttons...),
	)
	return container.NewPadded(body)
}

func (c *Chooser) browserButton(index int) fyne.CanvasObject {
	b := c.Browsers[index]
	onClick := func() {
		if c.OnPick != nil {
			c.OnPick(index)
		}
	}
	if res, err := loadIcon(b.Icon); err == nil {
		return widget.NewButtonWithIcon(b.Name, res, onClick)
	}
	return widget.NewButton(b.Name, onClick)
}

func loadIcon(path string) (fyne.Resource, error) {
	if path == "" {
		return nil, fmt.Errorf("no icon configured")
	}
	return fyne.LoadResourceFromPath(os.ExpandEnv(path))
}
