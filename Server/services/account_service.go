package services

import (
	"cardGame/model/packet"
	postgresRepo "cardGame/repository/postgres"
	"cardGame/utils"
	"context"
)

type AccountService struct {
	AccountRepo *postgresRepo.AccountRepo
}

func NewAccountService(accountRepo *postgresRepo.AccountRepo) *AccountService {
	return &AccountService{
		AccountRepo: accountRepo,
	}
}

func (s *AccountService) CreateAccount(ctx context.Context) {
}

func (s *AccountService) Login(ctx context.Context, data *packet.Login) (bool, error) {
	utils.Log().Infof("login accountId: %s", data.AccountId)
	return true, nil
}
