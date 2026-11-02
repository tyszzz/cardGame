package postgres

import "time"

type PlayerSummonPity struct {
	UID            string    `gorm:"column:uid;primaryKey"`
	PoolId         string    `gorm:"column:pool_id;primaryKey;size:64"`
	PullsSinceHigh int32     `gorm:"column:pulls_since_high;not null;default:0"`
	GuaranteedUp   bool      `gorm:"column:guaranteed_up;not null;default:false"`
	UpdatedAt      time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (PlayerSummonPity) TableName() string { return "player_summon_pity" }
