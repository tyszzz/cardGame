package postgres

import "time"

type PlayerItem struct {
	UID       string    `gorm:"column:uid;primaryKey"`
	ItemId    string    `gorm:"column:item_id;primaryKey;size:64"`
	Quantity  int64     `gorm:"column:quantity;not null;default:0"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime"`
}
