package postgres

import "time"

type Player struct {
	UID       string    `gorm:"primaryKey;column:uid"`
	Nickname  string    `gorm:"column:nickname;size:32;not null"`
	Level     int32     `gorm:"column:level;not null;default:1"`
	VipLevel  int32     `gorm:"column:vip_level;not null;default:0"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime"`
}
