//go:build windows

package browserchoose

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"golang.org/x/sys/windows/registry"
)

// Everything is written under HKCU, so registration needs no elevation. The
// HKCU\Software\Classes tree is merged into HKCR for the current user.
// See docs/uri-handler-registration.md for the layout and the Windows 8+
// restriction that still requires the user to confirm in Default Apps.
const (
	progID           = "BrowserChooseURL"
	capabilitiesPath = `SOFTWARE\BrowserChoose\Capabilities`
	appRegisteredKey = `SOFTWARE\RegisteredApplications`
)

// handledSchemes are the URL schemes BrowserChoose takes over.
var handledSchemes = []string{"http", "https", "ftp"}

// RegisterHandler makes BrowserChoose available as a default handler for
// http/https/ftp, then opens the Default Apps settings so the user can
// confirm it (Windows 8+ refuses programmatic default changes).
func RegisterHandler() error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	if exe, err = filepath.Abs(exe); err != nil {
		return err
	}

	if err := setKey(`SOFTWARE\Classes\`+progID, map[string]string{
		"":             "BrowserChoose URL",
		"URL Protocol": "",
	}); err != nil {
		return err
	}
	if err := setKey(`SOFTWARE\Classes\`+progID+`\DefaultIcon`, map[string]string{
		"": fmt.Sprintf(`"%s",0`, exe),
	}); err != nil {
		return err
	}
	if err := setKey(`SOFTWARE\Classes\`+progID+`\shell\open\command`, map[string]string{
		"": fmt.Sprintf(`"%s" "%%1"`, exe),
	}); err != nil {
		return err
	}
	if err := setKey(capabilitiesPath, map[string]string{
		"ApplicationName":        "BrowserChoose",
		"ApplicationDescription": "Routes each URL to a browser you pick on the fly.",
	}); err != nil {
		return err
	}

	assoc := map[string]string{}
	for _, scheme := range handledSchemes {
		assoc[scheme] = progID
	}
	if err := setKey(capabilitiesPath+`\URLAssociations`, assoc); err != nil {
		return err
	}

	if err := setKey(appRegisteredKey, map[string]string{
		"BrowserChoose": capabilitiesPath,
	}); err != nil {
		return err
	}

	return OpenDefaultAppsSettings()
}

// UnregisterHandler removes everything RegisterHandler wrote. Missing keys are
// ignored.
func UnregisterHandler() error {
	for _, p := range []string{
		`SOFTWARE\Classes\` + progID + `\shell\open\command`,
		`SOFTWARE\Classes\` + progID + `\shell\open`,
		`SOFTWARE\Classes\` + progID + `\shell`,
		`SOFTWARE\Classes\` + progID + `\DefaultIcon`,
		`SOFTWARE\Classes\` + progID,
		capabilitiesPath + `\URLAssociations`,
		capabilitiesPath,
		`SOFTWARE\BrowserChoose`,
	} {
		registry.DeleteKey(registry.CURRENT_USER, p) // best effort
	}

	key, err := registry.OpenKey(registry.CURRENT_USER, appRegisteredKey, registry.SET_VALUE)
	if err != nil {
		return nil
	}
	defer key.Close()
	key.DeleteValue("BrowserChoose") // best effort
	return nil
}

// OpenDefaultAppsSettings opens the Windows 10/11 "Default apps" page.
func OpenDefaultAppsSettings() error {
	return exec.Command("cmd", "/c", "start", "", "ms-settings:defaultapps").Start()
}

func setKey(path string, values map[string]string) error {
	key, _, err := registry.CreateKey(registry.CURRENT_USER, path, registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer key.Close()
	for name, value := range values {
		if err := key.SetStringValue(name, value); err != nil {
			return err
		}
	}
	return nil
}
