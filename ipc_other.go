//go:build !windows

package browserchoose

import "errors"

func ForwardURL(url string) bool { return false }

func ListenURLs(handle func(url string)) error {
	return errors.New("URL forwarding is only supported on Windows")
}
