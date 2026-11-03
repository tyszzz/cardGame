package growth

import (
	"context"
	"errors"
	"fmt"
)

type (
	// RawConfig 為從資料庫讀出、尚未驗證的設定。
	RawConfig struct {
		LevelExp   []LevelRow
		Stages     []StageRow
		StarLevels []StarLevelRow
		Rarities   []RarityLimitRow
		StarCosts  []StarCostRow
	}

	// Config 為所有「升」的設定。每個部分各自驗證、各自啟用：某一部分不合法時，
	// 只有該部分的功能不可用，不影響其他部分。
	Config struct {
		card      *CardGrowth
		cardErr   error
		player    *Table
		playerErr error
	}

	GrowthRepository interface {
		// LoadGrowth 讀取所有升級、進階、升星的設定。
		LoadGrowth(ctx context.Context) (RawConfig, error)
	}
)

var ErrUnavailable = errors.New("growth config unavailable")

// outer function

// Build 驗證並建立設定，永遠回傳 Config；不合法的部分以 Problems 回報，對應的 getter 會回傳 ErrUnavailable。
func Build(raw RawConfig) *Config {
	cfg := &Config{}

	byKind := make(map[string][]LevelRow)
	for _, row := range raw.LevelExp {
		byKind[row.Kind] = append(byKind[row.Kind], row)
	}

	cardLevels, err := BuildTable(KindCard, byKind[KindCard])
	if err == nil {
		cfg.card, err = buildCardGrowth(cardLevels, raw)
	}
	cfg.cardErr = err

	cfg.player, cfg.playerErr = BuildTable(KindPlayer, byKind[KindPlayer])
	return cfg
}

// Card 回傳卡片成長設定；該部分不合法時回傳 ErrUnavailable。
func (c *Config) Card() (*CardGrowth, error) {
	if c.card == nil {
		return nil, fmt.Errorf("%w: card: %v", ErrUnavailable, c.cardErr)
	}
	return c.card, nil
}

// Player 回傳玩家升級經驗表；該部分不合法時回傳 ErrUnavailable。
func (c *Config) Player() (*Table, error) {
	if c.player == nil {
		return nil, fmt.Errorf("%w: player: %v", ErrUnavailable, c.playerErr)
	}
	return c.player, nil
}

// Problems 列出不合法的部分與原因。
func (c *Config) Problems() []error {
	var problems []error
	if c.card == nil {
		problems = append(problems, fmt.Errorf("card: %w", c.cardErr))
	}
	if c.player == nil {
		problems = append(problems, fmt.Errorf("player: %w", c.playerErr))
	}
	return problems
}

// KeepValidFrom 讓新設定中不合法的部分沿用舊設定中合法的部分，回傳被沿用的部分名稱。
func (c *Config) KeepValidFrom(old *Config) []string {
	var kept []string
	if old == nil {
		return kept
	}
	if c.card == nil && old.card != nil {
		c.card = old.card
		kept = append(kept, "card")
	}
	if c.player == nil && old.player != nil {
		c.player = old.player
		kept = append(kept, "player")
	}
	return kept
}
