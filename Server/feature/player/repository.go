package player

import (
	"cardGame/model/constants"
	"context"
)

type PlayerRepository interface {
	// GetCurrencies 回傳玩家所有貨幣餘額，缺少的貨幣不會出現在結果中。
	GetCurrencies(ctx context.Context, uid string) (map[constants.CurrencyType]int64, error)
}
