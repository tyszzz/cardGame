package handler

import (
	"cardGame/model/constants"
	"cardGame/model/packet"
	"cardGame/utils"
	"context"

	"github.com/topfreegames/pitaya/v2/component"
)

type CardService interface {
	GetCards(ctx context.Context, uid string) (*packet.GetCardsResult, error)
}

type CardHandler struct {
	component.Base
	cs CardService
}

func NewCardHandler(cs CardService) *CardHandler {
	return &CardHandler{cs: cs}
}

// GetCards 回傳目前登入玩家持有的卡片，uid 以登入後綁定的 session 為準。
func (h *CardHandler) GetCards(ctx context.Context, _ *packet.Empty) (*packet.GetCardsResult, error) {
	uid := utils.GetSessionFromCtx(ctx).UID()
	if uid == "" {
		utils.Log().Warnf("GetCards rejected: session not logged in")
		return &packet.GetCardsResult{BaseResult: packet.BaseResult{ResultCode: constants.NotLoggedIn}, Cards: []packet.OwnedCard{}}, nil
	}
	result, err := h.cs.GetCards(ctx, uid)
	if err != nil {
		utils.Log().Warnf("UserUid: %s, GetCards failed, err: %v", uid, err)
		return &packet.GetCardsResult{BaseResult: packet.BaseResult{ResultCode: constants.PlayerQueryFailed}, Cards: []packet.OwnedCard{}}, nil
	}
	return result, nil
}
