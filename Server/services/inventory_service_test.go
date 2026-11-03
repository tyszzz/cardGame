package services

import (
	"cardGame/feature/growth"
	"cardGame/feature/inventory"
	"cardGame/model/cards"
	"cardGame/model/constants"
	"cardGame/model/packet"
	"context"
	"errors"
	"testing"
	"time"
)

type (
	fakeCard struct {
		inventory.OwnedCard
		rarity constants.Rarity
	}

	// fakeInventoryRepo 在記憶體中模擬卡片與背包，ApplyCardChange 依 plan 決定費用後原子扣除。
	fakeInventoryRepo struct {
		items map[string]int64
		cards map[string]*fakeCard
		err   error
	}

	fakeGrowth struct {
		card *growth.CardGrowth
		err  error
	}
)

var testBase = cards.BaseConfig{Attack: 500, AttackRatio: 1.2, Defense: 100, DefenseRatio: 1, Health: 8000, HealthRatio: 0.8}

func newFakeRepo() *fakeInventoryRepo {
	return &fakeInventoryRepo{items: map[string]int64{}, cards: map[string]*fakeCard{}}
}

func (f *fakeInventoryRepo) addCard(id string, rarity constants.Rarity, lv, star, stage int) *fakeCard {
	c := &fakeCard{
		OwnedCard: inventory.OwnedCard{CardId: id, CardName: id, Rarity: rarity, Level: int32(lv), Star: star, Stage: stage, Base: testBase},
		rarity:    rarity,
	}
	f.cards[id] = c
	return c
}

func (f *fakeInventoryRepo) ListItems(context.Context, string) ([]inventory.BackpackItem, error) {
	if f.err != nil {
		return nil, f.err
	}
	items := make([]inventory.BackpackItem, 0, len(f.items))
	for id, qty := range f.items {
		if qty > 0 {
			items = append(items, inventory.BackpackItem{ItemId: id, ItemName: id, ItemType: constants.ItemTypeShard, Quantity: qty})
		}
	}
	return items, nil
}

func (f *fakeInventoryRepo) ListCards(context.Context, string) ([]inventory.OwnedCard, error) {
	if f.err != nil {
		return nil, f.err
	}
	list := make([]inventory.OwnedCard, 0, len(f.cards))
	for _, c := range f.cards {
		list = append(list, c.OwnedCard)
	}
	return list, nil
}

func (f *fakeInventoryRepo) ApplyCardChange(_ context.Context, _, cardId string, plan inventory.CardChangePlan) (*inventory.ChangeResult, error) {
	if f.err != nil {
		return nil, f.err
	}
	c, ok := f.cards[cardId]
	if !ok {
		return nil, inventory.ErrCardNotOwned
	}
	change, err := plan(inventory.CardState{Rarity: c.rarity, Level: int(c.Level), Star: c.Star, Stage: c.Stage})
	if err != nil {
		return nil, err
	}
	merged := map[string]int64{}
	for _, cost := range change.Costs {
		merged[cost.ItemId] += cost.Quantity
	}
	for id, need := range merged {
		if f.items[id] < need {
			return nil, inventory.ErrNotEnoughItems
		}
	}
	remaining := map[string]int64{}
	for id, need := range merged {
		f.items[id] -= need
		remaining[id] = f.items[id]
	}
	c.Level += int32(change.Levels)
	c.Star += change.Stars
	c.Stage += change.Stages
	return &inventory.ChangeResult{
		State:     inventory.CardState{Rarity: c.rarity, Level: int(c.Level), Star: c.Star, Stage: c.Stage},
		Base:      c.Base,
		Remaining: remaining,
	}, nil
}

func (f fakeGrowth) Card(context.Context) (*growth.CardGrowth, error) { return f.card, f.err }

