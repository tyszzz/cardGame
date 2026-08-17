package constants

import "errors"

// database error codes
var (
	ErrCodeDuplicateKey = errors.New("account_already_exists")
)

// account error codes
var (
	ErrCodeInvalidAccountID = errors.New("invalid account id")
	ErrCodeInvalidPassword  = errors.New("invalid password")
	ErrCodeAccountIDLength  = errors.New("accountId length must be between 3 and 20")
	ErrCodePasswordLength   = errors.New("password length must be between 8 and 16")
)
