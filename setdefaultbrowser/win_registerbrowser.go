package setdefaultbrowser

import (
	"fmt"
	"log"

	"golang.org/x/sys/windows/registry"
)

// Approach is to set the app as the browser, not the _real_ selected browser.
// Then use a proxy launch / syscall to actually launch the browser.
// This means users can redirect url handlers to things like wget or shell scripts
// or specify a profile with arguments to the browser executable.
// or have a modal pop up to select the browser option per url call.

func registryquery() {
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\Microsoft\Windows NT\CurrentVersion`, registry.QUERY_VALUE)
	if err != nil {
		log.Fatal(err)
	}
	defer k.Close()

	s, _, err := k.GetStringValue("SystemRoot")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Windows system root is %q\n", s)
}
