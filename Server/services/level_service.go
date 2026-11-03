package services

import (
	"cardGame/feature/level"
	"cardGame/utils"
	"context"
	"fmt"
	"sync"
	"time"
)

type LevelService struct {
	repo level.LevelExpRepository
	now  func() time.Time
	warn func(format string, args ...any)

	mu       sync.Mutex
	tables   map[string]*level.Table
	loadedAt time.Time
}

const (
	levelCacheTTL        = 5 * time.Minute  // 經驗表定時重新載入的間隔
	levelReloadRetryWait = 30 * time.Second // 重新載入失敗並沿用舊資料時，下次重試的等待時間
)

func NewLevelService(repo level.LevelExpRepository) *LevelService {
	return &LevelService{
		repo: repo,
		now:  time.Now,
		warn: func(format string, args ...any) { utils.Log().Warnf(format, args...) },
	}
}

// inner function

// loadTables 載入並驗證所有經驗表；不合法的種類會被略過並記錄警告，不影響其他種類。
func (s *LevelService) loadTables(ctx context.Context) (map[string]*level.Table, error) {
	rows, err := s.repo.ListLevelExp(ctx)
	if err != nil {
		return nil, err
	}
	tables, err := level.BuildTables(rows)
	if err != nil {
		s.warn("some level exp tables skipped: %v", err)
	}
	return tables, nil
}

// getTable 取得指定種類的經驗表；資料超過 TTL 時重新載入，重新載入失敗則沿用舊資料。
func (s *LevelService) getTable(ctx context.Context, kind string) (*level.Table, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := s.now()
	if s.tables == nil || now.Sub(s.loadedAt) >= levelCacheTTL {
		tables, err := s.loadTables(ctx)
		switch {
		case err == nil:
			s.tables, s.loadedAt = tables, now
		case s.tables == nil:
			return nil, err
		default:
			s.warn("reload level exp tables failed, using stale data: %v", err)
			s.loadedAt = now.Add(levelReloadRetryWait - levelCacheTTL)
		}
	}

	table, ok := s.tables[kind]
	if !ok {
		return nil, fmt.Errorf("%w: %s", level.ErrTableUnavailable, kind)
	}
	return table, nil
}

// outer function

// CardTable 回傳卡片的升級經驗表。
func (s *LevelService) CardTable(ctx context.Context) (*level.Table, error) {
	return s.getTable(ctx, level.KindCard)
}
