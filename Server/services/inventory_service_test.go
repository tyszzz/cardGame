package services

import (
	"cardGame/feature/inventory"
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
}

func (f *fakeInventoryRepo) ListItems(context.Context, string) ([]inventory.BackpackItem, error) {
	return f.items, f.err
}

func (f *fakeInventoryRepo) ListCards(context.Context, string) ([]inventory.OwnedCard, error) {
	return f.cards, f.err
}

func TestInventoryServiceGetBackpack(t *testing.T) {
	s := NewInventoryService(&fakeInventoryRepo{items: []inventory.BackpackItem{
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
	s := NewInventoryService(&fakeInventoryRepo{})
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
	s := NewInventoryService(&fakeInventoryRepo{cards: []inventory.OwnedCard{
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
	s := NewInventoryService(&fakeInventoryRepo{err: boom})
	if _, err := s.GetBackpack(context.Background(), "u1"); !errors.Is(err, boom) {
		t.Errorf("GetBackpack err = %v", err)
	}
	if _, err := s.GetCards(context.Background(), "u1"); !errors.Is(err, boom) {
		t.Errorf("GetCards err = %v", err)
	}
}

func (f *fakeInventoryRepo) UpgradeCard(_ context.Context, _, _, expItemId string, cost inventory.LevelCostFunc) (*inventory.UpgradeResult, error) {
	if f.err != nil {
		return nil, f.err
	}
	need, err := cost(f.upgradeLevel)
	if err != nil {
		return nil, err
	}
	f.gotExpItem, f.gotNeed = expItemId, need
	return f.upgradeRes, nil
}

func TestInventoryServiceGetCardsStats(t *testing.T) {
	base := cards.BaseConfig{Attack: 500, AttackRatio: 1.2, Defense: 100, DefenseRatio: 1, Health: 8000, HealthRatio: 0.8}
	s := NewInventoryService(&fakeInventoryRepo{cards: []inventory.OwnedCard{
		{CardId: "a", Level: 2, Base: base},
		{CardId: "b", Level: constants.CardMaxLevel, Base: base},
	}})
	got, err := s.GetCards(context.Background(), "u1")
	if err != nil {
		t.Fatal(err)
	}
	if got.MaxLevel != constants.CardMaxLevel {
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
	s := NewInventoryService(&fakeInventoryRepo{cards: []inventory.OwnedCard{{CardId: "a", Level: 1, Skill: skill}}})
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
	s := NewInventoryService(repo)
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

func TestInventoryServiceUpgradeCardRejects(t *testing.T) {
	s := NewInventoryService(&fakeInventoryRepo{upgradeLevel: constants.CardMaxLevel})
	if _, err := s.UpgradeCard(context.Background(), "u1", &packet.UpgradeCard{CardId: "a"}); !errors.Is(err, inventory.ErrCardMaxLevel) {
		t.Errorf("max level err = %v", err)
	}
	for _, in := range []*packet.UpgradeCard{nil, {}} {
		if _, err := s.UpgradeCard(context.Background(), "u1", in); !errors.Is(err, inventory.ErrCardNotOwned) {
			t.Errorf("empty card id err = %v", err)
		}
	}
}
