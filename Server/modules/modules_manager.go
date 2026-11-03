package modules

import (
	postgresDb "cardGame/database/postgres"
	"cardGame/handler"
	postgresRepo "cardGame/repository/postgres"
	"cardGame/services"
)

type ModulesManager struct {
	accountHandler  *handler.AccountHandler
	summonHandler   *handler.SummonHandler
	playerHandler   *handler.PlayerHandler
	backpackHandler *handler.BackpackHandler
	cardHandler     *handler.CardHandler
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
		services.NewSummonService(postgresRepo.InitSummonRepo(db)),
	)
	// player
	mm.playerHandler = handler.NewPlayerHandler(
		services.NewPlayerService(postgresRepo.InitPlayerRepo(db)),
	)
	// inventory
	inventoryService := services.NewInventoryService(postgresRepo.InitInventoryRepo(db))
	mm.backpackHandler = handler.NewBackpackHandler(inventoryService)
	mm.cardHandler = handler.NewCardHandler(inventoryService)
	return mm
}

func (mm *ModulesManager) GetSummonHandler() *handler.SummonHandler {
	return mm.summonHandler
}

func (mm *ModulesManager) GetAccountHandler() *handler.AccountHandler {
	return mm.accountHandler
}

func (mm *ModulesManager) GetPlayerHandler() *handler.PlayerHandler {
	return mm.playerHandler
}

func (mm *ModulesManager) GetBackpackHandler() *handler.BackpackHandler {
	return mm.backpackHandler
}

func (mm *ModulesManager) GetCardHandler() *handler.CardHandler {
	return mm.cardHandler
}
