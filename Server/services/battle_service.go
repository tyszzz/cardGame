package services

import (
	"cardGame/feature/battle"
	"cardGame/feature/inventory"
	"cardGame/model/constants"
	"cardGame/model/packet"
	"context"
	"fmt"
	"math/rand"
	"time"
)

type (
	BattleService struct {
		lineupRepo    battle.LineupRepository
		inventoryRepo inventory.InventoryRepository
		newSeed       func() int64
	}
)

// 目前只有鏡像：對手的陣容與三圍和玩家自己完全相同。
const stageMirror = "mirror"

var battleStages = []packet.BattleStage{
	{StageId: stageMirror, Name: "鏡像", Desc: "對手的陣容與自己完全相同"},
}

func NewBattleService(lineupRepo battle.LineupRepository, inventoryRepo inventory.InventoryRepository) *BattleService {
	return &BattleService{
		lineupRepo:    lineupRepo,
		inventoryRepo: inventoryRepo,
		newSeed:       func() int64 { return rand.New(rand.NewSource(time.Now().UnixNano())).Int63() },
	}
}

// inner function

func teamName(t battle.Team) string {
	if t == 0 {
		return ""
	}
	return t.String()
}

// validateSlots 驗證位置範圍、位置與卡片不重複、至少一張，且每張卡都是玩家持有的。
func validateSlots(slots []packet.LineupSlot, owned map[string]inventory.OwnedCard) ([]battle.LineupSlot, error) {
	if len(slots) == 0 || len(slots) > battle.MaxLineupSize {
		return nil, fmt.Errorf("%w: need 1~%d cards", battle.ErrLineupInvalid, battle.MaxLineupSize)
	}
	positions := make(map[int]bool, len(slots))
	cardIds := make(map[string]bool, len(slots))
	out := make([]battle.LineupSlot, 0, len(slots))
	for _, s := range slots {
		switch {
		case s.Position < 1 || s.Position > battle.MaxLineupSize:
			return nil, fmt.Errorf("%w: position %d out of range", battle.ErrLineupInvalid, s.Position)
		case positions[s.Position]:
			return nil, fmt.Errorf("%w: position %d duplicated", battle.ErrLineupInvalid, s.Position)
		case cardIds[s.CardId]:
			return nil, fmt.Errorf("%w: card %q duplicated", battle.ErrLineupInvalid, s.CardId)
		}
		c, ok := owned[s.CardId]
		if !ok {
			return nil, fmt.Errorf("%w: card %q not owned", battle.ErrLineupInvalid, s.CardId)
		}
		if _, _, health := stats(c.Base, int(c.Level)); health <= 0 {
			return nil, fmt.Errorf("%w: card %q has no health and cannot fight", battle.ErrLineupInvalid, s.CardId)
		}
		positions[s.Position] = true
		cardIds[s.CardId] = true
		out = append(out, battle.LineupSlot{Position: s.Position, CardId: s.CardId})
	}
	return out, nil
}

func (s *BattleService) ownedCards(ctx context.Context, uid string) (map[string]inventory.OwnedCard, error) {
	cards, err := s.inventoryRepo.ListCards(ctx, uid)
	if err != nil {
		return nil, err
	}
	owned := make(map[string]inventory.OwnedCard, len(cards))
	for _, c := range cards {
		owned[c.CardId] = c
	}
	return owned, nil
}

func unitInputs(slots []battle.LineupSlot, owned map[string]inventory.OwnedCard) ([]battle.UnitInput, error) {
	units := make([]battle.UnitInput, 0, len(slots))
	for _, slot := range slots {
		c, ok := owned[slot.CardId]
		if !ok {
			return nil, fmt.Errorf("%w: card %q not owned", battle.ErrLineupInvalid, slot.CardId)
		}
		attack, defense, health := stats(c.Base, int(c.Level))
		units = append(units, battle.UnitInput{
			ID:       c.CardId,
			Position: slot.Position,
			Attack:   attack,
			Defense:  defense,
			Health:   health,
			Star:     c.Star,
			Level:    int(c.Level),
			Skills:   c.Skill,
		})
	}
	return units, nil
}

// outer function

func (s *BattleService) GetStages(_ context.Context) *packet.GetStagesResult {
	return &packet.GetStagesResult{Stages: battleStages}
}

func (s *BattleService) GetLineup(ctx context.Context, uid string) (*packet.GetLineupResult, error) {
	slots, err := s.lineupRepo.GetLineup(ctx, uid)
	if err != nil {
		return nil, err
	}
	result := &packet.GetLineupResult{Slots: make([]packet.LineupSlot, 0, len(slots))}
	for _, slot := range slots {
		result.Slots = append(result.Slots, packet.LineupSlot{Position: slot.Position, CardId: slot.CardId})
	}
	return result, nil
}

func (s *BattleService) SaveLineup(ctx context.Context, uid string, data *packet.SaveLineup) error {
	if data == nil {
		return fmt.Errorf("%w: empty request", battle.ErrLineupInvalid)
	}
	owned, err := s.ownedCards(ctx, uid)
	if err != nil {
		return err
	}
	slots, err := validateSlots(data.Slots, owned)
	if err != nil {
		return err
	}
	return s.lineupRepo.SaveLineup(ctx, uid, slots)
}

// StartBattle 以玩家儲存的陣容挑戰關卡；三圍依持有卡片的目前等級計算，隨機種子由 server 產生。
func (s *BattleService) StartBattle(ctx context.Context, uid string, data *packet.StartBattle) (*packet.StartBattleResult, error) {
	if data == nil || data.StageId != stageMirror {
		return nil, battle.ErrStageNotFound
	}
	slots, err := s.lineupRepo.GetLineup(ctx, uid)
	if err != nil {
		return nil, err
	}
	if len(slots) == 0 {
		return nil, battle.ErrNoLineup
	}
	owned, err := s.ownedCards(ctx, uid)
	if err != nil {
		return nil, err
	}
	ally, err := unitInputs(slots, owned)
	if err != nil {
		return nil, err
	}
	enemy := append([]battle.UnitInput(nil), ally...)

	res, err := battle.Run(battle.Input{Ally: ally, Enemy: enemy, Seed: s.newSeed()})
	if err != nil {
		return nil, err
	}

	out := &packet.StartBattleResult{
		StageId:       data.StageId,
		Winner:        res.Winner.String(),
		Rounds:        res.Rounds,
		RageThreshold: constants.RageCastThreshold,
		Units:         make([]packet.BattleUnit, 0, len(res.Initial)),
		Events:        make([]packet.BattleEvent, 0, len(res.Events)),
	}
	for _, u := range res.Initial {
		c := owned[u.ID]
		out.Units = append(out.Units, packet.BattleUnit{
			Team: u.Ref.Team.String(), Position: u.Ref.Position, CardId: u.ID, Name: c.CardName, Rarity: c.Rarity,
			MaxHealth: u.MaxHealth, Health: u.Health, Rage: u.Rage,
		})
	}
	for _, e := range res.Events {
		out.Events = append(out.Events, packet.BattleEvent{
			Round: e.Round, Type: e.Type.String(),
			ActorTeam: teamName(e.Actor.Team), ActorPos: e.Actor.Position,
			TargetTeam: teamName(e.Target.Team), TargetPos: e.Target.Position,
			Skill: e.Skill, Value: e.Value, Remain: e.Remain,
		})
	}
	return out, nil
}
