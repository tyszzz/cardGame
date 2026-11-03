package services

import (
	"cardGame/feature/inventory"
	"cardGame/model/constants"
	"cardGame/model/packet"
	"context"
)

type InventoryService struct {
	repo inventory.InventoryRepository
}

func NewInventoryService(repo inventory.InventoryRepository) *InventoryService {
	return &InventoryService{repo: repo}
}

func (s *InventoryService) GetBackpack(ctx context.Context, uid string) (*packet.GetBackpackResult, error) {
	items, err := s.repo.ListItems(ctx, uid)
	if err != nil {
		return nil, err
	}
	result := &packet.GetBackpackResult{
		Capacity: constants.BackpackCapacity,
		Used:     len(items),
		Items:    make([]packet.BackpackItem, 0, len(items)),
	}
	for _, it := range items {
		result.Items = append(result.Items, packet.BackpackItem{
			ItemId:   it.ItemId,
			ItemName: it.ItemName,
			ItemType: it.ItemType,
			Rarity:   it.Rarity,
			Quantity: it.Quantity,
		})
	}
	return result, nil
}

func (s *InventoryService) GetCards(ctx context.Context, uid string) (*packet.GetCardsResult, error) {
	cards, err := s.repo.ListCards(ctx, uid)
	if err != nil {
		return nil, err
	}
	result := &packet.GetCardsResult{Cards: make([]packet.OwnedCard, 0, len(cards))}
	for _, c := range cards {
		result.Cards = append(result.Cards, packet.OwnedCard{
			CardId:     c.CardId,
			CardName:   c.CardName,
			Rarity:     c.Rarity,
			Race:       c.Race,
			Level:      c.Level,
			IsLocked:   c.IsLocked,
			ObtainedAt: c.ObtainedAt.Unix(),
		})
	}
	return result, nil
}
