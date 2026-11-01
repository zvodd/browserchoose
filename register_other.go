//go:build !windows

package browserchoose

import "errors"

var errWindowsOnly = errors.New("URL-handler registration is only supported on Windows")

func RegisterHandler() error         { return errWindowsOnly }
func UnregisterHandler() error       { return errWindowsOnly }
func OpenDefaultAppsSettings() error { return errWindowsOnly }
