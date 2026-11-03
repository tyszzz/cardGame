package services

import (
	"cardGame/feature/growth"
	"cardGame/utils"
	"context"
	"sync"
	"time"
)

type GrowthService struct {
	repo growth.GrowthRepository
	now  func() time.Time
	warn func(format string, args ...any)

	mu       sync.Mutex
	config   *growth.Config
	loadedAt time.Time
}

const (
	growthCacheTTL        = 5 * time.Minute  // 升級、進階、升星設定定時重新載入的間隔
	growthReloadRetryWait = 30 * time.Second // 重新載入失敗並沿用舊資料時，下次重試的等待時間
)

func NewGrowthService(repo growth.GrowthRepository) *GrowthService {
	return &GrowthService{
		repo: repo,
		now:  time.Now,
		warn: func(format string, args ...any) { utils.Log().Warnf(format, args...) },
	}
}

// inner function

// load 從資料庫載入並驗證；不合法的部分記錄警告，並盡量沿用舊設定中合法的部分，
// 避免一次錯誤的資料修改讓已經運作的功能停用。
func (s *GrowthService) load(ctx context.Context) (*growth.Config, error) {
	raw, err := s.repo.LoadGrowth(ctx)
	if err != nil {
		return nil, err
	}
	cfg := growth.Build(raw)
	for _, problem := range cfg.Problems() {
		s.warn("growth config part invalid: %v", problem)
	}
	for _, part := range cfg.KeepValidFrom(s.config) {
		s.warn("growth config part %s kept from the previous valid data", part)
	}
	return cfg, nil
}

// get 取得設定；資料超過 TTL 時重新載入，重新載入失敗則沿用舊資料。
func (s *GrowthService) get(ctx context.Context) (*growth.Config, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := s.now()
	if s.config == nil || now.Sub(s.loadedAt) >= growthCacheTTL {
		cfg, err := s.load(ctx)
		switch {
		case err == nil:
			s.config, s.loadedAt = cfg, now
		case s.config == nil:
			return nil, err
		default:
			s.warn("reload growth config failed, using stale data: %v", err)
			s.loadedAt = now.Add(growthReloadRetryWait - growthCacheTTL)
		}
	}
	return s.config, nil
}

// outer function

// Card 取得卡片的成長設定（升級經驗、階、星數）；該部分不合法時回傳 growth.ErrUnavailable。
func (s *GrowthService) Card(ctx context.Context) (*growth.CardGrowth, error) {
	cfg, err := s.get(ctx)
	if err != nil {
		return nil, err
	}
	return cfg.Card()
}

// Player 取得玩家的升級經驗表；該部分不合法時回傳 growth.ErrUnavailable。
func (s *GrowthService) Player(ctx context.Context) (*growth.Table, error) {
	cfg, err := s.get(ctx)
	if err != nil {
		return nil, err
	}
	return cfg.Player()
}
