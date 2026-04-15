package summon

import "cardGame/model/constants"

type summonCoin struct {
}

func newSummonCoin() SummonManager {
	return &summonCoin{}
}

func (c *summonCoin) SummonOnce(summonType constants.SummonType) string {
	return "test"
}
func (c *summonCoin) SummonTen(summonType constants.SummonType) []string {
	return []string{"test"}
}
