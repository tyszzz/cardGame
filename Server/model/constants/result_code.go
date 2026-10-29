package constants

// result code
const (
	Success               = 0
	CreateAccountFailed   = 1
	LoginFailed           = 2
	AccountAlreadyExists  = 3
	NicknameAlreadyExists = 4

	NotLoggedIn             = 5
	SummonInvalidRequest    = 6 // 抽卡類型或次數不合法
	SummonNotEnoughCurrency = 7
	SummonFailed            = 8 // 伺服器端錯誤（卡池設定、資料庫等）
)
