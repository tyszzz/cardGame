package postgres

type (
	CardStarLevel struct {
		Star     int16 `gorm:"column:star;primaryKey"`
		MaxLevel int   `gorm:"column:max_level;not null"`
	}

	CardRarityLimit struct {
		Rarity  int16 `gorm:"column:rarity;primaryKey"`
		MaxStar int16 `gorm:"column:max_star;not null"`
	}

	CardStage struct {
		Stage            int16   `gorm:"column:stage;primaryKey"`
		MaxLevel         *int    `gorm:"column:max_level"`
		AscendItemId     *string `gorm:"column:ascend_item_id;size:64"`
		AscendItemQty    *int64  `gorm:"column:ascend_item_qty"`
		AscendExpPercent *int    `gorm:"column:ascend_exp_percent"`
	}

	CardStarCost struct {
		Rarity   int16  `gorm:"column:rarity;primaryKey"`
		Star     int16  `gorm:"column:star;primaryKey"`
		ItemId   string `gorm:"column:item_id;primaryKey;size:64"`
		Quantity int64  `gorm:"column:quantity;not null"`
	}
)

// outer function
func (CardStarLevel) TableName() string { return "card_star_level" }

func (CardRarityLimit) TableName() string { return "card_rarity_limit" }

func (CardStage) TableName() string { return "card_stage" }

func (CardStarCost) TableName() string { return "card_star_cost" }
