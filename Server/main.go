package main

import (
	"cardGame/modules"
	"cardGame/utils"

	"github.com/topfreegames/pitaya/v2/component"
)

// outer function

func main() {
	app := utils.App()
	defer app.Shutdown()

	mm := modules.NewModulesManager()
	app.Register(
		mm.GetSummonHandler(),
		component.WithName("SummonHandler"),
	)

	app.Register(
		mm.GetAccountHandler(),
		component.WithName("AccountHandler"),
	)
	app.Register(
		mm.GetPlayerHandler(),
		component.WithName("PlayerHandler"),
	)
	app.Register(
		mm.GetBackpackHandler(),
		component.WithName("BackpackHandler"),
	)

	app.Register(
		mm.GetCardHandler(),
		component.WithName("CardHandler"),
	)
	app.Start()
}
