package summon

import (
	"errors"
	"math/rand/v2"
	"reflect"
	"testing"
	"time"

	"cardGame/feature/card"
	"cardGame/model/constants"
)

func mkCard(id string, r constants.Rarity, limited bool) card.Card {
	return card.Card{CardId: id, CardName: id, Rarity: r, IsLimited: limited}
}

func testCards() []card.Card {
	return []card.Card{
		mkCard("r1", constants.RarityR, false),
		mkCard("r2", constants.RarityR, false),
		mkCard("sr1", constants.RaritySR, false),
		mkCard("ssr1", constants.RaritySSR, false),
		mkCard("ssr2", constants.RaritySSR, false),
		mkCard("ssr_limited", constants.RaritySSR, true),
		mkCard("ur1", constants.RarityUR, false),
	}
}

func normalDef() PoolDefinition {
	return PoolDefinition{
		PoolId: "normal", Currency: constants.CurrencyCoin,
		Costs:                map[int32]int64{1: 5000, 10: 50000},
		Weights:              []RarityWeight{{constants.RarityR, 70}, {constants.RaritySR, 30}},
		IncludeStandardCards: true, IsEnabled: true,
	}
}

func diamondDef() PoolDefinition {
	return PoolDefinition{
		PoolId: "diamond", Currency: constants.CurrencyDiamond,
		Costs: map[int32]int64{1: 200, 10: 1800},
		Weights: []RarityWeight{
			{constants.RarityR, 50}, {constants.RaritySR, 30}, {constants.RaritySSR, 18}, {constants.RarityUR, 2},
		},
		TenPullMinRarity:     constants.RaritySSR,
		IncludeStandardCards: true, IsEnabled: true,
	}
}

func upDef() PoolDefinition {
	d := diamondDef()
	d.PoolId = "up"
	d.UpRates = map[constants.Rarity]int{constants.RaritySSR: 50}
	d.Pity = &PityRule{Threshold: 20, MinRarity: constants.RaritySSR, UpGuarantee: true}
	return d
}

var upRefs = []PoolCardRef{{CardId: "ssr_limited", IsUp: true}}

func mustBuild(t *testing.T, def PoolDefinition, refs []PoolCardRef) *Pool {
	t.Helper()
	p, err := BuildPool(def, testCards(), refs)
	if err != nil {
		t.Fatal(err)
	}
	return p
}

// zeroRand 永遠回傳 0：稀有度落在權重表第一項、卡片取排序後第一張、UP 判斷為真。
type zeroRand struct{}

func (zeroRand) IntN(int) int { return 0 }

// seqRand 依序回傳預先指定的值，並檢查範圍。
type seqRand struct {
	t    *testing.T
	vals []int
	i    int
}

func (s *seqRand) IntN(n int) int {
	s.t.Helper()
	if s.i >= len(s.vals) {
		s.t.Fatalf("seqRand exhausted at call %d (n=%d)", s.i, n)
	}
	v := s.vals[s.i]
	s.i++
	if v >= n {
		s.t.Fatalf("seqRand value %d out of range n=%d (call %d)", v, n, s.i)
	}
	return v
}

func seeded(seed uint64) Rand { return rand.New(rand.NewPCG(seed, seed+1)) }

func TestBuildPoolMembership(t *testing.T) {
	std := mustBuild(t, diamondDef(), nil)
	if n := std.cardCount(constants.RaritySSR); n != 2 {
		t.Errorf("standard SSR count = %d, want 2 (limited excluded)", n)
	}

	withUp := mustBuild(t, upDef(), upRefs)
	if len(withUp.up[constants.RaritySSR]) != 1 || withUp.up[constants.RaritySSR][0].Card.CardId != "ssr_limited" {
		t.Errorf("up SSR = %+v", withUp.up[constants.RaritySSR])
	}
	if len(withUp.normal[constants.RaritySSR]) != 2 {
		t.Errorf("normal SSR = %d, want 2", len(withUp.normal[constants.RaritySSR]))
	}

	// 指定既有的標準卡為 UP：從一般卡移到 UP
	d := upDef()
	p := mustBuild(t, d, []PoolCardRef{{CardId: "ssr1", IsUp: true}})
	if len(p.up[constants.RaritySSR]) != 1 || len(p.normal[constants.RaritySSR]) != 1 {
		t.Errorf("override: up=%d normal=%d", len(p.up[constants.RaritySSR]), len(p.normal[constants.RaritySSR]))
	}

	// 不含標準卡：只有指定的卡
	only := PoolDefinition{
		PoolId: "only", Currency: constants.CurrencyDiamond, Costs: map[int32]int64{1: 100},
		Weights: []RarityWeight{{constants.RaritySSR, 100}}, UpRates: map[constants.Rarity]int{constants.RaritySSR: 100},
		IncludeStandardCards: false, IsEnabled: true,
	}
	op := mustBuild(t, only, upRefs)
	if op.cardCount(constants.RarityR) != 0 || op.cardCount(constants.RaritySSR) != 1 {
		t.Errorf("only pool counts wrong: R=%d SSR=%d", op.cardCount(constants.RarityR), op.cardCount(constants.RaritySSR))
	}

	// 指向未啟用（不在清單中）的卡會被略過，因此 UP 比例沒有對應的卡而驗證失敗
	if _, err := BuildPool(upDef(), testCards(), []PoolCardRef{{CardId: "ghost", IsUp: true}}); !errors.Is(err, ErrInvalidPool) {
		t.Errorf("err = %v, want ErrInvalidPool", err)
	}
}

