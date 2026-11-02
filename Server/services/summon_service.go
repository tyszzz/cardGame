package services

import (
	"cardGame/feature/summon"
	"cardGame/model/packet"
	"cardGame/utils"
	"context"
	"sync"
	"time"
)

const (
	poolCacheTTL        = 5 * time.Minute  // 卡池資料定時重新載入的間隔
	poolReloadRetryWait = 30 * time.Second // 重新載入失敗並沿用舊資料時，下次重試的等待時間
)

type (
	SummonService struct {
		repo summon.SummonRepository
		rnd  summon.Rand
		now  func() time.Time
		warn func(format string, args ...any)

		poolMu   sync.Mutex
		pools    map[string]*summon.Pool
		loadedAt time.Time
	}
)

func NewSummonService(repo summon.SummonRepository) *SummonService {
	return &SummonService{
		repo: repo,
		rnd:  summon.DefaultRand,
		now:  time.Now,
		warn: func(format string, args ...any) { utils.Log().Warnf(format, args...) },
	}
}

// loadPools 載入所有卡池並組裝；定義不合法的卡池會被略過並記錄警告，不影響其他卡池。
func (s *SummonService) loadPools(ctx context.Context) (map[string]*summon.Pool, error) {
	cards, err := s.repo.ListEnabledCards(ctx)
	if err != nil {
		return nil, err
	}
	defs, err := s.repo.ListPools(ctx)
	if err != nil {
		if len(defs) == 0 {
			return nil, err
		}
		s.warn("some summon pools skipped: %v", err)
	}
	refs, err := s.repo.ListPoolCards(ctx)
	if err != nil {
		return nil, err
	}

	pools := make(map[string]*summon.Pool, len(defs))
	for _, def := range defs {
		pool, err := summon.BuildPool(def, cards, refs[def.PoolId])
		if err != nil {
			s.warn("summon pool skipped: %v", err)
			continue
		}
		pools[def.PoolId] = pool
	}
	return pools, nil
}

// getPool 取得可抽的卡池；資料超過 TTL 時重新載入，重新載入失敗則沿用舊資料。
func (s *SummonService) getPool(ctx context.Context, poolId string) (*summon.Pool, error) {
	s.poolMu.Lock()
	defer s.poolMu.Unlock()

	now := s.now()
	if s.pools == nil || now.Sub(s.loadedAt) >= poolCacheTTL {
		pools, err := s.loadPools(ctx)
		switch {
		case err == nil:
			s.pools, s.loadedAt = pools, now
		case s.pools == nil:
			return nil, err
		default:
			s.warn("reload summon pools failed, using stale data: %v", err)
			s.loadedAt = now.Add(poolReloadRetryWait - poolCacheTTL)
		}
	}

	pool, ok := s.pools[poolId]
	if !ok || !pool.Available(now) {
		return nil, summon.ErrPoolUnavailable
	}
	return pool, nil
}

// SummonCard 驗證請求後，在單一 transaction 內扣款、抽卡並寫入結果。
// 卡池、費用與貨幣一律以伺服器端的卡池定義為準。
func (s *SummonService) SummonCard(ctx context.Context, uid string, data *packet.SummonCard) (*packet.SummonCardResult, error) {
	pool, err := s.getPool(ctx, data.PoolId)
	if err != nil {
		return nil, err
	}
	cost, err := pool.Cost(data.Times)
	if err != nil {
		return nil, err
	}

	outcome, err := s.repo.ApplySummon(ctx, uid, pool.PoolId, pool.Currency, cost, pool.TracksPity(),
		func(owned map[string]struct{}, pity summon.PityState) (summon.Outcome, error) {
			return pool.Summon(data.Times, owned, pity, s.rnd)
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
			IsUp:        r.IsUp,
			ShardId:     r.Shard.ItemId,
			ShardAmount: r.Shard.Amount,
		})
	}
	result := &packet.SummonCardResult{PoolId: pool.PoolId, Cards: items}
	if pool.Pity != nil {
		result.Pity = &packet.SummonPityInfo{
			PullsSinceHigh: int32(outcome.Pity.PullsSinceHigh),
			Threshold:      int32(pool.Pity.Threshold),
			GuaranteedUp:   outcome.Pity.GuaranteedUp,
		}
	}
	return result, nil
}
