//go:build !windows

package browserchoose

import (
	"fmt"
	"os"
)

func showError(title, msg string) {
	fmt.Fprintf(os.Stderr, "%s: %s\n", title, msg)
}