// 測試用設定：卡片經驗表 1~19 級（上限 20），第 lv 級升級需要 100+lv 經驗；
// 0 階上限 5、1 階上限 10、2 階不限；星數上限 8/12/20 級；R 最多 1 星、SR 最多 2 星。
func testGrowth() *growth.CardGrowth {
	levelExp := make([]growth.LevelRow, 0, 19)
	for lv := 1; lv <= 19; lv++ {
		levelExp = append(levelExp, growth.LevelRow{Kind: growth.KindCard, Level: lv, ExpRequired: int64(100 + lv)})
	}
	card, err := growth.Build(growth.RawConfig{
		LevelExp: levelExp,
		Stages: []growth.StageRow{
			{Stage: 0, MaxLevel: 5},
			{Stage: 1, MaxLevel: 10, ItemId: "ascend_pill", ItemName: "進階丹", ItemQty: 5, ExpPercent: 50},
			{Stage: 2, ItemId: "ascend_pill", ItemName: "進階丹", ItemQty: 10, ExpPercent: 50},
		},
		StarLevels: []growth.StarLevelRow{{Star: 0, MaxLevel: 8}, {Star: 1, MaxLevel: 12}, {Star: 2, MaxLevel: 20}},
		Rarities:   []growth.RarityLimitRow{{Rarity: constants.RarityR, MaxStar: 1}, {Rarity: constants.RaritySR, MaxStar: 2}},
		StarCosts: []growth.StarCostRow{
			{Rarity: constants.RarityR, Star: 1, ItemId: "shard_r", ItemName: "R 碎片", Quantity: 50},
			{Rarity: constants.RaritySR, Star: 1, ItemId: "shard_sr", ItemName: "SR 碎片", Quantity: 50},
			{Rarity: constants.RaritySR, Star: 1, ItemId: "shard_r", ItemName: "R 碎片", Quantity: 100},
			{Rarity: constants.RaritySR, Star: 2, ItemId: "shard_sr", ItemName: "SR 碎片", Quantity: 100},
			{Rarity: constants.RaritySR, Star: 2, ItemId: "shard_r", ItemName: "R 碎片", Quantity: 150},
		},
	}).Card()
	if err != nil {
		panic(err)
	}
	return card
}

func newTestInventory(repo inventory.InventoryRepository) *InventoryService {
	return NewInventoryService(repo, fakeGrowth{card: testGrowth()})
}

func TestInventoryServiceGetBackpack(t *testing.T) {
	repo := newFakeRepo()
	repo.items["shard_r"] = 60
	got, err := newTestInventory(repo).GetBackpack(context.Background(), "u1")
	if err != nil {
		t.Fatal(err)
	}
	if got.Capacity != constants.BackpackCapacity || got.Used != 1 || got.Items[0].ItemId != "shard_r" || got.Items[0].Quantity != 60 {
		t.Fatalf("unexpected result: %+v", got)
	}
}

func TestInventoryServiceEmpty(t *testing.T) {
	s := newTestInventory(newFakeRepo())
	b, err := s.GetBackpack(context.Background(), "u1")
	if err != nil || b.Used != 0 || b.Items == nil {
		t.Errorf("backpack = %+v, err %v (items must be non-nil)", b, err)
	}
	c, err := s.GetCards(context.Background(), "u1")
	if err != nil || len(c.Cards) != 0 || c.Cards == nil || c.Items == nil {
		t.Errorf("cards = %+v, err %v (cards and items must be non-nil)", c, err)
	}
}

func TestInventoryServiceGetCardsBasics(t *testing.T) {
	repo := newFakeRepo()
	c := repo.addCard("xian_mage_ur", constants.RarityUR, 2, 0, 0)
	c.Race, c.IsLocked, c.ObtainedAt = constants.RaceXian, true, time.Unix(1700000000, 0)
	repo.items[constants.ItemIdCardExp] = 777

	got, err := newTestInventory(repo).GetCards(context.Background(), "u1")
	if err != nil {
		t.Fatal(err)
	}
	if got.ExpOwned != 777 || got.Items[constants.ItemIdCardExp] != 777 {
		t.Errorf("owned items: %+v, exp %d", got.Items, got.ExpOwned)
	}
	card := got.Cards[0]
	if card.CardId != "xian_mage_ur" || card.Level != 2 || !card.IsLocked || card.ObtainedAt != 1700000000 {
		t.Errorf("unexpected card: %+v", card)
	}
	if card.Attack != 1100 || card.Defense != 200 || card.Health != 14400 || card.NextExp != 102 {
		t.Errorf("stats at level 2: %+v", card)
	}
}

