package packet

import "cardGame/model/constants"

type SummonCard struct {
	Times      int32                `json:"times"`
	SummonType constants.SummonType `json:"summonType"`
}
