package card

import "math"

// outer function

// StatAtLevel 回傳卡片在指定等級的屬性：基礎值 + (等級-1) × ceil(基礎值 × ratio)。
func StatAtLevel(base int, ratio float64, level int) int {
	if level < 1 {
		level = 1
	}
	// 容忍浮點誤差，避免 500×1.2=600.0000000000001 被誤進位成 601
	perLevel := int(math.Ceil(float64(base)*ratio - 1e-9))
	return base + (level-1)*perLevel
}
