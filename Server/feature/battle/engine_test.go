package battle

import (
	"reflect"
	"testing"

	"cardGame/model/cards"
	"cardGame/model/constants"
)

func dmgSkill(name string, pct int, shape constants.Target) cards.SkillDefinition {
	return cards.SkillDefinition{Name: name, Target: shape, Effects: []cards.SkillEffect{{
		Type: constants.EffectTypeDamage, Side: constants.SideEnemy, DamageType: constants.DamageTypePhysical,
		ScaleStat: constants.StatAttack, Value: pct,
	}}}
}

func skills(normal, active cards.SkillDefinition) cards.SkillConfig {
	return cards.SkillConfig{NormalAtk: normal, ActiveSkill: active}
}

func uin(id string, pos, atk, def, hp int) UnitInput {
	return UnitInput{ID: id, Position: pos, Attack: atk, Defense: def, Health: hp,
		Skills: skills(dmgSkill("atk", 100, constants.TargetSingle), dmgSkill("big", 300, constants.TargetSingle))}
}

func eventsOf(r *Result, t EventType) []Event {
	var out []Event
	for _, e := range r.Events {
		if e.Type == t {
			out = append(out, e)
		}
	}
	return out
}

func TestActionOrderInterleavesAllyFirst(t *testing.T) {
	b, err := newBattle(Input{
		Ally:  []UnitInput{uin("a1", 1, 1, 0, 9), uin("a2", 2, 1, 0, 9), uin("a3", 3, 1, 0, 9)},
		Enemy: []UnitInput{uin("e1", 1, 1, 0, 9), uin("e2", 4, 1, 0, 9)},
	})
	if err != nil {
		t.Fatal(err)
	}
	var got []string
	for _, u := range b.actionOrder() {
		got = append(got, u.id)
	}
	want := []string{"a1", "e1", "a2", "e2", "a3"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("order = %v, want %v", got, want)
	}

	b.enemyFirst = true
	got = nil
	for _, u := range b.actionOrder() {
		got = append(got, u.id)
	}
	want = []string{"e1", "a1", "e2", "a2", "a3"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("enemy first order = %v, want %v", got, want)
	}
}

func TestActionOrderSkipsDead(t *testing.T) {
	b, _ := newBattle(Input{
		Ally:  []UnitInput{uin("a1", 1, 1, 0, 9), uin("a2", 2, 1, 0, 9)},
		Enemy: []UnitInput{uin("e1", 1, 1, 0, 9), uin("e2", 2, 1, 0, 9)},
	})
	b.units[TeamAlly][0].health = 0
	var got []string
	for _, u := range b.actionOrder() {
		got = append(got, u.id)
	}
	if want := []string{"a2", "e1", "e2"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("order = %v, want %v", got, want)
	}
}

func TestRageCastsActiveWhenFull(t *testing.T) {
	// 初始 50：第 1 次行動普攻（+50 → 100），第 2 次行動主動（-100 → 0，+50）。
	r, err := Run(Input{
		Ally:  []UnitInput{uin("a", 1, 10, 0, 1000)},
		Enemy: []UnitInput{uin("e", 1, 1, 0, 1000)},
	})
	if err != nil {
		t.Fatal(err)
	}
	var kinds []EventType
	for _, e := range r.Events {
		if (e.Type == EventNormalAtk || e.Type == EventActiveSkill) && e.Actor.Team == TeamAlly {
			kinds = append(kinds, e.Type)
		}
	}
	if len(kinds) < 4 {
		t.Fatalf("kinds = %v", kinds)
	}
	want := []EventType{EventNormalAtk, EventActiveSkill, EventNormalAtk, EventActiveSkill}
	if !reflect.DeepEqual(kinds[:4], want) {
		t.Fatalf("kinds = %v, want %v", kinds[:4], want)
	}
}

