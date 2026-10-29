package services

import (
	"cardGame/feature/summon"
	"cardGame/model/packet"
	"context"
	"sync"
)

type (
	SummonService struct {
		repo summon.SummonRepository
		rnd  summon.Rand

		poolMu sync.Mutex
		pool   summon.CardPool // 啟動後第一次抽卡時載入，卡片異動需重啟 server
	}
)

func NewSummonService(repo summon.SummonRepository) *SummonService {
	return &SummonService{
		repo: repo,
		rnd:  summon.DefaultRand,
	}
}

func (s *SummonService) cardPool(ctx context.Context) (summon.CardPool, error) {
	s.poolMu.Lock()
	defer s.poolMu.Unlock()
	if s.pool != nil {
		return s.pool, nil
	}
	cards, err := s.repo.ListEnabledCards(ctx)
	if err != nil {
		return nil, err
	}
	if len(cards) == 0 {
		return nil, summon.ErrEmptyCardPool
	}
	s.pool = summon.NewCardPool(cards)
	return s.pool, nil
}

// SummonCard 驗證請求、抽卡後在單一 transaction 內扣款並寫入結果。
// 抽卡發生在 transaction 之前，卡池設定錯誤或餘額不足時都不會留下任何資料。
func (s *SummonService) SummonCard(ctx context.Context, uid string, data *packet.SummonCard) (*packet.SummonCardResult, error) {
	rule, err := summon.RuleOf(data.SummonType)
	if err != nil {
		return nil, err
	}
	cost, err := rule.Cost(data.Times)
	if err != nil {
		return nil, err
	}
	pool, err := s.cardPool(ctx)
	if err != nil {
		return nil, err
	}
	drawn, err := summon.Draw(rule, pool, data.Times, s.rnd)
	if err != nil {
		return nil, err
	}

	outcome, err := s.repo.ApplySummon(ctx, uid, rule.Currency, cost, func(owned map[string]struct{}) (summon.Outcome, error) {
		return summon.Resolve(drawn, owned)
	})
	if err != nil {
		return nil, err
	}

	items := make([]packet.SummonCardItem, 0, len(outcome.Results))
	for _, r := range outcome.Results {
		items = append(items, packet.SummonCardItem{
			CardId:      r.Card.CardId,
			CardName:    r.Card.CardName,
			Rarity:      r.Card.Rarity,
			IsNew:       r.IsNew,
			ShardId:     r.Shard.ItemId,
			ShardAmount: r.Shard.Amount,
		})
	}
	return &packet.SummonCardResult{Cards: items}, nil
}
