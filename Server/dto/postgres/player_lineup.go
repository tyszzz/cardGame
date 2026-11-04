package postgres

import "time"

type PlayerLineup struct {
	UID       string    `gorm:"column:uid;primaryKey"`
	Position  int16     `gorm:"column:position;primaryKey"`
	CardId    string    `gorm:"column:card_id;size:64;not null"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (PlayerLineup) TableName() string { return "player_lineup" }
