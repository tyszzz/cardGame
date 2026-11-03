package growth

import (
	"errors"
	"testing"
)

func withPlayerTable(raw RawConfig, needs ...int64) RawConfig {
	raw.LevelExp = append(raw.LevelExp, levelRows(KindPlayer, needs...)...)
	return raw
}

func TestBuildAllPartsValid(t *testing.T) {
	cfg := Build(withPlayerTable(validRaw(), 1, 2, 3))
	if _, err := cfg.Card(); err != nil {
		t.Errorf("card: %v", err)
	}
	player, err := cfg.Player()
	if err != nil || player.MaxLevel() != 4 {
		t.Errorf("player = %+v, err = %v", player, err)
	}
	if len(cfg.Problems()) != 0 {
		t.Errorf("problems = %v", cfg.Problems())
	}
}

func TestBuildPartsAreIndependent(t *testing.T) {
	// 玩家表不合法：只有玩家不可用，卡片不受影響
	cfg := Build(withPlayerTable(validRaw(), 5, 5))
	if _, err := cfg.Card(); err != nil {
		t.Errorf("a broken player table must not disable cards: %v", err)
	}
	if _, err := cfg.Player(); !errors.Is(err, ErrUnavailable) {
		t.Errorf("player err = %v, want ErrUnavailable", err)
	}
	if len(cfg.Problems()) != 1 {
		t.Errorf("problems = %v, want exactly the player problem", cfg.Problems())
	}

	// 卡片設定不合法：只有卡片不可用，玩家不受影響
	raw := withPlayerTable(validRaw(), 1, 2, 3)
	raw.Stages = raw.Stages[:1]
	cfg = Build(raw)
	if _, err := cfg.Card(); !errors.Is(err, ErrUnavailable) {
		t.Errorf("card err = %v, want ErrUnavailable", err)
	}
	if _, err := cfg.Player(); err != nil {
		t.Errorf("a broken card config must not disable the player table: %v", err)
	}
}

func TestBuildEmptyConfig(t *testing.T) {
	cfg := Build(RawConfig{})
	if _, err := cfg.Card(); !errors.Is(err, ErrUnavailable) {
		t.Errorf("card err = %v", err)
	}
	if _, err := cfg.Player(); !errors.Is(err, ErrUnavailable) {
		t.Errorf("player err = %v", err)
	}
	if len(cfg.Problems()) != 2 {
		t.Errorf("problems = %v", cfg.Problems())
	}
}

func TestKeepValidFrom(t *testing.T) {
	old := Build(withPlayerTable(validRaw(), 1, 2, 3))

	bad := validRaw()
	bad.Stages = bad.Stages[:1]
	next := Build(withPlayerTable(bad, 1, 2, 3, 4))
	kept := next.KeepValidFrom(old)
	if len(kept) != 1 || kept[0] != "card" {
		t.Fatalf("kept = %v, want only the card part", kept)
	}
	if _, err := next.Card(); err != nil {
		t.Errorf("card must fall back to the old valid part: %v", err)
	}
	player, _ := next.Player()
	if player.MaxLevel() != 5 {
		t.Errorf("the new valid player table must be used, max level = %d", player.MaxLevel())
	}

	if kept := Build(RawConfig{}).KeepValidFrom(nil); len(kept) != 0 {
		t.Errorf("nothing to keep without an old config, got %v", kept)
	}
}
