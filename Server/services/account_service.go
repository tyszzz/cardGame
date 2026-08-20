package services

import (
	"cardGame/feature/account"
	"cardGame/model/packet"
	"cardGame/utils"
	"context"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AccountService struct {
	AccountRepo account.AccountRepository
}

func NewAccountService(accountRepo account.AccountRepository) *AccountService {
	return &AccountService{
		AccountRepo: accountRepo,
	}
}

// inner function
func validateAccountID(accountId string) error {
	if len(accountId) < 3 || len(accountId) > 20 {
		return account.ErrAccountIDLength
	}
	return nil
}

func validatePassword(password string) error {
	if len(password) < 8 || len(password) > 16 {
		return account.ErrPasswordLength
	}
	return nil
}

// outer function
func (s *AccountService) CreateNewAccount(ctx context.Context, data *packet.CreateNewAccount) (bool, error) {
	utils.Log().Infof("create new account, accountId: %s", data.AccountId)
	if err := validateAccountID(data.AccountId); err != nil {
		return false, err
	}
	if err := validatePassword(data.Password); err != nil {
		return false, err
	}

	passwordHash, err := bcrypt.GenerateFromPassword(
		[]byte(data.Password),
		bcrypt.DefaultCost,
	)
	if err != nil {
		return false, err
	}

	account := &account.Account{
		AccountId:    data.AccountId,
		PasswordHash: string(passwordHash),
		UID:          uuid.NewString(),
		Status:       "active",
	}

	if err := s.AccountRepo.CreateNewAccount(ctx, account); err != nil {
		return false, err
	}

	return true, nil
}

func (s *AccountService) Login(ctx context.Context, data *packet.Login) (bool, string, error) {
	utils.Log().Infof("login accountId: %s", data.AccountId)
	accountData, err := s.AccountRepo.GetAccountByAccountId(ctx, data.AccountId)
	if err != nil {
		return false, "", err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(accountData.PasswordHash), []byte(data.Password)); err != nil {
		return false, "", account.ErrInvalidPassword
	}

	return true, accountData.UID, nil
}
