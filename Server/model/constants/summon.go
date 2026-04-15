package constants

import (
	"encoding/json"
	"fmt"
)

type SummonType int

const (
	NormalSummon  SummonType = 1
	FriendSummon  SummonType = 2
	DiamondSummon SummonType = 3
)

var summonTypeValueToName = map[SummonType]string{
	NormalSummon:  "Normal",
	FriendSummon:  "Friend",
	DiamondSummon: "Diamond",
}

var summonTypeNameToValue = map[string]SummonType{
	"Normal":  NormalSummon,
	"Friend":  FriendSummon,
	"Diamond": DiamondSummon,
}

func (s SummonType) String() string {
	if name, ok := summonTypeValueToName[s]; ok {
		return name
	}
	return fmt.Sprintf("SummonType(%d)", s)
}

func (s SummonType) Valid() bool {
	_, ok := summonTypeValueToName[s]
	return ok
}

func (s SummonType) MarshalJSON() ([]byte, error) {
	if name, ok := summonTypeValueToName[s]; ok {
		return json.Marshal(name)
	}
	return nil, fmt.Errorf("invalid SummonType value: %d", s)
}

func (s *SummonType) UnmarshalJSON(data []byte) error {
	var name string
	if err := json.Unmarshal(data, &name); err != nil {
		return err
	}
	if val, ok := summonTypeNameToValue[name]; ok {
		*s = val
		return nil
	}
	return fmt.Errorf("invalid SummonType name: %s", name)
}
