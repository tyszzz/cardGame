package cards

import "cardGame/model/constants"

// CardDefinition 對應 cards table 的一筆卡片定義，也是 seed JSON 的格式。
type CardDefinition struct {
	CardID      string           `json:"cardId"`
	CardName    string           `json:"cardName"`
	Rarity      constants.Rarity `json:"rarity"`
	Race        constants.Race   `json:"race"`
	BaseConfig  BaseConfig       `json:"baseConfig"`
	SkillConfig SkillConfig      `json:"skillConfig"`
	IsEnabled   bool             `json:"isEnabled"`
	IsLimited   bool             `json:"isLimited,omitempty"` // 限定卡不進入常駐卡池
}
