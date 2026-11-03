package growth

import (
	"cardGame/model/constants"
	"errors"
	"fmt"
	"sort"
)

type (
	// StageRow 為 card_stage 的一筆資料。MaxLevel 為 0 代表該階沒有等級上限；
	// Item* 與 ExpPercent 為「從上一階進入本階」的材料，第 0 階沒有。
	StageRow struct {
		Stage      int
		MaxLevel   int
		ItemId     string
		ItemName   string
		ItemQty    int64
		ExpPercent int
	}

	// StarLevelRow 為 card_star_level 的一筆資料。
	StarLevelRow struct {
		Star     int
		MaxLevel int
	}

	// RarityLimitRow 為 card_rarity_limit 的一筆資料。
	RarityLimitRow struct {
		Rarity  constants.Rarity
		MaxStar int
	}

	// StarCostRow 為 card_star_cost 的一筆資料。
	StarCostRow struct {
		Rarity   constants.Rarity
		Star     int
		ItemId   string
		ItemName string
		Quantity int64
	}

	// Cost 為一項道具消耗。
	Cost struct {
		ItemId   string
		ItemName string
		Quantity int64
	}

	// CardGrowth 為通過驗證的卡片成長設定：升級經驗表、階、星數與其費用。
	CardGrowth struct {
		levels    *Table
		stages    []StageRow
		starLevel []int
		maxStar   map[constants.Rarity]int
		starCosts map[constants.Rarity]map[int][]Cost
	}

	// Limit 為卡片目前的等級上限與無法再升級時的原因。
	Limit struct {
		MaxLevel int
		Reason   LimitReason
	}

	LimitReason int
)

const (
	ReasonStage LimitReason = iota + 1 // 受階限制，需要進階
	ReasonStar                         // 受星數限制，需要升星
	ReasonMax                          // 已達該稀有度的最大等級
)

var ErrInvalidConfig = errors.New("invalid growth config")

func NewConfigError(format string, args ...any) error {
	return fmt.Errorf("%w: %s", ErrInvalidConfig, fmt.Sprintf(format, args...))
}

// inner function

// buildCardGrowth 驗證並建立卡片成長設定：階與星從 0 連續、等級上限嚴格遞增且不超過經驗表的等級上限、
// 進階材料數量大於 0、每個稀有度的每一星都有升星材料。
func buildCardGrowth(levels *Table, raw RawConfig) (*CardGrowth, error) {
	cfg := &CardGrowth{
		levels:    levels,
		maxStar:   make(map[constants.Rarity]int),
		starCosts: make(map[constants.Rarity]map[int][]Cost),
	}

	stages := append([]StageRow(nil), raw.Stages...)
	sort.Slice(stages, func(i, j int) bool { return stages[i].Stage < stages[j].Stage })
	if len(stages) < 2 {
		return nil, NewConfigError("at least two stages are required")
	}
	for i, st := range stages {
		last := i == len(stages)-1
		switch {
		case st.Stage != i:
			return nil, NewConfigError("stage %d expected but found %d", i, st.Stage)
		case last && st.MaxLevel != 0:
			return nil, NewConfigError("the last stage %d must not have a max level", st.Stage)
		case !last && st.MaxLevel <= 0:
			return nil, NewConfigError("stage %d needs a max level", st.Stage)
		case i > 0 && !last && st.MaxLevel <= stages[i-1].MaxLevel:
			return nil, NewConfigError("stage %d max level %d must exceed stage %d", st.Stage, st.MaxLevel, st.Stage-1)
		case i > 0 && (st.ItemId == "" || st.ItemQty <= 0 || st.ExpPercent < 0 || st.ExpPercent > 100):
			return nil, NewConfigError("stage %d needs an ascend item, a positive quantity and an exp percent within 0~100", st.Stage)
		}
	}
	for _, st := range stages {
		if st.MaxLevel > levels.MaxLevel() {
			return nil, NewConfigError("stage %d max level %d exceeds the exp table max level %d", st.Stage, st.MaxLevel, levels.MaxLevel())
		}
	}
	cfg.stages = stages

	starLevels := append([]StarLevelRow(nil), raw.StarLevels...)
	sort.Slice(starLevels, func(i, j int) bool { return starLevels[i].Star < starLevels[j].Star })
	if len(starLevels) < 2 {
		return nil, NewConfigError("at least two star levels are required")
	}
	for i, sl := range starLevels {
		switch {
		case sl.Star != i:
			return nil, NewConfigError("star %d expected but found %d", i, sl.Star)
		case sl.MaxLevel <= 0:
			return nil, NewConfigError("star %d needs a positive max level", sl.Star)
		case i > 0 && sl.MaxLevel <= starLevels[i-1].MaxLevel:
			return nil, NewConfigError("star %d max level %d must exceed star %d", sl.Star, sl.MaxLevel, sl.Star-1)
		}
		if sl.MaxLevel > levels.MaxLevel() {
			return nil, NewConfigError("star %d max level %d exceeds the exp table max level %d", sl.Star, sl.MaxLevel, levels.MaxLevel())
		}
		cfg.starLevel = append(cfg.starLevel, sl.MaxLevel)
	}

	for _, r := range raw.Rarities {
		if r.MaxStar < 1 || r.MaxStar >= len(cfg.starLevel) {
			return nil, NewConfigError("rarity %v max star %d must be within 1~%d", r.Rarity, r.MaxStar, len(cfg.starLevel)-1)
		}
		cfg.maxStar[r.Rarity] = r.MaxStar
	}

	for _, c := range raw.StarCosts {
		if c.Quantity <= 0 || c.ItemId == "" {
			return nil, NewConfigError("rarity %v star %d has an invalid cost", c.Rarity, c.Star)
		}
		if cfg.starCosts[c.Rarity] == nil {
			cfg.starCosts[c.Rarity] = make(map[int][]Cost)
		}
		cfg.starCosts[c.Rarity][c.Star] = append(cfg.starCosts[c.Rarity][c.Star], Cost{ItemId: c.ItemId, ItemName: c.ItemName, Quantity: c.Quantity})
	}
	for rarity, maxStar := range cfg.maxStar {
		for star := 1; star <= maxStar; star++ {
			if len(cfg.starCosts[rarity][star]) == 0 {
				return nil, NewConfigError("rarity %v star %d has no star cost", rarity, star)
			}
		}
	}
	return cfg, nil
}

