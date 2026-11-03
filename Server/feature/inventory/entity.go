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
		Star       int
		Stage      int
		IsLocked   bool
		ObtainedAt time.Time
		Base       cards.BaseConfig // 升級時據此計算屬性
		Skill      cards.SkillConfig
	}

	// CardState 為卡片目前的進度，供 CardChangePlan 判斷與計算費用。
	CardState struct {
		Rarity constants.Rarity
		Level  int
		Star   int
		Stage  int
	}

	// ItemCost 為一項道具消耗。
	ItemCost struct {
		ItemId   string
		Quantity int64
	}

	// CardChange 為一次進度變更：等級、階、星數各增加多少，以及需要扣除的道具。
	CardChange struct {
		Levels int
		Stages int
		Stars  int
		Costs  []ItemCost
	}

	// CardChangePlan 在卡片被鎖定後依最新狀態決定這次變更與費用；回傳 error 會讓整筆交易回滾。
	CardChangePlan func(state CardState) (CardChange, error)

	// ChangeResult 為變更後的狀態。
	ChangeResult struct {
		State     CardState
		Base      cards.BaseConfig
		Remaining map[string]int64 // 已扣除道具的剩餘數量
	}

	InventoryRepository interface {
		// ApplyCardChange 在同一個 transaction 內鎖定卡片、由 plan 決定變更與費用、
		// 條件式扣除道具並更新等級、階與星數；任何一步失敗整筆回滾。
		ApplyCardChange(ctx context.Context, uid, cardId string, plan CardChangePlan) (*ChangeResult, error)
		ListItems(ctx context.Context, uid string) ([]BackpackItem, error)
		ListCards(ctx context.Context, uid string) ([]OwnedCard, error)
	}
)

var (
	ErrCardNotOwned         = errors.New("card not owned")
	ErrCardMaxLevel         = errors.New("card at max level")
	ErrCardNeedAscend       = errors.New("card needs ascend to level up")
	ErrCardNeedStar         = errors.New("card needs star up to level up")
	ErrAscendNotReady       = errors.New("card is below the stage max level")
	ErrAlreadyTopStage      = errors.New("card already at top stage")
	ErrStarMax              = errors.New("card already at max star")
	ErrNotEnoughItems       = errors.New("not enough items")
	ErrInvalidUpgradeLevels = errors.New("invalid upgrade levels")
)
