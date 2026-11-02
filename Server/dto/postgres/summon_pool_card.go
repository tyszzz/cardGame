package postgres

type SummonPoolCard struct {
	PoolId string `gorm:"column:pool_id;primaryKey;size:64"`
	CardId string `gorm:"column:card_id;primaryKey;size:64"`
	IsUp   bool   `gorm:"column:is_up;not null;default:false"`
}
