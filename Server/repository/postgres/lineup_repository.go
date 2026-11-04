package postgres

import (
	postgresDto "cardGame/dto/postgres"
	"cardGame/feature/battle"
	"context"

	"gorm.io/gorm"
)

type LineupRepo struct {
	db *gorm.DB
}

func InitLineupRepo(db *gorm.DB) *LineupRepo {
	return &LineupRepo{db: db}
}

// outer function

func (r *LineupRepo) GetLineup(ctx context.Context, uid string) ([]battle.LineupSlot, error) {
	var rows []postgresDto.PlayerLineup
	if err := r.db.WithContext(ctx).Where("uid = ?", uid).Order("position").Find(&rows).Error; err != nil {
		return nil, err
	}
	slots := make([]battle.LineupSlot, 0, len(rows))
	for _, row := range rows {
		slots = append(slots, battle.LineupSlot{Position: int(row.Position), CardId: row.CardId})
	}
	return slots, nil
}

func (r *LineupRepo) SaveLineup(ctx context.Context, uid string, slots []battle.LineupSlot) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("uid = ?", uid).Delete(&postgresDto.PlayerLineup{}).Error; err != nil {
			return err
		}
		if len(slots) == 0 {
			return nil
		}
		rows := make([]postgresDto.PlayerLineup, 0, len(slots))
		for _, s := range slots {
			rows = append(rows, postgresDto.PlayerLineup{UID: uid, Position: int16(s.Position), CardId: s.CardId})
		}
		return tx.Create(&rows).Error
	})
}
