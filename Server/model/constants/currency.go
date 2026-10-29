package constants

type CurrencyType int16

const (
	CurrencyCoin        CurrencyType = 1
	CurrencyDiamond     CurrencyType = 2
	CurrencyFriendPoint CurrencyType = 3
)

// AllCurrencyTypes 列出所有貨幣，建立玩家時據此產生初始餘額。
var AllCurrencyTypes = []CurrencyType{
	CurrencyCoin,
	CurrencyDiamond,
	CurrencyFriendPoint,
}
