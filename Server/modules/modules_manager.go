package modules

import (
	postgresDb "cardGame/database/postgres"
	"cardGame/handler"
	postgresRepo "cardGame/repository/postgres"
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

	db := postgresDb.Init()
	// account
	accountRepo := postgresRepo.InitAccountRepo(db)
	mm.accountHandler = handler.NewAccountHandler(
		services.NewAccountService(accountRepo),
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
