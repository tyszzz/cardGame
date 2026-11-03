package services

import (
	"cardGame/feature/inventory"
	"cardGame/feature/level"
	"cardGame/model/cards"
	"cardGame/model/constants"
	"cardGame/model/packet"
	"context"
	"errors"
	"testing"
	"time"
)

type fakeInventoryRepo struct {
	items []inventory.BackpackItem
	cards []inventory.OwnedCard
	err   error

	upgradeLevel int // UpgradeCard 時卡片目前等級
	upgradeExp   int64
	upgradeRes   *inventory.UpgradeResult
	gotExpItem   string
	gotNeed      int64
	gotLevels    int
}

func (f *fakeInventoryRepo) ListItems(context.Context, string) ([]inventory.BackpackItem, error) {
	return f.items, f.err
}

func (f *fakeInventoryRepo) ListCards(context.Context, string) ([]inventory.OwnedCard, error) {
	return f.cards, f.err
}

func TestInventoryServiceGetBackpack(t *testing.T) {
	s := newTestInventory(&fakeInventoryRepo{items: []inventory.BackpackItem{
		{ItemId: "shard_r", ItemName: "R 碎片", ItemType: constants.ItemTypeShard, Rarity: constants.RarityR, Quantity: 60},
		{ItemId: "shard_sr", ItemName: "SR 碎片", ItemType: constants.ItemTypeShard, Rarity: constants.RaritySR, Quantity: 30},
	}})
	got, err := s.GetBackpack(context.Background(), "u1")
	if err != nil {
		t.Fatal(err)
	}
	if got.Capacity != constants.BackpackCapacity || got.Used != 2 || len(got.Items) != 2 {
		t.Fatalf("unexpected result: %+v", got)
	}
	if got.Items[0].ItemId != "shard_r" || got.Items[0].Quantity != 60 {
		t.Errorf("unexpected first item: %+v", got.Items[0])
	}
}

func TestInventoryServiceEmpty(t *testing.T) {
	s := newTestInventory(&fakeInventoryRepo{})
	b, err := s.GetBackpack(context.Background(), "u1")
	if err != nil || b.Used != 0 || b.Items == nil {
		t.Errorf("backpack = %+v, err %v (items must be non-nil)", b, err)
	}
	c, err := s.GetCards(context.Background(), "u1")
	if err != nil || len(c.Cards) != 0 || c.Cards == nil {
		t.Errorf("cards = %+v, err %v (cards must be non-nil)", c, err)
	}
}

func TestInventoryServiceGetCards(t *testing.T) {
	at := time.Unix(1700000000, 0)
	s := newTestInventory(&fakeInventoryRepo{cards: []inventory.OwnedCard{
		{CardId: "xian_mage_ur", CardName: "仙族UR法師", Rarity: constants.RarityUR, Race: constants.RaceXian, Level: 5, IsLocked: true, ObtainedAt: at},
	}})
	got, err := s.GetCards(context.Background(), "u1")
	if err != nil {
		t.Fatal(err)
	}
	c := got.Cards[0]
	if c.CardId != "xian_mage_ur" || c.Level != 5 || !c.IsLocked || c.ObtainedAt != 1700000000 {
		t.Errorf("unexpected card: %+v", c)
	}
}

func TestInventoryServiceError(t *testing.T) {
	boom := errors.New("boom")
	s := newTestInventory(&fakeInventoryRepo{err: boom})
	if _, err := s.GetBackpack(context.Background(), "u1"); !errors.Is(err, boom) {
		t.Errorf("GetBackpack err = %v", err)
	}
	if _, err := s.GetCards(context.Background(), "u1"); !errors.Is(err, boom) {
		t.Errorf("GetCards err = %v", err)
	}
}

func (f *fakeInventoryRepo) UpgradeCard(_ context.Context, _, _, expItemId string, levels int, cost inventory.LevelCostFunc) (*inventory.UpgradeResult, error) {
	if f.err != nil {
		return nil, f.err
	}
	var need int64
	for i := 0; i < levels; i++ {
		next, err := cost(f.upgradeLevel + i)
		if err != nil {
			return nil, err
		}
		need += next
	}
	f.gotExpItem, f.gotNeed, f.gotLevels = expItemId, need, levels
	return f.upgradeRes, nil
}

