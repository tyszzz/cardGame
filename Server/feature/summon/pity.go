package summon

// PityState 為玩家在某個卡池的保底狀態，只有設定保底的卡池會使用。
type PityState struct {
	PullsSinceHigh int  // 自上次出保底稀有度以上的卡後已抽的次數
	GuaranteedUp   bool // 大保底：下一張保底稀有度以上的卡必為 UP 卡
}
