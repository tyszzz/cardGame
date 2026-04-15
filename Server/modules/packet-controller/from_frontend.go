package packetcontroller

import (
	"cardGame/model/packet"
	"cardGame/utils"
	"sync"
)

type PacketController struct {
	wg   *sync.WaitGroup
	inCh chan interface{}
	stop chan struct{}
}

func NewPacketController() *PacketController {
	utils.Log().Info("packet controller started")

	return &PacketController{
		wg:   &sync.WaitGroup{},
		inCh: make(chan interface{}, 512),
		stop: make(chan struct{}),
	}
}

// 外部用這個送封包
func (pc *PacketController) FirePacket(data interface{}) {
	select {
	case pc.inCh <- data:
	case <-pc.stop:
		// controller 停掉後忽略
	}
}

// 統一由manager呼叫開啟
func (pc *PacketController) StartLoop() {
	pc.wg.Add(1)

	go func() {
		defer pc.wg.Done()

		utils.Log().Info("packet controller logicLoop started")

		for {
			select {
			case msg, ok := <-pc.inCh:
				if !ok {
					utils.Log().Warn("packet controller input channel closed")
					return
				}

				switch msg := msg.(type) {
				case packet.SummonCard:
					// TODO: 執行邏輯
				default:
					utils.Log().Errorf("undefined packet: %+v", msg)
				}

			case <-pc.stop:
				utils.Log().Warn("packet controller stop signal received")
				return
			}
		}
	}()
}

// 統一由manager呼叫關閉
func (pc *PacketController) Stop() {
	close(pc.stop) // 通知 goroutine 停止
	close(pc.inCh) // 不再接收事件
	pc.wg.Wait()   // 等 goroutine 結束
	utils.Log().Info("packet controller stopped")
}
