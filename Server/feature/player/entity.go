package player

import "cardGame/model/constants"

type Player struct {
	UID        string
	Nickname   string
	Level      int32
	VipLevel   int32
	Currencies map[constants.CurrencyType]int64
}
