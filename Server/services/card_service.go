package services

import (
	cardfeature "cardGame/feature/card"
	"cardGame/feature/growth"
	"cardGame/feature/inventory"
	"cardGame/model/cards"
	"cardGame/model/constants"
	"cardGame/model/packet"
	"context"
)

type (
	// CardGrowth 提供卡片的成長設定：升級經驗、階、星數與其費用。
	CardGrowth interface {
		Card(ctx context.Context) (*growth.CardGrowth, error)
	}

	CardService struct {
		inventoryRepo inventory.InventoryRepository
		growth        CardGrowth
	}

	// cardView 為卡片在指定狀態下的等級上限與升級費用。
	cardView struct {
		maxLevel  int
		limitBy   string
		nextExp   int64
		levels10  int
		nextExp10 int64
	}
)

func NewCardService(inventoryRepo inventory.InventoryRepository, growth CardGrowth) *CardService {
	return &CardService{inventoryRepo: inventoryRepo, growth: growth}
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

// expForLevels 回傳從 lv 級連續升 n 級所需的總經驗；任何一級超過上限時回傳 0。
func expForLevels(card *growth.CardGrowth, lv, n int) int64 {
	var total int64
	for i := 0; i < n; i++ {
		need := card.ExpToNext(lv + i)
		if need == 0 {
			return 0
		}
		total += need
	}
	return total
}

// limitByName 把等級上限的原因轉成封包用的名稱，還沒到上限時由呼叫端傳入空字串。
func limitByName(reason growth.LimitReason) string {
	switch reason {
	case growth.ReasonStage:
		return "stage"
	case growth.ReasonStar:
		return "star"
	case growth.ReasonMax:
		return "max"
	}
	return ""
}

// limitError 把已到等級上限的原因轉成對應的錯誤。
func limitError(reason growth.LimitReason) error {
	switch reason {
	case growth.ReasonStage:
		return inventory.ErrCardNeedAscend
	case growth.ReasonStar:
		return inventory.ErrCardNeedStar
	}
	return inventory.ErrCardMaxLevel
}

// viewOf 計算卡片在指定狀態下的等級上限與升級費用。
func viewOf(card *growth.CardGrowth, st inventory.CardState) cardView {
	limit := card.LimitFor(st.Rarity, st.Stage, st.Star)
	v := cardView{maxLevel: limit.MaxLevel}
	room := limit.MaxLevel - st.Level
	if room <= 0 {
		v.limitBy = limitByName(limit.Reason)
		return v
	}
	v.nextExp = card.ExpToNext(st.Level)
	v.levels10 = min(constants.CardUpgradeBatchLevels, room)
	v.nextExp10 = expForLevels(card, st.Level, v.levels10)
	return v
}

func stats(base cards.BaseConfig, level int) (attack, defense, health int) {
	return cardfeature.StatAtLevel(base.Attack, base.AttackRatio, level),
		cardfeature.StatAtLevel(base.Defense, base.DefenseRatio, level),
		cardfeature.StatAtLevel(base.Health, base.HealthRatio, level)
}

func costItems(costs []growth.Cost) []packet.CostItem {
	items := make([]packet.CostItem, 0, len(costs))
	for _, c := range costs {
		items = append(items, packet.CostItem{ItemId: c.ItemId, ItemName: c.ItemName, Quantity: c.Quantity})
	}
	return items
}

// progressResult 把變更後的狀態整理成進階與升星的回應。
func progressResult(card *growth.CardGrowth, cardId string, res *inventory.ChangeResult) *packet.CardProgressResult {
	v := viewOf(card, res.State)
	attack, defense, health := stats(res.Base, res.State.Level)
	return &packet.CardProgressResult{
		CardId:   cardId,
		Level:    res.State.Level,
		Star:     res.State.Star,
		Stage:    res.State.Stage,
		MaxLevel: v.maxLevel,
		LimitBy:  v.limitBy,
		Attack:   attack,
		Defense:  defense,
		Health:   health,
	}
}

// outer function

func (s *CardService) GetCards(ctx context.Context, uid string) (*packet.GetCardsResult, error) {
	card, err := s.growth.Card(ctx)
	if err != nil {
		return nil, err
	}
	cards, err := s.inventoryRepo.ListCards(ctx, uid)
	if err != nil {
		return nil, err
	}
	items, err := s.inventoryRepo.ListItems(ctx, uid)
	if err != nil {
		return nil, err
	}
	owned := make(map[string]int64, len(items))
	for _, it := range items {
		owned[it.ItemId] = it.Quantity
	}
	result := &packet.GetCardsResult{
		MaxLevel: card.MaxLevel(),
		ExpOwned: owned[constants.ItemIdCardExp],
		Items:    owned,
		Cards:    make([]packet.OwnedCard, 0, len(cards)),
	}
	for _, c := range cards {
		st := inventory.CardState{Rarity: c.Rarity, Level: int(c.Level), Star: c.Star, Stage: c.Stage}
		v := viewOf(card, st)
		attack, defense, health := stats(c.Base, int(c.Level))
		oc := packet.OwnedCard{
			CardId:     c.CardId,
			CardName:   c.CardName,
			Rarity:     c.Rarity,
			Race:       c.Race,
			Level:      c.Level,
			Star:       c.Star,
			Stage:      c.Stage,
			MaxStar:    card.MaxStar(c.Rarity),
			MaxLevel:   v.maxLevel,
			LimitBy:    v.limitBy,
			IsLocked:   c.IsLocked,
			ObtainedAt: c.ObtainedAt.Unix(),
			Attack:     attack,
			Defense:    defense,
			Health:     health,
			NextExp:    v.nextExp,
			NextExp10:  v.nextExp10,
			Levels10:   v.levels10,
			Skills:     cardSkills(c.Skill),
		}
		if item, _, ok := card.AscendCost(c.Stage); ok {
			oc.Ascend = &packet.AscendCost{
				Exp:   card.AscendExp(c.Stage),
				Items: costItems([]growth.Cost{item}),
			}
			oc.CanAscend = int(c.Level) >= card.StageMaxLevel(c.Stage)
		}
		if costs, ok := card.StarCost(c.Rarity, c.Star+1); ok && c.Star < card.MaxStar(c.Rarity) {
			oc.StarUp = costItems(costs)
		}
		result.Cards = append(result.Cards, oc)
	}
	return result, nil
}

// UpgradeCard 消耗經驗道具將指定卡片升 1 級或 10 級；升級的級數受目前的等級上限截斷，
// 例如 21 級要升 10 級而上限是 30，就只升 9 級並只扣 9 級的經驗。
func (s *CardService) UpgradeCard(ctx context.Context, uid string, data *packet.UpgradeCard) (*packet.UpgradeCardResult, error) {
	if data == nil || data.CardId == "" {
		return nil, inventory.ErrCardNotOwned
	}
	levels := data.Levels
	if levels == 0 {
		levels = 1
	}
	if levels != 1 && levels != constants.CardUpgradeBatchLevels {
		return nil, inventory.ErrInvalidUpgradeLevels
	}
	card, err := s.growth.Card(ctx)
	if err != nil {
		return nil, err
	}

	var gained int
	res, err := s.inventoryRepo.ApplyCardChange(ctx, uid, data.CardId, func(st inventory.CardState) (inventory.CardChange, error) {
		limit := card.LimitFor(st.Rarity, st.Stage, st.Star)
		room := limit.MaxLevel - st.Level
		if room <= 0 {
			return inventory.CardChange{}, limitError(limit.Reason)
		}
		gained = min(levels, room)
		need := expForLevels(card, st.Level, gained)
		if need == 0 {
			return inventory.CardChange{}, inventory.ErrCardMaxLevel
		}
		return inventory.CardChange{
			Levels: gained,
			Costs:  []inventory.ItemCost{{ItemId: constants.ItemIdCardExp, Quantity: need}},
		}, nil
	})
	if err != nil {
		return nil, err
	}
	v := viewOf(card, res.State)
	attack, defense, health := stats(res.Base, res.State.Level)
	return &packet.UpgradeCardResult{
		CardId:       data.CardId,
		Level:        int32(res.State.Level),
		Levels:       gained,
		Star:         res.State.Star,
		Stage:        res.State.Stage,
		MaxLevel:     v.maxLevel,
		LimitBy:      v.limitBy,
		Attack:       attack,
		Defense:      defense,
		Health:       health,
		NextExp:      v.nextExp,
		NextExp10:    v.nextExp10,
		Levels10:     v.levels10,
		ExpRemaining: res.Remaining[constants.ItemIdCardExp],
	}, nil
}

// AscendCard 消耗進階丹與離開那一階的部分經驗，將卡片進階一階；進階需要等級已到當前階的上限。
func (s *CardService) AscendCard(ctx context.Context, uid string, data *packet.AscendCard) (*packet.CardProgressResult, error) {
	if data == nil || data.CardId == "" {
		return nil, inventory.ErrCardNotOwned
	}
	card, err := s.growth.Card(ctx)
	if err != nil {
		return nil, err
	}
	res, err := s.inventoryRepo.ApplyCardChange(ctx, uid, data.CardId, func(st inventory.CardState) (inventory.CardChange, error) {
		item, _, ok := card.AscendCost(st.Stage)
		if !ok {
			return inventory.CardChange{}, inventory.ErrAlreadyTopStage
		}
		if st.Level < card.StageMaxLevel(st.Stage) {
			return inventory.CardChange{}, inventory.ErrAscendNotReady
		}
		return inventory.CardChange{
			Stages: 1,
			Costs: []inventory.ItemCost{
				{ItemId: constants.ItemIdCardExp, Quantity: card.AscendExp(st.Stage)},
				{ItemId: item.ItemId, Quantity: item.Quantity},
			},
		}, nil
	})
	if err != nil {
		return nil, err
	}
	return progressResult(card, data.CardId, res), nil
}

// StarUpCard 消耗碎片將卡片升一星，星數不能超過稀有度的上限。
func (s *CardService) StarUpCard(ctx context.Context, uid string, data *packet.StarUpCard) (*packet.CardProgressResult, error) {
	if data == nil || data.CardId == "" {
		return nil, inventory.ErrCardNotOwned
	}
	card, err := s.growth.Card(ctx)
	if err != nil {
		return nil, err
	}
	res, err := s.inventoryRepo.ApplyCardChange(ctx, uid, data.CardId, func(st inventory.CardState) (inventory.CardChange, error) {
		if st.Star >= card.MaxStar(st.Rarity) {
			return inventory.CardChange{}, inventory.ErrStarMax
		}
		costs, ok := card.StarCost(st.Rarity, st.Star+1)
		if !ok {
			return inventory.CardChange{}, inventory.ErrStarMax
		}
		items := make([]inventory.ItemCost, 0, len(costs))
		for _, c := range costs {
			items = append(items, inventory.ItemCost{ItemId: c.ItemId, Quantity: c.Quantity})
		}
		return inventory.CardChange{Stars: 1, Costs: items}, nil
	})
	if err != nil {
		return nil, err
	}
	return progressResult(card, data.CardId, res), nil
}
