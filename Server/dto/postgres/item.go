package postgres

import (
	"cardGame/model/constants"
	"time"
)

type Item struct {
	ItemId    string             `gorm:"column:item_id;primaryKey"`
	ItemName  string             `gorm:"column:item_name;size:100;not null"`
	ItemType  constants.ItemType `gorm:"column:item_type;not null"`
	Rarity    constants.Rarity   `gorm:"column:rarity;not null"`
	Config    []byte             `gorm:"column:config;type:jsonb;not null;default:'{}'"`
	IsEnabled bool               `gorm:"column:is_enabled;not null;default:true"`
	CreatedAt time.Time          `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt time.Time          `gorm:"column:updated_at;autoUpdateTime"`
}
