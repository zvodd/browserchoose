package browserchoose

import (
	"fmt"
	"reflect"

	"browserchoose/icon"

	"github.com/getlantern/systray"
	"github.com/spf13/viper"
)

// Parse config
func Configure() *viper.Viper {
	cfg := viper.New()
	cfg.SetConfigName("browser_choose_config") // name of config file (without extension)
	cfg.SetConfigType("yaml")                  // REQUIRED if the config file does not have the extension in the name
	// TODO check OS
	// TODO Check file exists, or run in portable mode?
	cfg.AddConfigPath("$APPDATA/browserchoose.zvodd.github.com")
	cfg.AddConfigPath(".")    // optionally look for config in the working directory
	err := cfg.ReadInConfig() // Find and read the config file
	if err != nil {           // Handle errors reading the config file
		panic(fmt.Errorf("Fatal error config file: %s \n ", err))
	}

	return cfg
}

func Main() {
	cfg := Configure()
	browserList := NewBrowerListFromViper(cfg)
	go systray.Run(SysTrayOnReadyFactory(browserList, cfg.ConfigFileUsed()), systrayOnExit)
	HandleSingals()
}

// SysTrayOnReadyFactory builds the systray menu.
//
// The first entry is "Choose" (ask which browser to use for each URL); the rest
// map to browserList.List by index. Selecting an entry "locks" the mux to that
// browser. This lock is session-only: it starts on "Choose" and is never
// written back to the config, so it does not survive a restart.
func SysTrayOnReadyFactory(browserList *BrowserList, configPath string) func() {
	return func() {
		systray.SetIcon(icon.Data)
		systray.SetTitle("Browser Choose")
		systray.SetTooltip("Browser Choose")

		// items[0] is the "Choose" entry; items[1:] map to browserList.List.
		items := make([]*systray.MenuItem, 0, len(browserList.List)+1)

		mChoose := systray.AddMenuItem("Choose", "Ask which browser to use each time")
		mChoose.Check()
		items = append(items, mChoose)

		systray.AddSeparator()
		for _, entry := range browserList.List {
			mi := systray.AddMenuItem(entry.Name, fmt.Sprintf("Always open in %s", entry.Name))
			items = append(items, mi)
		}

		systray.AddSeparator()
		mConfig := systray.AddMenuItem("Open config", "Open the config file for editing")
		mQuit := systray.AddMenuItem("Quit", "Quit Browser Choose")

		// systray hands every menu item its own click channel. Fold them all
		// into a single reflect.Select so one goroutine can wait on the lot.
		cases := make([]reflect.SelectCase, 0, len(items)+2)
		for _, mi := range items {
			cases = append(cases, reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(mi.ClickedCh)})
		}
		configIndex := len(cases)
		cases = append(cases, reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(mConfig.ClickedCh)})
		quitIndex := len(cases)
		cases = append(cases, reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(mQuit.ClickedCh)})

		for loop := true; loop; {
			index, _, _ := reflect.Select(cases)
			switch {
			case index == quitIndex:
				systray.Quit()
				loop = false
			case index == configIndex:
				if err := OpenConfig(configPath); err != nil {
					fmt.Printf("open config: %v\n", err)
				}
			default:
				// A mode entry was clicked: check it and clear the others.
				for i, mi := range items {
					if i == index {
						mi.Check()
					} else {
						mi.Uncheck()
					}
				}
				if index == 0 {
					fmt.Println("mux mode: Choose (ask each time)")
				} else {
					fmt.Printf("mux mode: locked to %s\n", browserList.List[index-1].Name)
				}
			}
		}
	}
}

func systrayOnExit() {
	InnerSigChan <- InnerSigQuit
}
