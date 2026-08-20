package account

import "errors"

// database error
var (
	ErrDuplicateKey = errors.New("account_already_exists")
)

// account error
var (
	ErrInvalidAccountID = errors.New("invalid account id")
	ErrInvalidPassword  = errors.New("invalid password")
	ErrAccountIDLength  = errors.New("accountId length must be between 3 and 20")
	ErrPasswordLength   = errors.New("password length must be between 8 and 16")
)
