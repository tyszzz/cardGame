package services

import (
	"cardGame/feature/card"
	"cardGame/feature/summon"
	"cardGame/model/constants"
	"cardGame/model/packet"
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

type zeroRand struct{}

func (zeroRand) IntN(int) int { return 0 }

type applyCall struct {
	uid, poolId string
	currency    constants.CurrencyType
	cost        int64
	trackPity   bool
}

type fakeSummonRepo struct {
	cards     []card.Card
	pools     []summon.PoolDefinition
	poolCards map[string][]summon.PoolCardRef
	owned     map[string]struct{}
	pity      map[string]summon.PityState // pool id -> 保底狀態

	cardsErr  error
	poolsErr  error // 與 pools 同時設定代表部分卡池解析失敗
	applyErr  error
	loadCalls int
	calls     []applyCall
}

func (f *fakeSummonRepo) ListEnabledCards(context.Context) ([]card.Card, error) {
	f.loadCalls++
	return f.cards, f.cardsErr
}

func (f *fakeSummonRepo) ListPools(context.Context) ([]summon.PoolDefinition, error) {
	return f.pools, f.poolsErr
}

func (f *fakeSummonRepo) ListPoolCards(context.Context) (map[string][]summon.PoolCardRef, error) {
	return f.poolCards, nil
}

func (f *fakeSummonRepo) ApplySummon(_ context.Context, uid, poolId string, c constants.CurrencyType, cost int64, trackPity bool, resolve summon.ResolveFunc) (*summon.Outcome, error) {
	f.calls = append(f.calls, applyCall{uid, poolId, c, cost, trackPity})
	if f.applyErr != nil {
		return nil, f.applyErr
	}
	out, err := resolve(f.owned, f.pity[poolId])
	if err != nil {
		return nil, err
	}
	if trackPity {
		if f.pity == nil {
			f.pity = map[string]summon.PityState{}
		}
		f.pity[poolId] = out.Pity
	}
	return &out, nil
}

type testClock struct{ t time.Time }

func (c *testClock) now() time.Time { return c.t }

func newTestSummonService(repo *fakeSummonRepo) (*SummonService, *testClock, *[]string) {
	clock := &testClock{t: time.Date(2026, 11, 2, 12, 0, 0, 0, time.UTC)}
	var warns []string
	s := NewSummonService(repo)
	s.rnd = zeroRand{}
	s.now = clock.now
	s.warn = func(format string, args ...any) { warns = append(warns, format) }
	return s, clock, &warns
}

func mkCard(id string, r constants.Rarity, limited bool) card.Card {
	return card.Card{CardId: id, CardName: id, Rarity: r, IsLimited: limited}
}

func allCards() []card.Card {
	return []card.Card{
		mkCard("r1", constants.RarityR, false),
		mkCard("sr1", constants.RaritySR, false),
		mkCard("ssr1", constants.RaritySSR, false),
		mkCard("ssr_limited", constants.RaritySSR, true),
		mkCard("ur1", constants.RarityUR, false),
	}
}

func normalDef() summon.PoolDefinition {
	return summon.PoolDefinition{
		PoolId: "normal", Currency: constants.CurrencyCoin,
		Costs:                map[int32]int64{1: 5000, 10: 50000},
		Weights:              []summon.RarityWeight{{Rarity: constants.RarityR, Weight: 70}, {Rarity: constants.RaritySR, Weight: 30}},
		IncludeStandardCards: true, IsEnabled: true,
	}
}

func diamondDef() summon.PoolDefinition {
	return summon.PoolDefinition{
		PoolId: "diamond", Currency: constants.CurrencyDiamond,
		Costs: map[int32]int64{1: 200, 10: 1800},
		Weights: []summon.RarityWeight{
			{Rarity: constants.RarityR, Weight: 50}, {Rarity: constants.RaritySR, Weight: 30},
			{Rarity: constants.RaritySSR, Weight: 18}, {Rarity: constants.RarityUR, Weight: 2},
		},
		TenPullMinRarity:     constants.RaritySSR,
		IncludeStandardCards: true, IsEnabled: true,
	}
}

func upDef() summon.PoolDefinition {
	d := diamondDef()
	d.PoolId = "up"
	d.UpRates = map[constants.Rarity]int{constants.RaritySSR: 50}
	d.Pity = &summon.PityRule{Threshold: 20, MinRarity: constants.RaritySSR, UpGuarantee: true}
	return d
}

func summonReq(poolId string, times int32) *packet.SummonCard {
	return &packet.SummonCard{PoolId: poolId, Times: times}
}

func TestSummonCardChargesAndReturnsCards(t *testing.T) {
	repo := &fakeSummonRepo{cards: allCards(), pools: []summon.PoolDefinition{diamondDef()}, owned: map[string]struct{}{}}
	s, _, _ := newTestSummonService(repo)

	res, err := s.SummonCard(context.Background(), "u1", summonReq("diamond", 10))
	if err != nil {
		t.Fatal(err)
	}
	if want := (applyCall{"u1", "diamond", constants.CurrencyDiamond, 1800, false}); len(repo.calls) != 1 || repo.calls[0] != want {
		t.Errorf("apply calls = %+v, want %+v", repo.calls, want)
	}
	if res.PoolId != "diamond" || len(res.Cards) != 10 || res.Pity != nil {
		t.Fatalf("result = %+v", res)
	}
	// zeroRand：前 9 抽為 R，第 10 抽被十連保底改為 SSR；R 只有第一張是新卡
	if !res.Cards[0].IsNew || res.Cards[1].IsNew || res.Cards[1].ShardId != "shard_r" || res.Cards[1].ShardAmount != 30 {
		t.Errorf("unexpected dup handling: %+v %+v", res.Cards[0], res.Cards[1])
	}
	if last := res.Cards[9]; last.Rarity != constants.RaritySSR || !last.IsNew {
		t.Errorf("last = %+v, want new SSR", last)
	}
}

func TestSummonCardValidation(t *testing.T) {
	repo := &fakeSummonRepo{cards: allCards(), pools: []summon.PoolDefinition{normalDef()}}
	s, _, _ := newTestSummonService(repo)
	ctx := context.Background()

	for _, poolId := range []string{"", "nope"} {
		if _, err := s.SummonCard(ctx, "u1", summonReq(poolId, 1)); !errors.Is(err, summon.ErrPoolUnavailable) {
			t.Errorf("pool %q: err = %v, want ErrPoolUnavailable", poolId, err)
		}
	}
	for _, times := range []int32{0, 2, 11, -1} {
		if _, err := s.SummonCard(ctx, "u1", summonReq("normal", times)); !errors.Is(err, summon.ErrInvalidTimes) {
			t.Errorf("times %d: err = %v, want ErrInvalidTimes", times, err)
		}
	}
	if len(repo.calls) != 0 {
		t.Errorf("invalid requests must not charge: %+v", repo.calls)
	}
}

func TestSummonCardNotEnoughCurrency(t *testing.T) {
	repo := &fakeSummonRepo{cards: allCards(), pools: []summon.PoolDefinition{normalDef()}, applyErr: summon.ErrNotEnoughCurrency}
	s, _, _ := newTestSummonService(repo)
	if _, err := s.SummonCard(context.Background(), "u1", summonReq("normal", 1)); !errors.Is(err, summon.ErrNotEnoughCurrency) {
		t.Fatalf("err = %v, want ErrNotEnoughCurrency", err)
	}
}

func TestSummonCardPoolAvailability(t *testing.T) {
	start := time.Date(2026, 11, 2, 13, 0, 0, 0, time.UTC)
	end := start.Add(24 * time.Hour)
	timed := normalDef()
	timed.PoolId = "timed"
	timed.StartAt, timed.EndAt = &start, &end
	disabled := normalDef()
	disabled.PoolId = "off"
	disabled.IsEnabled = false

	repo := &fakeSummonRepo{cards: allCards(), pools: []summon.PoolDefinition{timed, disabled}}
	s, clock, _ := newTestSummonService(repo)
	ctx := context.Background()

	try := func(poolId string) error {
		_, err := s.SummonCard(ctx, "u1", summonReq(poolId, 1))
		return err
	}
	if err := try("timed"); !errors.Is(err, summon.ErrPoolUnavailable) {
		t.Errorf("before start: err = %v", err)
	}
	clock.t = start.Add(time.Minute)
	if err := try("timed"); err != nil {
		t.Errorf("inside window: err = %v", err)
	}
	clock.t = end
	if err := try("timed"); !errors.Is(err, summon.ErrPoolUnavailable) {
		t.Errorf("after end: err = %v", err)
	}
	if err := try("off"); !errors.Is(err, summon.ErrPoolUnavailable) {
		t.Errorf("disabled: err = %v", err)
	}
}

func TestSummonPoolsCachedAndReloaded(t *testing.T) {
	repo := &fakeSummonRepo{cards: allCards(), pools: []summon.PoolDefinition{normalDef()}, owned: map[string]struct{}{}}
	s, clock, warns := newTestSummonService(repo)
	ctx := context.Background()

	for i := 0; i < 3; i++ {
		if _, err := s.SummonCard(ctx, "u1", summonReq("normal", 1)); err != nil {
			t.Fatal(err)
		}
	}
	if repo.loadCalls != 1 {
		t.Errorf("loads within TTL = %d, want 1", repo.loadCalls)
	}

	// 超過 TTL 後重新載入，並看得到資料庫新增的卡池
	repo.pools = append(repo.pools, diamondDef())
	clock.t = clock.t.Add(poolCacheTTL)
	if _, err := s.SummonCard(ctx, "u1", summonReq("diamond", 1)); err != nil {
		t.Fatalf("new pool after reload: %v", err)
	}
	if repo.loadCalls != 2 {
		t.Errorf("loads after TTL = %d, want 2", repo.loadCalls)
	}

	// 重新載入失敗：沿用舊資料並記錄警告，之後 30 秒才會重試
	repo.cardsErr = errors.New("db down")
	clock.t = clock.t.Add(poolCacheTTL)
	if _, err := s.SummonCard(ctx, "u1", summonReq("normal", 1)); err != nil {
		t.Fatalf("stale data should still serve: %v", err)
	}
	if len(*warns) != 1 || repo.loadCalls != 3 {
		t.Errorf("warns=%v loads=%d, want 1 warning and 3 loads", *warns, repo.loadCalls)
	}
	clock.t = clock.t.Add(10 * time.Second)
	if _, err := s.SummonCard(ctx, "u1", summonReq("normal", 1)); err != nil || repo.loadCalls != 3 {
		t.Errorf("should not retry within wait: err=%v loads=%d", err, repo.loadCalls)
	}
	clock.t = clock.t.Add(poolReloadRetryWait)
	repo.cardsErr = nil
	if _, err := s.SummonCard(ctx, "u1", summonReq("normal", 1)); err != nil || repo.loadCalls != 4 {
		t.Errorf("should retry after wait: err=%v loads=%d", err, repo.loadCalls)
	}
}

func TestSummonFirstLoadFailureReturnsError(t *testing.T) {
	boom := errors.New("db down")
	repo := &fakeSummonRepo{cardsErr: boom}
	s, _, _ := newTestSummonService(repo)
	if _, err := s.SummonCard(context.Background(), "u1", summonReq("normal", 1)); !errors.Is(err, boom) {
		t.Fatalf("err = %v, want db error", err)
	}
}

func TestSummonSkipsInvalidAndUnparsablePools(t *testing.T) {
	// Diamond 池缺 SSR、UR（只有 R、SR 的卡）→ 被略過；Normal 池仍可抽
	repo := &fakeSummonRepo{
		cards:    []card.Card{mkCard("r1", constants.RarityR, false), mkCard("sr1", constants.RaritySR, false)},
		pools:    []summon.PoolDefinition{normalDef(), diamondDef()},
		poolsErr: errors.New("pool \"bad\": unparsable"),
		owned:    map[string]struct{}{},
	}
	s, _, warns := newTestSummonService(repo)
	ctx := context.Background()

	if _, err := s.SummonCard(ctx, "u1", summonReq("normal", 1)); err != nil {
		t.Fatalf("normal: %v", err)
	}
	if _, err := s.SummonCard(ctx, "u1", summonReq("diamond", 1)); !errors.Is(err, summon.ErrPoolUnavailable) {
		t.Errorf("diamond: err = %v, want ErrPoolUnavailable", err)
	}
	if len(*warns) != 2 {
		t.Errorf("warnings = %v, want one for unparsable and one for the invalid pool", *warns)
	}
}

func TestSummonUpPoolPity(t *testing.T) {
	repo := &fakeSummonRepo{
		cards:     allCards(),
		pools:     []summon.PoolDefinition{upDef()},
		poolCards: map[string][]summon.PoolCardRef{"up": {{CardId: "ssr_limited", IsUp: true}}},
		owned:     map[string]struct{}{},
		pity:      map[string]summon.PityState{"up": {PullsSinceHigh: 19}},
	}
	s, _, _ := newTestSummonService(repo)

	res, err := s.SummonCard(context.Background(), "u1", summonReq("up", 1))
	if err != nil {
		t.Fatal(err)
	}
	if call := repo.calls[0]; !call.trackPity || call.poolId != "up" {
		t.Errorf("call = %+v, want trackPity on pool up", call)
	}
	if c := res.Cards[0]; c.Rarity != constants.RaritySSR || !c.IsUp || c.CardId != "ssr_limited" {
		t.Errorf("card = %+v, want UP SSR forced by pity", c)
	}
	if res.Pity == nil || res.Pity.PullsSinceHigh != 0 || res.Pity.Threshold != 20 || res.Pity.GuaranteedUp {
		t.Errorf("pity = %+v", res.Pity)
	}
	if got := repo.pity["up"]; got.PullsSinceHigh != 0 {
		t.Errorf("stored pity = %+v, want reset", got)
	}
}

func TestSummonStandardPoolsExcludeLimitedCards(t *testing.T) {
	repo := &fakeSummonRepo{cards: allCards(), pools: []summon.PoolDefinition{diamondDef()}, owned: map[string]struct{}{}}
	s, _, _ := newTestSummonService(repo)
	res, err := s.SummonCard(context.Background(), "u1", summonReq("diamond", 10))
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range res.Cards {
		if strings.Contains(c.CardId, "limited") {
			t.Fatalf("limited card leaked into the standard pool: %+v", c)
		}
	}
}
