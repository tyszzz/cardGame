package postgres

import (
	"cardGame/model/cards"
	"cardGame/model/constants"
	"time"
)

type Card struct {
	CardId      string            `gorm:"column:card_id;primaryKey"`
	CardName    string            `gorm:"column:card_name;size:100;not null"`
	Rarity      constants.Rarity  `gorm:"column:rarity;not null"`
	Race        constants.Race    `gorm:"column:race;not null"`
	BaseConfig  cards.BaseConfig  `gorm:"column:base_config;type:jsonb;serializer:json"`
	SkillConfig cards.SkillConfig `gorm:"column:skill_config;type:jsonb;serializer:json"`
	IsEnabled   bool              `gorm:"column:is_enabled;not null;default:true"`
	CreatedAt   time.Time         `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt   time.Time         `gorm:"column:updated_at;autoUpdateTime"`
}
