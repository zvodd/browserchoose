package browserchoose

import (
	"fmt"
	"sort"
	"strconv"

	"github.com/spf13/viper"
)

type BrowserItem struct {
	Name string
	Icon string
	Path string
	Args []string
}

type BrowserList struct {
	List     []BrowserItem
	Selected int
}

func NewBrowerListFromViper(cfg *viper.Viper) *BrowserList {
	bList := new(BrowserList)

	//Set the selected item
	cfgCurDefault := cfg.GetString("selected")
	tmpCurrentDefault, err := strconv.ParseInt(cfgCurDefault, 10, 64)
	if err != nil {
		panic("invalid str int for config value 'selected'")
	}
	bList.Selected = int(tmpCurrentDefault)

	// Order keys - based on value
	browsersMap := cfg.GetStringMap("browsers")
	browserKeysInt := []int{}
	browserKeysStr := []string{}
	for keyStr := range browsersMap {
		keyNum, err := strconv.ParseInt(keyStr, 10, 64)
		if err != nil {
			panic("Bad config browsers[$KEY] :  $KEY must be a string containing only a positive decimal integer ")
		}
		browserKeysInt = append(browserKeysInt, int(keyNum))
	}
	sort.Ints(browserKeysInt)
	for _, key := range browserKeysInt {
		browserKeysStr = append(browserKeysStr, strconv.Itoa(key))
	}
	// end Order keys

	for _, k := range browserKeysStr {
		v := browsersMap[k]
		bsettings, ok := v.(map[string]interface{})
		if !ok {
			panic(fmt.Sprintf("Bad config browsers[%q] : expected a mapping of settings", k))
		}

		item := BrowserItem{}
		item.Name, _ = bsettings["name"].(string)
		item.Icon, _ = bsettings["icon"].(string)
		item.Path, _ = bsettings["path"].(string)
		if rawArgs, ok := bsettings["args"].([]interface{}); ok {
			for _, a := range rawArgs {
				if s, ok := a.(string); ok {
					item.Args = append(item.Args, s)
				}
			}
		}
		bList.List = append(bList.List, item)
	}

	return bList
}
