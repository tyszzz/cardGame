package summon

import "errors"

var (
	ErrPoolUnavailable   = errors.New("summon pool unavailable") // 不存在、未啟用或不在開放時間
	ErrInvalidTimes      = errors.New("invalid summon times")
	ErrNotEnoughCurrency = errors.New("not enough currency")
	ErrEmptyCardPool     = errors.New("card pool is empty")
	ErrInvalidPool       = errors.New("invalid summon pool")
	ErrNoShardReward     = errors.New("no shard reward for rarity")
)
