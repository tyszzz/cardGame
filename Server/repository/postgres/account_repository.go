package postgres

import (
	postgresDto "cardGame/dto/postgres"

	"gorm.io/gorm"
)

type AccountRepo struct {
	Db *gorm.DB
}

func InitAccountRepo(db *gorm.DB) *AccountRepo {
	return &AccountRepo{
		Db: db,
	}
}

func (r *AccountRepo) CreateNewAccount(account *postgresDto.Account) error {
	return r.Db.Create(account).Error
}
