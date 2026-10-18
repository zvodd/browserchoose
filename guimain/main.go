// Command guimain runs the pick-a-browser window on its own.
//
// It exists so the chooser GUI can be developed and tested independently of the
// systray. Pass a URL as the first argument to try a specific one:
//
//	go run ./guimain https://example.com
package main

import (
	"fmt"
	"os"

	"browserchoose"
	"browserchoose/gui"
)

func main() {
	cfg := browserchoose.Configure()
	list := browserchoose.NewBrowerListFromViper(cfg)

	url := "https://example.com"
	if len(os.Args) > 1 {
		url = os.Args[1]
	}
	fmt.Printf("Choosing a browser for %s\n", url)

	gui.RunChooser(url, list.List)
}
