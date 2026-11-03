package services

import (
	"cardGame/feature/inventory"
	"cardGame/model/constants"
	"context"
	"errors"
	"testing"
	"time"
)

type fakeInventoryRepo struct {
	items []inventory.BackpackItem
	cards []inventory.OwnedCard
	err   error
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
