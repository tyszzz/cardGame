package services

import (
	"cardGame/feature/growth"
	"cardGame/model/constants"
	"context"
	"errors"
	"testing"
	"time"
)

type fakeGrowthRepo struct {
	raw   growth.RawConfig
	err   error
	calls int
}

func (f *fakeGrowthRepo) LoadGrowth(context.Context) (growth.RawConfig, error) {
	f.calls++
	return f.raw, f.err
}

func levelRows(kind string, needs ...int64) []growth.LevelRow {
	rows := make([]growth.LevelRow, 0, len(needs))
	for i, n := range needs {
		rows = append(rows, growth.LevelRow{Kind: kind, Level: i + 1, ExpRequired: n})
	}
	return rows
}

func validGrowthRaw(cardNeeds, playerNeeds []int64) growth.RawConfig {
	return growth.RawConfig{
		LevelExp: append(levelRows(growth.KindCard, cardNeeds...), levelRows(growth.KindPlayer, playerNeeds...)...),
		Stages: []growth.StageRow{
			{Stage: 0, MaxLevel: 2},
			{Stage: 1, ItemId: "ascend_pill", ItemQty: 500, ExpPercent: 50},
		},
		StarLevels: []growth.StarLevelRow{{Star: 0, MaxLevel: 3}, {Star: 1, MaxLevel: 4}},
		Rarities:   []growth.RarityLimitRow{{Rarity: constants.RarityR, MaxStar: 1}},
		StarCosts:  []growth.StarCostRow{{Rarity: constants.RarityR, Star: 1, ItemId: "shard_r", Quantity: 50}},
	}
}

func newTestGrowthService(repo *fakeGrowthRepo, now *time.Time) (*GrowthService, *[]string) {
	var warnings []string
	s := NewGrowthService(repo)
	s.now = func() time.Time { return *now }
	s.warn = func(format string, _ ...any) { warnings = append(warnings, format) }
	return s, &warnings
}

func TestGrowthServiceCachesConfig(t *testing.T) {
	now := time.Unix(0, 0)
	repo := &fakeGrowthRepo{raw: validGrowthRaw([]int64{100, 120, 144}, []int64{1, 2})}
	s, _ := newTestGrowthService(repo, &now)

	card, err := s.Card(context.Background())
	if err != nil || card.MaxLevel() != 4 {
		t.Fatalf("card = %+v, err = %v", card, err)
	}
	player, err := s.Player(context.Background())
	if err != nil || player.MaxLevel() != 3 {
		t.Fatalf("player = %+v, err = %v", player, err)
	}
	if repo.calls != 1 {
		t.Errorf("card and player must share one load, calls = %d", repo.calls)
	}

	repo.raw = validGrowthRaw([]int64{100, 120, 144, 200}, []int64{1, 2})
	now = now.Add(growthCacheTTL)
	card, err = s.Card(context.Background())
	if err != nil || card.MaxLevel() != 5 || repo.calls != 2 {
		t.Errorf("must reload after the TTL: max level %d, calls %d, err %v", card.MaxLevel(), repo.calls, err)
	}
}

func TestGrowthServiceKeepsStaleOnReloadFailure(t *testing.T) {
	now := time.Unix(0, 0)
	repo := &fakeGrowthRepo{raw: validGrowthRaw([]int64{100, 120, 144}, []int64{1, 2})}
	s, warnings := newTestGrowthService(repo, &now)
	s.Card(context.Background())

	repo.err = errors.New("db down")
	now = now.Add(growthCacheTTL)
	if card, err := s.Card(context.Background()); err != nil || card.MaxLevel() != 4 {
		t.Fatalf("expected the stale config, got %+v, err %v", card, err)
	}
	if len(*warnings) == 0 {
		t.Error("a warning must be logged when falling back to stale data")
	}
	calls := repo.calls
	now = now.Add(growthReloadRetryWait - time.Second)
	s.Card(context.Background())
	if repo.calls != calls {
		t.Error("must not retry before the retry wait has passed")
	}
	now = now.Add(2 * time.Second)
	s.Card(context.Background())
	if repo.calls != calls+1 {
		t.Error("must retry after the retry wait")
	}
}

func TestGrowthServiceFirstLoadFailure(t *testing.T) {
	now := time.Unix(0, 0)
	boom := errors.New("db down")
	s, _ := newTestGrowthService(&fakeGrowthRepo{err: boom}, &now)
	if _, err := s.Card(context.Background()); !errors.Is(err, boom) {
		t.Errorf("err = %v, want %v", err, boom)
	}
}

func TestGrowthServicePartsAreIndependent(t *testing.T) {
	now := time.Unix(0, 0)
	// 玩家表不合法（經驗沒有遞增）：只有玩家不可用，卡片照常
	repo := &fakeGrowthRepo{raw: validGrowthRaw([]int64{100, 120, 144}, []int64{5, 5})}
	s, warnings := newTestGrowthService(repo, &now)
	if _, err := s.Card(context.Background()); err != nil {
		t.Errorf("a broken player table must not disable cards: %v", err)
	}
	if _, err := s.Player(context.Background()); !errors.Is(err, growth.ErrUnavailable) {
		t.Errorf("player err = %v, want ErrUnavailable", err)
	}
	if len(*warnings) == 0 {
		t.Error("an invalid part must be reported")
	}

	// 空資料：兩個部分都不可用
	s2, _ := newTestGrowthService(&fakeGrowthRepo{}, &now)
	if _, err := s2.Card(context.Background()); !errors.Is(err, growth.ErrUnavailable) {
		t.Errorf("card err = %v", err)
	}
}

func TestGrowthServiceKeepsPreviousValidPartOnBadReload(t *testing.T) {
	now := time.Unix(0, 0)
	repo := &fakeGrowthRepo{raw: validGrowthRaw([]int64{100, 120, 144}, []int64{1, 2})}
	s, warnings := newTestGrowthService(repo, &now)
	s.Card(context.Background())

	// 資料被改壞：卡片經驗表不遞增。重新載入後仍沿用上一份合法的卡片設定。
	repo.raw = validGrowthRaw([]int64{100, 90, 144}, []int64{1, 2, 3})
	now = now.Add(growthCacheTTL)
	card, err := s.Card(context.Background())
	if err != nil || card.MaxLevel() != 4 {
		t.Fatalf("expected the previous card config, got %+v, err %v", card, err)
	}
	player, err := s.Player(context.Background())
	if err != nil || player.MaxLevel() != 4 {
		t.Errorf("the new valid player table must be used: %+v, err %v", player, err)
	}
	if len(*warnings) < 2 {
		t.Errorf("expected warnings for the invalid part and the fallback, got %v", *warnings)
	}
}
