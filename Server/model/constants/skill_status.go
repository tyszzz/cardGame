package constants

import (
	"encoding/json"
	"fmt"
)

type Status int

const (
	StatusPoison Status = 1 // 中毒
	StatusBurn   Status = 2 // 燃燒
	StatusStun   Status = 3 // 暈眩
)

var statusValueToName = map[Status]string{
	StatusPoison: "Poison",
	StatusBurn:   "Burn",
	StatusStun:   "Stun",
}

var statusNameToValue = map[string]Status{
	"Poison": StatusPoison,
	"Burn":   StatusBurn,
	"Stun":   StatusStun,
}

func (v Status) String() string {
	if name, ok := statusValueToName[v]; ok {
		return name
	}
	return fmt.Sprintf("Status(%d)", v)
}

func (v Status) Valid() bool {
	_, ok := statusValueToName[v]
	return ok
}

func (v Status) MarshalJSON() ([]byte, error) {
	if name, ok := statusValueToName[v]; ok {
		return json.Marshal(name)
	}
	return nil, fmt.Errorf("invalid Status value: %d", v)
}

func (v *Status) UnmarshalJSON(data []byte) error {
	var name string
	if err := json.Unmarshal(data, &name); err != nil {
		return err
	}
	if val, ok := statusNameToValue[name]; ok {
		*v = val
		return nil
	}
	return fmt.Errorf("invalid Status name: %s", name)
}
