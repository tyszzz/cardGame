package summon

import "cardGame/model/constants"

type summonFriend struct {
}

func newSummonFriend() SummonManager {
	return &summonFriend{}
}

func (f *summonFriend) SummonOnce(summonType constants.SummonType) string {
	return "test"
}
func (f *summonFriend) SummonTen(summonType constants.SummonType) []string {
	return []string{"test"}
}
