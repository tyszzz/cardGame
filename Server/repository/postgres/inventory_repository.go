package postgres

import (
	"cardGame/feature/inventory"
	cardsModel "cardGame/model/cards"
	"cardGame/model/constants"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type InventoryRepo struct {
	db *gorm.DB
}

func InitInventoryRepo(db *gorm.DB) *InventoryRepo {
	return &InventoryRepo{db: db}
}

// outer function

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
		CardId      string
		CardName    string
		Rarity      constants.Rarity
		Race        constants.Race
		Level       int32
		IsLocked    bool
		ObtainedAt  time.Time
		BaseConfig  []byte
		SkillConfig []byte
	}
	err := r.db.WithContext(ctx).
		Table("player_cards AS pc").
		Select("pc.card_id, c.card_name, c.rarity, c.race, pc.level, pc.is_locked, pc.obtained_at, c.base_config, c.skill_config").
		Joins("JOIN cards c ON c.card_id = pc.card_id").
		Where("pc.uid = ?", uid).
		Order("c.rarity DESC, pc.card_id").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	cards := make([]inventory.OwnedCard, 0, len(rows))
	for _, row := range rows {
		var base cardsModel.BaseConfig
		if err := json.Unmarshal(row.BaseConfig, &base); err != nil {
			return nil, fmt.Errorf("card %s: bad base_config: %w", row.CardId, err)
		}
		var skill cardsModel.SkillConfig
		if err := json.Unmarshal(row.SkillConfig, &skill); err != nil {
			return nil, fmt.Errorf("card %s: bad skill_config: %w", row.CardId, err)
		}
		cards = append(cards, inventory.OwnedCard{
			CardId:     row.CardId,
			CardName:   row.CardName,
			Rarity:     row.Rarity,
			Race:       row.Race,
			Level:      row.Level,
			IsLocked:   row.IsLocked,
			ObtainedAt: row.ObtainedAt,
			Base:       base,
			Skill:      skill,
		})
	}
	return cards, nil
}

func (r *InventoryRepo) UpgradeCard(
	ctx context.Context,
	uid, cardId, expItemId string,
	levels int,
	cost inventory.LevelCostFunc,
) (*inventory.UpgradeResult, error) {
	var result inventory.UpgradeResult
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 鎖定卡片列：同一張卡的並發升級在此排隊，才不會依過期等級計算費用
		var level int32
		lock := tx.Raw("SELECT level FROM player_cards WHERE uid = ? AND card_id = ? FOR UPDATE", uid, cardId).Scan(&level)
		if lock.Error != nil {
			return lock.Error
		}
		if lock.RowsAffected == 0 {
			return inventory.ErrCardNotOwned
		}

		// 逐級累加費用，任何一級超過上限都整筆失敗
		var need int64
		for i := 0; i < levels; i++ {
			next, err := cost(int(level) + i)
			if err != nil {
				return err
			}
			need += next
		}

		// 條件式更新原子地檢查經驗道具數量
		deduct := tx.Exec(
			"UPDATE player_items SET quantity = quantity - ?, updated_at = CURRENT_TIMESTAMP WHERE uid = ? AND item_id = ? AND quantity >= ?",
			need, uid, expItemId, need,
		)
		if deduct.Error != nil {
			return deduct.Error
		}
		if deduct.RowsAffected == 0 {
			return inventory.ErrNotEnoughExp
		}

		if err := tx.Exec(
			"UPDATE player_cards SET level = level + ?, updated_at = CURRENT_TIMESTAMP WHERE uid = ? AND card_id = ?",
			levels, uid, cardId,
		).Error; err != nil {
			return err
		}

		var row struct {
			Level      int32
			BaseConfig []byte
		}
		if err := tx.Raw(
			"SELECT pc.level, c.base_config FROM player_cards pc JOIN cards c ON c.card_id = pc.card_id WHERE pc.uid = ? AND pc.card_id = ?",
			uid, cardId,
		).Scan(&row).Error; err != nil {
			return err
		}
		if err := json.Unmarshal(row.BaseConfig, &result.Base); err != nil {
			return fmt.Errorf("card %s: bad base_config: %w", cardId, err)
		}
		result.Level = row.Level

		return tx.Raw("SELECT quantity FROM player_items WHERE uid = ? AND item_id = ?", uid, expItemId).
			Scan(&result.ExpRemaining).Error
	})
	if err != nil {
		return nil, err
	}
	return &result, nil
}
