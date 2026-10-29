package constants

import (
	"encoding/json"
	"fmt"
)

type Stat int

const (
	StatHealth  Stat = 1 // 生命
	StatAttack  Stat = 2 // 攻擊
	StatDefense Stat = 3 // 防禦
)

var statValueToName = map[Stat]string{
	StatHealth:  "Health",
	StatAttack:  "Attack",
	StatDefense: "Defense",
}

var statNameToValue = map[string]Stat{
	"Health":  StatHealth,
	"Attack":  StatAttack,
	"Defense": StatDefense,
}

func (v Stat) String() string {
	if name, ok := statValueToName[v]; ok {
		return name
	}
	return fmt.Sprintf("Stat(%d)", v)
}

func (v Stat) Valid() bool {
	_, ok := statValueToName[v]
	return ok
}

func (v Stat) MarshalJSON() ([]byte, error) {
	if name, ok := statValueToName[v]; ok {
		return json.Marshal(name)
	}
	return nil, fmt.Errorf("invalid Stat value: %d", v)
}

func (v *Stat) UnmarshalJSON(data []byte) error {
	var name string
	if err := json.Unmarshal(data, &name); err != nil {
		return err
	}
	if val, ok := statNameToValue[name]; ok {
		*v = val
		return nil
	}
	return fmt.Errorf("invalid Stat name: %s", name)
}
