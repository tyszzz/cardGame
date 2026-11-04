package constants

// 怒氣規則：所有卡片共用，不放在卡片定義上。
// 怒氣達到門檻時，輪到行動會施放主動技能並扣除門檻、保留剩餘；
// 每次行動後增加 RageGainOnAction；怒氣上限與受傷增加怒氣等尚未決定。
const (
	RageCastThreshold = 100 // 施放主動技能所需怒氣
	RageCostOnCast    = RageCastThreshold
	RageInitial       = 50 // 開場初始怒氣
	RageGainOnAction  = 50 // 每次行動後增加的怒氣
)
