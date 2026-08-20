package postgres

import (
	postgresDto "cardGame/dto/postgres"
	"cardGame/feature/account"
	"context"
	"errors"

	"gorm.io/gorm"
)

type AccountRepo struct {
	db *gorm.DB
}

func InitAccountRepo(db *gorm.DB) *AccountRepo {
	return &AccountRepo{
		db: db,
	}
}

// inner function
func toAccountEntity(data *postgresDto.Account) *account.Account {
	return &account.Account{
		AccountId:    data.AccountId,
		PasswordHash: data.PasswordHash,
		UID:          data.UID,
		Status:       data.Status,
	}
}

func toAccountDto(data *account.Account) *postgresDto.Account {
	return &postgresDto.Account{
		AccountId:    data.AccountId,
		PasswordHash: data.PasswordHash,
		UID:          data.UID,
		Status:       data.Status,
	}
}

// interface implementation
func (r *AccountRepo) CreateNewAccount(ctx context.Context, accountData *account.Account) error {
	if err := r.db.WithContext(ctx).Create(toAccountDto(accountData)).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return account.ErrDuplicateKey
		}
		return err
	}
	return nil
}

func (r *AccountRepo) GetAccountByAccountId(ctx context.Context, accountId string) (*account.Account, error) {
	var accountData postgresDto.Account
	if err := r.db.WithContext(ctx).Where("account_id = ?", accountId).First(&accountData).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, account.ErrInvalidAccountID
		}
		return nil, err
	}
	return toAccountEntity(&accountData), nil
}
