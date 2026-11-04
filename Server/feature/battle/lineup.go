package battle

import (
	"context"
	"errors"
)

type (
	// LineupSlot 是陣容中的一格：位置 1~3 為前排、4~6 為後排。
	LineupSlot struct {
		Position int
		CardId   string
	}

	LineupRepository interface {
		// GetLineup 回傳玩家儲存的陣容（依位置排序），沒有陣容時回傳空清單。
		GetLineup(ctx context.Context, uid string) ([]LineupSlot, error)
		// SaveLineup 在單一 transaction 內整批覆蓋玩家的陣容。
		SaveLineup(ctx context.Context, uid string, slots []LineupSlot) error
	}
)

var (
	ErrLineupInvalid = errors.New("invalid lineup")
	ErrNoLineup      = errors.New("player has no lineup")
	ErrStageNotFound = errors.New("battle stage not found")
)

// MaxLineupSize 是陣容最多的卡片數（位置 1~6）。
const MaxLineupSize = maxPosition
