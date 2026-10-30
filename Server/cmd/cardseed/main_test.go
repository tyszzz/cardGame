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

func buff(mutate func(*cards.SkillEffect)) cards.SkillEffect {
	e := cards.SkillEffect{Type: constants.EffectTypeBuff, Name: "n", Side: constants.SideAlly, Stat: constants.StatAttack, Value: 20, Duration: 2}
	mutate(&e)
	return e
}

func shield(mutate func(*cards.SkillEffect)) cards.SkillEffect {
	e := cards.SkillEffect{
		Type: constants.EffectTypeShield, Name: "聖光護盾", Side: constants.SideAlly,
		ScaleStat: constants.StatHealth, Value: 20, Duration: 3,
	}
	mutate(&e)
	return e
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
			NormalAtk:    skill("normal", 0, damageEffect()),
			ActiveSkill:  skill("active", 0, damageEffect()),
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
		{"bad trigger", strings.Replace(validStr, `"BeforeRoundStart"`, `"Never"`, 1), "Trigger"},
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
		{"normal atk with trigger", func(d *cards.CardDefinition) { d.SkillConfig.NormalAtk.Trigger = constants.TriggerAfterAction }, "trigger is not allowed"},
		{"active with trigger", func(d *cards.CardDefinition) { d.SkillConfig.ActiveSkill.Trigger = constants.TriggerBeforeAction }, "trigger is not allowed"},
		{"passive without trigger", func(d *cards.CardDefinition) { d.SkillConfig.PassiveSkill.Trigger = 0 }, "invalid trigger"},
		{"talent without trigger", func(d *cards.CardDefinition) {
			d.Rarity = constants.RaritySSR
			t := talent("c")
			t.Trigger = 0
			d.SkillConfig.TalentSkill = []cards.SkillDefinition{talent("a"), talent("b"), t}
		}, "invalid trigger"},
		{"battle start trigger ok", func(d *cards.CardDefinition) { d.SkillConfig.PassiveSkill.Trigger = constants.TriggerBattleStart }, ""},
		{"opposite on enemy ok", func(d *cards.CardDefinition) { d.SkillConfig.NormalAtk.Effects[0].Target = constants.TargetOpposite }, ""},
		{"opposite as skill default on enemy ok", func(d *cards.CardDefinition) { d.SkillConfig.NormalAtk.Target = constants.TargetOpposite }, ""},
		{"opposite on ally rejected", func(d *cards.CardDefinition) {
			d.SkillConfig.ActiveSkill.Effects = []cards.SkillEffect{{
				Type: constants.EffectTypeHeal, Side: constants.SideAlly, Target: constants.TargetOpposite, ScaleStat: constants.StatAttack, Value: 80,
			}}
		}, "Opposite is only allowed"},
		{"opposite skill default with ally effect rejected", func(d *cards.CardDefinition) {
			d.SkillConfig.ActiveSkill.Target = constants.TargetOpposite
			d.SkillConfig.ActiveSkill.Effects = []cards.SkillEffect{{
				Type: constants.EffectTypeHeal, Side: constants.SideAlly, ScaleStat: constants.StatAttack, Value: 80,
			}}
		}, "Opposite is only allowed"},
		{"opposite skill default with self effect ok", func(d *cards.CardDefinition) {
			d.SkillConfig.ActiveSkill.Target = constants.TargetOpposite
			d.SkillConfig.ActiveSkill.Effects = []cards.SkillEffect{{
				Type: constants.EffectTypeHeal, Side: constants.SideSelf, ScaleStat: constants.StatAttack, Value: 80,
			}}
		}, ""},
		{"buff missing name", func(d *cards.CardDefinition) {
			d.SkillConfig.ActiveSkill.Effects = []cards.SkillEffect{buff(func(e *cards.SkillEffect) { e.Name = "" })}
		}, "name is required for Buff"},
		{"buff blank name", func(d *cards.CardDefinition) {
			d.SkillConfig.ActiveSkill.Effects = []cards.SkillEffect{buff(func(e *cards.SkillEffect) { e.Name = "   " })}
		}, "name must not be blank"},
		{"buff name 20 chars ok", func(d *cards.CardDefinition) {
			d.SkillConfig.ActiveSkill.Effects = []cards.SkillEffect{buff(func(e *cards.SkillEffect) { e.Name = strings.Repeat("狂", 20) })}
		}, ""},
		{"buff name 21 chars", func(d *cards.CardDefinition) {
			d.SkillConfig.ActiveSkill.Effects = []cards.SkillEffect{buff(func(e *cards.SkillEffect) { e.Name = strings.Repeat("狂", 21) })}
		}, "at most 20 characters"},
		{"status missing name", func(d *cards.CardDefinition) {
			d.SkillConfig.ActiveSkill.Effects = []cards.SkillEffect{{
				Type: constants.EffectTypeStatus, Side: constants.SideEnemy, Status: constants.StatusStun, Duration: 1,
			}}
		}, "name is required for Status"},
		{"two poisons with different names ok", func(d *cards.CardDefinition) {
			poison := func(name string, value int) cards.SkillEffect {
				return cards.SkillEffect{
					Type: constants.EffectTypeStatus, Name: name, Side: constants.SideEnemy, Status: constants.StatusPoison,
					Duration: 3, ScaleStat: constants.StatAttack, Value: value,
				}
			}
			d.SkillConfig.ActiveSkill.Effects = []cards.SkillEffect{poison("A毒", 30), poison("B毒", 50)}
		}, ""},
		{"damage with name", func(d *cards.CardDefinition) { d.SkillConfig.NormalAtk.Effects[0].Name = "x" }, "name is not allowed"},
		{"rage with name", func(d *cards.CardDefinition) {
			d.SkillConfig.NormalAtk.Effects = []cards.SkillEffect{{Type: constants.EffectTypeRage, Name: "x", Side: constants.SideSelf, Value: 20}}
		}, "name is not allowed"},
		{"dispel with name", func(d *cards.CardDefinition) {
			d.SkillConfig.ActiveSkill.Effects = []cards.SkillEffect{{Type: constants.EffectTypeDispel, Name: "x", Side: constants.SideEnemy}}
		}, "name is not allowed"},
		{"shield ok", func(d *cards.CardDefinition) {
			d.SkillConfig.ActiveSkill.Effects = []cards.SkillEffect{shield(func(*cards.SkillEffect) {})}
		}, ""},
		{"shield on self ok", func(d *cards.CardDefinition) {
			d.SkillConfig.ActiveSkill.Target = 0
			d.SkillConfig.ActiveSkill.Effects = []cards.SkillEffect{shield(func(e *cards.SkillEffect) { e.Side = constants.SideSelf })}
		}, ""},
		{"shield permanent nonRemovable ok", func(d *cards.CardDefinition) {
			d.SkillConfig.ActiveSkill.Effects = []cards.SkillEffect{shield(func(e *cards.SkillEffect) {
				e.Duration, e.Permanent, e.NonRemovable = 0, true, true
			})}
		}, ""},
		{"shield permanent with duration", func(d *cards.CardDefinition) {
			d.SkillConfig.ActiveSkill.Effects = []cards.SkillEffect{shield(func(e *cards.SkillEffect) { e.Permanent = true })}
		}, "duration must not be set when permanent"},
		{"shield on enemy", func(d *cards.CardDefinition) {
			d.SkillConfig.ActiveSkill.Effects = []cards.SkillEffect{shield(func(e *cards.SkillEffect) { e.Side = constants.SideEnemy })}
		}, "Shield is only allowed"},
		{"shield missing name", func(d *cards.CardDefinition) {
			d.SkillConfig.ActiveSkill.Effects = []cards.SkillEffect{shield(func(e *cards.SkillEffect) { e.Name = "" })}
		}, "name is required for Shield"},
		{"shield missing scaleStat", func(d *cards.CardDefinition) {
			d.SkillConfig.ActiveSkill.Effects = []cards.SkillEffect{shield(func(e *cards.SkillEffect) { e.ScaleStat = 0 })}
		}, "scaleStat is required"},
		{"shield missing value", func(d *cards.CardDefinition) {
			d.SkillConfig.ActiveSkill.Effects = []cards.SkillEffect{shield(func(e *cards.SkillEffect) { e.Value = 0 })}
		}, "value is required"},
		{"shield negative value", func(d *cards.CardDefinition) {
			d.SkillConfig.ActiveSkill.Effects = []cards.SkillEffect{shield(func(e *cards.SkillEffect) { e.Value = -5 })}
		}, "value must be positive"},
		{"shield missing duration", func(d *cards.CardDefinition) {
			d.SkillConfig.ActiveSkill.Effects = []cards.SkillEffect{shield(func(e *cards.SkillEffect) { e.Duration = 0 })}
		}, "duration is required"},
		{"shield with stat", func(d *cards.CardDefinition) {
			d.SkillConfig.ActiveSkill.Effects = []cards.SkillEffect{shield(func(e *cards.SkillEffect) { e.Stat = constants.StatAttack })}
		}, "stat is not allowed"},
		{"shield with maxStack", func(d *cards.CardDefinition) {
			d.SkillConfig.ActiveSkill.Effects = []cards.SkillEffect{shield(func(e *cards.SkillEffect) { e.MaxStack = 2 })}
		}, "maxStack is not allowed"},
		{"shield with damageType", func(d *cards.CardDefinition) {
			d.SkillConfig.ActiveSkill.Effects = []cards.SkillEffect{shield(func(e *cards.SkillEffect) { e.DamageType = constants.DamageTypeMagic })}
		}, "damageType is not allowed"},
		{"buff permanent ok", func(d *cards.CardDefinition) {
			d.SkillConfig.ActiveSkill.Effects = []cards.SkillEffect{buff(func(e *cards.SkillEffect) { e.Duration = 0; e.Permanent = true })}
		}, ""},
		{"buff permanent with duration", func(d *cards.CardDefinition) {
			d.SkillConfig.ActiveSkill.Effects = []cards.SkillEffect{buff(func(e *cards.SkillEffect) { e.Permanent = true })}
		}, "duration must not be set when permanent"},
		{"buff nonRemovable ok", func(d *cards.CardDefinition) {
			d.SkillConfig.ActiveSkill.Effects = []cards.SkillEffect{buff(func(e *cards.SkillEffect) { e.NonRemovable = true })}
		}, ""},
		{"damage with permanent", func(d *cards.CardDefinition) { d.SkillConfig.NormalAtk.Effects[0].Permanent = true }, "permanent is not allowed"},
		{"damage with nonRemovable", func(d *cards.CardDefinition) { d.SkillConfig.NormalAtk.Effects[0].NonRemovable = true }, "nonRemovable is not allowed"},
		{"stun nonRemovable ok", func(d *cards.CardDefinition) {
			d.SkillConfig.ActiveSkill.Effects = []cards.SkillEffect{{
				Type: constants.EffectTypeStatus, Name: "n", Side: constants.SideEnemy, Status: constants.StatusStun, Duration: 1, NonRemovable: true,
			}}
		}, ""},
		{"poison nonRemovable ok", func(d *cards.CardDefinition) {
			d.SkillConfig.ActiveSkill.Effects = []cards.SkillEffect{{
				Type: constants.EffectTypeStatus, Name: "n", Side: constants.SideEnemy, Status: constants.StatusPoison, Duration: 3,
				ScaleStat: constants.StatAttack, Value: 30, NonRemovable: true,
			}}
		}, ""},
		{"status with permanent", func(d *cards.CardDefinition) {
			d.SkillConfig.ActiveSkill.Effects = []cards.SkillEffect{{
				Type: constants.EffectTypeStatus, Name: "n", Side: constants.SideEnemy, Status: constants.StatusStun, Duration: 1, Permanent: true,
			}}
		}, "duration must not be set when permanent"},
		{"dispel enemy ok", func(d *cards.CardDefinition) {
			d.SkillConfig.ActiveSkill.Effects = []cards.SkillEffect{{Type: constants.EffectTypeDispel, Side: constants.SideEnemy}}
		}, ""},
		{"dispel with count ok", func(d *cards.CardDefinition) {
			d.SkillConfig.ActiveSkill.Effects = []cards.SkillEffect{{Type: constants.EffectTypeDispel, Side: constants.SideEnemy, Count: 2}}
		}, ""},
		{"dispel negative count", func(d *cards.CardDefinition) {
			d.SkillConfig.ActiveSkill.Effects = []cards.SkillEffect{{Type: constants.EffectTypeDispel, Side: constants.SideEnemy, Count: -1}}
		}, "must not be negative"},
		{"dispel on ally", func(d *cards.CardDefinition) {
			d.SkillConfig.ActiveSkill.Effects = []cards.SkillEffect{{Type: constants.EffectTypeDispel, Side: constants.SideAlly}}
		}, "Dispel is only allowed"},
		{"dispel on self", func(d *cards.CardDefinition) {
			d.SkillConfig.ActiveSkill.Effects = []cards.SkillEffect{{Type: constants.EffectTypeDispel, Side: constants.SideSelf}}
		}, "Dispel is only allowed"},
		{"dispel with value", func(d *cards.CardDefinition) {
			d.SkillConfig.ActiveSkill.Effects = []cards.SkillEffect{{Type: constants.EffectTypeDispel, Side: constants.SideEnemy, Value: 10}}
		}, "value is not allowed"},
		{"cleanse ally ok", func(d *cards.CardDefinition) {
			d.SkillConfig.ActiveSkill.Effects = []cards.SkillEffect{{Type: constants.EffectTypeCleanse, Side: constants.SideAlly, Count: 1}}
		}, ""},
		{"cleanse self ok", func(d *cards.CardDefinition) {
			d.SkillConfig.ActiveSkill.Target = 0
			d.SkillConfig.ActiveSkill.Effects = []cards.SkillEffect{{Type: constants.EffectTypeCleanse, Side: constants.SideSelf}}
		}, ""},
		{"cleanse on enemy", func(d *cards.CardDefinition) {
			d.SkillConfig.ActiveSkill.Effects = []cards.SkillEffect{{Type: constants.EffectTypeCleanse, Side: constants.SideEnemy}}
		}, "Cleanse is only allowed"},
		{"count on damage", func(d *cards.CardDefinition) { d.SkillConfig.NormalAtk.Effects[0].Count = 1 }, "count is not allowed"},
		{"rage gain ok", func(d *cards.CardDefinition) {
			d.SkillConfig.NormalAtk.Effects = []cards.SkillEffect{{Type: constants.EffectTypeRage, Side: constants.SideSelf, Value: 20}}
		}, ""},
		{"rage drain enemy ok", func(d *cards.CardDefinition) {
			d.SkillConfig.ActiveSkill.Effects = []cards.SkillEffect{{Type: constants.EffectTypeRage, Side: constants.SideEnemy, Target: constants.TargetAll, Value: -30}}
		}, ""},
		{"rage zero value", func(d *cards.CardDefinition) {
			d.SkillConfig.NormalAtk.Effects = []cards.SkillEffect{{Type: constants.EffectTypeRage, Side: constants.SideSelf}}
		}, "value is required"},
		{"rage with scaleStat", func(d *cards.CardDefinition) {
			d.SkillConfig.NormalAtk.Effects = []cards.SkillEffect{{Type: constants.EffectTypeRage, Side: constants.SideSelf, Value: 20, ScaleStat: constants.StatAttack}}
		}, "scaleStat is not allowed"},
		{"rage with duration", func(d *cards.CardDefinition) {
			d.SkillConfig.NormalAtk.Effects = []cards.SkillEffect{{Type: constants.EffectTypeRage, Side: constants.SideSelf, Value: 20, Duration: 2}}
		}, "duration is not allowed"},
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
				Type: constants.EffectTypeBuff, Name: "n", Side: constants.SideEnemy, Stat: constants.StatDefense, Value: -20, Duration: 2, MaxStack: 3,
			}}
		}, ""},
		{"buff without duration", func(d *cards.CardDefinition) {
			d.SkillConfig.ActiveSkill.Effects = []cards.SkillEffect{{
				Type: constants.EffectTypeBuff, Name: "n", Side: constants.SideEnemy, Stat: constants.StatDefense, Value: -20,
			}}
		}, "duration is required"},
		{"buff with scaleStat", func(d *cards.CardDefinition) {
			d.SkillConfig.ActiveSkill.Effects = []cards.SkillEffect{{
				Type: constants.EffectTypeBuff, Name: "n", Side: constants.SideEnemy, Stat: constants.StatDefense, Value: -20, Duration: 2, ScaleStat: constants.StatAttack,
			}}
		}, "scaleStat is not allowed"},
		{"stun ok", func(d *cards.CardDefinition) {
			d.SkillConfig.ActiveSkill.Effects = []cards.SkillEffect{{
				Type: constants.EffectTypeStatus, Name: "n", Side: constants.SideEnemy, Status: constants.StatusStun, Duration: 1,
			}}
		}, ""},
		{"stun with value", func(d *cards.CardDefinition) {
			d.SkillConfig.ActiveSkill.Effects = []cards.SkillEffect{{
				Type: constants.EffectTypeStatus, Name: "n", Side: constants.SideEnemy, Status: constants.StatusStun, Duration: 1, Value: 10,
			}}
		}, "value is not allowed"},
		{"poison ok", func(d *cards.CardDefinition) {
			d.SkillConfig.ActiveSkill.Effects = []cards.SkillEffect{{
				Type: constants.EffectTypeStatus, Name: "n", Side: constants.SideEnemy, Status: constants.StatusPoison, Duration: 3, ScaleStat: constants.StatAttack, Value: 30,
			}}
		}, ""},
		{"poison without value", func(d *cards.CardDefinition) {
			d.SkillConfig.ActiveSkill.Effects = []cards.SkillEffect{{
				Type: constants.EffectTypeStatus, Name: "n", Side: constants.SideEnemy, Status: constants.StatusPoison, Duration: 3,
			}}
		}, "required"},
		{"status effect without status", func(d *cards.CardDefinition) {
			d.SkillConfig.ActiveSkill.Effects = []cards.SkillEffect{{
				Type: constants.EffectTypeStatus, Name: "n", Side: constants.SideEnemy, Duration: 3,
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
	for _, want := range []string{"'O''Brien'", "VALUES ('card_001', 'O''Brien', 2, 4,", `"trigger":"BeforeRoundStart"`, "ON CONFLICT (card_id) DO UPDATE"} {
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
