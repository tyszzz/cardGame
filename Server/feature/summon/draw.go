package summon

import (
	"math/rand/v2"

	"cardGame/model/constants"
)

type (
	// Rand 抽象化亂數來源，讓測試可以注入固定結果。
	Rand interface {
		IntN(n int) int
	}

	globalRand struct{}
)

// DefaultRand 使用 math/rand/v2 的全域亂數，可併發使用。
var DefaultRand Rand = globalRand{}

// inner function

func (p *Pool) pickCard(rarity constants.Rarity, state PityState, rnd Rand) PoolCard {
	ups, normal := p.up[rarity], p.normal[rarity]
	if len(ups) == 0 {
		return normal[rnd.IntN(len(normal))]
	}
	forced := p.Pity != nil && p.Pity.UpGuarantee && state.GuaranteedUp && rarity >= p.Pity.MinRarity
	if forced || len(normal) == 0 || rnd.IntN(100) < p.UpRates[rarity] {
		return ups[rnd.IntN(len(ups))]
	}
	return normal[rnd.IntN(len(normal))]
}

// rollRarity 依權重抽稀有度，只考慮 minRarity 以上的項目（0 代表不限制）。
func rollRarity(weights []RarityWeight, minRarity constants.Rarity, rnd Rand) (constants.Rarity, error) {
	total := 0
	for _, w := range weights {
		if w.Rarity >= minRarity {
			total += w.Weight
		}
	}
	if total <= 0 {
		return 0, ErrInvalidPool
	}
	n := rnd.IntN(total)
	for _, w := range weights {
		if w.Rarity < minRarity {
			continue
		}
		if n < w.Weight {
			return w.Rarity, nil
		}
		n -= w.Weight
	}
	return 0, ErrInvalidPool
}

// outer function

func (globalRand) IntN(n int) int { return rand.IntN(n) }

// Draw 依卡池規則抽 times 張卡，回傳抽到的卡與抽完後的保底狀態。
// 每一抽先決定稀有度（累積保底與十連保底會限制最低稀有度），再依 UP 比例選卡。
func (p *Pool) Draw(times int32, state PityState, rnd Rand) ([]PoolCard, PityState, error) {
	if times <= 0 {
		return nil, state, ErrInvalidTimes
	}

	tenSatisfied := p.TenPullMinRarity == 0
	drawn := make([]PoolCard, 0, times)
	for i := int32(0); i < times; i++ {
		var minRarity constants.Rarity
		if p.Pity != nil && state.PullsSinceHigh+1 >= p.Pity.Threshold {
			minRarity = p.Pity.MinRarity
		}
		if times == TenPullTimes && i == times-1 && !tenSatisfied && p.TenPullMinRarity > minRarity {
			minRarity = p.TenPullMinRarity
		}

		rarity, err := rollRarity(p.Weights, minRarity, rnd)
		if err != nil {
			return nil, state, err
		}
		pc := p.pickCard(rarity, state, rnd)
		drawn = append(drawn, pc)

		if p.TenPullMinRarity == 0 || rarity >= p.TenPullMinRarity {
			tenSatisfied = true
		}
		if p.Pity != nil {
			if rarity >= p.Pity.MinRarity {
				state.PullsSinceHigh = 0
				// 大保底只看有 UP 卡的稀有度；其他稀有度不改變狀態
				if p.Pity.UpGuarantee && len(p.up[rarity]) > 0 {
					state.GuaranteedUp = !pc.IsUp
				}
			} else {
				state.PullsSinceHigh++
			}
		}
	}
	return drawn, state, nil
}
