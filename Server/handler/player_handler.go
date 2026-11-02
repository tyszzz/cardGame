package handler

import (
	"cardGame/model/constants"
	"cardGame/model/packet"
	"cardGame/utils"
	"context"

	"github.com/topfreegames/pitaya/v2/component"
)

type PlayerService interface {
	GetCurrencies(ctx context.Context, uid string) (*packet.GetCurrenciesResult, error)
}

type PlayerHandler struct {
	component.Base
	ps PlayerService
}

func NewPlayerHandler(ps PlayerService) *PlayerHandler {
	return &PlayerHandler{ps: ps}
}

// GetCurrencies 回傳目前登入玩家的貨幣餘額，uid 以登入後綁定的 session 為準。
func (h *PlayerHandler) GetCurrencies(ctx context.Context, _ *packet.GetCurrencies) (*packet.GetCurrenciesResult, error) {
	uid := utils.GetSessionFromCtx(ctx).UID()
	if uid == "" {
		utils.Log().Warnf("GetCurrencies rejected: session not logged in")
		return &packet.GetCurrenciesResult{BaseResult: packet.BaseResult{ResultCode: constants.NotLoggedIn}}, nil
	}
	result, err := h.ps.GetCurrencies(ctx, uid)
	if err != nil {
		utils.Log().Warnf("UserUid: %s, GetCurrencies failed, err: %v", uid, err)
		return &packet.GetCurrenciesResult{BaseResult: packet.BaseResult{ResultCode: constants.PlayerQueryFailed}}, nil
	}
	return result, nil
}
