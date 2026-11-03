package postgres

import (
	postgresDto "cardGame/dto/postgres"
	"cardGame/feature/level"
	"context"

	"gorm.io/gorm"
)

type LevelExpRepo struct {
	db *gorm.DB
}

func InitLevelExpRepo(db *gorm.DB) *LevelExpRepo {
	return &LevelExpRepo{db: db}
}

// outer function

func (r *LevelExpRepo) ListLevelExp(ctx context.Context) ([]level.Row, error) {
	var rows []postgresDto.LevelExp
	if err := r.db.WithContext(ctx).Order("kind, level").Find(&rows).Error; err != nil {
		return nil, err
	}
	result := make([]level.Row, 0, len(rows))
	for _, row := range rows {
		result = append(result, level.Row{Kind: row.Kind, Level: row.Level, ExpRequired: row.ExpRequired})
	}
	return result, nil
}
