package browserchoose

import (
	"fmt"
	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// Chooser is the window that asks which browser should handle a URL.
type Chooser struct {
	URL      string
	Browsers []BrowserItem
	OnPick   func(index int)

	window fyne.Window
}

// NewChooser builds a chooser window for url. Call Show to display it.
func NewChooser(a fyne.App, url string, browsers []BrowserItem, onPick func(index int)) *Chooser {
	c := &Chooser{URL: url, Browsers: browsers, OnPick: onPick}
	c.window = a.NewWindow("Open URL")
	c.window.SetContent(c.build())
	c.window.Resize(fyne.NewSize(420, 200))
	c.window.CenterOnScreen()
	return c
}

// Show displays the chooser window without blocking.
func (c *Chooser) Show() {
	c.window.Show()
}

// RunChooser is the standalone path used when BrowserChoose is invoked as a
// URL handler and no daemon is running. It blocks until the window closes,
// which happens once a browser has been picked or the window is dismissed.
func RunChooser(url string, browsers []BrowserItem) {
	a := app.New()
	c := NewChooser(a, url, browsers, func(i int) {
		if err := browsers[i].Launch(url); err != nil {
			fmt.Fprintf(os.Stderr, "launch %s: %v\n", browsers[i].Name, err)
		}
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
		c.window.Close()
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
