package packet

// account packet
type (
	LoginResult struct {
		ResultCode int32  `json:"resultCode"`
		AccountId  string `json:"accountId"`
	}
)

// summon packet
type (
	SummonCardResult struct {
		CardId   string `json:"cardId"`
		CardName string `json:"cardName"`
	}
)
