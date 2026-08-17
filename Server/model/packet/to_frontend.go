package packet

// base packet
type (
	BaseResult struct {
		ResultCode int32 `json:"resultCode"`
	}
)

// account packet
type (
	CreateAccountResult struct {
		BaseResult
		AccountId string `json:"accountId"`
	}

	LoginResult struct {
		BaseResult
		AccountId string `json:"accountId"`
	}
)

// summon packet
type (
	SummonCardResult struct {
		CardId   string `json:"cardId"`
		CardName string `json:"cardName"`
	}
)
