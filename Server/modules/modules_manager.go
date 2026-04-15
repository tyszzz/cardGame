package modules

import (
	"cardGame/model/constants"
	packetcontroller "cardGame/modules/packet-controller"
)

type ModulesManager struct {
	ModulesPool map[constants.ControllerType]ModulesFunc
}

type ModulesFunc interface {
	StartLoop()
	Stop()
	FirePacket(data interface{})
}

func NewModulesManager() *ModulesManager {
	mm := &ModulesManager{
		ModulesPool: make(map[string]ModulesFunc),
	}
	pc := packetcontroller.NewPacketController()
	mm.ModulesPool[constants.PACKETCONTROLLER] = pc
	return mm
}
func InitModules() {

}
