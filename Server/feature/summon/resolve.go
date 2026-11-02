package summon

import (
	"fmt"

	"cardGame/feature/card"
	"cardGame/model/constants"
)

// ShardReward 為重複卡片轉換的碎片道具與數量。
type ShardReward struct {
	ItemId string
	Amount int64
}

var shardRewards = map[constants.Rarity]ShardReward{
	constants.RarityR:   {"shard_r", 30},
	constants.RaritySR:  {"shard_sr", 30},
	constants.RaritySSR: {"shard_ssr", 10},
	constants.RarityUR:  {"shard_ur", 5},
}

// Result 為單次抽取的結果；重複卡片 IsNew 為 false 並帶有碎片獎勵。
type Result struct {
	Card  card.Card
	IsUp  bool
	IsNew bool
	Shard ShardReward
}

type Outcome struct {
	Results    []Result
	NewCardIds []string         // 需寫入玩家卡片的新卡
	Shards     map[string]int64 // 碎片道具 -> 總數量
	Pity       PityState        // 抽完後的保底狀態
}

// ResolveFunc 由 repository 在 transaction 內以玩家已擁有的卡片與保底狀態呼叫。
type ResolveFunc func(owned map[string]struct{}, pity PityState) (Outcome, error)

// Summon 抽卡並判斷新卡或重複卡，不會修改傳入的 owned。
func (p *Pool) Summon(times int32, owned map[string]struct{}, pity PityState, rnd Rand) (Outcome, error) {
	drawn, newPity, err := p.Draw(times, pity, rnd)
	if err != nil {
		return Outcome{}, err
	}
	outcome, err := Resolve(drawn, owned)
	if err != nil {
		return Outcome{}, err
	}
	outcome.Pity = newPity
	return outcome, nil
}

// Resolve 判斷每次抽取是新卡或重複卡，同一次抽取中先抽到的卡也算已擁有。
func Resolve(drawn []PoolCard, owned map[string]struct{}) (Outcome, error) {
	have := make(map[string]struct{}, len(owned)+len(drawn))
	for id := range owned {
		have[id] = struct{}{}
	}

	outcome := Outcome{
		Results: make([]Result, 0, len(drawn)),
		Shards:  make(map[string]int64),
	}
	for _, pc := range drawn {
		c := pc.Card
		if _, ok := have[c.CardId]; !ok {
			have[c.CardId] = struct{}{}
			outcome.NewCardIds = append(outcome.NewCardIds, c.CardId)
			outcome.Results = append(outcome.Results, Result{Card: c, IsUp: pc.IsUp, IsNew: true})
			continue
		}
		shard, ok := shardRewards[c.Rarity]
		if !ok {
			return Outcome{}, fmt.Errorf("%w: %s", ErrNoShardReward, c.Rarity)
		}
		outcome.Shards[shard.ItemId] += shard.Amount
		outcome.Results = append(outcome.Results, Result{Card: c, IsUp: pc.IsUp, Shard: shard})
	}
	return outcome, nil
}