func TestDamageFormula(t *testing.T) {
	// 普攻 100% × 攻擊 10 − 防禦 3 = 7；主動 300% × 10 − 3 = 27。
	r, _ := Run(Input{
		Ally:  []UnitInput{uin("a", 1, 10, 0, 1000)},
		Enemy: []UnitInput{uin("e", 1, 1, 3, 1000)},
	})
	d := eventsOf(r, EventDamage)
	if d[0].Value != 7 || d[0].Remain != 993 {
		t.Fatalf("first damage = %+v", d[0])
	}
	// d[1] 為敵方的普攻，d[2] 為我方主動技能。
	if d[2].Value != 27 {
		t.Fatalf("active damage = %+v", d[2])
	}
}

func TestDamageMinimumOneAndTrueIgnoresDefense(t *testing.T) {
	r, _ := Run(Input{
		Ally:  []UnitInput{uin("a", 1, 1, 0, 1000)},
		Enemy: []UnitInput{uin("e", 1, 1, 99, 1000)},
	})
	if v := eventsOf(r, EventDamage)[0].Value; v != 1 {
		t.Fatalf("min damage = %d", v)
	}

	trueSkill := dmgSkill("true", 100, constants.TargetSingle)
	trueSkill.Effects[0].DamageType = constants.DamageTypeTrue
	a := uin("a", 1, 10, 0, 1000)
	a.Skills = skills(trueSkill, trueSkill)
	r, _ = Run(Input{Ally: []UnitInput{a}, Enemy: []UnitInput{uin("e", 1, 1, 99, 1000)}})
	if v := eventsOf(r, EventDamage)[0].Value; v != 10 {
		t.Fatalf("true damage = %d", v)
	}
}

func TestWinnerAndDeath(t *testing.T) {
	r, _ := Run(Input{
		Ally:  []UnitInput{uin("a", 1, 100, 0, 1000)},
		Enemy: []UnitInput{uin("e", 1, 1, 0, 50)},
	})
	if r.Winner != TeamAlly || r.Rounds != 1 {
		t.Fatalf("winner = %v rounds = %d", r.Winner, r.Rounds)
	}
	if len(eventsOf(r, EventDeath)) != 1 {
		t.Fatalf("death events = %v", eventsOf(r, EventDeath))
	}
	last := r.Events[len(r.Events)-1]
	if last.Type != EventBattleEnd || last.Actor.Team != TeamAlly {
		t.Fatalf("last = %+v", last)
	}
}

func TestTimeoutDefenderWins(t *testing.T) {
	// 雙方都打不死對方：1 傷害 vs 極高血量。
	r, _ := Run(Input{
		Ally:  []UnitInput{uin("a", 1, 1, 0, 100000)},
		Enemy: []UnitInput{uin("e", 1, 1, 0, 100000)},
	})
	if r.Winner != TeamEnemy || r.Rounds != MaxRounds {
		t.Fatalf("winner = %v rounds = %d", r.Winner, r.Rounds)
	}
}

func TestFirstMoverDecidesTie(t *testing.T) {
	in := Input{
		Ally:  []UnitInput{uin("a", 1, 100, 0, 100)},
		Enemy: []UnitInput{uin("e", 1, 100, 0, 100)},
	}
	if r, _ := Run(in); r.Winner != TeamAlly {
		t.Fatalf("ally first winner = %v", r.Winner)
	}
	in.EnemyFirst = true
	if r, _ := Run(in); r.Winner != TeamEnemy {
		t.Fatalf("enemy first winner = %v", r.Winner)
	}
}

func TestSameSeedSameResult(t *testing.T) {
	random := dmgSkill("rnd", 100, constants.TargetRandom3)
	mk := func() Input {
		var ally, enemy []UnitInput
		for i := 1; i <= 6; i++ {
			a := uin("a", i, 30, 5, 400)
			a.Skills = skills(random, random)
			ally = append(ally, a)
			enemy = append(enemy, uin("e", i, 30, 5, 400))
		}
		return Input{Ally: ally, Enemy: enemy, Seed: 42}
	}
	r1, _ := Run(mk())
	r2, _ := Run(mk())
	if !reflect.DeepEqual(r1, r2) {
		t.Fatal("same seed produced different results")
	}
}

