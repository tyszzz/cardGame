package handler

import (
	"cardGame/model/packet"
	"context"

	"github.com/topfreegames/pitaya/v2/component"
)

type SummonService interface {
	SummonCard(ctx context.Context, data *packet.SummonCard)
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

func (h *SummonHandler) SummonCard(ctx context.Context, data *packet.SummonCard) {
	h.ss.SummonCard(ctx, data)
}
