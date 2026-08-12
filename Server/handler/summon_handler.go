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
	result, err := h.ss.SummonCard(ctx, data)
	if err != nil {
		utils.Log().Errorf("SummonCard error, UID: %s, Error: %v", data.UID, err)
		return nil, err
	}
	utils.Log().Infof("SummonCard result, UID: %s, CardId: %s, CardName: %s", result.UID, result.CardId, result.CardName)
	return &packet.SummonCardResult{
		UID:      result.UID,
		CardId:   result.CardId,
		CardName: result.CardName,
	}, nil
}