func TestInvalidInput(t *testing.T) {
	good := []UnitInput{uin("a", 1, 1, 0, 1)}
	cases := map[string]Input{
		"empty ally":    {Enemy: good},
		"empty enemy":   {Ally: good},
		"bad position":  {Ally: []UnitInput{uin("a", 7, 1, 0, 1)}, Enemy: good},
		"duplicate pos": {Ally: []UnitInput{uin("a", 1, 1, 0, 1), uin("b", 1, 1, 0, 1)}, Enemy: good},
		"zero health":   {Ally: []UnitInput{uin("a", 1, 1, 0, 0)}, Enemy: good},
		"too many":      {Ally: make([]UnitInput, 7), Enemy: good},
	}
	for name, in := range cases {
		if _, err := Run(in); err == nil {
			t.Errorf("%s: expected error", name)
		}
	}
}

func passive(trigger constants.Trigger, e cards.SkillEffect) cards.SkillDefinition {
	return cards.SkillDefinition{Name: "p", Trigger: trigger, Target: constants.TargetAll, Effects: []cards.SkillEffect{e}}
}

func TestBattleStartAndRoundEndPassives(t *testing.T) {
	rage := cards.SkillEffect{Type: constants.EffectTypeRage, Side: constants.SideSelf, Value: 25}
	a := uin("a", 1, 1, 0, 100000)
	a.Skills.PassiveSkill = passive(constants.TriggerBattleStart, rage)
	a.Skills.TalentSkill = []cards.SkillDefinition{passive(constants.TriggerRoundEnd, rage)}
	r, _ := Run(Input{Ally: []UnitInput{a}, Enemy: []UnitInput{uin("e", 1, 1, 0, 100000)}})

	rages := eventsOf(r, EventRage)
	if len(rages) != 1+MaxRounds {
		t.Fatalf("rage events = %d, want %d", len(rages), 1+MaxRounds)
	}
	if rages[0].Round != 0 || rages[0].Remain != constants.RageInitial+25 {
		t.Fatalf("battle start rage = %+v", rages[0])
	}
	// 戰鬥開始的被動讓初始怒氣變 75，第 1 次行動即 75 < 100 普攻，+50 = 125。
	if len(eventsOf(r, EventPassive)) != 1+MaxRounds {
		t.Fatalf("passive events = %d", len(eventsOf(r, EventPassive)))
	}
}

func TestBeforeAndAfterActionHooks(t *testing.T) {
	rage := cards.SkillEffect{Type: constants.EffectTypeRage, Side: constants.SideSelf, Value: 1}
	a := uin("a", 1, 1, 0, 100000)
	a.Skills.PassiveSkill = passive(constants.TriggerBeforeAction, rage)
	a.Skills.TalentSkill = []cards.SkillDefinition{passive(constants.TriggerAfterAction, rage)}
	r, _ := Run(Input{Ally: []UnitInput{a}, Enemy: []UnitInput{uin("e", 1, 1, 0, 100000)}})
	if got := len(eventsOf(r, EventRage)); got != 2*MaxRounds {
		t.Fatalf("rage events = %d, want %d", got, 2*MaxRounds)
	}
}

func TestTalentLockedUntilUnlocked(t *testing.T) {
	rage := cards.SkillEffect{Type: constants.EffectTypeRage, Side: constants.SideSelf, Value: 10}
	locked := passive(constants.TriggerBattleStart, rage)
	locked.Unlock = &cards.UnlockRule{Star: 5}
	a := uin("a", 1, 1, 0, 100000)
	a.Skills.TalentSkill = []cards.SkillDefinition{locked}
	in := Input{Ally: []UnitInput{a}, Enemy: []UnitInput{uin("e", 1, 1, 0, 100000)}}
	if r, _ := Run(in); len(eventsOf(r, EventRage)) != 0 {
		t.Fatal("locked talent fired")
	}
	in.Ally[0].Star = 5
	if r, _ := Run(in); len(eventsOf(r, EventRage)) != 1 {
		t.Fatal("unlocked talent did not fire")
	}
}

