package services

import (
	"cardGame/feature/battle"
	"cardGame/model/cards"
	"cardGame/model/constants"
	"cardGame/model/packet"
	"context"
	"errors"
	"sort"
	"testing"
)

type fakeLineupRepo struct {
	slots []battle.LineupSlot
	err   error
}

func (f *fakeLineupRepo) GetLineup(context.Context, string) ([]battle.LineupSlot, error) {
	return f.slots, f.err
}

func (f *fakeLineupRepo) SaveLineup(_ context.Context, _ string, slots []battle.LineupSlot) error {
	if f.err != nil {
		return f.err
	}
	f.slots = append([]battle.LineupSlot(nil), slots...)
	sort.Slice(f.slots, func(i, j int) bool { return f.slots[i].Position < f.slots[j].Position })
	return nil
}

func newBattleSvc(t *testing.T) (*BattleService, *fakeInventoryRepo, *fakeLineupRepo) {
	t.Helper()
	inv := newFakeRepo()
	attack := cards.SkillDefinition{Name: "斬擊", Target: constants.TargetSingle, Effects: []cards.SkillEffect{{
		Type: constants.EffectTypeDamage, Side: constants.SideEnemy, DamageType: constants.DamageTypePhysical,
		ScaleStat: constants.StatAttack, Value: 3000,
	}}}
	for _, id := range []string{"c1", "c2", "c3"} {
		c := inv.addCard(id, constants.RarityR, 1, 0, 0)
		c.Skill = cards.SkillConfig{NormalAtk: attack, ActiveSkill: attack}
	}
	lineup := &fakeLineupRepo{}
	svc := NewBattleService(lineup, inv)
	svc.newSeed = func() int64 { return 1 }
	return svc, inv, lineup
}

func slots(pairs ...any) []packet.LineupSlot {
	var out []packet.LineupSlot
	for i := 0; i < len(pairs); i += 2 {
		out = append(out, packet.LineupSlot{Position: pairs[i].(int), CardId: pairs[i+1].(string)})
	}
	return out
}

func TestBattleServiceStages(t *testing.T) {
	svc, _, _ := newBattleSvc(t)
	res := svc.GetStages(context.Background())
	if len(res.Stages) != 1 || res.Stages[0].StageId != "mirror" {
		t.Fatalf("stages = %+v", res.Stages)
	}
}

func TestBattleServiceSaveLineupValidation(t *testing.T) {
	svc, inv, _ := newBattleSvc(t)
	ctx := context.Background()
	bad := map[string][]packet.LineupSlot{
		"empty":         {},
		"too many":      slots(1, "c1", 2, "c2", 3, "c3", 4, "x", 5, "y", 6, "z", 1, "w"),
		"position low":  slots(0, "c1"),
		"position high": slots(7, "c1"),
		"dup position":  slots(1, "c1", 1, "c2"),
		"dup card":      slots(1, "c1", 2, "c1"),
		"not owned":     slots(1, "nope"),
	}
	for name, s := range bad {
		err := svc.SaveLineup(ctx, "u", &packet.SaveLineup{Slots: s})
		if !errors.Is(err, battle.ErrLineupInvalid) {
			t.Errorf("%s: err = %v", name, err)
		}
	}
	zero := inv.addCard("zero", constants.RarityR, 1, 0, 0)
	zero.Base = cards.BaseConfig{}
	if err := svc.SaveLineup(ctx, "u", &packet.SaveLineup{Slots: slots(1, "zero")}); !errors.Is(err, battle.ErrLineupInvalid) {
		t.Errorf("zero health card err = %v", err)
	}
	if err := svc.SaveLineup(ctx, "u", nil); !errors.Is(err, battle.ErrLineupInvalid) {
		t.Errorf("nil request err = %v", err)
	}

	if err := svc.SaveLineup(ctx, "u", &packet.SaveLineup{Slots: slots(4, "c2", 1, "c1")}); err != nil {
		t.Fatal(err)
	}
	got, _ := svc.GetLineup(ctx, "u")
	if len(got.Slots) != 2 || got.Slots[0] != (packet.LineupSlot{Position: 1, CardId: "c1"}) || got.Slots[1].Position != 4 {
		t.Fatalf("lineup = %+v", got.Slots)
	}

	inv.err = errors.New("db down")
	if err := svc.SaveLineup(ctx, "u", &packet.SaveLineup{Slots: slots(1, "c1")}); err == nil || errors.Is(err, battle.ErrLineupInvalid) {
		t.Fatalf("repo error should pass through, got %v", err)
	}
}

func TestBattleServiceStartBattle(t *testing.T) {
	svc, _, _ := newBattleSvc(t)
	ctx := context.Background()

	if _, err := svc.StartBattle(ctx, "u", &packet.StartBattle{StageId: "nope"}); !errors.Is(err, battle.ErrStageNotFound) {
		t.Fatalf("unknown stage err = %v", err)
	}
	if _, err := svc.StartBattle(ctx, "u", nil); !errors.Is(err, battle.ErrStageNotFound) {
		t.Fatalf("nil err = %v", err)
	}
	if _, err := svc.StartBattle(ctx, "u", &packet.StartBattle{StageId: "mirror"}); !errors.Is(err, battle.ErrNoLineup) {
		t.Fatalf("no lineup err = %v", err)
	}

	if err := svc.SaveLineup(ctx, "u", &packet.SaveLineup{Slots: slots(1, "c1", 4, "c2")}); err != nil {
		t.Fatal(err)
	}
	res, err := svc.StartBattle(ctx, "u", &packet.StartBattle{StageId: "mirror"})
	if err != nil {
		t.Fatal(err)
	}
	if res.Winner != "Ally" || res.RageThreshold != constants.RageCastThreshold || res.Rounds < 1 {
		t.Fatalf("winner=%s rounds=%d threshold=%d", res.Winner, res.Rounds, res.RageThreshold)
	}
	if len(res.Units) != 4 {
		t.Fatalf("units = %+v", res.Units)
	}
	atk, def, hp := stats(testBase, 1)
	u := res.Units[0]
	if u.Team != "Ally" || u.Position != 1 || u.CardId != "c1" || u.Name != "c1" || u.MaxHealth != hp || u.Rage != constants.RageInitial {
		t.Fatalf("unit = %+v (stats %d %d %d)", u, atk, def, hp)
	}
	if res.Units[2].Team != "Enemy" || res.Units[2].MaxHealth != hp {
		t.Fatalf("enemy mirror = %+v", res.Units[2])
	}
	first, last := res.Events[0], res.Events[len(res.Events)-1]
	if first.Type != "BattleStart" || last.Type != "BattleEnd" || last.ActorTeam != "Ally" {
		t.Fatalf("first=%+v last=%+v", first, last)
	}
	for _, e := range res.Events {
		if e.Type == "Damage" && (e.ActorTeam == "" || e.TargetTeam == "" || e.Value < 1) {
			t.Fatalf("bad damage event %+v", e)
		}
	}
}

func TestBattleServiceStartBattleRepoErrors(t *testing.T) {
	svc, _, lineup := newBattleSvc(t)
	lineup.err = errors.New("db down")
	if _, err := svc.StartBattle(context.Background(), "u", &packet.StartBattle{StageId: "mirror"}); err == nil || errors.Is(err, battle.ErrNoLineup) {
		t.Fatalf("err = %v", err)
	}
	if _, err := svc.GetLineup(context.Background(), "u"); err == nil {
		t.Fatal("expected error")
	}
}
