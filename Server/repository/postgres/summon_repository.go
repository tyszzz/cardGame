package postgres

import (
	postgresDto "cardGame/dto/postgres"
	"cardGame/feature/card"
	"cardGame/feature/summon"
	"cardGame/model/constants"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"

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

// inner function
func parseRarity(name string) (constants.Rarity, error) {
	var r constants.Rarity
	if err := json.Unmarshal([]byte(strconv.Quote(name)), &r); err != nil {
		return 0, err
	}
	return r, nil
}

func parseRarityMap(raw []byte) (map[constants.Rarity]int, error) {
	var byName map[string]int
	if err := json.Unmarshal(raw, &byName); err != nil {
		return nil, err
	}
	result := make(map[constants.Rarity]int, len(byName))
	for name, v := range byName {
		r, err := parseRarity(name)
		if err != nil {
			return nil, err
		}
		result[r] = v
	}
	return result, nil
}

func toPoolDefinition(row *postgresDto.SummonPool) (summon.PoolDefinition, error) {
	var rawCosts map[string]int64
	if err := json.Unmarshal(row.Costs, &rawCosts); err != nil {
		return summon.PoolDefinition{}, fmt.Errorf("costs: %w", err)
	}
	costs := make(map[int32]int64, len(rawCosts))
	for k, v := range rawCosts {
		times, err := strconv.ParseInt(k, 10, 32)
		if err != nil {
			return summon.PoolDefinition{}, fmt.Errorf("costs key %q: %w", k, err)
		}
		costs[int32(times)] = v
	}

	weightMap, err := parseRarityMap(row.RarityWeights)
	if err != nil {
		return summon.PoolDefinition{}, fmt.Errorf("rarity_weights: %w", err)
	}
	weights := make([]summon.RarityWeight, 0, len(weightMap))
	for r, w := range weightMap {
		weights = append(weights, summon.RarityWeight{Rarity: r, Weight: w})
	}
	sort.Slice(weights, func(i, j int) bool { return weights[i].Rarity < weights[j].Rarity })

	upRates, err := parseRarityMap(row.UpRates)
	if err != nil {
		return summon.PoolDefinition{}, fmt.Errorf("up_rates: %w", err)
	}

	def := summon.PoolDefinition{
		PoolId:               row.PoolId,
		PoolName:             row.PoolName,
		Currency:             row.CurrencyType,
		Costs:                costs,
		Weights:              weights,
		UpRates:              upRates,
		IncludeStandardCards: row.IncludeStandardCards,
		StartAt:              row.StartAt,
		EndAt:                row.EndAt,
		IsEnabled:            row.IsEnabled,
	}
	if row.TenPullMinRarity != nil {
		def.TenPullMinRarity = *row.TenPullMinRarity
	}
	if row.PityThreshold != nil && row.PityMinRarity != nil {
		def.Pity = &summon.PityRule{
			Threshold:   int(*row.PityThreshold),
			MinRarity:   *row.PityMinRarity,
			UpGuarantee: row.UpGuarantee,
		}
	}
	return def, nil
}

// interface implementation
func (r *SummonRepo) ListEnabledCards(ctx context.Context) ([]card.Card, error) {
	var rows []postgresDto.Card
	if err := r.db.WithContext(ctx).
		Select("card_id", "card_name", "rarity", "is_limited").
		Where("is_enabled = ?", true).
		Find(&rows).Error; err != nil {
		return nil, err
	}
	cards := make([]card.Card, 0, len(rows))
	for _, row := range rows {
		cards = append(cards, card.Card{
			CardId:    row.CardId,
			CardName:  row.CardName,
			Rarity:    row.Rarity,
			IsLimited: row.IsLimited,
		})
	}
	return cards, nil
}

// ListPools 解析失敗的卡池會被略過，錯誤以 errors.Join 回傳，其餘卡池仍會回傳。
func (r *SummonRepo) ListPools(ctx context.Context) ([]summon.PoolDefinition, error) {
	var rows []postgresDto.SummonPool
	if err := r.db.WithContext(ctx).Order("pool_id").Find(&rows).Error; err != nil {
		return nil, err
	}
	defs := make([]summon.PoolDefinition, 0, len(rows))
	var parseErrs []error
	for i := range rows {
		def, err := toPoolDefinition(&rows[i])
		if err != nil {
			parseErrs = append(parseErrs, fmt.Errorf("pool %q: %w", rows[i].PoolId, err))
			continue
		}
		defs = append(defs, def)
	}
	return defs, errors.Join(parseErrs...)
}

func (r *SummonRepo) ListPoolCards(ctx context.Context) (map[string][]summon.PoolCardRef, error) {
	var rows []postgresDto.SummonPoolCard
	if err := r.db.WithContext(ctx).Order("pool_id, card_id").Find(&rows).Error; err != nil {
		return nil, err
	}
	refs := make(map[string][]summon.PoolCardRef)
	for _, row := range rows {
		refs[row.PoolId] = append(refs[row.PoolId], summon.PoolCardRef{CardId: row.CardId, IsUp: row.IsUp})
	}
	return refs, nil
}

func (r *SummonRepo) ApplySummon(
	ctx context.Context,
	uid string,
	poolId string,
	currency constants.CurrencyType,
	cost int64,
	trackPity bool,
	resolve summon.ResolveFunc,
) (*summon.Outcome, error) {
	var outcome summon.Outcome
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 先鎖住玩家：同一玩家的所有抽卡（不論貨幣或卡池）在此排隊，
		// 之後讀取的已擁有卡片與保底狀態才不會過時。
		var locked int
		lock := tx.Raw("SELECT 1 FROM players WHERE uid = ? FOR UPDATE", uid).Scan(&locked)
		if lock.Error != nil {
			return lock.Error
		}
		if lock.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}

		// 條件式更新原子地檢查餘額
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

		var pity summon.PityState
		if trackPity {
			var row postgresDto.PlayerSummonPity
			err := tx.Where("uid = ? AND pool_id = ?", uid, poolId).Take(&row).Error
			switch {
			case err == nil:
				pity = summon.PityState{PullsSinceHigh: int(row.PullsSinceHigh), GuaranteedUp: row.GuaranteedUp}
			case !errors.Is(err, gorm.ErrRecordNotFound):
				return err
			}
		}

		result, err := resolve(owned, pity)
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

		if trackPity {
			if err := tx.Clauses(clause.OnConflict{
				Columns: []clause.Column{{Name: "uid"}, {Name: "pool_id"}},
				DoUpdates: clause.Assignments(map[string]interface{}{
					"pulls_since_high": result.Pity.PullsSinceHigh,
					"guaranteed_up":    result.Pity.GuaranteedUp,
					"updated_at":       gorm.Expr("CURRENT_TIMESTAMP"),
				}),
			}).Create(&postgresDto.PlayerSummonPity{
				UID:            uid,
				PoolId:         poolId,
				PullsSinceHigh: int32(result.Pity.PullsSinceHigh),
				GuaranteedUp:   result.Pity.GuaranteedUp,
			}).Error; err != nil {
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
