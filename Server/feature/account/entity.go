package account

import "time"

type Account struct {
	AccountId    string
	PasswordHash string
	UID          string
	Status       string
	LastLoginAt  *time.Time
}
