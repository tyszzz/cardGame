package packet

import "cardGame/model/constants"

// account packet
type (
	CreateNewAccount struct {
		AccountId string `json:"accountId"`
		Password  string `json:"password"`
	}

	Login struct {
		AccountId string `json:"accountId"`
		Password  string `json:"password"`
	}
)

// summon packet
type (
	SummonCard struct {
		Times      int32                `json:"times"`
		SummonType constants.SummonType `json:"summonType"`
	}
)
