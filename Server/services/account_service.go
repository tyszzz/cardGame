package services

import (
	"cardGame/feature/account"
	"cardGame/feature/player"
	"cardGame/model/constants"
	"cardGame/model/packet"
	"cardGame/utils"
	"context"
	"regexp"
	"time"
	"unicode/utf8"

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

var (
	accountIDPattern = regexp.MustCompile(`^[A-Za-z0-9]+$`)
	nicknamePattern  = regexp.MustCompile(`^[A-Za-z0-9\p{Han}]+$`)
	passwordPattern  = regexp.MustCompile(`^[A-Za-z0-9@#$%^&*()_+=!?.\-,]+$`)
)

// inner function
func validateAccountID(accountId string) error {
	if len(accountId) < 3 || len(accountId) > 20 {
		return account.ErrAccountIDLength
	}
	if !accountIDPattern.MatchString(accountId) {
		return account.ErrAccountIDFormat
	}
	return nil
}

func validateNickname(nickname string) error {
	if n := utf8.RuneCountInString(nickname); n < 2 || n > 12 {
		return account.ErrNicknameLength
	}
	if !nicknamePattern.MatchString(nickname) {
		return account.ErrNicknameFormat
	}
	return nil
}

func newPlayer(uid, nickname string) *player.Player {
	currencies := make(map[constants.CurrencyType]int64, len(constants.AllCurrencyTypes))
	for _, currencyType := range constants.AllCurrencyTypes {
		currencies[currencyType] = 0
	}
	return &player.Player{
		UID:        uid,
		Nickname:   nickname,
		Level:      1,
		VipLevel:   0,
		Currencies: currencies,
	}
}

func validatePassword(password string) error {
	if len(password) < 8 || len(password) > 16 {
		return account.ErrPasswordLength
	}
	if !passwordPattern.MatchString(password) {
		return account.ErrPasswordFormat
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

	if err := validateNickname(data.Nickname); err != nil {
		return false, err
	}

	passwordHash, err := bcrypt.GenerateFromPassword(
		[]byte(data.Password),
		bcrypt.DefaultCost,
	)
	if err != nil {
		return false, err
	}

	uid := uuid.NewString()
	account := &account.Account{
		AccountId:    data.AccountId,
		PasswordHash: string(passwordHash),
		UID:          uid,
		Status:       "active",
	}

	if err := s.AccountRepo.CreateNewAccount(ctx, account, newPlayer(uid, data.Nickname)); err != nil {
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

	if accountData.Status != "active" {
		return false, "", account.ErrAccountDisabled
	}

	// last_login_at 為統計資料，更新失敗不阻擋登入
	if err := s.AccountRepo.UpdateLastLogin(ctx, accountData.AccountId, time.Now()); err != nil {
		utils.Log().Warnf("update last login failed, accountId: %s, err: %v", data.AccountId, err)
	}

	return true, accountData.UID, nil
}
