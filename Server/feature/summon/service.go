package summon

import "cardGame/model/constants"

type SummonStruct struct {
	summonPool map[constants.SummonType]SummonManager
}

func NewSummonManager() *SummonStruct {
	var (
		normalSummon  = newSummonCoin()
		friendSummon  = newSummonFriend()
		diamondSummon = newSummonDiamond()

		summonPool = map[constants.SummonType]SummonManager{
			constants.NormalSummon:  normalSummon,
			constants.FriendSummon:  friendSummon,
			constants.DiamondSummon: diamondSummon,
		}
	)
	return &SummonStruct{
		summonPool: summonPool,
	}
}

// 從召喚池里看玩家是用哪個類型召喚
func (s *SummonStruct) SummonOnce(summonType constants.SummonType) string {
	return s.summonPool[summonType].SummonOnce(summonType)
}

func (s *SummonStruct) SummonTen(summonType constants.SummonType) []string {
	return s.summonPool[summonType].SummonTen(summonType)
}
