package packet

// account packet
type (
	LoginResult struct {
		ResultCode int32  `json:"resultCode"`
		AccountId  string `json:"accountId"`
		UID        string `json:"uid"`
	}
)

// summon packet
type (
	SummonCardResult struct {
		UID      string `json:"uid"`
		CardId   string `json:"cardId"`
		CardName string `json:"cardName"`
	}
)
