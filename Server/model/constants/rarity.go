package constants

import (
	"encoding/json"
	"fmt"
)

type Rarity int

const (
	RarityR   Rarity = 1
	RaritySR  Rarity = 2
	RaritySSR Rarity = 3
	RarityUR  Rarity = 4
)

var rarityValueToName = map[Rarity]string{
	RarityR:   "R",
	RaritySR:  "SR",
	RaritySSR: "SSR",
	RarityUR:  "UR",
}

var rarityNameToValue = map[string]Rarity{
	"R":   RarityR,
	"SR":  RaritySR,
	"SSR": RaritySSR,
	"UR":  RarityUR,
}

func (r Rarity) String() string {
	if name, ok := rarityValueToName[r]; ok {
		return name
	}
	return fmt.Sprintf("Rarity(%d)", r)
}

func (r Rarity) Valid() bool {
	_, ok := rarityValueToName[r]
	return ok
}

func (r Rarity) MarshalJSON() ([]byte, error) {
	if name, ok := rarityValueToName[r]; ok {
		return json.Marshal(name)
	}
	return nil, fmt.Errorf("invalid Rarity value: %d", r)
}

func (r *Rarity) UnmarshalJSON(data []byte) error {
	var name string
	if err := json.Unmarshal(data, &name); err != nil {
		return err
	}
	if val, ok := rarityNameToValue[name]; ok {
		*r = val
		return nil
	}
	return fmt.Errorf("invalid Rarity name: %s", name)
}
