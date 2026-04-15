package utils

import (
	"cardGame/global"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/viper"
)

func initEnv() {
	// 變數設定 讀取 game.yaml ，若有變更 可以使用環境變數方式修改
	// 若環境變數有設定(全大寫 底線會被置換為 "." ) 以環境變數為主
	// 若沒有設定，以game.yaml 為主，
	// 若 game.yaml 也沒有，程式執行會出錯， game.yaml 就是預設值，無論如何都需要被設定
	vp := viper.GetViper()
	vp.SetConfigFile(".config/game.yaml")
	vp.ReadInConfig() // 讀取config到viper
	vp.AutomaticEnv() // 如果沒有找到設定值，會去找有沒有符合的環境變數
	vp.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	err := vp.Unmarshal(&global.GameConf)
	if err != nil {
		Log().Errorf("unable to decode into config struct, %v", err)
	}
	// prepare serveridx
	if hostname, err := os.Hostname(); err == nil {
		global.GameConf.Game.HostName = hostname
		vals := strings.Split(hostname, "-")
		last := vals[len(vals)-1]
		if idx, err := strconv.Atoi(last); err == nil {
			global.GameConf.Game.Serveridx = uint32(idx)
		}
	}
}
