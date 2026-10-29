package constants

import (
	"encoding/json"
	"testing"
)

func TestRaceJSONNames(t *testing.T) {
	for name, want := range map[string]Race{"Human": RaceHuman, "Xian": RaceXian, "Mo": RaceMo, "Dragon": RaceDragon, "Hybrid": RaceHybrid} {
		var r Race
		if err := json.Unmarshal([]byte(`"`+name+`"`), &r); err != nil || r != want {
			t.Errorf("unmarshal %q = %v, %v; want %v", name, r, err, want)
		}
		data, err := json.Marshal(want)
		if err != nil || string(data) != `"`+name+`"` {
			t.Errorf("marshal %v = %s, %v", want, data, err)
		}
	}
	for _, old := range []string{"Immortal", "Demon"} {
		var r Race
		if err := json.Unmarshal([]byte(`"`+old+`"`), &r); err == nil {
			t.Errorf("old name %q should be rejected", old)
		}
	}
}
