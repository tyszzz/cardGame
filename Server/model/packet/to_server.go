package packet

import "cardGame/model/constants"

// account packet
type (
	Login struct {
		AccountId string `json:"accountId"`
		Password  string `json:"password"`
	}
)

// summon packet
type (
	SummonCard struct {
		UID        string               `json:"uid"`
		Times      int32                `json:"times"`
		SummonType constants.SummonType `json:"summonType"`
	}
)
