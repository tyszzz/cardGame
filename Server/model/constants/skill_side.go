package constants

import (
	"encoding/json"
	"fmt"
)

type Side int

const (
	SideEnemy Side = 1 // 敵方
	SideAlly  Side = 2 // 我方
	SideSelf  Side = 3 // 自己
)

var sideValueToName = map[Side]string{
	SideEnemy: "Enemy",
	SideAlly:  "Ally",
	SideSelf:  "Self",
}

var sideNameToValue = map[string]Side{
	"Enemy": SideEnemy,
	"Ally":  SideAlly,
	"Self":  SideSelf,
}

func (v Side) String() string {
	if name, ok := sideValueToName[v]; ok {
		return name
	}
	return fmt.Sprintf("Side(%d)", v)
}

func (v Side) Valid() bool {
	_, ok := sideValueToName[v]
	return ok
}

func (v Side) MarshalJSON() ([]byte, error) {
	if name, ok := sideValueToName[v]; ok {
		return json.Marshal(name)
	}
	return nil, fmt.Errorf("invalid Side value: %d", v)
}

func (v *Side) UnmarshalJSON(data []byte) error {
	var name string
	if err := json.Unmarshal(data, &name); err != nil {
		return err
	}
	if val, ok := sideNameToValue[name]; ok {
		*v = val
		return nil
	}
	return fmt.Errorf("invalid Side name: %s", name)
}
