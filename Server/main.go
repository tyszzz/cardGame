package main

import (
	"cardGame/modules"
	"cardGame/utils"

	"github.com/topfreegames/pitaya/v2/component"
)

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
	app.Start()
}