func TestPoolValidation(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(*PoolDefinition) []PoolCardRef
		cards   []card.Card
		wantErr error
	}{
		{"valid up pool", func(d *PoolDefinition) []PoolCardRef { *d = upDef(); return upRefs }, nil, nil},
		{"empty id", func(d *PoolDefinition) []PoolCardRef { d.PoolId = ""; return nil }, nil, ErrInvalidPool},
		{"unknown currency", func(d *PoolDefinition) []PoolCardRef { d.Currency = 99; return nil }, nil, ErrInvalidPool},
		{"no costs", func(d *PoolDefinition) []PoolCardRef { d.Costs = nil; return nil }, nil, ErrInvalidPool},
		{"cost for 5 pulls", func(d *PoolDefinition) []PoolCardRef { d.Costs = map[int32]int64{5: 10}; return nil }, nil, ErrInvalidPool},
		{"zero cost", func(d *PoolDefinition) []PoolCardRef { d.Costs = map[int32]int64{1: 0}; return nil }, nil, ErrInvalidPool},
		{"no weights", func(d *PoolDefinition) []PoolCardRef { d.Weights = nil; return nil }, nil, ErrInvalidPool},
		{"zero weight", func(d *PoolDefinition) []PoolCardRef { d.Weights[0].Weight = 0; return nil }, nil, ErrInvalidPool},
		{"duplicate rarity", func(d *PoolDefinition) []PoolCardRef {
			d.Weights = append(d.Weights, RarityWeight{constants.RarityR, 5})
			return nil
		}, nil, ErrInvalidPool},
		{"rarity without cards", func(d *PoolDefinition) []PoolCardRef { return nil }, []card.Card{mkCard("r1", constants.RarityR, false)}, ErrEmptyCardPool},
		{"ten pull min unreachable", func(d *PoolDefinition) []PoolCardRef {
			*d = normalDef() // 只有 R、SR
			d.TenPullMinRarity = constants.RaritySSR
			return nil
		}, nil, ErrInvalidPool},
		{"up rate without up card", func(d *PoolDefinition) []PoolCardRef {
			*d = diamondDef()
			d.UpRates = map[constants.Rarity]int{constants.RaritySSR: 50}
			return nil
		}, nil, ErrInvalidPool},
		{"up card without up rate", func(d *PoolDefinition) []PoolCardRef { *d = diamondDef(); return upRefs }, nil, ErrInvalidPool},
		{"up rate zero", func(d *PoolDefinition) []PoolCardRef { *d = upDef(); d.UpRates[constants.RaritySSR] = 0; return upRefs }, nil, ErrInvalidPool},
		{"up rate over 100", func(d *PoolDefinition) []PoolCardRef {
			*d = upDef()
			d.UpRates[constants.RaritySSR] = 101
			return upRefs
		}, nil, ErrInvalidPool},
		{"pity threshold zero", func(d *PoolDefinition) []PoolCardRef { *d = upDef(); d.Pity.Threshold = 0; return upRefs }, nil, ErrInvalidPool},
		{"pity min rarity unreachable", func(d *PoolDefinition) []PoolCardRef {
			*d = normalDef() // 只有 R、SR
			d.Pity = &PityRule{Threshold: 10, MinRarity: constants.RaritySSR}
			return nil
		}, nil, ErrInvalidPool},
		{"up guarantee without up at or above min", func(d *PoolDefinition) []PoolCardRef {
			*d = upDef()
			d.UpRates = map[constants.Rarity]int{constants.RaritySR: 50}
			return []PoolCardRef{{CardId: "sr1", IsUp: true}}
		}, nil, ErrInvalidPool},
		{"end before start", func(d *PoolDefinition) []PoolCardRef {
			start := time.Now()
			end := start.Add(-time.Hour)
			d.StartAt, d.EndAt = &start, &end
			return nil
		}, nil, ErrInvalidPool},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			def := diamondDef()
			def.Weights = append([]RarityWeight(nil), def.Weights...)
			refs := tt.mutate(&def)
			cards := tt.cards
			if cards == nil {
				cards = testCards()
			}
			_, err := BuildPool(def, cards, refs)
			if tt.wantErr == nil {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("err = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestAvailableWindow(t *testing.T) {
	now := time.Date(2026, 11, 2, 12, 0, 0, 0, time.UTC)
	hour := time.Hour
	past, future := now.Add(-hour), now.Add(hour)
	tests := []struct {
		name       string
		enabled    bool
		start, end *time.Time
		want       bool
	}{
		{"always open", true, nil, nil, true},
		{"disabled", false, nil, nil, false},
		{"inside window", true, &past, &future, true},
		{"not started", true, &future, nil, false},
		{"ended", true, nil, &past, false},
		{"starts exactly now", true, &now, nil, true},
		{"ends exactly now", true, nil, &now, false},
	}
	for _, tt := range tests {
		d := PoolDefinition{IsEnabled: tt.enabled, StartAt: tt.start, EndAt: tt.end}
		if got := d.Available(now); got != tt.want {
			t.Errorf("%s: Available = %v, want %v", tt.name, got, tt.want)
		}
	}
}

func TestCost(t *testing.T) {
	d := diamondDef()
	for times, want := range map[int32]int64{1: 200, 10: 1800} {
		if got, err := d.Cost(times); err != nil || got != want {
			t.Errorf("Cost(%d) = %d, %v; want %d", times, got, err, want)
		}
	}
	for _, times := range []int32{0, 2, 5, 11, -1} {
		if _, err := d.Cost(times); !errors.Is(err, ErrInvalidTimes) {
			t.Errorf("Cost(%d) err = %v, want ErrInvalidTimes", times, err)
		}
	}
}

func rarityCounts(t *testing.T, p *Pool, n int, rnd Rand) map[constants.Rarity]int {
	t.Helper()
	counts := map[constants.Rarity]int{}
	for i := 0; i < n; i++ {
		drawn, _, err := p.Draw(1, PityState{}, rnd)
		if err != nil {
			t.Fatal(err)
		}
		counts[drawn[0].Card.Rarity]++
	}
	return counts
}

func TestDrawNormalOnlyRAndSR(t *testing.T) {
	const n = 20000
	counts := rarityCounts(t, mustBuild(t, normalDef(), nil), n, seeded(1))
	if counts[constants.RaritySSR] != 0 || counts[constants.RarityUR] != 0 {
		t.Fatalf("normal pool drew SSR/UR: %v", counts)
	}
	if r := float64(counts[constants.RarityR]) / n; r < 0.68 || r > 0.72 {
		t.Errorf("R ratio = %.3f, want about 0.70", r)
	}
}

func TestDrawDiamondDistribution(t *testing.T) {
	const n = 50000
	counts := rarityCounts(t, mustBuild(t, diamondDef(), nil), n, seeded(2))
	want := map[constants.Rarity]float64{constants.RarityR: 0.50, constants.RaritySR: 0.30, constants.RaritySSR: 0.18, constants.RarityUR: 0.02}
	for rarity, w := range want {
		if got := float64(counts[rarity]) / n; got < w-0.01 || got > w+0.01 {
			t.Errorf("%v ratio = %.3f, want about %.2f", rarity, got, w)
		}
	}
}

func TestDrawTenPullGuaranteesSSR(t *testing.T) {
	p := mustBuild(t, diamondDef(), nil)

	drawn, _, err := p.Draw(TenPullTimes, PityState{}, zeroRand{})
	if err != nil {
		t.Fatal(err)
	}
	if len(drawn) != 10 || drawn[9].Card.Rarity != constants.RaritySSR {
		t.Fatalf("last = %+v, want SSR forced", drawn[len(drawn)-1])
	}
	for _, pc := range drawn[:9] {
		if pc.Card.Rarity != constants.RarityR {
			t.Fatalf("first nine should be R, got %+v", pc)
		}
	}

	rnd := seeded(3)
	for i := 0; i < 5000; i++ {
		drawn, _, err := p.Draw(TenPullTimes, PityState{}, rnd)
		if err != nil {
			t.Fatal(err)
		}
		ok := false
		for _, pc := range drawn {
			ok = ok || pc.Card.Rarity >= constants.RaritySSR
		}
		if !ok {
			t.Fatalf("ten pull %d has no SSR or above", i)
		}
	}
}

func TestDrawNoGuaranteeForSingleOrNormalPool(t *testing.T) {
	drawn, _, err := mustBuild(t, normalDef(), nil).Draw(TenPullTimes, PityState{}, zeroRand{})
	if err != nil || len(drawn) != 10 {
		t.Fatal(err, len(drawn))
	}
	for _, pc := range drawn {
		if pc.Card.Rarity != constants.RarityR {
			t.Fatalf("unexpected %+v", pc)
		}
	}
	// Diamond 單抽沒有保底
	single, _, err := mustBuild(t, diamondDef(), nil).Draw(1, PityState{}, zeroRand{})
	if err != nil || single[0].Card.Rarity != constants.RarityR {
		t.Fatalf("single = %+v, %v", single, err)
	}
	if _, _, err := mustBuild(t, normalDef(), nil).Draw(0, PityState{}, zeroRand{}); !errors.Is(err, ErrInvalidTimes) {
		t.Fatalf("err = %v, want ErrInvalidTimes", err)
	}
}

func TestDrawUpRate(t *testing.T) {
	d := upDef()
	d.Pity = nil // 只測 UP 比例
	p := mustBuild(t, d, upRefs)
	rnd := seeded(4)
	ssr, up := 0, 0
	for i := 0; i < 100000; i++ {
		drawn, _, err := p.Draw(1, PityState{}, rnd)
		if err != nil {
			t.Fatal(err)
		}
		if drawn[0].Card.Rarity == constants.RaritySSR {
			ssr++
			if drawn[0].IsUp {
				up++
			}
		}
	}
	if r := float64(up) / float64(ssr); r < 0.48 || r > 0.52 {
		t.Errorf("up ratio among SSR = %.3f (%d/%d), want about 0.50", r, up, ssr)
	}
}

func TestPityForcesTwentiethPull(t *testing.T) {
	p := mustBuild(t, upDef(), upRefs)
	state := PityState{}
	for pull := 1; pull <= 20; pull++ {
		drawn, next, err := p.Draw(1, state, zeroRand{})
		if err != nil {
			t.Fatal(err)
		}
		state = next
		if pull < 20 {
			if drawn[0].Card.Rarity != constants.RarityR || state.PullsSinceHigh != pull {
				t.Fatalf("pull %d: %+v state %+v", pull, drawn[0], state)
			}
		} else if drawn[0].Card.Rarity != constants.RaritySSR || state.PullsSinceHigh != 0 {
			t.Fatalf("pull 20: %+v state %+v, want SSR and reset", drawn[0], state)
		}
	}
}

func TestPityInsideTenPull(t *testing.T) {
	p := mustBuild(t, upDef(), upRefs)
	drawn, state, err := p.Draw(TenPullTimes, PityState{PullsSinceHigh: 15}, zeroRand{})
	if err != nil {
		t.Fatal(err)
	}
	for i, pc := range drawn {
		want := constants.RarityR
		if i == 4 { // 第 5 抽累積到第 20 抽
			want = constants.RaritySSR
		}
		if pc.Card.Rarity != want {
			t.Errorf("pull %d rarity = %v, want %v", i, pc.Card.Rarity, want)
		}
	}
	// 保底已出 SSR，十連保底視為滿足，第 10 抽不再被強制
	if state.PullsSinceHigh != 5 {
		t.Errorf("PullsSinceHigh = %d, want 5", state.PullsSinceHigh)
	}
}

func TestUpGuarantee(t *testing.T) {
	p := mustBuild(t, upDef(), upRefs)
	// 權重 R50 SR30 SSR18 UR2：80~97 為 SSR，98~99 為 UR；SSR 非 UP 為 ssr1、ssr2
	const ssrRoll, urRoll = 90, 99

	// A：出非 UP 的 SSR → 取得大保底
	drawn, state, err := p.Draw(1, PityState{}, &seqRand{t: t, vals: []int{ssrRoll, 99, 0}})
	if err != nil || drawn[0].IsUp || drawn[0].Card.CardId != "ssr1" || !state.GuaranteedUp || state.PullsSinceHigh != 0 {
		t.Fatalf("A: %+v state %+v err %v", drawn, state, err)
	}
	// B：大保底生效，必為 UP 且不再抽 UP 判斷，之後解除
	drawn, state, err = p.Draw(1, state, &seqRand{t: t, vals: []int{ssrRoll, 0}})
	if err != nil || !drawn[0].IsUp || drawn[0].Card.CardId != "ssr_limited" || state.GuaranteedUp {
		t.Fatalf("B: %+v state %+v err %v", drawn, state, err)
	}
	// C：沒有大保底時，UP 判斷成功也是 UP，維持無大保底
	drawn, state, err = p.Draw(1, state, &seqRand{t: t, vals: []int{ssrRoll, 10, 0}})
	if err != nil || !drawn[0].IsUp || state.GuaranteedUp {
		t.Fatalf("C: %+v state %+v err %v", drawn, state, err)
	}
	// D：保底稀有度以下的卡不影響大保底
	drawn, state, err = p.Draw(1, PityState{GuaranteedUp: true}, &seqRand{t: t, vals: []int{0, 0}})
	if err != nil || drawn[0].Card.Rarity != constants.RarityR || !state.GuaranteedUp || state.PullsSinceHigh != 1 {
		t.Fatalf("D: %+v state %+v err %v", drawn, state, err)
	}
	// E：出了沒有 UP 卡的稀有度（UR），照一般機率，也不改變大保底
	drawn, state, err = p.Draw(1, PityState{GuaranteedUp: true}, &seqRand{t: t, vals: []int{urRoll, 0}})
	if err != nil || drawn[0].Card.Rarity != constants.RarityUR || !state.GuaranteedUp || state.PullsSinceHigh != 0 {
		t.Fatalf("E: %+v state %+v err %v", drawn, state, err)
	}
}

func TestSummonComposesDrawAndResolve(t *testing.T) {
	p := mustBuild(t, normalDef(), nil)
	owned := map[string]struct{}{"r1": {}}
	out, err := p.Summon(3, owned, PityState{}, zeroRand{}) // zeroRand 永遠抽到 r1
	if err != nil {
		t.Fatal(err)
	}
	if len(out.NewCardIds) != 0 || out.Shards["shard_r"] != 90 || len(out.Results) != 3 {
		t.Errorf("outcome = %+v", out)
	}
	if len(owned) != 1 {
		t.Error("owned was mutated")
	}
}

func TestResolve(t *testing.T) {
	by := map[string]PoolCard{}
	for _, c := range testCards() {
		by[c.CardId] = PoolCard{Card: c}
	}
	up := by["ssr1"]
	up.IsUp = true
	owned := map[string]struct{}{"r1": {}}
	drawn := []PoolCard{by["r1"], by["sr1"], by["sr1"], by["ur1"], by["r1"], up}

	outcome, err := Resolve(drawn, owned)
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{"sr1", "ur1", "ssr1"}; !reflect.DeepEqual(outcome.NewCardIds, want) {
		t.Errorf("NewCardIds = %v, want %v", outcome.NewCardIds, want)
	}
	wantShards := map[string]int64{"shard_r": 60, "shard_sr": 30}
	if !reflect.DeepEqual(outcome.Shards, wantShards) {
		t.Errorf("Shards = %v, want %v", outcome.Shards, wantShards)
	}
	isNew := []bool{false, true, false, true, false, true}
	for i, r := range outcome.Results {
		if r.IsNew != isNew[i] {
			t.Errorf("result %d IsNew = %v, want %v", i, r.IsNew, isNew[i])
		}
		if r.IsNew != (r.Shard == ShardReward{}) {
			t.Errorf("result %d shard %+v inconsistent with IsNew", i, r.Shard)
		}
	}
	if !outcome.Results[5].IsUp || outcome.Results[1].IsUp {
		t.Error("IsUp not carried to results")
	}
	if len(owned) != 1 {
		t.Errorf("owned was mutated: %v", owned)
	}
}

func TestShardRewardsCoverAllRarities(t *testing.T) {
	want := map[constants.Rarity]ShardReward{
		constants.RarityR: {"shard_r", 30}, constants.RaritySR: {"shard_sr", 30},
		constants.RaritySSR: {"shard_ssr", 10}, constants.RarityUR: {"shard_ur", 5},
	}
	if !reflect.DeepEqual(shardRewards, want) {
		t.Errorf("shardRewards = %v", shardRewards)
	}
}
