package services

import (
	postgresDto "cardGame/dto/postgres"
	"cardGame/model/constants"
	"cardGame/model/packet"
	postgresRepo "cardGame/repository/postgres"
	"cardGame/utils"
	"context"
	"errors"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AccountService struct {
	AccountRepo *postgresRepo.AccountRepo
}

func NewAccountService(accountRepo *postgresRepo.AccountRepo) *AccountService {
	return &AccountService{
		AccountRepo: accountRepo,
	}
}

// inner function
func validateAccountID(accountId string) error {
	if len(accountId) < 3 || len(accountId) > 20 {
		return errors.New(constants.ErrCodeAccountIDLength)
	}
	return nil
}

// outer function
func (s *AccountService) CreateNewAccount(ctx context.Context, data *packet.CreateNewAccount) (bool, error) {
	utils.Log().Infof("create new account, accountId: %s", data.AccountId)
	if err := validateAccountID(data.AccountId); err != nil {
		return false, err
	}

	passwordHash, err := bcrypt.GenerateFromPassword(
		[]byte(data.Password),
		bcrypt.DefaultCost,
	)
	if err != nil {
		return false, err
	}

	account := &postgresDto.Account{
		AccountId:    data.AccountId,
		PasswordHash: string(passwordHash),
		UID:          uuid.NewString(),
		Status:       "active",
	}

	if err := s.AccountRepo.CreateNewAccount(ctx, account); err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return false, errors.New(constants.ErrCodeDuplicateKey)
		}
		return false, err
	}

	return true, nil
}

func (s *AccountService) Login(ctx context.Context, data *packet.Login) (bool, error) {
	utils.Log().Infof("login accountId: %s", data.AccountId)
	return true, nil
}
