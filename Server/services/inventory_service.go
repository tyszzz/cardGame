package services

import (
	"cardGame/feature/inventory"
	"cardGame/model/constants"
	"cardGame/model/packet"
	"context"
)

type InventoryService struct {
	inventoryRepo inventory.InventoryRepository
}

func NewInventoryService(inventoryRepo inventory.InventoryRepository) *InventoryService {
	return &InventoryService{inventoryRepo: inventoryRepo}
}

// outer function

func (s *InventoryService) GetBackpack(ctx context.Context, uid string) (*packet.GetBackpackResult, error) {
	items, err := s.inventoryRepo.ListItems(ctx, uid)
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
