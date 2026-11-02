package postgres

import (
	"cardGame/model/constants"
	"time"
)

type SummonPool struct {
	PoolId               string                 `gorm:"column:pool_id;primaryKey"`
	PoolName             string                 `gorm:"column:pool_name;size:100;not null"`
	CurrencyType         constants.CurrencyType `gorm:"column:currency_type;not null"`
	Costs                []byte                 `gorm:"column:costs;type:jsonb;not null"`
	RarityWeights        []byte                 `gorm:"column:rarity_weights;type:jsonb;not null"`
	TenPullMinRarity     *constants.Rarity      `gorm:"column:ten_pull_min_rarity"`
	UpRates              []byte                 `gorm:"column:up_rates;type:jsonb;not null;default:'{}'"`
	PityThreshold        *int32                 `gorm:"column:pity_threshold"`
	PityMinRarity        *constants.Rarity      `gorm:"column:pity_min_rarity"`
	UpGuarantee          bool                   `gorm:"column:up_guarantee;not null;default:false"`
	IncludeStandardCards bool                   `gorm:"column:include_standard_cards;not null;default:true"`
	StartAt              *time.Time             `gorm:"column:start_at"`
	EndAt                *time.Time             `gorm:"column:end_at"`
	IsEnabled            bool                   `gorm:"column:is_enabled;not null;default:true"`
	CreatedAt            time.Time              `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt            time.Time              `gorm:"column:updated_at;autoUpdateTime"`
}
