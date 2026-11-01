package browserchoose

import (
	"os"
	"os/signal"
	"syscall"

	"fyne.io/fyne/v2"
)

// installSignalHandler quits the daemon on Ctrl-C or SIGTERM.
func installSignalHandler(a fyne.App) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		fyne.Do(a.Quit)
	}()
}
