package handler

import (
	"cardGame/model/packet"
	"context"

	"github.com/topfreegames/pitaya/v2/component"
)

type (
	TestHandler struct {
		component.Base
	}
)

func NewTestHandler() *TestHandler {
	return &TestHandler{}
}

func (h *TestHandler) SummonCard(ctx context.Context, data *packet.SummonCard) {

}
