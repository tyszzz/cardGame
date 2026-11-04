package battle

import (
	"cardGame/model/cards"
	"cardGame/model/constants"
)

// cast 施放一個技能：依序處理每個效果，目標由效果的陣營與形狀決定（效果未填形狀時沿用技能的）。
func (b *battle) cast(caster *unit, skill cards.SkillDefinition, kind EventType) {
	b.emit(Event{Type: kind, Actor: caster.ref, Skill: skill.Name})
	// 同一次施放內，相同「陣營 + 形狀」只選一次目標，之後的效果沿用同一批人（已死亡的略過）。
	type key struct {
		side  constants.Side
		shape constants.Target
	}
	chosen := map[key][]*unit{}
	for _, eff := range skill.Effects {
		shape := eff.Target
		if shape == 0 {
			shape = skill.Target
		}
		k := key{eff.Side, shape}
		targets, ok := chosen[k]
		if !ok {
			targets = b.resolveTargets(caster, eff.Side, shape)
			chosen[k] = targets
		}
		for _, t := range targets {
			if t.alive() {
				b.applyEffect(caster, t, skill.Name, eff)
			}
		}
	}
}

// inner function

func (b *battle) applyEffect(caster, target *unit, skill string, e cards.SkillEffect) {
	switch e.Type {
	case constants.EffectTypeDamage:
		b.damage(caster, target, skill, e)
	case constants.EffectTypeHeal:
		b.heal(caster, target, skill, e)
	case constants.EffectTypeRage:
		b.addRage(caster, target, skill, e)
	default:
		// Buff、Status、Shield、Dispel、Cleanse 第一版尚未實作，之後在這裡接上。
	}
}

// scaled 回傳施放者 scaleStat（未填視為攻擊）的 value%。
func scaled(caster *unit, e cards.SkillEffect) int {
	stat := e.ScaleStat
	if stat == 0 {
		stat = constants.StatAttack
	}
	return caster.stat(stat) * e.Value / 100
}

func (b *battle) damage(caster, target *unit, skill string, e cards.SkillEffect) {
	dmg := scaled(caster, e)
	if e.DamageType != constants.DamageTypeTrue {
		dmg -= target.defense
	}
	if dmg < 1 {
		dmg = 1
	}
	target.health -= dmg
	if target.health < 0 {
		target.health = 0
	}
	b.emit(Event{Type: EventDamage, Actor: caster.ref, Target: target.ref, Skill: skill, Value: dmg, Remain: target.health})
	if !target.alive() {
		b.emit(Event{Type: EventDeath, Actor: caster.ref, Target: target.ref, Skill: skill})
	}
}

func (b *battle) heal(caster, target *unit, skill string, e cards.SkillEffect) {
	amount := scaled(caster, e)
	if room := target.maxHealth - target.health; amount > room {
		amount = room
	}
	if amount < 0 {
		amount = 0
	}
	target.health += amount
	b.emit(Event{Type: EventHeal, Actor: caster.ref, Target: target.ref, Skill: skill, Value: amount, Remain: target.health})
}

func (b *battle) addRage(caster, target *unit, skill string, e cards.SkillEffect) {
	target.rage += e.Value
	if target.rage < 0 {
		target.rage = 0
	}
	b.emit(Event{Type: EventRage, Actor: caster.ref, Target: target.ref, Skill: skill, Value: e.Value, Remain: target.rage})
}
