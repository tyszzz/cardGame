package summon

import (
	"errors"
	"math/rand/v2"
	"reflect"
	"testing"

	"cardGame/feature/card"
	"cardGame/model/constants"
)

func testCards() []card.Card {
	return []card.Card{
		{CardId: "r1", CardName: "R1", Rarity: constants.RarityR},
		{CardId: "r2", CardName: "R2", Rarity: constants.RarityR},
		{CardId: "sr1", CardName: "SR1", Rarity: constants.RaritySR},
		{CardId: "ssr1", CardName: "SSR1", Rarity: constants.RaritySSR},
		{CardId: "ur1", CardName: "UR1", Rarity: constants.RarityUR},
	}
}

// zeroRand 永遠回傳 0，讓每次都落在權重表第一項、卡池第一張。
type zeroRand struct{}

func (zeroRand) IntN(int) int { return 0 }

func seeded(seed uint64) Rand { return rand.New(rand.NewPCG(seed, seed+1)) }

func TestRuleCost(t *testing.T) {
	tests := []struct {
		st       constants.SummonType
		times    int32
		currency constants.CurrencyType
		cost     int64
	}{
		{constants.NormalSummon, 1, constants.CurrencyCoin, 5000},
		{constants.NormalSummon, 10, constants.CurrencyCoin, 50000},
		{constants.FriendSummon, 1, constants.CurrencyFriendPoint, 20},
		{constants.FriendSummon, 10, constants.CurrencyFriendPoint, 200},
		{constants.DiamondSummon, 1, constants.CurrencyDiamond, 200},
		{constants.DiamondSummon, 10, constants.CurrencyDiamond, 1800},
	}
	for _, tt := range tests {
		rule, err := RuleOf(tt.st)
		if err != nil {
			t.Fatal(err)
		}
		cost, err := rule.Cost(tt.times)
		if err != nil || cost != tt.cost || rule.Currency != tt.currency {
			t.Errorf("%v x%d = %d (%v), currency %v; want %d, %v", tt.st, tt.times, cost, err, rule.Currency, tt.cost, tt.currency)
		}
	}

	rule, _ := RuleOf(constants.NormalSummon)
	for _, times := range []int32{0, 2, 5, 11, -1} {
		if _, err := rule.Cost(times); !errors.Is(err, ErrInvalidTimes) {
			t.Errorf("times %d: err = %v, want ErrInvalidTimes", times, err)
		}
	}
	for _, st := range []constants.SummonType{0, 99} {
		if _, err := RuleOf(st); !errors.Is(err, ErrInvalidSummonType) {
			t.Errorf("type %d: err = %v, want ErrInvalidSummonType", st, err)
		}
	}
}

func TestRuleWeightsSum100(t *testing.T) {
	for st, rule := range rules {
		total := 0
		for _, w := range rule.Weights {
			total += w.Weight
		}
		if total != 100 {
			t.Errorf("%v weights sum = %d, want 100", st, total)
		}
	}
}

func TestDrawNormalOnlyRAndSR(t *testing.T) {
	rule, _ := RuleOf(constants.NormalSummon)
	pool := NewCardPool(testCards())
	rnd := seeded(1)
	counts := map[constants.Rarity]int{}
	const n = 20000
	for i := 0; i < n; i++ {
		drawn, err := Draw(rule, pool, 1, rnd)
		if err != nil {
			t.Fatal(err)
		}
		counts[drawn[0].Rarity]++
	}
	if counts[constants.RaritySSR] != 0 || counts[constants.RarityUR] != 0 {
		t.Fatalf("normal pool drew SSR/UR: %v", counts)
	}
	if r := float64(counts[constants.RarityR]) / n; r < 0.68 || r > 0.72 {
		t.Errorf("R ratio = %.3f, want about 0.70", r)
	}
}

