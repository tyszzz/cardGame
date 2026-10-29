package account

import (
	"cardGame/feature/player"
	"context"
	"time"
)

type AccountRepository interface {
	CreateNewAccount(ctx context.Context, account *Account, player *player.Player) error
	GetAccountByAccountId(ctx context.Context, accountId string) (*Account, error)
	UpdateLastLogin(ctx context.Context, accountId string, loginAt time.Time) error
}
