package constants

import (
	"encoding/json"
	"fmt"
)

// -------------------- Race --------------------
type Race int

const (
	RaceHuman    Race = 1 // 人族
	RaceImmortal Race = 2 // 仙族
	RaceDemon    Race = 3 // 魔族
	RaceDragon   Race = 4 // 龍族
	RaceHybrid   Race = 5 // 雜種
)

var raceValueToName = map[Race]string{
	RaceHuman:    "Human",
	RaceImmortal: "Immortal",
	RaceDemon:    "Demon",
	RaceDragon:   "Dragon",
	RaceHybrid:   "Hybrid",
}

var raceNameToValue = map[string]Race{
	"Human":    RaceHuman,
	"Immortal": RaceImmortal,
	"Demon":    RaceDemon,
	"Dragon":   RaceDragon,
	"Hybrid":   RaceHybrid,
}

func (r Race) String() string {
	if name, ok := raceValueToName[r]; ok {
		return name
	}
	return fmt.Sprintf("Race(%d)", r)
}

func (r Race) Valid() bool {
	_, ok := raceValueToName[r]
	return ok
}

func (r Race) MarshalJSON() ([]byte, error) {
	if name, ok := raceValueToName[r]; ok {
		return json.Marshal(name)
	}
	return nil, fmt.Errorf("invalid Race value: %d", r)
}

func (r *Race) UnmarshalJSON(data []byte) error {
	var name string
	if err := json.Unmarshal(data, &name); err != nil {
		return err
	}
	if val, ok := raceNameToValue[name]; ok {
		*r = val
		return nil
	}
	return fmt.Errorf("invalid Race name: %s", name)
}
