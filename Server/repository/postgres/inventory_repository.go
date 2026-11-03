package postgres

import (
	"cardGame/feature/inventory"
	"cardGame/model/constants"
	"context"
	"time"

	"gorm.io/gorm"
)

type InventoryRepo struct {
	db *gorm.DB
}

func InitInventoryRepo(db *gorm.DB) *InventoryRepo {
	return &InventoryRepo{db: db}
}

func (r *InventoryRepo) ListItems(ctx context.Context, uid string) ([]inventory.BackpackItem, error) {
	var rows []struct {
		ItemId   string
		ItemName string
		ItemType constants.ItemType
		Rarity   constants.Rarity
		Quantity int64
	}
	err := r.db.WithContext(ctx).
		Table("player_items AS pi").
		Select("pi.item_id, i.item_name, i.item_type, i.rarity, pi.quantity").
		Joins("JOIN items i ON i.item_id = pi.item_id").
		Where("pi.uid = ? AND pi.quantity > 0", uid).
		Order("pi.item_id").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	items := make([]inventory.BackpackItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, inventory.BackpackItem{
			ItemId:   row.ItemId,
			ItemName: row.ItemName,
			ItemType: row.ItemType,
			Rarity:   row.Rarity,
			Quantity: row.Quantity,
		})
	}
	return items, nil
}

func (r *InventoryRepo) ListCards(ctx context.Context, uid string) ([]inventory.OwnedCard, error) {
	var rows []struct {
		CardId     string
		CardName   string
		Rarity     constants.Rarity
		Race       constants.Race
		Level      int32
		IsLocked   bool
		ObtainedAt time.Time
	}
	err := r.db.WithContext(ctx).
		Table("player_cards AS pc").
		Select("pc.card_id, c.card_name, c.rarity, c.race, pc.level, pc.is_locked, pc.obtained_at").
		Joins("JOIN cards c ON c.card_id = pc.card_id").
		Where("pc.uid = ?", uid).
		Order("c.rarity DESC, pc.card_id").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	cards := make([]inventory.OwnedCard, 0, len(rows))
	for _, row := range rows {
		cards = append(cards, inventory.OwnedCard{
			CardId:     row.CardId,
			CardName:   row.CardName,
			Rarity:     row.Rarity,
			Race:       row.Race,
			Level:      row.Level,
			IsLocked:   row.IsLocked,
			ObtainedAt: row.ObtainedAt,
		})
	}
	return cards, nil
}
