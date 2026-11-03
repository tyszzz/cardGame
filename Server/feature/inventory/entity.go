package inventory

import (
	"cardGame/model/constants"
	"context"
	"time"
)

// BackpackItem 為玩家持有的道具，名稱與類型來自道具定義。
type BackpackItem struct {
	ItemId   string
	ItemName string
	ItemType constants.ItemType
	Rarity   constants.Rarity
	Quantity int64
}

// OwnedCard 為玩家持有的卡片，名稱、稀有度與種族來自卡片定義。
type OwnedCard struct {
	CardId     string
	CardName   string
	Rarity     constants.Rarity
	Race       constants.Race
	Level      int32
	IsLocked   bool
	ObtainedAt time.Time
}

type InventoryRepository interface {
	ListItems(ctx context.Context, uid string) ([]BackpackItem, error)
	ListCards(ctx context.Context, uid string) ([]OwnedCard, error)
}
