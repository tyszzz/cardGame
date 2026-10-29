package constants

import (
	"encoding/json"
	"testing"
)

func TestSkillEnumsJSONRoundTrip(t *testing.T) {
	var (
		trigger = TriggerRoundEnd
		target  = TargetColumn
		typ     = EffectTypeBuff
		stat    = StatDefense
		status  = StatusBurn
		dmg     = DamageTypeTrue
		side    = SideSelf
	)
	cases := []struct {
		name string
		v    any
		want string
		dst  any
	}{
		{"trigger", trigger, `"RoundEnd"`, new(Trigger)},
		{"target", target, `"Column"`, new(Target)},
		{"effectType", typ, `"Buff"`, new(EffectType)},
		{"stat", stat, `"Defense"`, new(Stat)},
		{"status", status, `"Burn"`, new(Status)},
		{"damageType", dmg, `"True"`, new(DamageType)},
		{"side", side, `"Self"`, new(Side)},
	}
	for _, c := range cases {
		data, err := json.Marshal(c.v)
		if err != nil || string(data) != c.want {
			t.Fatalf("%s: marshal = %s, %v; want %s", c.name, data, err, c.want)
		}
		if err := json.Unmarshal(data, c.dst); err != nil {
			t.Fatalf("%s: unmarshal: %v", c.name, err)
		}
	}
}

func TestSkillEnumsRejectInvalid(t *testing.T) {
	if (Trigger(0)).Valid() || (Side(0)).Valid() || (Stat(0)).Valid() {
		t.Fatal("zero value must be invalid")
	}
	var s Status
	if err := json.Unmarshal([]byte(`"Frozen"`), &s); err == nil {
		t.Fatal("expected error for unknown name")
	}
	if _, err := json.Marshal(Target(99)); err == nil {
		t.Fatal("expected error for invalid value")
	}
}
