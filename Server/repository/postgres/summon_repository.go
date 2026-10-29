package postgres

import (
	postgresDto "cardGame/dto/postgres"
	"cardGame/feature/card"
	"cardGame/feature/summon"
	"cardGame/model/constants"
	"context"
	"sort"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type SummonRepo struct {
	db *gorm.DB
}

func InitSummonRepo(db *gorm.DB) *SummonRepo {
	return &SummonRepo{
		db: db,
	}
}

// interface implementation
func (r *SummonRepo) ListEnabledCards(ctx context.Context) ([]card.Card, error) {
	var rows []postgresDto.Card
	if err := r.db.WithContext(ctx).
		Select("card_id", "card_name", "rarity").
		Where("is_enabled = ?", true).
		Find(&rows).Error; err != nil {
		return nil, err
	}
	cards := make([]card.Card, 0, len(rows))
	for _, row := range rows {
		cards = append(cards, card.Card{
			CardId:   row.CardId,
			CardName: row.CardName,
			Rarity:   row.Rarity,
		})
	}
	return cards, nil
}

func (r *SummonRepo) ApplySummon(
	ctx context.Context,
	uid string,
	currency constants.CurrencyType,
	cost int64,
	resolve summon.ResolveFunc,
) (*summon.Outcome, error) {
	var outcome summon.Outcome
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 先扣款：條件式更新原子地檢查餘額，並鎖住該玩家的貨幣列，
		// 使同一玩家的併發抽卡在此排隊，後續讀取已擁有卡片不會過時。
		deduct := tx.Model(&postgresDto.PlayerCurrency{}).
			Where("uid = ? AND currency_type = ? AND amount >= ?", uid, currency, cost).
			Update("amount", gorm.Expr("amount - ?", cost))
		if deduct.Error != nil {
			return deduct.Error
		}
		if deduct.RowsAffected == 0 {
			return summon.ErrNotEnoughCurrency
		}

		var ownedIds []string
		if err := tx.Model(&postgresDto.PlayerCard{}).
			Where("uid = ?", uid).
			Pluck("card_id", &ownedIds).Error; err != nil {
			return err
		}
		owned := make(map[string]struct{}, len(ownedIds))
		for _, id := range ownedIds {
			owned[id] = struct{}{}
		}

		result, err := resolve(owned)
		if err != nil {
			return err
		}

		if len(result.NewCardIds) > 0 {
			newCards := make([]postgresDto.PlayerCard, 0, len(result.NewCardIds))
			for _, id := range result.NewCardIds {
				newCards = append(newCards, postgresDto.PlayerCard{UID: uid, CardId: id})
			}
			if err := tx.Create(&newCards).Error; err != nil {
				return err
			}
		}

		if len(result.Shards) > 0 {
			// 依 item_id 排序，讓併發交易以相同順序取得鎖
			itemIds := make([]string, 0, len(result.Shards))
			for itemId := range result.Shards {
				itemIds = append(itemIds, itemId)
			}
			sort.Strings(itemIds)
			items := make([]postgresDto.PlayerItem, 0, len(itemIds))
			for _, itemId := range itemIds {
				items = append(items, postgresDto.PlayerItem{UID: uid, ItemId: itemId, Quantity: result.Shards[itemId]})
			}
			if err := tx.Clauses(clause.OnConflict{
				Columns: []clause.Column{{Name: "uid"}, {Name: "item_id"}},
				DoUpdates: clause.Assignments(map[string]interface{}{
					"quantity":   gorm.Expr("player_items.quantity + EXCLUDED.quantity"),
					"updated_at": gorm.Expr("CURRENT_TIMESTAMP"),
				}),
			}).Create(&items).Error; err != nil {
				return err
			}
		}

		outcome = result
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &outcome, nil
}
