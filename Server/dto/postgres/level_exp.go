package postgres

type LevelExp struct {
	Kind        string `gorm:"column:kind;primaryKey;size:16"`
	Level       int    `gorm:"column:level;primaryKey"`
	ExpRequired int64  `gorm:"column:exp_required;not null"`
}

func (LevelExp) TableName() string { return "level_exp" }