func TestDrawDiamondDistribution(t *testing.T) {
	rule, _ := RuleOf(constants.DiamondSummon)
	pool := NewCardPool(testCards())
	rnd := seeded(2)
	counts := map[constants.Rarity]int{}
	const n = 50000
	for i := 0; i < n; i++ {
		drawn, err := Draw(rule, pool, 1, rnd) // 單抽無保底
		if err != nil {
			t.Fatal(err)
		}
		counts[drawn[0].Rarity]++
	}
	want := map[constants.Rarity]float64{constants.RarityR: 0.50, constants.RaritySR: 0.30, constants.RaritySSR: 0.18, constants.RarityUR: 0.02}
	for rarity, w := range want {
		if got := float64(counts[rarity]) / n; got < w-0.01 || got > w+0.01 {
			t.Errorf("%v ratio = %.3f, want about %.2f", rarity, got, w)
		}
	}
}

func TestDrawDiamondTenPullGuaranteesSSR(t *testing.T) {
	rule, _ := RuleOf(constants.DiamondSummon)
	pool := NewCardPool(testCards())

	// 前 9 抽都是 R，第 10 抽必須被強制改為 SSR 以上
	drawn, err := Draw(rule, pool, TenPullTimes, zeroRand{})
	if err != nil {
		t.Fatal(err)
	}
	if len(drawn) != 10 || drawn[9].Rarity != constants.RaritySSR {
		t.Fatalf("last = %+v, want SSR forced", drawn[len(drawn)-1])
	}
	for _, c := range drawn[:9] {
		if c.Rarity != constants.RarityR {
			t.Fatalf("first nine should be R, got %+v", c)
		}
	}

	// 大量十連都至少有一張 SSR 以上
	rnd := seeded(3)
	for i := 0; i < 5000; i++ {
		drawn, err := Draw(rule, pool, TenPullTimes, rnd)
		if err != nil {
			t.Fatal(err)
		}
		ok := false
		for _, c := range drawn {
			if c.Rarity >= constants.RaritySSR {
				ok = true
			}
		}
		if !ok {
			t.Fatalf("ten pull %d has no SSR or above: %+v", i, drawn)
		}
	}
}

func TestDrawNormalTenPullHasNoGuarantee(t *testing.T) {
	rule, _ := RuleOf(constants.NormalSummon)
	drawn, err := Draw(rule, NewCardPool(testCards()), TenPullTimes, zeroRand{})
	if err != nil || len(drawn) != 10 {
		t.Fatal(err, len(drawn))
	}
	for _, c := range drawn {
		if c.Rarity != constants.RarityR {
			t.Fatalf("unexpected %+v", c)
		}
	}
}

func TestDrawEmptyPool(t *testing.T) {
	rule, _ := RuleOf(constants.DiamondSummon)
	pool := NewCardPool(testCards()[:3]) // 沒有 SSR、UR
	if _, err := Draw(rule, pool, 1, zeroRand{}); !errors.Is(err, ErrEmptyCardPool) {
		t.Fatalf("err = %v, want ErrEmptyCardPool", err)
	}
	if _, err := Draw(rule, pool, 0, zeroRand{}); !errors.Is(err, ErrInvalidTimes) {
		t.Fatalf("err = %v, want ErrInvalidTimes", err)
	}
}

func TestResolve(t *testing.T) {
	c := map[string]card.Card{}
	for _, x := range testCards() {
		c[x.CardId] = x
	}
	owned := map[string]struct{}{"r1": {}}
	drawn := []card.Card{c["r1"], c["sr1"], c["sr1"], c["ur1"], c["r1"]}

	outcome, err := Resolve(drawn, owned)
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{"sr1", "ur1"}; !reflect.DeepEqual(outcome.NewCardIds, want) {
		t.Errorf("NewCardIds = %v, want %v", outcome.NewCardIds, want)
	}
	wantShards := map[string]int64{"shard_r": 60, "shard_sr": 30}
	if !reflect.DeepEqual(outcome.Shards, wantShards) {
		t.Errorf("Shards = %v, want %v", outcome.Shards, wantShards)
	}
	isNew := []bool{false, true, false, true, false}
	for i, r := range outcome.Results {
		if r.IsNew != isNew[i] {
			t.Errorf("result %d IsNew = %v, want %v", i, r.IsNew, isNew[i])
		}
		if r.IsNew != (r.Shard == ShardReward{}) {
			t.Errorf("result %d shard %+v inconsistent with IsNew", i, r.Shard)
		}
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
