package postgres

import (
	"cardGame/feature/inventory"
	cardsModel "cardGame/model/cards"
	"cardGame/model/constants"
	"context"
	"encoding/json"
	"fmt"
	"sort"
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
		Star        int
		Stage       int
		IsLocked    bool
		ObtainedAt  time.Time
		BaseConfig  []byte
		SkillConfig []byte
	}
	err := r.db.WithContext(ctx).
		Table("player_cards AS pc").
		Select("pc.card_id, c.card_name, c.rarity, c.race, pc.level, pc.star, pc.stage, pc.is_locked, pc.obtained_at, c.base_config, c.skill_config").
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
			Star:       row.Star,
			Stage:      row.Stage,
			IsLocked:   row.IsLocked,
			ObtainedAt: row.ObtainedAt,
			Base:       base,
			Skill:      skill,
		})
	}
	return cards, nil
}

func (r *InventoryRepo) ApplyCardChange(
	ctx context.Context,
	uid, cardId string,
	plan inventory.CardChangePlan,
) (*inventory.ChangeResult, error) {
	var result inventory.ChangeResult
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 鎖定卡片列：同一張卡的並發變更在此排隊，才不會依過期的等級、階或星數計算
		var row struct {
			Rarity     constants.Rarity
			Level      int
			Star       int
			Stage      int
			BaseConfig []byte
		}
		lock := tx.Raw(
			"SELECT c.rarity, pc.level, pc.star, pc.stage, c.base_config FROM player_cards pc "+
				"JOIN cards c ON c.card_id = pc.card_id WHERE pc.uid = ? AND pc.card_id = ? FOR UPDATE OF pc",
			uid, cardId,
		).Scan(&row)
		if lock.Error != nil {
			return lock.Error
		}
		if lock.RowsAffected == 0 {
			return inventory.ErrCardNotOwned
		}
		state := inventory.CardState{Rarity: row.Rarity, Level: row.Level, Star: row.Star, Stage: row.Stage}

		change, err := plan(state)
		if err != nil {
			return err
		}

		// 同一個道具合併後依 item_id 排序扣除，鎖定順序固定以避免死結
		merged := make(map[string]int64, len(change.Costs))
		for _, c := range change.Costs {
			merged[c.ItemId] += c.Quantity
		}
		itemIds := make([]string, 0, len(merged))
		for id := range merged {
			itemIds = append(itemIds, id)
		}
		sort.Strings(itemIds)
		for _, id := range itemIds {
			need := merged[id]
			if need <= 0 {
				continue
			}
			// 條件式更新原子地檢查道具數量
			deduct := tx.Exec(
				"UPDATE player_items SET quantity = quantity - ?, updated_at = CURRENT_TIMESTAMP WHERE uid = ? AND item_id = ? AND quantity >= ?",
				need, uid, id, need,
			)
			if deduct.Error != nil {
				return deduct.Error
			}
			if deduct.RowsAffected == 0 {
				return inventory.ErrNotEnoughItems
			}
		}

		if err := tx.Exec(
			"UPDATE player_cards SET level = level + ?, stage = stage + ?, star = star + ?, updated_at = CURRENT_TIMESTAMP WHERE uid = ? AND card_id = ?",
			change.Levels, change.Stages, change.Stars, uid, cardId,
		).Error; err != nil {
			return err
		}

		if err := json.Unmarshal(row.BaseConfig, &result.Base); err != nil {
			return fmt.Errorf("card %s: bad base_config: %w", cardId, err)
		}
		result.State = inventory.CardState{
			Rarity: row.Rarity,
			Level:  row.Level + change.Levels,
			Star:   row.Star + change.Stars,
			Stage:  row.Stage + change.Stages,
		}
		result.Remaining = make(map[string]int64, len(itemIds))
		for _, id := range itemIds {
			var quantity int64
			if err := tx.Raw("SELECT quantity FROM player_items WHERE uid = ? AND item_id = ?", uid, id).Scan(&quantity).Error; err != nil {
				return err
			}
			result.Remaining[id] = quantity
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &result, nil
}
