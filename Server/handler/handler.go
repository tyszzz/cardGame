package handler

import (
	"cardGame/model/packet"
	"cardGame/modules"
	"context"

	"github.com/topfreegames/pitaya/v2/component"
)

type (
	HandlerService struct {
		component.Base
		modulesManager *modules.ModulesManager
	}
)

func New(mm *modules.ModulesManager) *HandlerService {
	return &HandlerService{
		modulesManager: mm,
	}
}

func (h *HandlerService) SummonCard(ctx context.Context, data *packet.SummonCard) {

}