func TestInventoryServiceGetCardsStats(t *testing.T) {
	base := cards.BaseConfig{Attack: 500, AttackRatio: 1.2, Defense: 100, DefenseRatio: 1, Health: 8000, HealthRatio: 0.8}
	s := newTestInventory(&fakeInventoryRepo{cards: []inventory.OwnedCard{
		{CardId: "a", Level: 2, Base: base},
		{CardId: "b", Level: testMaxLevel, Base: base},
	}})
	got, err := s.GetCards(context.Background(), "u1")
	if err != nil {
		t.Fatal(err)
	}
	if got.MaxLevel != testMaxLevel {
		t.Errorf("MaxLevel = %d", got.MaxLevel)
	}
	a := got.Cards[0]
	if a.Attack != 1100 || a.Defense != 200 || a.Health != 14400 || a.NextExp != 120 {
		t.Errorf("level 2 card: %+v", a)
	}
	if b := got.Cards[1]; b.NextExp != 0 {
		t.Errorf("max level card must have no next exp: %+v", b)
	}
}

func TestInventoryServiceGetCardsSkills(t *testing.T) {
	skill := cards.SkillConfig{
		NormalAtk:    cards.SkillDefinition{Name: "普攻", Desc: "n"},
		ActiveSkill:  cards.SkillDefinition{Name: "主動", Desc: "a"},
		PassiveSkill: cards.SkillDefinition{},
		TalentSkill:  []cards.SkillDefinition{{Name: "天賦一", Desc: "t1", Unlock: &cards.UnlockRule{Star: 3}}, {Name: "天賦二", Desc: "t2", Unlock: &cards.UnlockRule{Level: 10}}},
	}
	s := newTestInventory(&fakeInventoryRepo{cards: []inventory.OwnedCard{{CardId: "a", Level: 1, Skill: skill}}})
	got, err := s.GetCards(context.Background(), "u1")
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

func TestInventoryServiceUpgradeCard(t *testing.T) {
	base := cards.BaseConfig{Attack: 500, AttackRatio: 1.2, Defense: 100, DefenseRatio: 1, Health: 8000, HealthRatio: 0.8}
	repo := &fakeInventoryRepo{upgradeLevel: 1, upgradeRes: &inventory.UpgradeResult{Level: 2, Base: base, ExpRemaining: 40}}
	s := newTestInventory(repo)
	got, err := s.UpgradeCard(context.Background(), "u1", &packet.UpgradeCard{CardId: "a"})
	if err != nil {
		t.Fatal(err)
	}
	if repo.gotExpItem != constants.ItemIdCardExp || repo.gotNeed != 100 {
		t.Errorf("charged %d of %q, want 100 of exp item", repo.gotNeed, repo.gotExpItem)
	}
	if got.Level != 2 || got.Attack != 1100 || got.NextExp != 120 || got.ExpRemaining != 40 {
		t.Errorf("unexpected result: %+v", got)
	}
}

func TestInventoryServiceUpgradeCardInvalidLevels(t *testing.T) {
	s := newTestInventory(&fakeInventoryRepo{upgradeLevel: 1})
	for _, levels := range []int{-1, 2, 9, 11} {
		if _, err := s.UpgradeCard(context.Background(), "u1", &packet.UpgradeCard{CardId: "a", Levels: levels}); !errors.Is(err, inventory.ErrInvalidUpgradeLevels) {
			t.Errorf("levels %d err = %v", levels, err)
		}
	}
}

func TestInventoryServiceUpgradeCardBatchBeyondMax(t *testing.T) {
	// 測試經驗表只有 4 級，從 1 級連升 10 級會超過上限
	s := newTestInventory(&fakeInventoryRepo{upgradeLevel: 1})
	if _, err := s.UpgradeCard(context.Background(), "u1", &packet.UpgradeCard{CardId: "a", Levels: 10}); !errors.Is(err, inventory.ErrCardMaxLevel) {
		t.Errorf("err = %v", err)
	}
}

func TestInventoryServiceUpgradeCardBatchCost(t *testing.T) {
	table, err := level.BuildTable(level.KindCard, tenLevelRows())
	if err != nil {
		t.Fatal(err)
	}
	base := cards.BaseConfig{Attack: 500, AttackRatio: 1.2}
	repo := &fakeInventoryRepo{upgradeLevel: 1, upgradeRes: &inventory.UpgradeResult{Level: 11, Base: base, ExpRemaining: 5}}
	s := NewInventoryService(repo, fakeCardLevels{table: table})
	got, err := s.UpgradeCard(context.Background(), "u1", &packet.UpgradeCard{CardId: "a", Levels: 10})
	if err != nil {
		t.Fatal(err)
	}
	// 1→2 到 10→11 依序 101~110，共 1055
	if repo.gotLevels != 10 || repo.gotNeed != 1055 {
		t.Errorf("levels = %d, need = %d", repo.gotLevels, repo.gotNeed)
	}
	if got.Level != 11 || got.ExpRemaining != 5 {
		t.Errorf("unexpected result: %+v", got)
	}
}

func TestInventoryServiceGetCardsExp(t *testing.T) {
	table, err := level.BuildTable(level.KindCard, tenLevelRows())
	if err != nil {
		t.Fatal(err)
	}
	repo := &fakeInventoryRepo{
		items: []inventory.BackpackItem{{ItemId: constants.ItemIdCardExp, Quantity: 1500}},
		cards: []inventory.OwnedCard{{CardId: "a", Level: 1}, {CardId: "b", Level: 2}},
	}
	got, err := NewInventoryService(repo, fakeCardLevels{table: table}).GetCards(context.Background(), "u1")
	if err != nil {
		t.Fatal(err)
	}
	if got.ExpOwned != 1500 {
		t.Errorf("ExpOwned = %d", got.ExpOwned)
	}
	// 1 級連升 10 級為 101~110，共 1055；2 級為 102~111，共 1065
	if got.Cards[0].NextExp10 != 1055 || got.Cards[1].NextExp10 != 1065 {
		t.Errorf("NextExp10 = %d, %d", got.Cards[0].NextExp10, got.Cards[1].NextExp10)
	}

	// 沒有經驗道具時 ExpOwned 為 0；接近上限時沒有 10 級的選項
	repo = &fakeInventoryRepo{cards: []inventory.OwnedCard{{CardId: "c", Level: 15}}}
	got, err = NewInventoryService(repo, fakeCardLevels{table: table}).GetCards(context.Background(), "u1")
	if err != nil {
		t.Fatal(err)
	}
	if got.ExpOwned != 0 || got.Cards[0].NextExp != 115 || got.Cards[0].NextExp10 != 0 {
		t.Errorf("unexpected result: %+v", got)
	}
}

// tenLevelRows 為 1~19 級的經驗表，第 lv 級升級需要 100+lv 經驗（等級上限 20）。
func tenLevelRows() []level.Row {
	rows := make([]level.Row, 0, 19)
	for lv := 1; lv <= 19; lv++ {
		rows = append(rows, level.Row{Kind: level.KindCard, Level: lv, ExpRequired: int64(100 + lv)})
	}
	return rows
}

func TestInventoryServiceUpgradeCardRejects(t *testing.T) {
	s := newTestInventory(&fakeInventoryRepo{upgradeLevel: testMaxLevel})
	if _, err := s.UpgradeCard(context.Background(), "u1", &packet.UpgradeCard{CardId: "a"}); !errors.Is(err, inventory.ErrCardMaxLevel) {
		t.Errorf("max level err = %v", err)
	}
	for _, in := range []*packet.UpgradeCard{nil, {}} {
		if _, err := s.UpgradeCard(context.Background(), "u1", in); !errors.Is(err, inventory.ErrCardNotOwned) {
			t.Errorf("empty card id err = %v", err)
		}
	}
}

// 測試用經驗表：1→2 需要 100、2→3 需要 120、3→4 需要 144，等級上限 4。
const testMaxLevel = 4

type fakeCardLevels struct {
	table *level.Table
	err   error
}

func (f fakeCardLevels) CardTable(context.Context) (*level.Table, error) { return f.table, f.err }

func TestInventoryServiceCardTableUnavailable(t *testing.T) {
	s := NewInventoryService(&fakeInventoryRepo{}, fakeCardLevels{err: level.ErrTableUnavailable})
	if _, err := s.GetCards(context.Background(), "u1"); !errors.Is(err, level.ErrTableUnavailable) {
		t.Errorf("GetCards err = %v", err)
	}
	if _, err := s.UpgradeCard(context.Background(), "u1", &packet.UpgradeCard{CardId: "a"}); !errors.Is(err, level.ErrTableUnavailable) {
		t.Errorf("UpgradeCard err = %v", err)
	}
}

func newTestInventory(repo inventory.InventoryRepository) *InventoryService {
	table, err := level.BuildTable(level.KindCard, []level.Row{
		{Kind: level.KindCard, Level: 1, ExpRequired: 100},
		{Kind: level.KindCard, Level: 2, ExpRequired: 120},
		{Kind: level.KindCard, Level: 3, ExpRequired: 144},
	})
	if err != nil {
		panic(err)
	}
	return NewInventoryService(repo, fakeCardLevels{table: table})
}
