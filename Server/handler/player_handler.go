package handler

import (
	"cardGame/model/constants"
	"cardGame/model/packet"
	"cardGame/utils"
	"context"

	"github.com/topfreegames/pitaya/v2/component"
)

type PlayerService interface {
	GetInfo(ctx context.Context, uid string) (*packet.GetInfoResult, error)
}

type PlayerHandler struct {
	component.Base
	ps PlayerService
}

func NewPlayerHandler(ps PlayerService) *PlayerHandler {
	return &PlayerHandler{ps: ps}
}

// GetInfo 回傳目前登入玩家的基本資料與貨幣，uid 以登入後綁定的 session 為準。
func (h *PlayerHandler) GetInfo(ctx context.Context, _ *packet.Empty) (*packet.GetInfoResult, error) {
	uid := utils.GetSessionFromCtx(ctx).UID()
	if uid == "" {
		utils.Log().Warnf("GetInfo rejected: session not logged in")
		return &packet.GetInfoResult{BaseResult: packet.BaseResult{ResultCode: constants.NotLoggedIn}}, nil
	}
	result, err := h.ps.GetInfo(ctx, uid)
	if err != nil {
		utils.Log().Warnf("UserUid: %s, GetInfo failed, err: %v", uid, err)
		return &packet.GetInfoResult{BaseResult: packet.BaseResult{ResultCode: constants.PlayerQueryFailed}}, nil
	}
	return result, nil
}
