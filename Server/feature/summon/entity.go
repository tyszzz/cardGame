package summon

import "cardGame/model/constants"

type SummonManager interface {
	SummonOnce(summonType constants.SummonType) string
	SummonTen(summonType constants.SummonType) []string
}
