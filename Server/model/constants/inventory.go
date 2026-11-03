package constants

type ItemType int16

// 背包容量上限，由 service 在新增時檢查
const (
	BackpackCapacity = 300 // 背包最多可持有的道具種類數
)

const (
	ItemTypeShard    ItemType = 1 // 碎片
	ItemTypeMaterial ItemType = 2 // 素材（例如卡片升級經驗）
)
