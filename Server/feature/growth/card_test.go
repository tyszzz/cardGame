package growth

import (
	"cardGame/model/constants"
	"errors"
	"testing"
)

// 測試用設定：卡片經驗表 1~19 級（上限 20），第 lv 級升級需要 100+lv 經驗。
func validRaw() RawConfig {
	levelExp := make([]LevelRow, 0, 19)
	for lv := 1; lv <= 19; lv++ {
		levelExp = append(levelExp, LevelRow{Kind: KindCard, Level: lv, ExpRequired: int64(100 + lv)})
	}
	return RawConfig{
		LevelExp: levelExp,
		Stages: []StageRow{
			{Stage: 0, MaxLevel: 3},
			{Stage: 1, MaxLevel: 5, ItemId: "ascend_pill", ItemName: "進階丹", ItemQty: 500, ExpPercent: 50},
			{Stage: 2, MaxLevel: 0, ItemId: "ascend_pill", ItemName: "進階丹", ItemQty: 1000, ExpPercent: 50},
		},
		StarLevels: []StarLevelRow{{0, 5}, {1, 8}, {2, 12}},
		Rarities:   []RarityLimitRow{{constants.RarityR, 1}, {constants.RaritySR, 2}},
		StarCosts: []StarCostRow{
			{constants.RarityR, 1, "shard_r", "R 碎片", 50},
			{constants.RaritySR, 1, "shard_sr", "SR 碎片", 50},
			{constants.RaritySR, 1, "shard_r", "R 碎片", 100},
			{constants.RaritySR, 2, "shard_sr", "SR 碎片", 100},
			{constants.RaritySR, 2, "shard_r", "R 碎片", 150},
		},
	}
}

func mustBuildCard(t *testing.T) *CardGrowth {
	t.Helper()
	card, err := Build(validRaw()).Card()
	if err != nil {
		t.Fatal(err)
	}
	return card
}

func TestBuildCardRejectsInvalid(t *testing.T) {
	cases := map[string]func(*RawConfig){
		"stage gap":              func(r *RawConfig) { r.Stages[2].Stage = 3 },
		"single stage":           func(r *RawConfig) { r.Stages = r.Stages[:1] },
		"last stage has a cap":   func(r *RawConfig) { r.Stages[2].MaxLevel = 9 },
		"middle stage no cap":    func(r *RawConfig) { r.Stages[1].MaxLevel = 0 },
		"stage caps not rising":  func(r *RawConfig) { r.Stages[1].MaxLevel = 3 },
		"stage cap over table":   func(r *RawConfig) { r.Stages[1].MaxLevel = 21 },
		"ascend item missing":    func(r *RawConfig) { r.Stages[1].ItemId = "" },
		"ascend qty zero":        func(r *RawConfig) { r.Stages[2].ItemQty = 0 },
		"percent over 100":       func(r *RawConfig) { r.Stages[1].ExpPercent = 101 },
		"star gap":               func(r *RawConfig) { r.StarLevels[2].Star = 3 },
		"star levels not rising": func(r *RawConfig) { r.StarLevels[1].MaxLevel = 5 },
		"star cap over table":    func(r *RawConfig) { r.StarLevels[2].MaxLevel = 21 },
		"max star out of range":  func(r *RawConfig) { r.Rarities[0].MaxStar = 3 },
		"zero max star":          func(r *RawConfig) { r.Rarities[0].MaxStar = 0 },
		"cost qty zero":          func(r *RawConfig) { r.StarCosts[0].Quantity = 0 },
		"missing star cost":      func(r *RawConfig) { r.StarCosts = r.StarCosts[:3] },
		"no card exp table":      func(r *RawConfig) { r.LevelExp = nil },
		"card exp not rising":    func(r *RawConfig) { r.LevelExp[3].ExpRequired = 1 },
	}
	for name, mutate := range cases {
		raw := validRaw()
		mutate(&raw)
		if _, err := Build(raw).Card(); !errors.Is(err, ErrUnavailable) {
			t.Errorf("%s: err = %v, want ErrUnavailable", name, err)
		}
	}
}

func TestLimitFor(t *testing.T) {
	card := mustBuildCard(t)
	cases := []struct {
		rarity      constants.Rarity
		stage, star int
		want        Limit
	}{
		{constants.RarityR, 0, 0, Limit{3, ReasonStage}}, // 階上限 3 < 星上限 5
		{constants.RarityR, 1, 0, Limit{5, ReasonStage}}, // 階上限 5 == 星上限 5：先進階
		{constants.RarityR, 2, 0, Limit{5, ReasonStar}},  // 不限階，受 0 星限制
		{constants.RarityR, 2, 1, Limit{8, ReasonMax}},   // R 最大 1 星
		{constants.RaritySR, 2, 1, Limit{8, ReasonStar}}, // SR 還能升到 2 星
		{constants.RaritySR, 2, 2, Limit{12, ReasonMax}},
	}
	for _, c := range cases {
		if got := card.LimitFor(c.rarity, c.stage, c.star); got != c.want {
			t.Errorf("LimitFor(%v, stage %d, star %d) = %+v, want %+v", c.rarity, c.stage, c.star, got, c.want)
		}
	}
}

func TestExpToNextAndMaxLevel(t *testing.T) {
	card := mustBuildCard(t)
	if card.MaxLevel() != 20 {
		t.Errorf("MaxLevel = %d, want 20", card.MaxLevel())
	}
	for level, want := range map[int]int64{1: 101, 19: 119, 20: 0, 0: 0} {
		if got := card.ExpToNext(level); got != want {
			t.Errorf("ExpToNext(%d) = %d, want %d", level, got, want)
		}
	}
}

func TestAscendCostAndExp(t *testing.T) {
	card := mustBuildCard(t)

	item, percent, ok := card.AscendCost(0)
	if !ok || item.ItemId != "ascend_pill" || item.Quantity != 500 || percent != 50 {
		t.Errorf("AscendCost(0) = %+v, %d, %v", item, percent, ok)
	}
	if _, _, ok := card.AscendCost(card.TopStage()); ok {
		t.Error("the top stage cannot ascend")
	}

	// 0 階為 1~3 級：101 + 102 = 203，50% 向上取整為 102
	if got := card.AscendExp(0); got != 102 {
		t.Errorf("AscendExp(0) = %d, want 102", got)
	}
	// 1 階為 3~5 級：103 + 104 = 207，50% 向上取整為 104
	if got := card.AscendExp(1); got != 104 {
		t.Errorf("AscendExp(1) = %d, want 104", got)
	}
	if got := card.AscendExp(card.TopStage()); got != 0 {
		t.Errorf("AscendExp(top) = %d, want 0", got)
	}
}

func TestStarCostAndLookups(t *testing.T) {
	card := mustBuildCard(t)
	costs, ok := card.StarCost(constants.RaritySR, 2)
	if !ok || len(costs) != 2 {
		t.Fatalf("StarCost(SR, 2) = %+v, %v", costs, ok)
	}
	if _, ok := card.StarCost(constants.RarityR, 2); ok {
		t.Error("R has no star 2")
	}
	if card.MaxStar(constants.RaritySR) != 2 || card.MaxStar(constants.RarityUR) != 0 {
		t.Error("unexpected max star")
	}
	if card.StarLevelCap(1) != 8 || card.StarLevelCap(9) != 0 || card.StarLevelCap(-1) != 0 {
		t.Error("unexpected star level cap")
	}
	if card.StageMaxLevel(0) != 3 || card.StageMaxLevel(2) != 0 || card.StageMaxLevel(7) != 0 {
		t.Error("unexpected stage max level")
	}
}
