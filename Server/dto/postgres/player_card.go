package postgres

import "time"

type PlayerCard struct {
	Id         int64     `gorm:"column:id;primaryKey;autoIncrement"`
	UID        string    `gorm:"column:uid;not null;index:idx_player_cards_uid"`
	CardId     string    `gorm:"column:card_id;size:64;not null"`
	Level      int32     `gorm:"column:level;not null;default:1"`
	IsLocked   bool      `gorm:"column:is_locked;not null;default:false"`
	ObtainedAt time.Time `gorm:"column:obtained_at;autoCreateTime"`
	UpdatedAt  time.Time `gorm:"column:updated_at;autoUpdateTime"`
}
