package card

import "cardGame/model/constants"

// Card 為抽卡等流程需要的卡片資料，不含技能設定。
type Card struct {
	CardId    string
	CardName  string
	Rarity    constants.Rarity
	IsLimited bool // 限定卡不進入常駐卡池
}
