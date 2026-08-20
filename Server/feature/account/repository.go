package account

import "context"

type AccountRepository interface {
	CreateNewAccount(ctx context.Context, account *Account) error
	GetAccountByAccountId(ctx context.Context, accountId string) (*Account, error)
}
