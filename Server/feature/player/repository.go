package player

import (
	"context"
	"errors"
)

type PlayerRepository interface {
	// GetPlayer 回傳玩家基本資料與所有貨幣餘額，缺少的貨幣不會出現在 Currencies 中。
	GetPlayer(ctx context.Context, uid string) (*Player, error)
}

var ErrPlayerNotFound = errors.New("player not found")
