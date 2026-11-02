package summon

import (
	"context"

	"cardGame/feature/card"
	"cardGame/model/constants"
)

type SummonRepository interface {
	// ListEnabledCards 取得所有啟用中的卡片。
	ListEnabledCards(ctx context.Context) ([]card.Card, error)
	// ListPools 取得所有卡池定義（含未啟用、不在開放時間的，是否可抽由 Available 判斷）。
	ListPools(ctx context.Context) ([]PoolDefinition, error)
	// ListPoolCards 取得各卡池額外指定的卡片，key 為 pool id。
	ListPoolCards(ctx context.Context) (map[string][]PoolCardRef, error)
	// ApplySummon 在單一 transaction 內鎖定玩家、扣款、讀取已擁有的卡片與保底狀態（trackPity 時）、
	// 以 resolve 抽卡並決定新卡與碎片後寫入；貨幣不足時回傳 ErrNotEnoughCurrency 且不寫入任何資料。
	ApplySummon(ctx context.Context, uid, poolId string, currency constants.CurrencyType, cost int64, trackPity bool, resolve ResolveFunc) (*Outcome, error)
}
