package constants

// 背包容量上限，由 service 在新增時檢查
const (
	BackpackCapacity = 300 // 背包最多可持有的道具種類數
)

type ItemType int16

const (
	ItemTypeShard ItemType = 1 // 碎片
)
