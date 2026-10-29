package account

import "errors"

// database error
var (
	ErrDuplicateKey         = errors.New("account_already_exists")
	ErrNicknameDuplicateKey = errors.New("nickname_already_exists")
)

// account error
var (
	ErrInvalidAccountID = errors.New("invalid account id")
	ErrInvalidPassword  = errors.New("invalid password")
	ErrAccountIDLength  = errors.New("accountId length must be between 3 and 20")
	ErrAccountIDFormat  = errors.New("accountId only allows letters and digits")
	ErrPasswordLength   = errors.New("password length must be between 8 and 16")
	ErrPasswordFormat   = errors.New("password contains unsupported characters")
	ErrNicknameLength   = errors.New("nickname length must be between 2 and 12")
	ErrNicknameFormat   = errors.New("nickname only allows letters, digits and Chinese characters")
	ErrAccountDisabled  = errors.New("account is not active")
)
