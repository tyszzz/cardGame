package postgres

import (
	postgresDto "cardGame/dto/postgres"
	"context"

	"gorm.io/gorm"
)

type AccountRepo struct {
	Db *gorm.DB
}

func InitAccountRepo(db *gorm.DB) *AccountRepo {
	return &AccountRepo{
		Db: db,
	}
}

func (r *AccountRepo) CreateNewAccount(ctx context.Context, account *postgresDto.Account) error {
	return r.Db.WithContext(ctx).Create(account).Error
}

func (r *AccountRepo) GetAccountByAccountId(ctx context.Context, accountId string) (*postgresDto.Account, error) {
	var account postgresDto.Account
	if err := r.Db.WithContext(ctx).Where("account_id = ?", accountId).First(&account).Error; err != nil {
		return nil, err
	}
	return &account, nil
}
