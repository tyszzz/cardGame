package constants

import (
	"encoding/json"
	"fmt"
)

type DamageType int

const (
	DamageTypePhysical DamageType = 1 // 物理
	DamageTypeMagic    DamageType = 2 // 魔法
	DamageTypeTrue     DamageType = 3 // 真實
)

var damageTypeValueToName = map[DamageType]string{
	DamageTypePhysical: "Physical",
	DamageTypeMagic:    "Magic",
	DamageTypeTrue:     "True",
}

var damageTypeNameToValue = map[string]DamageType{
	"Physical": DamageTypePhysical,
	"Magic":    DamageTypeMagic,
	"True":     DamageTypeTrue,
}

func (v DamageType) String() string {
	if name, ok := damageTypeValueToName[v]; ok {
		return name
	}
	return fmt.Sprintf("DamageType(%d)", v)
}

func (v DamageType) Valid() bool {
	_, ok := damageTypeValueToName[v]
	return ok
}

func (v DamageType) MarshalJSON() ([]byte, error) {
	if name, ok := damageTypeValueToName[v]; ok {
		return json.Marshal(name)
	}
	return nil, fmt.Errorf("invalid DamageType value: %d", v)
}

func (v *DamageType) UnmarshalJSON(data []byte) error {
	var name string
	if err := json.Unmarshal(data, &name); err != nil {
		return err
	}
	if val, ok := damageTypeNameToValue[name]; ok {
		*v = val
		return nil
	}
	return fmt.Errorf("invalid DamageType name: %s", name)
}
