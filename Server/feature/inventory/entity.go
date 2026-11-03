package inventory

import (
	"cardGame/model/cards"
	"cardGame/model/constants"
	"context"
	"errors"
	"time"
)

type (
	// BackpackItem 為玩家持有的道具，名稱與類型來自道具定義。
	BackpackItem struct {
		ItemId   string
		ItemName string
		ItemType constants.ItemType
		Rarity   constants.Rarity
		Quantity int64
	}

	// OwnedCard 為玩家持有的卡片，名稱、稀有度與種族來自卡片定義。
	OwnedCard struct {
		CardId     string
		CardName   string
		Rarity     constants.Rarity
		Race       constants.Race
		Level      int32
		IsLocked   bool
		ObtainedAt time.Time
		Base       cards.BaseConfig // 升級時據此計算屬性
		Skill      cards.SkillConfig
	}

	// UpgradeResult 為升級後的狀態。
	UpgradeResult struct {
		Level        int32
		Base         cards.BaseConfig
		ExpRemaining int64
	}

	// LevelCostFunc 依卡片目前等級回傳升一級所需的經驗；已達上限時回傳 ErrCardMaxLevel。
	LevelCostFunc func(level int) (int64, error)

	InventoryRepository interface {
		// UpgradeCard 在同一個 transaction 內鎖定卡片、一次扣除升 levels 級所需的經驗道具並升級。
		UpgradeCard(ctx context.Context, uid, cardId, expItemId string, levels int, cost LevelCostFunc) (*UpgradeResult, error)
		ListItems(ctx context.Context, uid string) ([]BackpackItem, error)
		ListCards(ctx context.Context, uid string) ([]OwnedCard, error)
	}
)

var (
	ErrCardNotOwned         = errors.New("card not owned")
	ErrCardMaxLevel         = errors.New("card at max level")
	ErrNotEnoughExp         = errors.New("not enough exp items")
	ErrInvalidUpgradeLevels = errors.New("invalid upgrade levels")
)
