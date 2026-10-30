package main

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"cardGame/model/cards"
	"cardGame/model/constants"
)

// talentCountByRarity 為各稀有度固定的天賦技能數量，未列出的稀有度視為 0。
var talentCountByRarity = map[constants.Rarity]int{
	constants.RaritySSR: 3,
	constants.RarityUR:  4,
}

// maxEffectNameRunes 為 Buff 與 Status 顯示名稱的長度上限（以字元計）。
const maxEffectNameRunes = 20

type effectField uint

const (
	fName effectField = 1 << iota
	fDamageType
	fStat
	fStatus
	fScaleStat
	fValue
	fDuration
	fMaxStack
	fPermanent
	fNonRemovable
	fCount
)

var effectFieldNames = []struct {
	bit  effectField
	name string
}{
	{fName, "name"},
	{fDamageType, "damageType"},
	{fStat, "stat"},
	{fStatus, "status"},
	{fScaleStat, "scaleStat"},
	{fValue, "value"},
	{fDuration, "duration"},
	{fMaxStack, "maxStack"},
	{fPermanent, "permanent"},
	{fNonRemovable, "nonRemovable"},
	{fCount, "count"},
}

// slotRule 描述技能欄位的差異：普攻與主動由行動流程和怒氣決定發動，不填 trigger；
// 被動與天賦由 trigger 決定發動時機；只有天賦需要解鎖條件。
type slotRule struct {
	trigger bool
	talent  bool
}

var (
	normalSlot  = slotRule{}
	activeSlot  = slotRule{}
	passiveSlot = slotRule{trigger: true}
	talentSlot  = slotRule{trigger: true, talent: true}
)

func validateSkills(rarity constants.Rarity, cfg *cards.SkillConfig) error {
	if err := validateSkill(&cfg.NormalAtk, normalSlot); err != nil {
		return fmt.Errorf("skillConfig.normalAtk: %w", err)
	}
	if err := validateSkill(&cfg.ActiveSkill, activeSlot); err != nil {
		return fmt.Errorf("skillConfig.activeSkill: %w", err)
	}
	if err := validateSkill(&cfg.PassiveSkill, passiveSlot); err != nil {
		return fmt.Errorf("skillConfig.passiveSkill: %w", err)
	}
	if want := talentCountByRarity[rarity]; len(cfg.TalentSkill) != want {
		return fmt.Errorf("skillConfig.talentSkill: %s requires %d talents, got %d", rarity, want, len(cfg.TalentSkill))
	}
	for i := range cfg.TalentSkill {
		if err := validateSkill(&cfg.TalentSkill[i], talentSlot); err != nil {
			return fmt.Errorf("skillConfig.talentSkill[%d]: %w", i, err)
		}
	}
	return nil
}

func validateSkill(s *cards.SkillDefinition, slot slotRule) error {
	if s.Name == "" {
		return fmt.Errorf("name must not be empty")
	}
	switch {
	case slot.trigger && !s.Trigger.Valid():
		return fmt.Errorf("invalid trigger: %d", s.Trigger)
	case !slot.trigger && s.Trigger != 0:
		return fmt.Errorf("trigger is not allowed on normalAtk/activeSkill")
	}
	if s.Target != 0 && !s.Target.Valid() {
		return fmt.Errorf("invalid target: %d", s.Target)
	}
	if err := validateUnlock(s.Unlock, slot.talent); err != nil {
		return err
	}
	if len(s.Effects) == 0 {
		return fmt.Errorf("at least one effect is required")
	}
	for i := range s.Effects {
		if err := validateEffect(&s.Effects[i], s.Target); err != nil {
			return fmt.Errorf("effects[%d]: %w", i, err)
		}
	}
	return nil
}

func validateUnlock(u *cards.UnlockRule, talent bool) error {
	if !talent {
		if u != nil {
			return fmt.Errorf("unlock is only allowed on talent skills")
		}
		return nil
	}
	if u == nil || (u.Star == 0 && u.Level == 0) {
		return fmt.Errorf("talent skill requires unlock star or level")
	}
	if u.Star < 0 || u.Level < 0 {
		return fmt.Errorf("unlock star/level must not be negative")
	}
	return nil
}

