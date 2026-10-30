package constants

// 怒氣規則：所有卡片共用，不放在卡片定義上。
// 怒氣達到門檻時，輪到行動會施放主動技能並扣除門檻、保留剩餘；
// 怒氣上限與累積方式（普攻、受傷等）待戰鬥系統設計時決定。
const (
	RageCastThreshold = 100 // 施放主動技能所需怒氣
	RageCostOnCast    = RageCastThreshold
	RageInitial       = 50 // 開場初始怒氣
)
