package packet

// account packet
type (
	LoginResult struct {
		ResultCode int32  `json:"resultCode"`
		AccountId  string `json:"accountId"`
		UID        string `json:"uid"`
	}
)