func validateEffect(e *cards.SkillEffect, skillTarget constants.Target) error {
	if !e.Type.Valid() {
		return fmt.Errorf("invalid type: %d", e.Type)
	}
	if !e.Side.Valid() {
		return fmt.Errorf("invalid side: %d", e.Side)
	}
	if e.Target != 0 && !e.Target.Valid() {
		return fmt.Errorf("invalid target: %d", e.Target)
	}
	switch {
	case e.Type == constants.EffectTypeDispel && e.Side != constants.SideEnemy:
		return fmt.Errorf("Dispel is only allowed when side is Enemy")
	case e.Type == constants.EffectTypeCleanse && e.Side == constants.SideEnemy:
		return fmt.Errorf("Cleanse is only allowed when side is Ally or Self")
	case e.Type == constants.EffectTypeShield && e.Side == constants.SideEnemy:
		return fmt.Errorf("Shield is only allowed when side is Ally or Self")
	}
	if e.Side == constants.SideSelf {
		if e.Target != 0 {
			return fmt.Errorf("target must not be set when side is Self")
		}
	} else if e.Target == 0 && skillTarget == 0 {
		return fmt.Errorf("target is required (effect or skill level) when side is %s", e.Side)
	}
	// 對位指正對面的敵人，對我方沒有意義
	effectiveTarget := e.Target
	if effectiveTarget == 0 {
		effectiveTarget = skillTarget
	}
	if e.Side == constants.SideAlly && effectiveTarget == constants.TargetOpposite {
		return fmt.Errorf("target Opposite is only allowed when side is Enemy")
	}
	if e.Duration < 0 || e.MaxStack < 0 || e.Count < 0 {
		return fmt.Errorf("duration/maxStack/count must not be negative")
	}
	if e.Permanent && e.Duration != 0 {
		return fmt.Errorf("duration must not be set when permanent")
	}

	required, optional, err := effectRule(e)
	if err != nil {
		return err
	}
	present := presentFields(e)
	for _, f := range effectFieldNames {
		switch {
		case required&f.bit != 0 && present&f.bit == 0:
			return fmt.Errorf("%s is required for %s effect", f.name, e.Type)
		case (required|optional)&f.bit == 0 && present&f.bit != 0:
			return fmt.Errorf("%s is not allowed for %s effect", f.name, e.Type)
		}
	}

	if present&fName != 0 {
		if strings.TrimSpace(e.Name) == "" {
			return fmt.Errorf("name must not be blank")
		}
		if n := utf8.RuneCountInString(e.Name); n > maxEffectNameRunes {
			return fmt.Errorf("name must be at most %d characters, got %d", maxEffectNameRunes, n)
		}
	}

	switch e.Type {
	case constants.EffectTypeDamage, constants.EffectTypeHeal, constants.EffectTypeShield:
		if e.Value <= 0 {
			return fmt.Errorf("value must be positive")
		}
	case constants.EffectTypeStatus:
		if e.Duration <= 0 {
			return fmt.Errorf("duration must be positive")
		}
		if present&fValue != 0 && e.Value <= 0 {
			return fmt.Errorf("value must be positive")
		}
	}
	if e.Type == constants.EffectTypeBuff || e.Type == constants.EffectTypeShield {
		if !e.Permanent && e.Duration <= 0 {
			return fmt.Errorf("duration must be positive")
		}
	}
	return nil
}

func effectRule(e *cards.SkillEffect) (required, optional effectField, err error) {
	switch e.Type {
	case constants.EffectTypeDamage:
		return fDamageType | fScaleStat | fValue, 0, nil
	case constants.EffectTypeHeal:
		return fScaleStat | fValue, 0, nil
	case constants.EffectTypeBuff:
		buffOptional := fMaxStack | fPermanent | fNonRemovable
		if e.Permanent { // 永久 Buff 不填 duration
			return fName | fStat | fValue, buffOptional, nil
		}
		return fName | fStat | fValue | fDuration, buffOptional, nil
	case constants.EffectTypeShield:
		// 額度 = 施放者 scaleStat 的 value%；永久護盾不填 duration
		shieldOptional := fPermanent | fNonRemovable
		if e.Permanent {
			return fName | fScaleStat | fValue, shieldOptional, nil
		}
		return fName | fScaleStat | fValue | fDuration, shieldOptional, nil
	case constants.EffectTypeDispel, constants.EffectTypeCleanse:
		return 0, fCount, nil // count 未填代表全部移除
	case constants.EffectTypeRage:
		return fValue, 0, nil // value 為固定怒氣點數，可為負
	case constants.EffectTypeStatus:
		if !e.Status.Valid() {
			return 0, 0, fmt.Errorf("invalid status: %d", e.Status)
		}
		if e.Status == constants.StatusStun {
			return fName | fStatus | fDuration, fMaxStack | fNonRemovable, nil
		}
		// 持續傷害類狀態需要傷害百分比與基準屬性
		return fName | fStatus | fDuration | fScaleStat | fValue, fMaxStack | fNonRemovable, nil
	}
	return 0, 0, fmt.Errorf("unsupported type: %s", e.Type)
}

func presentFields(e *cards.SkillEffect) effectField {
	var m effectField
	set := func(ok bool, bit effectField) {
		if ok {
			m |= bit
		}
	}
	set(e.Name != "", fName)
	set(e.DamageType != 0, fDamageType)
	set(e.Stat != 0, fStat)
	set(e.Status != 0, fStatus)
	set(e.ScaleStat != 0, fScaleStat)
	set(e.Value != 0, fValue)
	set(e.Duration != 0, fDuration)
	set(e.MaxStack != 0, fMaxStack)
	set(e.Permanent, fPermanent)
	set(e.NonRemovable, fNonRemovable)
	set(e.Count != 0, fCount)
	return m
}
