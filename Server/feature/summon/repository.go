package summon

import (
	"context"

	"cardGame/feature/card"
	"cardGame/model/constants"
)

type SummonRepository interface {
	// ListEnabledCards 取得所有啟用中的卡片。
	ListEnabledCards(ctx context.Context) ([]card.Card, error)
	// ApplySummon 在單一 transaction 內扣款、讀取玩家已擁有的卡片、
	// 以 resolve 決定新卡與碎片後寫入；貨幣不足時回傳 ErrNotEnoughCurrency 且不寫入任何資料。
	ApplySummon(ctx context.Context, uid string, currency constants.CurrencyType, cost int64, resolve ResolveFunc) (*Outcome, error)
}
