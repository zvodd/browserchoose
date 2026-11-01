// Command guimain runs the pick-a-browser window on its own, without the tray
// daemon. It is a development aid for the chooser. Pass a URL as the first
// argument to try a specific one:
//
//	go run ./guimain https://example.com
package main

import (
	"fmt"
	"os"

	"browserchoose"
)

func main() {
	cfg := browserchoose.Configure()
	list := browserchoose.NewBrowerListFromViper(cfg)

	url := "https://example.com"
	if len(os.Args) > 1 {
		url = os.Args[1]
	}
	fmt.Printf("Choosing a browser for %s\n", url)

	browserchoose.RunChooser(url, list.List)
}
