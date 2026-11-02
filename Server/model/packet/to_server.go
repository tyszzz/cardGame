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

// GetCurrencies 不需要參數，玩家以 session 為準
type GetCurrencies struct{}
