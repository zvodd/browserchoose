package browserchoose

import (
	"fmt"
	"log"
	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"browserchoose/icon"
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
	c.window.Resize(fyne.NewSize(560, 220))
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
	a.SetIcon(icon.AppIcon)
	c := NewChooser(a, url, browsers, func(i int) {
		log.Printf("launching %s for %s", browsers[i].Name, url)
		if err := browsers[i].Launch(url); err != nil {
			log.Printf("launch %s: %v", browsers[i].Name, err)
		}
	})
	c.Show()
	a.Run()
}

func (c *Chooser) build() fyne.CanvasObject {
	title := widget.NewLabelWithStyle("Which browser should open this URL?", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	// A plain label, not a disabled Entry: disabled entries are drawn faded
	// and become hard to read. Long URLs wrap at any character.
	link := widget.NewLabel(c.URL)
	link.Wrapping = fyne.TextWrapBreak
	link.TextStyle = fyne.TextStyle{Monospace: true}

	buttons := make([]fyne.CanvasObject, 0, len(c.Browsers))
	for i := range c.Browsers {
		buttons = append(buttons, c.browserButton(i))
	}

	return container.NewPadded(container.NewVBox(
		title,
		link,
		widget.NewSeparator(),
		container.NewGridWrap(fyne.NewSize(160, 48), buttons...),
	))
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
