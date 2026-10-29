package handler

import (
	"cardGame/feature/summon"
	"cardGame/model/constants"
	"cardGame/model/packet"
	"cardGame/utils"
	"context"
	"errors"

	"github.com/topfreegames/pitaya/v2/component"
)

type SummonService interface {
	SummonCard(ctx context.Context, uid string, data *packet.SummonCard) (*packet.SummonCardResult, error)
}
type (
	SummonHandler struct {
		component.Base
		ss SummonService
	}
)

func NewSummonHandler(ss SummonService) *SummonHandler {
	return &SummonHandler{
		ss: ss,
	}
}

func summonFailed(resultCode int32) *packet.SummonCardResult {
	return &packet.SummonCardResult{
		BaseResult: packet.BaseResult{ResultCode: resultCode},
		Cards:      []packet.SummonCardItem{},
	}
}

func (h *SummonHandler) SummonCard(
	ctx context.Context,
	data *packet.SummonCard,
) (*packet.SummonCardResult, error) {
	// uid 一律以登入後綁定的 session 為準，不採用封包內容
	uid := utils.GetSessionFromCtx(ctx).UID()
	if uid == "" {
		utils.Log().Warnf("SummonCard rejected: session not logged in")
		return summonFailed(constants.NotLoggedIn), nil
	}

	result, err := h.ss.SummonCard(ctx, uid, data)
	if err != nil {
		utils.Log().Warnf("UserUid: %s, SummonCard failed, type: %v, times: %d, err: %v", uid, data.SummonType, data.Times, err)
		resultCode := int32(constants.SummonFailed)
		switch {
		case errors.Is(err, summon.ErrInvalidSummonType), errors.Is(err, summon.ErrInvalidTimes):
			resultCode = constants.SummonInvalidRequest
		case errors.Is(err, summon.ErrNotEnoughCurrency):
			resultCode = constants.SummonNotEnoughCurrency
		}
		return summonFailed(resultCode), nil
	}
	result.ResultCode = constants.Success
	utils.Log().Infof("UserUid: %s, SummonCard success, type: %v, times: %d, cards: %d", uid, data.SummonType, data.Times, len(result.Cards))
	return result, nil
}
