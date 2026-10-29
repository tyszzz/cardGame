package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"cardGame/model/cards"
	"cardGame/model/constants"
)

func damageEffect() cards.SkillEffect {
	return cards.SkillEffect{
		Type:       constants.EffectTypeDamage,
		Side:       constants.SideEnemy,
		DamageType: constants.DamageTypePhysical,
		ScaleStat:  constants.StatAttack,
		Value:      150,
	}
}

func skill(name string, trigger constants.Trigger, effects ...cards.SkillEffect) cards.SkillDefinition {
	return cards.SkillDefinition{Name: name, Trigger: trigger, Target: constants.TargetSingle, Effects: effects}
}

func talent(name string) cards.SkillDefinition {
	s := skill(name, constants.TriggerRoundEnd, damageEffect())
	s.Unlock = &cards.UnlockRule{Star: 1}
	return s
}

func validDef() cards.CardDefinition {
	return cards.CardDefinition{
		CardID:   "card_001",
		CardName: "O'Brien",
		Rarity:   constants.RaritySR,
		Race:     constants.RaceDragon,
		BaseConfig: cards.BaseConfig{
			Attack: 10, AttackRatio: 0.1, Defense: 5, DefenseRatio: 0.05, Health: 100, HealthRatio: 0.2,
		},
		SkillConfig: cards.SkillConfig{
			NormalAtk:    skill("normal", constants.TriggerBeforeAction, damageEffect()),
			ActiveSkill:  skill("active", constants.TriggerAfterAction, damageEffect()),
			PassiveSkill: skill("passive", constants.TriggerBeforeRoundStart, damageEffect()),
		},
		IsEnabled: true,
	}
}

func parseDef(t *testing.T, def cards.CardDefinition) (*parsedCard, error) {
	t.Helper()
	data, err := json.Marshal(def)
	if err != nil {
		t.Fatal(err)
	}
	return parseCard(data)
}

func TestParseCard(t *testing.T) {
	valid, err := json.Marshal(validDef())
	if err != nil {
		t.Fatal(err)
	}
	validStr := string(valid)

	tests := []struct {
		name    string
		json    string
		wantErr string
	}{
		{"valid", validStr, ""},
		{"bad rarity", strings.Replace(validStr, `"SR"`, `"XX"`, 1), "Rarity"},
		{"missing rarity", strings.Replace(validStr, `"rarity":"SR",`, "", 1), "rarity"},
		{"bad race", strings.Replace(validStr, `"Dragon"`, `"Cat"`, 1), "Race"},
		{"unknown field", strings.Replace(validStr, `"cardName"`, `"cardNmae"`, 1), "unknown field"},
		{"removed condition field", strings.Replace(validStr, `"name":"normal"`, `"name":"normal","condition":{}`, 1), "unknown field"},
		{"bad trigger", strings.Replace(validStr, `"BeforeAction"`, `"Never"`, 1), "Trigger"},
		{"negative attack", strings.Replace(validStr, `"attack":10`, `"attack":-1`, 1), "negative"},
		{"negative ratio", strings.Replace(validStr, `"healthRatio":0.2`, `"healthRatio":-0.2`, 1), "negative"},
		{"empty id", strings.Replace(validStr, `"card_001"`, `""`, 1), "cardId"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseCard([]byte(tt.json))
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("error = %v, want containing %q", err, tt.wantErr)
			}
		})
	}
}

