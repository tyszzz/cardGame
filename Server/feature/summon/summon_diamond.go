package summon

import "cardGame/model/constants"

type summonDiamond struct {
}

func newSummonDiamond() SummonManager {
	return &summonDiamond{}
}

func (d *summonDiamond) SummonOnce(summonType constants.SummonType) string {
	return "test"
}
func (d *summonDiamond) SummonTen(summonType constants.SummonType) []string {
	return []string{"test"}
}
