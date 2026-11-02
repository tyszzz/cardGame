package postgres

import (
	postgresDto "cardGame/dto/postgres"
	"cardGame/model/constants"
	"context"

	"gorm.io/gorm"
)

type PlayerRepo struct {
	db *gorm.DB
}

func InitPlayerRepo(db *gorm.DB) *PlayerRepo {
	return &PlayerRepo{db: db}
}

func (r *PlayerRepo) GetCurrencies(ctx context.Context, uid string) (map[constants.CurrencyType]int64, error) {
	var rows []postgresDto.PlayerCurrency
	if err := r.db.WithContext(ctx).Where("uid = ?", uid).Find(&rows).Error; err != nil {
		return nil, err
	}
	currencies := make(map[constants.CurrencyType]int64, len(rows))
	for _, row := range rows {
		currencies[row.CurrencyType] = row.Amount
	}
	return currencies, nil
}
