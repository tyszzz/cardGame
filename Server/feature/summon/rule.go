package summon

import "cardGame/model/constants"

// TenPullTimes 為十連抽的次數，保底只在十連抽生效。
const TenPullTimes int32 = 10

type RarityWeight struct {
	Rarity constants.Rarity
	Weight int // 整數權重，同一個 Rule 內的權重總和即為分母
}

type Rule struct {
	Currency         constants.CurrencyType
	Costs            map[int32]int64 // 抽數 -> 總費用
	Weights          []RarityWeight
	TenPullMinRarity constants.Rarity // 十連抽至少一張此稀有度以上，0 代表無保底
}

var rules = map[constants.SummonType]Rule{
	constants.NormalSummon: {
		Currency: constants.CurrencyCoin,
		Costs:    map[int32]int64{1: 5000, 10: 50000},
		Weights: []RarityWeight{
			{constants.RarityR, 70},
			{constants.RaritySR, 30},
		},
	},
	constants.FriendSummon: {
		Currency: constants.CurrencyFriendPoint,
		Costs:    map[int32]int64{1: 20, 10: 200},
		Weights: []RarityWeight{
			{constants.RarityR, 70},
			{constants.RaritySR, 30},
		},
	},
	constants.DiamondSummon: {
		Currency: constants.CurrencyDiamond,
		Costs:    map[int32]int64{1: 200, 10: 1800}, // 十連 9 折
		Weights: []RarityWeight{
			{constants.RarityR, 50},
			{constants.RaritySR, 30},
			{constants.RaritySSR, 18},
			{constants.RarityUR, 2},
		},
		TenPullMinRarity: constants.RaritySSR,
	},
}

func RuleOf(summonType constants.SummonType) (Rule, error) {
	rule, ok := rules[summonType]
	if !ok {
		return Rule{}, ErrInvalidSummonType
	}
	return rule, nil
}

func (r Rule) Cost(times int32) (int64, error) {
	cost, ok := r.Costs[times]
	if !ok {
		return 0, ErrInvalidTimes
	}
	return cost, nil
}
