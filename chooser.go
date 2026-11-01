package browserchoose

import (
	"fmt"
	"os"
	"path/filepath"

	"fyne.io/fyne/v2/app"
	"github.com/spf13/viper"
)

// appID is the Fyne app ID and the name of the per-user config directory.
const appID = "browserchoose.zvodd.github.com"

// Configure loads browser_choose_config.yaml.
func Configure() *viper.Viper {
	cfg := viper.New()
	cfg.SetConfigName("browser_choose_config")
	cfg.SetConfigType("yaml")
	cfg.AddConfigPath(filepath.Join(os.Getenv("APPDATA"), appID))
	cfg.AddConfigPath(".") // portable mode
	if err := cfg.ReadInConfig(); err != nil {
		panic(fmt.Errorf("fatal error config file: %w", err))
	}
	return cfg
}

// Run is the single entry point for the BrowserChoose binary.
//
//	BrowserChoose.exe             run the tray daemon
//	BrowserChoose.exe <url>       handle one URL: forward it to the daemon if one
//	                              is running, otherwise ask and exit
//	BrowserChoose.exe -register   register as an http/https/ftp handler
//	BrowserChoose.exe -unregister remove the registration
func Run(args []string) {
	if len(args) > 1 {
		switch args[1] {
		case "-register", "--register":
			exitOnErr(RegisterHandler())
			fmt.Println("Registered. Select BrowserChoose for http/https/ftp in the Default Apps window that just opened.")
			return
		case "-unregister", "--unregister":
			exitOnErr(UnregisterHandler())
			fmt.Println("Unregistered.")
			return
		}
	}

	cfg := Configure()
	browsers := NewBrowerListFromViper(cfg)

	// Invoked by Windows as the protocol handler.
	if len(args) > 1 {
		url := args[1]
		if ForwardURL(url) {
			return
		}
		RunChooser(url, browsers.List)
		return
	}

	// Daemon mode: tray and URL listener share Fyne's event loop. Claim the
	// pipe first so a second daemon exits before it shows a tray icon.
	a := app.NewWithID(appID)
	tray := newTrayApp(a, browsers, cfg.ConfigFileUsed())
	if err := ListenURLs(tray.handleURL); err != nil {
		exitOnErr(fmt.Errorf("cannot accept URLs (is BrowserChoose already running?): %w", err))
	}
	exitOnErr(tray.install())
	installSignalHandler(a)
	a.Run()
}

func exitOnErr(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, "browserchoose:", err)
		os.Exit(1)
	}
}
