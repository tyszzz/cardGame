package constants

// database error codes
const (
	ErrCodeDuplicateKey = "account_already_exists"
)

// account error codes
const (
	ErrCodeInvalidAccountID = "invalid account id"
	ErrCodeInvalidPassword  = "invalid password"
	ErrCodeAccountIDLength  = "accountId length must be between 3 and 20"
)
