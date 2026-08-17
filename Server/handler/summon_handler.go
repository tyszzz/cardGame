package handler

import (
	"cardGame/model/packet"
	"cardGame/utils"
	"context"

	"github.com/topfreegames/pitaya/v2/component"
)

type SummonService interface {
	SummonCard(ctx context.Context, data *packet.SummonCard) (*packet.SummonCardResult, error)
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

func (h *SummonHandler) SummonCard(
	ctx context.Context,
	data *packet.SummonCard,
) (*packet.SummonCardResult, error) {
	session := utils.GetSessionFromCtx(ctx)
	result, err := h.ss.SummonCard(ctx, data)
	if err != nil {
		utils.Log().Errorf("SummonCard error, Error: %v", err)
		return nil, err
	}
	utils.Log().Infof("UserUid: %s, SummonCard result, CardId: %s, CardName: %s", session.UID(), result.CardId, result.CardName)
	return &packet.SummonCardResult{
		CardId:   result.CardId,
		CardName: result.CardName,
	}, nil
}
