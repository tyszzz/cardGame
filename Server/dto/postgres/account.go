package postgres

import "time"

type Account struct {
	AccountId    string     `gorm:"primaryKey;column:account_id"`
	PasswordHash string     `gorm:"column:password_hash;not null"`
	UID          string     `gorm:"column:uid;uniqueIndex;not null"`
	Status       string     `gorm:"column:status;size:20;not null;default:active"`
	CreatedAt    time.Time  `gorm:"column:created_at;autoCreateTime"`
	LastLoginAt  *time.Time `gorm:"column:last_login_at"`
	UpdatedAt    time.Time  `gorm:"column:updated_at;autoUpdateTime"`
}
