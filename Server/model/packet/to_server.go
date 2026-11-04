package packet

type (
	// account packet
	CreateNewAccount struct {
		AccountId string `json:"accountId"`
		Password  string `json:"password"`
		Nickname  string `json:"nickname"`
	}

	Login struct {
		AccountId string `json:"accountId"`
		Password  string `json:"password"`
	}

	// summon packet
	SummonCard struct {
		PoolId string `json:"poolId"`
		Times  int32  `json:"times"`
	}

	// 查詢玩家資料、背包、卡片皆不需要參數，玩家以 session 為準
	Empty struct{}

	// AscendCard 為進階指定卡片一階
	AscendCard struct {
		CardId string `json:"cardId"`
	}

	// StarUpCard 為升星指定卡片一星
	StarUpCard struct {
		CardId string `json:"cardId"`
	}

	// UpgradeCard 為升級指定卡片，Levels 為 1 或 10，省略視為 1
	UpgradeCard struct {
		CardId string `json:"cardId"`
		Levels int    `json:"levels,omitempty"`
	}

	// LineupSlot 為陣容中的一格，位置 1~3 為前排、4~6 為後排
	LineupSlot struct {
		Position int    `json:"position"`
		CardId   string `json:"cardId"`
	}

	// SaveLineup 整批覆蓋玩家的陣容
	SaveLineup struct {
		Slots []LineupSlot `json:"slots"`
	}

	// StartBattle 以玩家目前儲存的陣容挑戰指定關卡
	StartBattle struct {
		StageId string `json:"stageId"`
	}
)
