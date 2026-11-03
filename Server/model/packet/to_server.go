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

	// UpgradeCard 為升級指定卡片，Levels 為 1 或 10，省略視為 1
	UpgradeCard struct {
		CardId string `json:"cardId"`
		Levels int    `json:"levels,omitempty"`
	}
)
