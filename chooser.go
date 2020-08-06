package main

import (
	"fmt"
	"sort"
	"strconv"

	"example.com/icon"
	"github.com/getlantern/systray"
	"github.com/spf13/viper"
)

func configure() *viper.Viper {
	cfg := viper.New()
	cfg.SetConfigName("config") // name of config file (without extension)
	cfg.SetConfigType("yaml")   // REQUIRED if the config file does not have the extension in the name
	cfg.AddConfigPath("$APPDATA/xyz.browserchooser")
	cfg.AddConfigPath(".")    // optionally look for config in the working directory
	err := cfg.ReadInConfig() // Find and read the config file
	if err != nil {           // Handle errors reading the config file
		panic(fmt.Errorf("Fatal error config file: %s \n ", err))
	}
	return cfg
}

func main() {
	cfg := configure()
	// x := cfg.Sub("browsers") //.GetStringSlice("args")
	// fmt.Printf("%+v\n\n\n", x)
	// y := cfg.GetStringMap("browsers")
	// fmt.Printf("%+v\n", y)
	systray.Run(onReadyFactory(cfg), onExit)
}

var (
	browserItems = map[string]*systray.MenuItem{}
)

func onReadyFactory(cfg *viper.Viper) func() {
	return func() {
		systray.SetIcon(icon.Data)
		systray.SetTitle("Awesome App")
		systray.SetTooltip(cfg.GetString("unused"))

		// Order keys - based on value
		// TODO refactor into configure()
		browserKeysInt := []int{}
		browserKeysStr := []string{}
		for k := range cfg.GetStringMap("browsers") {
			intOfK, err := strconv.ParseInt(k, 10, 64)
			if err != nil {
				panic("Bad config browsers[$KEY] :  $KEY must be string form of integer number")
			}
			browserKeysInt = append(browserKeysInt, int(intOfK))
		}
		sort.Ints(browserKeysInt)

		for _, key := range browserKeysInt {
			browserKeysStr = append(browserKeysStr, strconv.Itoa(key))
		}
		// end Order keys

		currentDefault := cfg.GetString("currentDefault")
		browsersMap := cfg.GetStringMap("browsers")
		for _, k := range browserKeysStr {
			v := browsersMap[k]
			bsettings := v.(map[string]interface{})
			name := bsettings["name"]
			mi := systray.AddMenuItem(name.(string), fmt.Sprintf("Set browser to %s", k))
			browserItems[k] = mi

			if k == currentDefault {
				mi.Check()
			}
		}

		systray.AddSeparator()
		mQuit := systray.AddMenuItem("Quit", "Quit BrowserChooser")
		// Sets the icon of a menu item. Only available on Mac and Windows.
		mQuit.SetIcon(icon.Data)

		for {
			select {
			case <-mQuit.ClickedCh:
				systray.Quit()
			}
		}
	}
}

func onExit() {
	// clean up here

}