func TestHealCappedAtMaxHealth(t *testing.T) {
	heal := cards.SkillEffect{Type: constants.EffectTypeHeal, Side: constants.SideSelf, ScaleStat: constants.StatHealth, Value: 50}
	a := uin("a", 1, 1, 0, 100)
	a.Skills.PassiveSkill = passive(constants.TriggerRoundEnd, heal)
	// 敵人每次普攻 1 傷害（攻擊 1、我方防禦 0）。
	r, _ := Run(Input{Ally: []UnitInput{a}, Enemy: []UnitInput{uin("e", 1, 1, 0, 100000)}})
	h := eventsOf(r, EventHeal)
	if len(h) == 0 {
		t.Fatal("no heal events")
	}
	for _, e := range h {
		if e.Remain > 100 {
			t.Fatalf("heal over max: %+v", e)
		}
	}
}

func twoEffectSkill(shape constants.Target, firstPct, secondPct int) cards.SkillDefinition {
	eff := func(pct int) cards.SkillEffect {
		return cards.SkillEffect{Type: constants.EffectTypeDamage, Side: constants.SideEnemy,
			DamageType: constants.DamageTypeTrue, ScaleStat: constants.StatAttack, Value: pct}
	}
	return cards.SkillDefinition{Name: "two", Target: shape, Effects: []cards.SkillEffect{eff(firstPct), eff(secondPct)}}
}

func sixUnits(prefix string, hp int) []UnitInput {
	var out []UnitInput
	for i := 1; i <= 6; i++ {
		out = append(out, uin(prefix, i, 10, 0, hp))
	}
	return out
}

func TestCastReusesTargetsAcrossEffects(t *testing.T) {
	ally := sixUnits("a", 100000)
	ally[0].Skills = skills(twoEffectSkill(constants.TargetRandom3, 100, 100), twoEffectSkill(constants.TargetRandom3, 100, 100))
	r, _ := Run(Input{Ally: ally[:1], Enemy: sixUnits("e", 100000), Seed: 7})

	var first []Event
	for _, e := range r.Events {
		if e.Type == EventDamage && e.Actor.Team == TeamAlly {
			first = append(first, e)
			if len(first) == 6 {
				break
			}
		}
	}
	for i := 0; i < 3; i++ {
		if first[i].Target != first[i+3].Target {
			t.Fatalf("effect 2 targets differ: %v vs %v", first[:3], first[3:])
		}
	}
}

func TestCastSkipsDeadAndDoesNotReselect(t *testing.T) {
	// 效果 1 打死單體目標（血 5，傷害 10），效果 2 不應重選其他人，也不應再打已死的人。
	ally := sixUnits("a", 100000)
	ally[0].Skills = skills(twoEffectSkill(constants.TargetSingle, 100, 100), twoEffectSkill(constants.TargetSingle, 100, 100))
	enemy := sixUnits("e", 100000)
	enemy[0].Health = 5
	r, _ := Run(Input{Ally: ally[:1], Enemy: enemy})

	var firstCast []Event
	for _, e := range r.Events {
		if e.Actor.Team == TeamAlly && (e.Type == EventDamage || e.Type == EventDeath) {
			firstCast = append(firstCast, e)
		}
		if e.Actor.Team == TeamEnemy && e.Type == EventNormalAtk {
			break
		}
	}
	if len(firstCast) != 2 || firstCast[0].Type != EventDamage || firstCast[1].Type != EventDeath {
		t.Fatalf("first cast events = %+v", firstCast)
	}
}
