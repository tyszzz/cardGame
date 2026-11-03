package packet

// account packet
type (
	CreateNewAccount struct {
		AccountId string `json:"accountId"`
		Password  string `json:"password"`
		Nickname  string `json:"nickname"`
	}

	Login struct {
		AccountId string `json:"accountId"`
		Password  string `json:"password"`
	}
)

// summon packet
type (
	SummonCard struct {
		PoolId string `json:"poolId"`
		Times  int32  `json:"times"`
	}
)

// 查詢玩家資料、背包、卡片皆不需要參數，玩家以 session 為準
type Empty struct{}