// outer function

// TopStage 為最後一階。
func (c *CardGrowth) TopStage() int { return len(c.stages) - 1 }

// StageMaxLevel 為指定階的等級上限，0 代表不限。
func (c *CardGrowth) StageMaxLevel(stage int) int {
	if stage < 0 || stage >= len(c.stages) {
		return 0
	}
	return c.stages[stage].MaxLevel
}

// MaxStar 為稀有度的星數上限，未設定的稀有度回傳 0。
func (c *CardGrowth) MaxStar(rarity constants.Rarity) int { return c.maxStar[rarity] }

// StarLevelCap 為指定星數的等級上限；星數超出範圍時回傳 0。
func (c *CardGrowth) StarLevelCap(star int) int {
	if star < 0 || star >= len(c.starLevel) {
		return 0
	}
	return c.starLevel[star]
}

// LimitFor 回傳卡片目前的等級上限（階與星數兩者取較小）與到達上限時的原因。
func (c *CardGrowth) LimitFor(rarity constants.Rarity, stage, star int) Limit {
	starCap := c.StarLevelCap(star)
	stageCap := c.StageMaxLevel(stage)
	switch {
	case stageCap != 0 && stageCap <= starCap:
		return Limit{MaxLevel: stageCap, Reason: ReasonStage}
	case star < c.MaxStar(rarity):
		return Limit{MaxLevel: starCap, Reason: ReasonStar}
	default:
		return Limit{MaxLevel: starCap, Reason: ReasonMax}
	}
}

// AscendCost 回傳從 stage 進入 stage+1 所需的進階道具與經驗比例；已是最高階時 ok 為 false。
func (c *CardGrowth) AscendCost(stage int) (item Cost, expPercent int, ok bool) {
	if stage < 0 || stage >= c.TopStage() {
		return Cost{}, 0, false
	}
	next := c.stages[stage+1]
	return Cost{ItemId: next.ItemId, ItemName: next.ItemName, Quantity: next.ItemQty}, next.ExpPercent, true
}

// AscendExp 回傳離開 stage 時需要的經驗：該階等級範圍內的升級經驗總和 × 進階經驗比例，向上取整。
// 該階範圍為「上一階的上限（第 0 階從 1 級）」到「本階上限」。
func (c *CardGrowth) AscendExp(stage int) int64 {
	_, percent, ok := c.AscendCost(stage)
	if !ok {
		return 0
	}
	from := 1
	if stage > 0 {
		from = c.StageMaxLevel(stage - 1)
	}
	var total int64
	for lv := from; lv < c.StageMaxLevel(stage); lv++ {
		total += c.levels.ExpToNext(lv)
	}
	return (total*int64(percent) + 99) / 100
}

// StarCost 回傳升到 star 星所需的道具；沒有設定時 ok 為 false。
func (c *CardGrowth) StarCost(rarity constants.Rarity, star int) ([]Cost, bool) {
	costs := c.starCosts[rarity][star]
	return costs, len(costs) > 0
}

// MaxLevel 為經驗表的等級上限（不考慮階與星數）。
func (c *CardGrowth) MaxLevel() int { return c.levels.MaxLevel() }

// ExpToNext 回傳指定等級升到下一級所需經驗；已達上限或等級不合法時回傳 0。
func (c *CardGrowth) ExpToNext(level int) int64 { return c.levels.ExpToNext(level) }