func TestInventoryServiceGetCardsLimits(t *testing.T) {
	repo := newFakeRepo()
	repo.addCard("a", constants.RaritySR, 3, 0, 0) // 0 階上限 5：還能升 2 級
	repo.addCard("b", constants.RaritySR, 5, 0, 0) // 到 0 階上限：需要進階
	repo.addCard("c", constants.RaritySR, 8, 0, 2) // 不限階，星 0 上限 8：需要升星
	repo.addCard("d", constants.RarityR, 12, 1, 2) // R 最多 1 星，上限 12：已滿
	got, err := newTestInventory(repo).GetCards(context.Background(), "u1")
	if err != nil {
		t.Fatal(err)
	}
	byID := map[string]packet.OwnedCard{}
	for _, c := range got.Cards {
		byID[c.CardId] = c
	}

	a := byID["a"]
	if a.MaxLevel != 5 || a.LimitBy != "" || a.Levels10 != 2 || a.NextExp10 != 103+104 {
		t.Errorf("a = %+v", a)
	}
	if a.Ascend == nil || a.Ascend.Items[0].ItemId != "ascend_pill" || a.Ascend.Items[0].Quantity != 5 {
		t.Errorf("a ascend = %+v", a.Ascend)
	}
	// 0 階為 1~5 級：101+102+103+104 = 410，50% = 205
	if a.Ascend.Exp != 205 {
		t.Errorf("a ascend exp = %d, want 205", a.Ascend.Exp)
	}
	if a.CanAscend {
		t.Error("a is below the stage cap, it cannot ascend yet")
	}
	if a.MaxStar != 2 || len(a.StarUp) != 2 {
		t.Errorf("a star up = %+v (maxStar %d)", a.StarUp, a.MaxStar)
	}

	if !byID["b"].CanAscend {
		t.Error("b is at the stage cap, it can ascend")
	}
	if b := byID["b"]; b.MaxLevel != 5 || b.LimitBy != "stage" || b.NextExp != 0 || b.Levels10 != 0 {
		t.Errorf("b = %+v", b)
	}
	if c := byID["c"]; c.MaxLevel != 8 || c.LimitBy != "star" || c.Ascend != nil {
		t.Errorf("c = %+v (top stage has no ascend cost)", c)
	}
	d := byID["d"]
	if d.MaxLevel != 12 || d.LimitBy != "max" || d.StarUp != nil || d.Ascend != nil {
		t.Errorf("d = %+v", d)
	}
}

func TestInventoryServiceUpgradeCard(t *testing.T) {
	repo := newFakeRepo()
	repo.addCard("a", constants.RarityR, 1, 0, 0)
	repo.items[constants.ItemIdCardExp] = 500
	got, err := newTestInventory(repo).UpgradeCard(context.Background(), "u1", &packet.UpgradeCard{CardId: "a"})
	if err != nil {
		t.Fatal(err)
	}
	if got.Level != 2 || got.Levels != 1 || got.ExpRemaining != 399 || got.Attack != 1100 || got.NextExp != 102 {
		t.Errorf("unexpected result: %+v", got)
	}
	if repo.items[constants.ItemIdCardExp] != 399 { // 1→2 級需要 101
		t.Errorf("exp left = %d, want 399", repo.items[constants.ItemIdCardExp])
	}
}

func TestInventoryServiceUpgradeCardClampsToCap(t *testing.T) {
	repo := newFakeRepo()
	repo.addCard("a", constants.RarityR, 3, 0, 0) // 0 階上限 5，連升 10 級只能升 2 級
	repo.items[constants.ItemIdCardExp] = 10_000
	got, err := newTestInventory(repo).UpgradeCard(context.Background(), "u1", &packet.UpgradeCard{CardId: "a", Levels: 10})
	if err != nil {
		t.Fatal(err)
	}
	if got.Level != 5 || got.Levels != 2 || got.LimitBy != "stage" {
		t.Errorf("unexpected result: %+v", got)
	}
	if want := int64(10_000 - 103 - 104); repo.items[constants.ItemIdCardExp] != want {
		t.Errorf("exp left = %d, want %d (only the 2 real levels are charged)", repo.items[constants.ItemIdCardExp], want)
	}
}

func TestInventoryServiceUpgradeCardBlockedAtCap(t *testing.T) {
	cases := []struct {
		name               string
		rarity             constants.Rarity
		level, star, stage int
		want               error
	}{
		{"stage cap", constants.RaritySR, 5, 0, 0, inventory.ErrCardNeedAscend},
		{"star cap", constants.RaritySR, 8, 0, 2, inventory.ErrCardNeedStar},
		{"fully maxed", constants.RarityR, 12, 1, 2, inventory.ErrCardMaxLevel},
	}
	for _, c := range cases {
		repo := newFakeRepo()
		repo.addCard("a", c.rarity, c.level, c.star, c.stage)
		repo.items[constants.ItemIdCardExp] = 10_000
		_, err := newTestInventory(repo).UpgradeCard(context.Background(), "u1", &packet.UpgradeCard{CardId: "a"})
		if !errors.Is(err, c.want) {
			t.Errorf("%s: err = %v, want %v", c.name, err, c.want)
		}
		if repo.items[constants.ItemIdCardExp] != 10_000 {
			t.Errorf("%s: nothing may be charged", c.name)
		}
	}
}

