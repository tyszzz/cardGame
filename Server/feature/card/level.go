package card

import (
	"cardGame/model/constants"
	"math"
)

// ExpToNext 回傳目前等級升到下一級所需的經驗；已達上限或等級不合法時回傳 0。
// 1 級升 2 級需要 CardExpBase，之後每級為上一級需求乘 1.2 後向上取整（以整數運算避免浮點誤差）。
func ExpToNext(level int) int64 {
	if level < 1 || level >= constants.CardMaxLevel {
		return 0
	}
	need := int64(constants.CardExpBase)
	for l := 1; l < level; l++ {
		need = (need*12 + 9) / 10
	}
	return need
}

// StatAtLevel 回傳卡片在指定等級的屬性：基礎值 + (等級-1) × ceil(基礎值 × ratio)。
func StatAtLevel(base int, ratio float64, level int) int {
	if level < 1 {
		level = 1
	}
	// 容忍浮點誤差，避免 500×1.2=600.0000000000001 被誤進位成 601
	perLevel := int(math.Ceil(float64(base)*ratio - 1e-9))
	return base + (level-1)*perLevel
}