func TestSkillRules(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(*cards.CardDefinition)
		wantErr string
	}{
		{"valid", func(d *cards.CardDefinition) {}, ""},
		{"missing normal atk", func(d *cards.CardDefinition) { d.SkillConfig.NormalAtk = cards.SkillDefinition{} }, "normalAtk"},
		{"missing active", func(d *cards.CardDefinition) { d.SkillConfig.ActiveSkill = cards.SkillDefinition{} }, "activeSkill"},
		{"missing passive", func(d *cards.CardDefinition) { d.SkillConfig.PassiveSkill = cards.SkillDefinition{} }, "passiveSkill"},
		{"no effects", func(d *cards.CardDefinition) { d.SkillConfig.ActiveSkill.Effects = nil }, "at least one effect"},
		{"SR with talent", func(d *cards.CardDefinition) { d.SkillConfig.TalentSkill = []cards.SkillDefinition{talent("t")} }, "requires 0 talents"},
		{"SSR with 2 talents", func(d *cards.CardDefinition) {
			d.Rarity = constants.RaritySSR
			d.SkillConfig.TalentSkill = []cards.SkillDefinition{talent("a"), talent("b")}
		}, "requires 3 talents"},
		{"SSR with 3 talents", func(d *cards.CardDefinition) {
			d.Rarity = constants.RaritySSR
			d.SkillConfig.TalentSkill = []cards.SkillDefinition{talent("a"), talent("b"), talent("c")}
		}, ""},
		{"UR with 3 talents", func(d *cards.CardDefinition) {
			d.Rarity = constants.RarityUR
			d.SkillConfig.TalentSkill = []cards.SkillDefinition{talent("a"), talent("b"), talent("c")}
		}, "requires 4 talents"},
		{"talent without unlock", func(d *cards.CardDefinition) {
			d.Rarity = constants.RaritySSR
			t := talent("c")
			t.Unlock = nil
			d.SkillConfig.TalentSkill = []cards.SkillDefinition{talent("a"), talent("b"), t}
		}, "requires unlock"},
		{"unlock on non-talent", func(d *cards.CardDefinition) { d.SkillConfig.NormalAtk.Unlock = &cards.UnlockRule{Star: 1} }, "only allowed on talent"},
		{"damage without damageType", func(d *cards.CardDefinition) { d.SkillConfig.NormalAtk.Effects[0].DamageType = 0 }, "damageType is required"},
		{"damage without scaleStat", func(d *cards.CardDefinition) { d.SkillConfig.NormalAtk.Effects[0].ScaleStat = 0 }, "scaleStat is required"},
		{"damage with duration", func(d *cards.CardDefinition) { d.SkillConfig.NormalAtk.Effects[0].Duration = 2 }, "duration is not allowed"},
		{"damage zero value", func(d *cards.CardDefinition) { d.SkillConfig.NormalAtk.Effects[0].Value = 0 }, "value is required"},
		{"no side", func(d *cards.CardDefinition) { d.SkillConfig.NormalAtk.Effects[0].Side = 0 }, "invalid side"},
		{"no target anywhere", func(d *cards.CardDefinition) { d.SkillConfig.NormalAtk.Target = 0 }, "target is required"},
		{"effect target overrides", func(d *cards.CardDefinition) {
			d.SkillConfig.NormalAtk.Target = 0
			d.SkillConfig.NormalAtk.Effects[0].Target = constants.TargetAll
		}, ""},
		{"self with target", func(d *cards.CardDefinition) {
			d.SkillConfig.NormalAtk.Effects[0].Side = constants.SideSelf
			d.SkillConfig.NormalAtk.Effects[0].Target = constants.TargetAll
		}, "Self"},
		{"self without target", func(d *cards.CardDefinition) {
			d.SkillConfig.NormalAtk.Target = 0
			d.SkillConfig.NormalAtk.Effects[0].Side = constants.SideSelf
		}, ""},
		{"heal ok", func(d *cards.CardDefinition) {
			d.SkillConfig.ActiveSkill.Effects = []cards.SkillEffect{{
				Type: constants.EffectTypeHeal, Side: constants.SideAlly, ScaleStat: constants.StatAttack, Value: 80,
			}}
		}, ""},
		{"buff negative ok", func(d *cards.CardDefinition) {
			d.SkillConfig.ActiveSkill.Effects = []cards.SkillEffect{{
				Type: constants.EffectTypeBuff, Side: constants.SideEnemy, Stat: constants.StatDefense, Value: -20, Duration: 2, MaxStack: 3,
			}}
		}, ""},
		{"buff without duration", func(d *cards.CardDefinition) {
			d.SkillConfig.ActiveSkill.Effects = []cards.SkillEffect{{
				Type: constants.EffectTypeBuff, Side: constants.SideEnemy, Stat: constants.StatDefense, Value: -20,
			}}
		}, "duration is required"},
		{"buff with scaleStat", func(d *cards.CardDefinition) {
			d.SkillConfig.ActiveSkill.Effects = []cards.SkillEffect{{
				Type: constants.EffectTypeBuff, Side: constants.SideEnemy, Stat: constants.StatDefense, Value: -20, Duration: 2, ScaleStat: constants.StatAttack,
			}}
		}, "scaleStat is not allowed"},
		{"stun ok", func(d *cards.CardDefinition) {
			d.SkillConfig.ActiveSkill.Effects = []cards.SkillEffect{{
				Type: constants.EffectTypeStatus, Side: constants.SideEnemy, Status: constants.StatusStun, Duration: 1,
			}}
		}, ""},
		{"stun with value", func(d *cards.CardDefinition) {
			d.SkillConfig.ActiveSkill.Effects = []cards.SkillEffect{{
				Type: constants.EffectTypeStatus, Side: constants.SideEnemy, Status: constants.StatusStun, Duration: 1, Value: 10,
			}}
		}, "value is not allowed"},
		{"poison ok", func(d *cards.CardDefinition) {
			d.SkillConfig.ActiveSkill.Effects = []cards.SkillEffect{{
				Type: constants.EffectTypeStatus, Side: constants.SideEnemy, Status: constants.StatusPoison, Duration: 3, ScaleStat: constants.StatAttack, Value: 30,
			}}
		}, ""},
		{"poison without value", func(d *cards.CardDefinition) {
			d.SkillConfig.ActiveSkill.Effects = []cards.SkillEffect{{
				Type: constants.EffectTypeStatus, Side: constants.SideEnemy, Status: constants.StatusPoison, Duration: 3,
			}}
		}, "required"},
		{"status effect without status", func(d *cards.CardDefinition) {
			d.SkillConfig.ActiveSkill.Effects = []cards.SkillEffect{{
				Type: constants.EffectTypeStatus, Side: constants.SideEnemy, Duration: 3,
			}}
		}, "invalid status"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			def := validDef()
			tt.mutate(&def)
			_, err := parseDef(t, def)
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("error = %v, want containing %q", err, tt.wantErr)
			}
		})
	}
}

func TestSQLEscapesAndMapsEnums(t *testing.T) {
	c, err := parseDef(t, validDef())
	if err != nil {
		t.Fatal(err)
	}
	got := c.sql()
	for _, want := range []string{"'O''Brien'", "VALUES ('card_001', 'O''Brien', 2, 4,", `"trigger":"BeforeAction"`, "ON CONFLICT (card_id) DO UPDATE"} {
		if !strings.Contains(got, want) {
			t.Errorf("sql missing %q:\n%s", want, got)
		}
	}
}

func TestGenerateDuplicateID(t *testing.T) {
	data, err := json.Marshal(validDef())
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	for _, name := range []string{"a.json", "b.json"} {
		if err := os.WriteFile(filepath.Join(dir, name), data, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	err = generate(dir, filepath.Join(t.TempDir(), "cards.sql"))
	if err == nil || !strings.Contains(err.Error(), "duplicate") {
		t.Fatalf("error = %v, want duplicate", err)
	}
}
