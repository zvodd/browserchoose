package browserchoose

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/getlantern/systray"
)

type InnerSignal int

const (
	InnerSigQuit = iota
)

var (
	InnerSigChan = make(chan InnerSignal)
)

// Select on various channels to handle shutdown
func HandleSingals() {
	OsSigs := make(chan os.Signal, 1)
	signal.Notify(OsSigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-OsSigs
		fmt.Printf("\n Recived OS signal '%v'\n", sig)
		InnerSigChan <- InnerSigQuit
	}()

	fmt.Println("awaiting Signals")
	for loop := true; loop; {

		switch <-InnerSigChan {
		case InnerSigQuit:
			systray.Quit()
			loop = false
			break
		}
	}
	fmt.Println("exiting")
}
