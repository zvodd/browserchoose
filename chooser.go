package browserchoose

import (
	"fmt"
	"time"

	"reflect"

	"browserchoose/icon"

	"github.com/getlantern/systray"
	"github.com/spf13/viper"
)

//
// Some Config and Defaults
//
var (
	browserItems = map[string]*systray.MenuItem{}
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
	// x := cfg.Sub("browsers") //.GetStringSlice("args")
	// fmt.Printf("%+v\n\n\n", x)
	// y := cfg.GetStringMap("browsers")
	// fmt.Printf("%+v\n", y)
	go systray.Run(SysTrayOnReadyFactory(browserList), systrayOnExit)
	HandleSingals()
}

func SysTrayOnReadyFactory(browserList *BrowserList) func() {
	return func() {
		systray.SetIcon(icon.Data)
		systray.SetTitle("Browser Choose")
		// systray.SetTooltip(cfg.GetString("unused"))
		rflslcs := make([]reflect.SelectCase, len(browserList.List))

		for i, entry := range browserList.List {
			mi := systray.AddMenuItem(entry.Name, fmt.Sprintf("Set browser to %s", entry.Name))
			if i == browserList.Selected {
				mi.Check()
			}
			browserItems[entry.Name] = mi
			rflslcs[i] = reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(mi.ClickedCh)}
		}

		systray.AddSeparator()
		mQuit := systray.AddMenuItem("Quit", "Quit BrowserChooser")
		// Sets the icon of a menu item. Only available on Mac and Windows.
		mQuit.SetIcon(icon.Data)

		// append here so quit index is == len(browserList.List)
		rflslcs = append(rflslcs, reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(mQuit.ClickedCh)})

		for loop := true; loop; {

			// I wish that stray used a differnt mechanism for handling clicks, that isnt just every item has its own channel
			// e.g. specify one channel and and an ID for each item.
			// select {
			// case <-mQuit.ClickedCh:
			// 	systray.Quit()
			// 	loop = false
			// }
			// fmt.Println("SystrayLoop Canary")

			time.Sleep(1 * time.Second)
			index, value, recvOK := reflect.Select(rflslcs)
			if index == len(browserItems) {
				systray.Quit()
				loop = false
			} else {
				fmt.Print("Menu Item Clicked = ")
				fmt.Printf("%d, %v, %v \n", index, value, recvOK)

				fmt.Printf("\"%s\" \n\n", browserList.List[index].Name)
			}
		}
	}
}

func systrayOnExit() {
	InnerSigChan <- InnerSigQuit
}
