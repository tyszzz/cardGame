package postgres

import (
	"cardGame/feature/growth"
	"cardGame/model/constants"
	"context"

	"gorm.io/gorm"
)

type GrowthRepo struct {
	db *gorm.DB
}

func InitGrowthRepo(db *gorm.DB) *GrowthRepo {
	return &GrowthRepo{db: db}
}

// outer function

func (r *GrowthRepo) LoadGrowth(ctx context.Context) (growth.RawConfig, error) {
	db := r.db.WithContext(ctx)
	var raw growth.RawConfig

	var levelExp []struct {
		Kind        string
		Level       int
		ExpRequired int64
	}
	if err := db.Table("level_exp").Order("kind, level").Scan(&levelExp).Error; err != nil {
		return raw, err
	}
	for _, row := range levelExp {
		raw.LevelExp = append(raw.LevelExp, growth.LevelRow{Kind: row.Kind, Level: row.Level, ExpRequired: row.ExpRequired})
	}

	var itemNames []struct {
		ItemId   string
		ItemName string
	}
	if err := db.Table("items").Select("item_id, item_name").Scan(&itemNames).Error; err != nil {
		return raw, err
	}
	names := make(map[string]string, len(itemNames))
	for _, it := range itemNames {
		names[it.ItemId] = it.ItemName
	}

	var stages []struct {
		Stage            int
		MaxLevel         *int
		AscendItemId     *string
		AscendItemQty    *int64
		AscendExpPercent *int
	}
	if err := db.Table("card_stage").Order("stage").Scan(&stages).Error; err != nil {
		return raw, err
	}
	for _, st := range stages {
		row := growth.StageRow{Stage: st.Stage}
		if st.MaxLevel != nil {
			row.MaxLevel = *st.MaxLevel
		}
		if st.AscendItemId != nil {
			row.ItemId = *st.AscendItemId
			row.ItemName = names[row.ItemId]
		}
		if st.AscendItemQty != nil {
			row.ItemQty = *st.AscendItemQty
		}
		if st.AscendExpPercent != nil {
			row.ExpPercent = *st.AscendExpPercent
		}
		raw.Stages = append(raw.Stages, row)
	}

	var starLevels []struct {
		Star     int
		MaxLevel int
	}
	if err := db.Table("card_star_level").Order("star").Scan(&starLevels).Error; err != nil {
		return raw, err
	}
	for _, sl := range starLevels {
		raw.StarLevels = append(raw.StarLevels, growth.StarLevelRow{Star: sl.Star, MaxLevel: sl.MaxLevel})
	}

	var rarities []struct {
		Rarity  constants.Rarity
		MaxStar int
	}
	if err := db.Table("card_rarity_limit").Order("rarity").Scan(&rarities).Error; err != nil {
		return raw, err
	}
	for _, rl := range rarities {
		raw.Rarities = append(raw.Rarities, growth.RarityLimitRow{Rarity: rl.Rarity, MaxStar: rl.MaxStar})
	}

	var costs []struct {
		Rarity   constants.Rarity
		Star     int
		ItemId   string
		Quantity int64
	}
	if err := db.Table("card_star_cost").Order("rarity, star, item_id").Scan(&costs).Error; err != nil {
		return raw, err
	}
	for _, c := range costs {
		raw.StarCosts = append(raw.StarCosts, growth.StarCostRow{
			Rarity: c.Rarity, Star: c.Star, ItemId: c.ItemId, ItemName: names[c.ItemId], Quantity: c.Quantity,
		})
	}
	return raw, nil
}
