package browserchoose

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"fyne.io/fyne/v2/app"
	"github.com/spf13/viper"
)

// appID is the Fyne app ID and the name of the per-user data directory.
const appID = "browserchoose.zvodd.github.com"

// Configure loads browser_choose_config.yaml. It is searched for in the
// per-user data directory, next to the executable, and in the working
// directory. The working directory is not reliable: when Windows launches the
// URL handler it can be anywhere.
func Configure() (*viper.Viper, error) {
	cfg := viper.New()
	cfg.SetConfigName("browser_choose_config")
	cfg.SetConfigType("yaml")
	cfg.AddConfigPath(filepath.Join(os.Getenv("APPDATA"), appID))
	if exe, err := os.Executable(); err == nil {
		cfg.AddConfigPath(filepath.Dir(exe))
	}
	cfg.AddConfigPath(".")
	if err := cfg.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("could not read browser_choose_config.yaml: %w", err)
	}
	return cfg, nil
}

// Run is the single entry point for the BrowserChoose binary.
//
//	BrowserChoose.exe             run the tray daemon
//	BrowserChoose.exe <url>       handle one URL: forward it to the daemon if one
//	                              is running, otherwise ask and exit
//	BrowserChoose.exe -register   register as an http/https/ftp handler
//	BrowserChoose.exe -unregister remove the registration
func Run(args []string) {
	setupLogging()
	defer func() {
		if r := recover(); r != nil {
			fail(fmt.Errorf("unexpected error: %v", r))
		}
	}()

	if len(args) > 1 {
		switch args[1] {
		case "-register", "--register":
			check(RegisterHandler())
			fmt.Println("Registered. Select BrowserChoose for http/https/ftp in the Default Apps window that just opened.")
			return
		case "-unregister", "--unregister":
			check(UnregisterHandler())
			fmt.Println("Unregistered.")
			return
		}
	}

	cfg, err := Configure()
	check(err)
	browsers := NewBrowerListFromViper(cfg)
	log.Printf("loaded config %s (%d browsers)", cfg.ConfigFileUsed(), len(browsers.List))

	// Invoked by Windows as the protocol handler.
	if len(args) > 1 {
		url := args[1]
		log.Printf("handling URL %s", url)
		if ForwardURL(url) {
			log.Printf("forwarded to running daemon")
			return
		}
		log.Printf("no daemon running; showing chooser")
		RunChooser(url, browsers.List)
		return
	}

	// Daemon mode: tray and URL listener share Fyne's event loop. Claim the
	// pipe first so a second daemon exits before it shows a tray icon.
	a := app.NewWithID(appID)
	tray := newTrayApp(a, browsers, cfg.ConfigFileUsed())
	check(ListenURLs(tray.handleURL))
	check(tray.install())
	installSignalHandler(a)
	log.Printf("daemon started")
	a.Run()
	log.Printf("daemon stopped")
}

// setupLogging sends the log to browserchoose.log in the per-user data
// directory. The GUI build has no console, so this is the only place
// diagnostics can be read. If the file cannot be opened, logging stays on
// its default output.
func setupLogging() {
	dir := filepath.Join(os.Getenv("APPDATA"), appID)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return
	}
	f, err := os.OpenFile(filepath.Join(dir, "browserchoose.log"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return
	}
	log.SetOutput(f)
}

// check exits through fail when err is non-nil.
func check(err error) {
	if err != nil {
		fail(err)
	}
}

// fail logs err, tells the user, and exits. A dialog is needed because a
// failure in the GUI build is otherwise invisible.
func fail(err error) {
	log.Printf("fatal: %v", err)
	showError("BrowserChoose", err.Error())
	os.Exit(1)
}
