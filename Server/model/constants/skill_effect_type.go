package constants

import (
	"encoding/json"
	"fmt"
)

type EffectType int

const (
	EffectTypeDamage EffectType = 1 // 傷害
	EffectTypeStatus EffectType = 2 // 狀態
	EffectTypeHeal   EffectType = 3 // 回血
	EffectTypeBuff   EffectType = 4 // 增減屬性
)

var effectTypeValueToName = map[EffectType]string{
	EffectTypeDamage: "Damage",
	EffectTypeStatus: "Status",
	EffectTypeHeal:   "Heal",
	EffectTypeBuff:   "Buff",
}

var effectTypeNameToValue = map[string]EffectType{
	"Damage": EffectTypeDamage,
	"Status": EffectTypeStatus,
	"Heal":   EffectTypeHeal,
	"Buff":   EffectTypeBuff,
}

func (v EffectType) String() string {
	if name, ok := effectTypeValueToName[v]; ok {
		return name
	}
	return fmt.Sprintf("EffectType(%d)", v)
}

func (v EffectType) Valid() bool {
	_, ok := effectTypeValueToName[v]
	return ok
}

func (v EffectType) MarshalJSON() ([]byte, error) {
	if name, ok := effectTypeValueToName[v]; ok {
		return json.Marshal(name)
	}
	return nil, fmt.Errorf("invalid EffectType value: %d", v)
}

func (v *EffectType) UnmarshalJSON(data []byte) error {
	var name string
	if err := json.Unmarshal(data, &name); err != nil {
		return err
	}
	if val, ok := effectTypeNameToValue[name]; ok {
		*v = val
		return nil
	}
	return fmt.Errorf("invalid EffectType name: %s", name)
}