func TestInventoryServiceUpgradeCardRejects(t *testing.T) {
	repo := newFakeRepo()
	repo.addCard("a", constants.RarityR, 1, 0, 0)
	s := newTestInventory(repo)
	for _, levels := range []int{-1, 2, 9, 11} {
		if _, err := s.UpgradeCard(context.Background(), "u1", &packet.UpgradeCard{CardId: "a", Levels: levels}); !errors.Is(err, inventory.ErrInvalidUpgradeLevels) {
			t.Errorf("levels %d err = %v", levels, err)
		}
	}
	for _, in := range []*packet.UpgradeCard{nil, {}, {CardId: "missing"}} {
		if _, err := s.UpgradeCard(context.Background(), "u1", in); !errors.Is(err, inventory.ErrCardNotOwned) {
			t.Errorf("input %+v err = %v", in, err)
		}
	}
	if _, err := s.UpgradeCard(context.Background(), "u1", &packet.UpgradeCard{CardId: "a"}); !errors.Is(err, inventory.ErrNotEnoughItems) {
		t.Errorf("no exp err = %v", err)
	}
}

func TestInventoryServiceAscendCard(t *testing.T) {
	repo := newFakeRepo()
	repo.addCard("a", constants.RarityR, 5, 0, 0)
	repo.items[constants.ItemIdCardExp] = 1000
	repo.items["ascend_pill"] = 7
	s := newTestInventory(repo)

	got, err := s.AscendCard(context.Background(), "u1", &packet.AscendCard{CardId: "a"})
	if err != nil {
		t.Fatal(err)
	}
	if got.Stage != 1 || got.Level != 5 || got.MaxLevel != 8 || got.LimitBy != "" {
		t.Errorf("unexpected result: %+v", got)
	}
	// 1 階上限 10 但 0 星上限 8，所以等級上限變成 8（尚未到上限，LimitBy 為空）；經驗 205、進階丹 5
	if repo.items[constants.ItemIdCardExp] != 1000-205 || repo.items["ascend_pill"] != 2 {
		t.Errorf("items = %v", repo.items)
	}
	if up, err := s.UpgradeCard(context.Background(), "u1", &packet.UpgradeCard{CardId: "a"}); err != nil || up.Level != 6 {
		t.Errorf("after ascend the card must level up: %+v, %v", up, err)
	}
}

func TestInventoryServiceAscendCardRejects(t *testing.T) {
	repo := newFakeRepo()
	repo.addCard("low", constants.RarityR, 4, 0, 0)  // 還沒到 0 階上限
	repo.addCard("top", constants.RarityR, 12, 1, 2) // 已是最高階
	repo.addCard("ready", constants.RarityR, 5, 0, 0)
	repo.items[constants.ItemIdCardExp] = 100 // 少於 205
	repo.items["ascend_pill"] = 5
	s := newTestInventory(repo)

	for id, want := range map[string]error{
		"low":     inventory.ErrAscendNotReady,
		"top":     inventory.ErrAlreadyTopStage,
		"ready":   inventory.ErrNotEnoughItems,
		"missing": inventory.ErrCardNotOwned,
	} {
		if _, err := s.AscendCard(context.Background(), "u1", &packet.AscendCard{CardId: id}); !errors.Is(err, want) {
			t.Errorf("%s: err = %v, want %v", id, err, want)
		}
	}
	if repo.items["ascend_pill"] != 5 || repo.items[constants.ItemIdCardExp] != 100 || repo.cards["ready"].Stage != 0 {
		t.Errorf("failed ascends must not change anything: %v, stage %d", repo.items, repo.cards["ready"].Stage)
	}
	if _, err := s.AscendCard(context.Background(), "u1", nil); !errors.Is(err, inventory.ErrCardNotOwned) {
		t.Errorf("nil request err = %v", err)
	}
}

func TestInventoryServiceStarUpCard(t *testing.T) {
	repo := newFakeRepo()
	repo.addCard("a", constants.RaritySR, 1, 0, 0)
	repo.items["shard_sr"] = 120
	repo.items["shard_r"] = 300
	s := newTestInventory(repo)

	got, err := s.StarUpCard(context.Background(), "u1", &packet.StarUpCard{CardId: "a"})
	if err != nil {
		t.Fatal(err)
	}
	if got.Star != 1 || got.MaxLevel != 5 {
		t.Errorf("unexpected result: %+v", got)
	}
	if repo.items["shard_sr"] != 70 || repo.items["shard_r"] != 200 {
		t.Errorf("shards = %v", repo.items)
	}
	if got, err := s.StarUpCard(context.Background(), "u1", &packet.StarUpCard{CardId: "a"}); !errors.Is(err, inventory.ErrNotEnoughItems) {
		t.Errorf("star 2 needs 100 SR (have 70): %+v, %v", got, err)
	}
}

