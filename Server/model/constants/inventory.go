package constants

// 背包與卡片容量上限，由 service 在新增時檢查
const (
	BackpackCapacity = 300 // 背包最多可持有的道具種類數
	CardCapacity     = 300 // 玩家最多可持有的卡片張數
)

type ItemType int16
