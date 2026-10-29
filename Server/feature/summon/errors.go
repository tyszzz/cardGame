package summon

import "errors"

var (
	ErrInvalidSummonType = errors.New("invalid summon type")
	ErrInvalidTimes      = errors.New("invalid summon times")
	ErrNotEnoughCurrency = errors.New("not enough currency")
	ErrEmptyCardPool     = errors.New("card pool is empty")
	ErrInvalidRule       = errors.New("invalid summon rule")
	ErrNoShardReward     = errors.New("no shard reward for rarity")
)
