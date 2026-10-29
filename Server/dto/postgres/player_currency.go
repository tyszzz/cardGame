package postgres

import (
	"cardGame/model/constants"
	"time"
)

type PlayerCurrency struct {
	UID          string                 `gorm:"primaryKey;column:uid"`
	CurrencyType constants.CurrencyType `gorm:"primaryKey;column:currency_type"`
	Amount       int64                  `gorm:"column:amount;not null;default:0"`
	UpdatedAt    time.Time              `gorm:"column:updated_at;autoUpdateTime"`
}
