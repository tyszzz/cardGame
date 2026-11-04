package services

import (
	"cardGame/model/constants"
	"context"
	"errors"
	"testing"
)

func TestInventoryServiceGetBackpack(t *testing.T) {
	repo := newFakeRepo()
	repo.items["shard_r"] = 60
	got, err := NewInventoryService(repo).GetBackpack(context.Background(), "u1")
	if err != nil {
		t.Fatal(err)
	}
	if got.Capacity != constants.BackpackCapacity || got.Used != 1 || got.Items[0].ItemId != "shard_r" || got.Items[0].Quantity != 60 {
		t.Fatalf("unexpected result: %+v", got)
	}
}

func TestInventoryServiceEmpty(t *testing.T) {
	b, err := NewInventoryService(newFakeRepo()).GetBackpack(context.Background(), "u1")
	if err != nil || b.Used != 0 || b.Items == nil {
		t.Errorf("backpack = %+v, err %v (items must be non-nil)", b, err)
	}
}

func TestInventoryServiceRepoError(t *testing.T) {
	boom := errors.New("boom")
	repo := newFakeRepo()
	repo.err = boom
	if _, err := NewInventoryService(repo).GetBackpack(context.Background(), "u1"); !errors.Is(err, boom) {
		t.Errorf("GetBackpack err = %v", err)
	}
}
