package services

import (
	"cardGame/feature/card"
	"cardGame/feature/inventory"
	"cardGame/model/cards"
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

// inner function

// cardSkills 依普攻、主動、被動、天賦的順序整理技能顯示資料，未設定名稱的技能略過。
func cardSkills(cfg cards.SkillConfig) []packet.CardSkill {
	skills := make([]packet.CardSkill, 0, 3+len(cfg.TalentSkill))
	add := func(skillType string, def cards.SkillDefinition) {
		if def.Name == "" {
			return
		}
		skill := packet.CardSkill{Type: skillType, Name: def.Name, Desc: def.Desc}
		if def.Unlock != nil {
			skill.UnlockStar = def.Unlock.Star
			skill.UnlockLevel = def.Unlock.Level
		}
		skills = append(skills, skill)
	}
	add("normalAtk", cfg.NormalAtk)
	add("activeSkill", cfg.ActiveSkill)
	add("passiveSkill", cfg.PassiveSkill)
	for _, talent := range cfg.TalentSkill {
		add("talent", talent)
	}
	return skills
}

// upgradeCost 為升一級所需經驗，已達上限時回傳 ErrCardMaxLevel。
func upgradeCost(level int) (int64, error) {
	need := card.ExpToNext(level)
	if need == 0 {
		return 0, inventory.ErrCardMaxLevel
	}
	return need, nil
}

// outer function

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
	result := &packet.GetCardsResult{MaxLevel: constants.CardMaxLevel, Cards: make([]packet.OwnedCard, 0, len(cards))}
	for _, c := range cards {
		result.Cards = append(result.Cards, packet.OwnedCard{
			CardId:     c.CardId,
			CardName:   c.CardName,
			Rarity:     c.Rarity,
			Race:       c.Race,
			Level:      c.Level,
			IsLocked:   c.IsLocked,
			ObtainedAt: c.ObtainedAt.Unix(),
			Attack:     card.StatAtLevel(c.Base.Attack, c.Base.AttackRatio, int(c.Level)),
			Defense:    card.StatAtLevel(c.Base.Defense, c.Base.DefenseRatio, int(c.Level)),
			Health:     card.StatAtLevel(c.Base.Health, c.Base.HealthRatio, int(c.Level)),
			NextExp:    card.ExpToNext(int(c.Level)),
			Skills:     cardSkills(c.Skill),
		})
	}
	return result, nil
}

// UpgradeCard 消耗經驗道具將指定卡片升一級。
func (s *InventoryService) UpgradeCard(ctx context.Context, uid string, data *packet.UpgradeCard) (*packet.UpgradeCardResult, error) {
	if data == nil || data.CardId == "" {
		return nil, inventory.ErrCardNotOwned
	}
	res, err := s.repo.UpgradeCard(ctx, uid, data.CardId, constants.ItemIdCardExp, upgradeCost)
	if err != nil {
		return nil, err
	}
	level := int(res.Level)
	return &packet.UpgradeCardResult{
		CardId:       data.CardId,
		Level:        res.Level,
		Attack:       card.StatAtLevel(res.Base.Attack, res.Base.AttackRatio, level),
		Defense:      card.StatAtLevel(res.Base.Defense, res.Base.DefenseRatio, level),
		Health:       card.StatAtLevel(res.Base.Health, res.Base.HealthRatio, level),
		NextExp:      card.ExpToNext(level),
		ExpRemaining: res.ExpRemaining,
	}, nil
}
