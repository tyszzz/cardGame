package modules

import (
	"cardGame/handler"
	"cardGame/services"
)

type ModulesManager struct {
	accountHandler *handler.AccountHandler
	summonHandler  *handler.SummonHandler
}

type ModulesFunc interface {
	StartLoop()
	Stop()
	FirePacket(data interface{})
}

func NewModulesManager() *ModulesManager {
	mm := &ModulesManager{}

	// account
	mm.accountHandler = handler.NewAccountHandler(
		services.NewAccountService(),
	)

	// summon
	mm.summonHandler = handler.NewSummonHandler(
		services.NewSummonService(),
	)
	return mm
}

func (mm *ModulesManager) GetSummonHandler() *handler.SummonHandler {
	return mm.summonHandler
}

func (mm *ModulesManager) GetAccountHandler() *handler.AccountHandler {
	return mm.accountHandler
}
