package handler

import (
	"cardGame/feature/battle"
	"cardGame/model/constants"
	"cardGame/model/packet"
	"cardGame/utils"
	"context"
	"errors"

	"github.com/topfreegames/pitaya/v2/component"
)

type (
	BattleService interface {
		GetStages(ctx context.Context) *packet.GetStagesResult
		GetLineup(ctx context.Context, uid string) (*packet.GetLineupResult, error)
		SaveLineup(ctx context.Context, uid string, data *packet.SaveLineup) error
		StartBattle(ctx context.Context, uid string, data *packet.StartBattle) (*packet.StartBattleResult, error)
	}

	BattleHandler struct {
		component.Base
		bs BattleService
	}
)

func NewBattleHandler(bs BattleService) *BattleHandler {
	return &BattleHandler{bs: bs}
}

// inner function

// battleResultCode 把戰鬥相關錯誤轉成 resultCode，其餘視為伺服器端錯誤，不回傳原始錯誤。
func battleResultCode(err error) int32 {
	switch {
	case errors.Is(err, battle.ErrLineupInvalid), errors.Is(err, battle.ErrStageNotFound):
		return constants.BattleInvalidRequest
	case errors.Is(err, battle.ErrNoLineup):
		return constants.BattleNoLineup
	}
	return constants.BattleFailed
}

// outer function

func (h *BattleHandler) GetStages(ctx context.Context, _ *packet.Empty) (*packet.GetStagesResult, error) {
	if utils.GetSessionFromCtx(ctx).UID() == "" {
		utils.Log().Warnf("GetStages rejected: session not logged in")
		return &packet.GetStagesResult{BaseResult: packet.BaseResult{ResultCode: constants.NotLoggedIn}, Stages: []packet.BattleStage{}}, nil
	}
	return h.bs.GetStages(ctx), nil
}

func (h *BattleHandler) GetLineup(ctx context.Context, _ *packet.Empty) (*packet.GetLineupResult, error) {
	uid := utils.GetSessionFromCtx(ctx).UID()
	if uid == "" {
		utils.Log().Warnf("GetLineup rejected: session not logged in")
		return &packet.GetLineupResult{BaseResult: packet.BaseResult{ResultCode: constants.NotLoggedIn}, Slots: []packet.LineupSlot{}}, nil
	}
	result, err := h.bs.GetLineup(ctx, uid)
	if err != nil {
		utils.Log().Warnf("UserUid: %s, GetLineup failed, err: %v", uid, err)
		return &packet.GetLineupResult{BaseResult: packet.BaseResult{ResultCode: constants.BattleFailed}, Slots: []packet.LineupSlot{}}, nil
	}
	return result, nil
}

func (h *BattleHandler) SaveLineup(ctx context.Context, data *packet.SaveLineup) (*packet.SaveLineupResult, error) {
	uid := utils.GetSessionFromCtx(ctx).UID()
	if uid == "" {
		utils.Log().Warnf("SaveLineup rejected: session not logged in")
		return &packet.SaveLineupResult{BaseResult: packet.BaseResult{ResultCode: constants.NotLoggedIn}}, nil
	}
	if err := h.bs.SaveLineup(ctx, uid, data); err != nil {
		utils.Log().Warnf("UserUid: %s, SaveLineup failed, err: %v", uid, err)
		return &packet.SaveLineupResult{BaseResult: packet.BaseResult{ResultCode: battleResultCode(err)}}, nil
	}
	return &packet.SaveLineupResult{}, nil
}

func (h *BattleHandler) StartBattle(ctx context.Context, data *packet.StartBattle) (*packet.StartBattleResult, error) {
	uid := utils.GetSessionFromCtx(ctx).UID()
	if uid == "" {
		utils.Log().Warnf("StartBattle rejected: session not logged in")
		return &packet.StartBattleResult{BaseResult: packet.BaseResult{ResultCode: constants.NotLoggedIn}, Units: []packet.BattleUnit{}, Events: []packet.BattleEvent{}}, nil
	}
	result, err := h.bs.StartBattle(ctx, uid, data)
	if err != nil {
		utils.Log().Warnf("UserUid: %s, StartBattle failed, err: %v", uid, err)
		return &packet.StartBattleResult{BaseResult: packet.BaseResult{ResultCode: battleResultCode(err)}, Units: []packet.BattleUnit{}, Events: []packet.BattleEvent{}}, nil
	}
	return result, nil
}
