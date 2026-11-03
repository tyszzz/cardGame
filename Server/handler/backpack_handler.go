package handler

import (
	"cardGame/model/constants"
	"cardGame/model/packet"
	"cardGame/utils"
	"context"

	"github.com/topfreegames/pitaya/v2/component"
)

type BackpackService interface {
	GetBackpack(ctx context.Context, uid string) (*packet.GetBackpackResult, error)
}

type BackpackHandler struct {
	component.Base
	bs BackpackService
}

func NewBackpackHandler(bs BackpackService) *BackpackHandler {
	return &BackpackHandler{bs: bs}
}

// GetBackpack 回傳目前登入玩家的背包道具，uid 以登入後綁定的 session 為準。
func (h *BackpackHandler) GetBackpack(ctx context.Context, _ *packet.Empty) (*packet.GetBackpackResult, error) {
	uid := utils.GetSessionFromCtx(ctx).UID()
	if uid == "" {
		utils.Log().Warnf("GetBackpack rejected: session not logged in")
		return &packet.GetBackpackResult{BaseResult: packet.BaseResult{ResultCode: constants.NotLoggedIn}, Items: []packet.BackpackItem{}}, nil
	}
	result, err := h.bs.GetBackpack(ctx, uid)
	if err != nil {
		utils.Log().Warnf("UserUid: %s, GetBackpack failed, err: %v", uid, err)
		return &packet.GetBackpackResult{BaseResult: packet.BaseResult{ResultCode: constants.PlayerQueryFailed}, Items: []packet.BackpackItem{}}, nil
	}
	return result, nil
}
