package services

import (
	"cardGame/model/packet"
	"cardGame/utils"
	"context"
)

type AccountService struct {
}

func NewAccountService() *AccountService {
	return &AccountService{}
}

func (s *AccountService) CreateAccount(ctx context.Context) {
}

func (s *AccountService) Login(ctx context.Context, data *packet.Login) (bool, error) {
	utils.Log().Infof("login accountId: %s, password: %s", data.AccountId, data.Password)
	return true, nil
}
