package postgres

import (
	postgresDto "cardGame/dto/postgres"
	"cardGame/feature/account"
	"cardGame/feature/player"
	"context"
	"errors"
	"time"

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
		LastLoginAt:  data.LastLoginAt,
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

func toPlayerDto(data *player.Player) *postgresDto.Player {
	return &postgresDto.Player{
		UID:      data.UID,
		Nickname: data.Nickname,
		Level:    data.Level,
		VipLevel: data.VipLevel,
	}
}

func toPlayerCurrencyDtos(data *player.Player) []postgresDto.PlayerCurrency {
	currencies := make([]postgresDto.PlayerCurrency, 0, len(data.Currencies))
	for currencyType, amount := range data.Currencies {
		currencies = append(currencies, postgresDto.PlayerCurrency{
			UID:          data.UID,
			CurrencyType: currencyType,
			Amount:       amount,
		})
	}
	return currencies
}

// interface implementation
func (r *AccountRepo) CreateNewAccount(ctx context.Context, accountData *account.Account, playerData *player.Player) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(toAccountDto(accountData)).Error; err != nil {
			if errors.Is(err, gorm.ErrDuplicatedKey) {
				return account.ErrDuplicateKey
			}
			return err
		}
		// accounts 已通過唯一性檢查，此處的唯一鍵衝突只可能來自暱稱
		if err := tx.Create(toPlayerDto(playerData)).Error; err != nil {
			if errors.Is(err, gorm.ErrDuplicatedKey) {
				return account.ErrNicknameDuplicateKey
			}
			return err
		}
		currencies := toPlayerCurrencyDtos(playerData)
		if len(currencies) > 0 {
			if err := tx.Create(&currencies).Error; err != nil {
				return err
			}
		}
		return nil
	})
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

func (r *AccountRepo) UpdateLastLogin(ctx context.Context, accountId string, loginAt time.Time) error {
	return r.db.WithContext(ctx).
		Model(&postgresDto.Account{}).
		Where("account_id = ?", accountId).
		Update("last_login_at", loginAt).Error
}
