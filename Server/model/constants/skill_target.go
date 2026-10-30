package constants

import (
	"encoding/json"
	"fmt"
)

type Target int

const (
	TargetSingle   Target = 1 // 單體
	TargetRandom3  Target = 2 // 隨機3人
	TargetAll      Target = 3 // 全體
	TargetFrontRow Target = 4 // 前排
	TargetBackRow  Target = 5 // 後排
	TargetColumn   Target = 6 // 縱排
	TargetOpposite Target = 7 // 對位（正對面的敵人）
)

var targetValueToName = map[Target]string{
	TargetSingle:   "Single",
	TargetRandom3:  "Random3",
	TargetAll:      "All",
	TargetFrontRow: "FrontRow",
	TargetBackRow:  "BackRow",
	TargetColumn:   "Column",
	TargetOpposite: "Opposite",
}

var targetNameToValue = map[string]Target{
	"Single":   TargetSingle,
	"Random3":  TargetRandom3,
	"All":      TargetAll,
	"FrontRow": TargetFrontRow,
	"BackRow":  TargetBackRow,
	"Column":   TargetColumn,
	"Opposite": TargetOpposite,
}

func (v Target) String() string {
	if name, ok := targetValueToName[v]; ok {
		return name
	}
	return fmt.Sprintf("Target(%d)", v)
}

func (v Target) Valid() bool {
	_, ok := targetValueToName[v]
	return ok
}

func (v Target) MarshalJSON() ([]byte, error) {
	if name, ok := targetValueToName[v]; ok {
		return json.Marshal(name)
	}
	return nil, fmt.Errorf("invalid Target value: %d", v)
}

func (v *Target) UnmarshalJSON(data []byte) error {
	var name string
	if err := json.Unmarshal(data, &name); err != nil {
		return err
	}
	if val, ok := targetNameToValue[name]; ok {
		*v = val
		return nil
	}
	return fmt.Errorf("invalid Target name: %s", name)
}
