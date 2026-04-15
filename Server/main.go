package main

import (
	"cardGame/handler"
	"cardGame/modules"
	"cardGame/utils"

	"github.com/topfreegames/pitaya/v2/component"
)

func main() {
	app := utils.App()
	defer app.Shutdown()

	mm := modules.NewModulesManager()
	app.Register(handler.New(mm),
		component.WithName("CardHandler"),
	)
	app.Start()
	mm.ModulesPool["test"].Stop()
}
