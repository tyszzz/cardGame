package constants

import (
	"encoding/json"
	"fmt"
)

type Trigger int

const (
	TriggerBeforeRoundStart Trigger = 1 // 回合開始前
	TriggerRoundEnd         Trigger = 2 // 回合結束時
	TriggerBeforeAction     Trigger = 3 // 行動前
	TriggerAfterAction      Trigger = 4 // 行動後
)

var triggerValueToName = map[Trigger]string{
	TriggerBeforeRoundStart: "BeforeRoundStart",
	TriggerRoundEnd:         "RoundEnd",
	TriggerBeforeAction:     "BeforeAction",
	TriggerAfterAction:      "AfterAction",
}

var triggerNameToValue = map[string]Trigger{
	"BeforeRoundStart": TriggerBeforeRoundStart,
	"RoundEnd":         TriggerRoundEnd,
	"BeforeAction":     TriggerBeforeAction,
	"AfterAction":      TriggerAfterAction,
}

func (v Trigger) String() string {
	if name, ok := triggerValueToName[v]; ok {
		return name
	}
	return fmt.Sprintf("Trigger(%d)", v)
}

func (v Trigger) Valid() bool {
	_, ok := triggerValueToName[v]
	return ok
}

func (v Trigger) MarshalJSON() ([]byte, error) {
	if name, ok := triggerValueToName[v]; ok {
		return json.Marshal(name)
	}
	return nil, fmt.Errorf("invalid Trigger value: %d", v)
}

func (v *Trigger) UnmarshalJSON(data []byte) error {
	var name string
	if err := json.Unmarshal(data, &name); err != nil {
		return err
	}
	if val, ok := triggerNameToValue[name]; ok {
		*v = val
		return nil
	}
	return fmt.Errorf("invalid Trigger name: %s", name)
}
