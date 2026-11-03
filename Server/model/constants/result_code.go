package constants

// result code
const (
	Success               = 0
	CreateAccountFailed   = 1
	LoginFailed           = 2
	AccountAlreadyExists  = 3
	NicknameAlreadyExists = 4

	NotLoggedIn             = 5
	SummonInvalidRequest    = 6 // 抽卡類型或次數不合法
	SummonNotEnoughCurrency = 7
	SummonFailed            = 8 // 伺服器端錯誤（卡池設定、資料庫等）
	SummonPoolUnavailable   = 9 // 卡池不存在、未啟用或不在開放時間
)

const PlayerQueryFailed = 10 // 查詢玩家資料失敗（伺服器端錯誤）

// 卡片升級
const (
	CardUpgradeInvalidRequest = 11 // 卡片 ID 無效或玩家未持有該卡
	CardUpgradeMaxLevel       = 12 // 已達等級上限
	CardUpgradeNotEnoughExp   = 13 // 經驗道具不足
	CardUpgradeFailed         = 14 // 伺服器端錯誤
	CardNeedAscend            = 15 // 等級已到當前階的上限，需要進階
	CardNeedStar              = 16 // 等級已到當前星數的上限，需要升星
	CardAscendNotReady        = 17 // 進階需要等級到達當前階的上限
	CardAlreadyTopStage       = 18 // 已是最高階
	CardStarMax               = 19 // 已達星數上限
	CardNotEnoughMaterial     = 20 // 進階或升星的材料不足
)
