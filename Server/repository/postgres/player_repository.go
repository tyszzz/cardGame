package postgres

import (
	postgresDto "cardGame/dto/postgres"
	"cardGame/feature/player"
	"cardGame/model/constants"
	"context"
	"errors"

	"gorm.io/gorm"
)

type PlayerRepo struct {
	db *gorm.DB
}

func InitPlayerRepo(db *gorm.DB) *PlayerRepo {
	return &PlayerRepo{db: db}
}

// outer function

func (r *PlayerRepo) GetPlayer(ctx context.Context, uid string) (*player.Player, error) {
	db := r.db.WithContext(ctx)

	var row postgresDto.Player
	if err := db.Where("uid = ?", uid).Take(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, player.ErrPlayerNotFound
		}
		return nil, err
	}

	var currencyRows []postgresDto.PlayerCurrency
	if err := db.Where("uid = ?", uid).Find(&currencyRows).Error; err != nil {
		return nil, err
	}
	currencies := make(map[constants.CurrencyType]int64, len(currencyRows))
	for _, c := range currencyRows {
		currencies[c.CurrencyType] = c.Amount
	}
	return &player.Player{
		UID:        row.UID,
		Nickname:   row.Nickname,
		Level:      row.Level,
		VipLevel:   row.VipLevel,
		Currencies: currencies,
	}, nil
}
