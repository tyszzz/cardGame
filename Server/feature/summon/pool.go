package summon

import (
	"fmt"
	"slices"
	"sort"
	"time"

	"cardGame/feature/card"
	"cardGame/model/constants"
)

// TenPullTimes 為十連抽的次數，十連保底只在十連抽生效。
const TenPullTimes int32 = 10

type RarityWeight struct {
	Rarity constants.Rarity
	Weight int // 整數權重，同一個卡池內的權重總和即為分母
}

// PityRule 為累積保底：自上次出 MinRarity 以上後，第 Threshold 抽必出 MinRarity 以上。
// UpGuarantee 為大保底：出了非 UP 的 MinRarity 以上後，下一張必為 UP 卡。
type PityRule struct {
	Threshold   int
	MinRarity   constants.Rarity
	UpGuarantee bool
}

// PoolDefinition 為 summon_pools 的一筆資料。
type PoolDefinition struct {
	PoolId               string
	PoolName             string
	Currency             constants.CurrencyType
	Costs                map[int32]int64          // 抽數 -> 總費用
	Weights              []RarityWeight           // 稀有度機率
	TenPullMinRarity     constants.Rarity         // 十連抽至少一張此稀有度以上，0 代表無保底
	UpRates              map[constants.Rarity]int // 出該稀有度時為 UP 卡的百分比
	Pity                 *PityRule                // nil 代表無累積保底
	IncludeStandardCards bool                     // 是否包含所有啟用且非限定的卡
	StartAt              *time.Time
	EndAt                *time.Time
	IsEnabled            bool
}

// PoolCardRef 為 summon_pool_cards 的一筆資料。
type PoolCardRef struct {
	CardId string
	IsUp   bool
}

type PoolCard struct {
	Card card.Card
	IsUp bool
}

// Pool 為組裝完成且通過驗證的卡池。
type Pool struct {
	PoolDefinition
	up     map[constants.Rarity][]PoolCard
	normal map[constants.Rarity][]PoolCard
}

// Available 判斷卡池目前是否可抽，時間以呼叫端提供的伺服器時間為準。
func (d PoolDefinition) Available(now time.Time) bool {
	if !d.IsEnabled {
		return false
	}
	if d.StartAt != nil && now.Before(*d.StartAt) {
		return false
	}
	if d.EndAt != nil && !now.Before(*d.EndAt) {
		return false
	}
	return true
}

func (d PoolDefinition) Cost(times int32) (int64, error) {
	cost, ok := d.Costs[times]
	if !ok {
		return 0, ErrInvalidTimes
	}
	return cost, nil
}

// TracksPity 表示此卡池需要讀寫玩家的保底狀態。
func (d PoolDefinition) TracksPity() bool { return d.Pity != nil }

// BuildPool 以所有啟用中的卡片與卡池額外指定的卡組裝卡池並驗證。
// 限定卡只有被 refs 明確指定時才會進入；refs 指向未啟用的卡會被略過。
func BuildPool(def PoolDefinition, enabledCards []card.Card, refs []PoolCardRef) (*Pool, error) {
	members := make(map[string]PoolCard)
	byId := make(map[string]card.Card, len(enabledCards))
	for _, c := range enabledCards {
		byId[c.CardId] = c
		if def.IncludeStandardCards && !c.IsLimited {
			members[c.CardId] = PoolCard{Card: c}
		}
	}
	for _, ref := range refs {
		if c, ok := byId[ref.CardId]; ok {
			members[ref.CardId] = PoolCard{Card: c, IsUp: ref.IsUp}
		}
	}

	pool := &Pool{
		PoolDefinition: def,
		up:             make(map[constants.Rarity][]PoolCard),
		normal:         make(map[constants.Rarity][]PoolCard),
	}
	pool.Weights = slices.Clone(def.Weights)
	sort.Slice(pool.Weights, func(i, j int) bool { return pool.Weights[i].Rarity < pool.Weights[j].Rarity })

	ids := make([]string, 0, len(members))
	for id := range members {
		ids = append(ids, id)
	}
	sort.Strings(ids) // 固定順序，讓固定亂數的結果可重現
	for _, id := range ids {
		pc := members[id]
		if pc.IsUp {
			pool.up[pc.Card.Rarity] = append(pool.up[pc.Card.Rarity], pc)
		} else {
			pool.normal[pc.Card.Rarity] = append(pool.normal[pc.Card.Rarity], pc)
		}
	}
	if err := pool.validate(); err != nil {
		return nil, fmt.Errorf("pool %q: %w", def.PoolId, err)
	}
	return pool, nil
}

func (p *Pool) cardCount(r constants.Rarity) int { return len(p.up[r]) + len(p.normal[r]) }

func (p *Pool) maxWeightedRarity() constants.Rarity {
	var max constants.Rarity
	for _, w := range p.Weights {
		if w.Weight > 0 && w.Rarity > max {
			max = w.Rarity
		}
	}
	return max
}

func (p *Pool) validate() error {
	invalid := func(format string, args ...any) error {
		return fmt.Errorf("%w: "+format, append([]any{ErrInvalidPool}, args...)...)
	}
	if p.PoolId == "" {
		return invalid("pool id must not be empty")
	}
	currencyOK := false
	for _, c := range constants.AllCurrencyTypes {
		currencyOK = currencyOK || c == p.Currency
	}
	if !currencyOK {
		return invalid("unknown currency %d", p.Currency)
	}
	if len(p.Costs) == 0 {
		return invalid("costs must not be empty")
	}
	for times, cost := range p.Costs {
		if (times != 1 && times != TenPullTimes) || cost <= 0 {
			return invalid("cost %d for %d pulls is not allowed", cost, times)
		}
	}
	if p.StartAt != nil && p.EndAt != nil && !p.EndAt.After(*p.StartAt) {
		return invalid("endAt must be after startAt")
	}

	seen := make(map[constants.Rarity]bool)
	total := 0
	for _, w := range p.Weights {
		if !w.Rarity.Valid() || w.Weight <= 0 || seen[w.Rarity] {
			return invalid("bad rarity weight %+v", w)
		}
		seen[w.Rarity] = true
		total += w.Weight
	}
	if total == 0 {
		return invalid("rarity weights must not be empty")
	}
	for _, w := range p.Weights {
		if p.cardCount(w.Rarity) == 0 {
			return fmt.Errorf("%w: %s", ErrEmptyCardPool, w.Rarity)
		}
	}

	max := p.maxWeightedRarity()
	if p.TenPullMinRarity != 0 && (!p.TenPullMinRarity.Valid() || p.TenPullMinRarity > max) {
		return invalid("ten pull min rarity %v cannot be satisfied", p.TenPullMinRarity)
	}
	for r, rate := range p.UpRates {
		if !seen[r] || rate <= 0 || rate > 100 {
			return invalid("up rate %d for %v is not allowed", rate, r)
		}
		if len(p.up[r]) == 0 {
			return invalid("up rate for %v but no up card", r)
		}
	}
	for r := range p.up {
		if p.UpRates[r] == 0 {
			return invalid("up cards for %v but no up rate", r)
		}
	}
	if pity := p.Pity; pity != nil {
		if pity.Threshold <= 0 || !pity.MinRarity.Valid() || pity.MinRarity > max {
			return invalid("bad pity rule %+v", *pity)
		}
		if pity.UpGuarantee {
			hasUp := false
			for r := range p.up {
				hasUp = hasUp || r >= pity.MinRarity
			}
			if !hasUp {
				return invalid("up guarantee needs an up card at %v or above", pity.MinRarity)
			}
		}
	}
	return nil
}
