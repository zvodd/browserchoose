package browserchoose

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// URLToken is the placeholder in a browser's configured args that is replaced
// with the URL being opened.
const URLToken = "@URL"

// Launch starts this browser for the given URL. Environment variables in the
// configured path (e.g. ${PROGRAMFILES(X86)}) are expanded, and URLToken in the
// configured args is replaced with url.
//
// It returns as soon as the process has been started; it does not wait for the
// browser to exit.
func (b BrowserItem) Launch(url string) error {
	path := os.ExpandEnv(b.Path)
	if path == "" {
		return fmt.Errorf("browser %q has no path configured", b.Name)
	}

	args := make([]string, 0, len(b.Args))
	for _, a := range b.Args {
		args = append(args, strings.ReplaceAll(a, URLToken, url))
	}

	return exec.Command(path, args...).Start()
}

// OpenConfig opens the given file with the OS default handler so the user can
// edit it. On Windows this goes through the shell's `start`, which also offers
// an "Open with" chooser when there is no association for .yaml.
func OpenConfig(path string) error {
	if path == "" {
		return fmt.Errorf("no config file to open")
	}
	return exec.Command("cmd", "/c", "start", "", path).Start()
}
