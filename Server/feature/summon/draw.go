package summon

import (
	"fmt"
	"math/rand/v2"

	"cardGame/feature/card"
	"cardGame/model/constants"
)

// Rand 抽象化亂數來源，讓測試可以注入固定結果。
type Rand interface {
	IntN(n int) int
}

type globalRand struct{}

func (globalRand) IntN(n int) int { return rand.IntN(n) }

// DefaultRand 使用 math/rand/v2 的全域亂數，可併發使用。
var DefaultRand Rand = globalRand{}

type CardPool map[constants.Rarity][]card.Card

func NewCardPool(cards []card.Card) CardPool {
	pool := make(CardPool)
	for _, c := range cards {
		pool[c.Rarity] = append(pool[c.Rarity], c)
	}
	return pool
}

// Draw 依規則抽出 times 張卡，同稀有度內機率均等。
// 抽之前先檢查所有可能出現的稀有度都有卡，避免結果取決於運氣。
func Draw(rule Rule, pool CardPool, times int32, rnd Rand) ([]card.Card, error) {
	if times <= 0 {
		return nil, ErrInvalidTimes
	}
	for _, w := range rule.Weights {
		if w.Weight > 0 && len(pool[w.Rarity]) == 0 {
			return nil, fmt.Errorf("%w: %s", ErrEmptyCardPool, w.Rarity)
		}
	}

	guaranteed := rule.TenPullMinRarity == 0
	drawn := make([]card.Card, 0, times)
	for i := int32(0); i < times; i++ {
		weights := rule.Weights
		if times == TenPullTimes && i == times-1 && !guaranteed {
			weights = weightsAtLeast(rule.Weights, rule.TenPullMinRarity)
		}
		rarity, err := rollRarity(weights, rnd)
		if err != nil {
			return nil, err
		}
		if rarity >= rule.TenPullMinRarity {
			guaranteed = true
		}
		cards := pool[rarity]
		drawn = append(drawn, cards[rnd.IntN(len(cards))])
	}
	return drawn, nil
}

func weightsAtLeast(weights []RarityWeight, min constants.Rarity) []RarityWeight {
	filtered := make([]RarityWeight, 0, len(weights))
	for _, w := range weights {
		if w.Rarity >= min {
			filtered = append(filtered, w)
		}
	}
	return filtered
}

func rollRarity(weights []RarityWeight, rnd Rand) (constants.Rarity, error) {
	total := 0
	for _, w := range weights {
		if w.Weight > 0 {
			total += w.Weight
		}
	}
	if total <= 0 {
		return 0, ErrInvalidRule
	}
	n := rnd.IntN(total)
	for _, w := range weights {
		if w.Weight <= 0 {
			continue
		}
		if n < w.Weight {
			return w.Rarity, nil
		}
		n -= w.Weight
	}
	return 0, ErrInvalidRule
}