func TestInventoryServiceStarUpCardRejects(t *testing.T) {
	repo := newFakeRepo()
	repo.addCard("max", constants.RarityR, 1, 1, 0)
	repo.addCard("poor", constants.RaritySR, 1, 0, 0)
	repo.items["shard_sr"] = 999 // 缺 R 碎片：必須整筆失敗，不能只扣一部分
	s := newTestInventory(repo)

	if _, err := s.StarUpCard(context.Background(), "u1", &packet.StarUpCard{CardId: "max"}); !errors.Is(err, inventory.ErrStarMax) {
		t.Errorf("max star err = %v", err)
	}
	if _, err := s.StarUpCard(context.Background(), "u1", &packet.StarUpCard{CardId: "poor"}); !errors.Is(err, inventory.ErrNotEnoughItems) {
		t.Errorf("poor err = %v", err)
	}
	if repo.items["shard_sr"] != 999 || repo.cards["poor"].Star != 0 {
		t.Errorf("nothing may change on failure: %v, star %d", repo.items, repo.cards["poor"].Star)
	}
	for _, in := range []*packet.StarUpCard{nil, {}, {CardId: "missing"}} {
		if _, err := s.StarUpCard(context.Background(), "u1", in); !errors.Is(err, inventory.ErrCardNotOwned) {
			t.Errorf("input %+v err = %v", in, err)
		}
	}
}

func TestInventoryServiceRulesUnavailable(t *testing.T) {
	repo := newFakeRepo()
	repo.addCard("a", constants.RarityR, 1, 0, 0)
	s := NewInventoryService(repo, fakeGrowth{err: growth.ErrUnavailable})
	if _, err := s.GetCards(context.Background(), "u1"); err == nil {
		t.Error("GetCards must fail")
	}
	if _, err := s.UpgradeCard(context.Background(), "u1", &packet.UpgradeCard{CardId: "a"}); err == nil {
		t.Error("UpgradeCard must fail")
	}
	if _, err := s.AscendCard(context.Background(), "u1", &packet.AscendCard{CardId: "a"}); err == nil {
		t.Error("AscendCard must fail")
	}
	if _, err := s.StarUpCard(context.Background(), "u1", &packet.StarUpCard{CardId: "a"}); err == nil {
		t.Error("StarUpCard must fail")
	}
}

func TestInventoryServiceRepoError(t *testing.T) {
	boom := errors.New("boom")
	repo := newFakeRepo()
	repo.err = boom
	s := newTestInventory(repo)
	if _, err := s.GetBackpack(context.Background(), "u1"); !errors.Is(err, boom) {
		t.Errorf("GetBackpack err = %v", err)
	}
	if _, err := s.GetCards(context.Background(), "u1"); !errors.Is(err, boom) {
		t.Errorf("GetCards err = %v", err)
	}
}

func TestInventoryServiceGetCardsSkills(t *testing.T) {
	skill := cards.SkillConfig{
		NormalAtk:    cards.SkillDefinition{Name: "普攻", Desc: "n"},
		ActiveSkill:  cards.SkillDefinition{Name: "主動", Desc: "a"},
		PassiveSkill: cards.SkillDefinition{},
		TalentSkill:  []cards.SkillDefinition{{Name: "天賦一", Desc: "t1", Unlock: &cards.UnlockRule{Star: 3}}, {Name: "天賦二", Desc: "t2", Unlock: &cards.UnlockRule{Level: 10}}},
	}
	repo := newFakeRepo()
	repo.addCard("a", constants.RarityR, 1, 0, 0).Skill = skill
	got, err := newTestInventory(repo).GetCards(context.Background(), "u1")
	if err != nil {
		t.Fatal(err)
	}
	want := []packet.CardSkill{
		{Type: "normalAtk", Name: "普攻", Desc: "n"},
		{Type: "activeSkill", Name: "主動", Desc: "a"},
		{Type: "talent", Name: "天賦一", Desc: "t1", UnlockStar: 3},
		{Type: "talent", Name: "天賦二", Desc: "t2", UnlockLevel: 10},
	}
	if len(got.Cards[0].Skills) != len(want) {
		t.Fatalf("skills = %+v", got.Cards[0].Skills)
	}
	for i, w := range want {
		if got.Cards[0].Skills[i] != w {
			t.Errorf("skill %d = %+v, want %+v", i, got.Cards[0].Skills[i], w)
		}
	}
}
