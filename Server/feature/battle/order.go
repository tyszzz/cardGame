package battle

// actionOrder 回傳本回合的行動順序：先手方 1 號、後手方 1 號、先手方 2 號…；
// 某方人數較少時，剩下的單位接著排。依回合開始時存活的單位計算，行動時再檢查是否已死亡。
func (b *battle) actionOrder() []*unit {
	first, second := TeamAlly, TeamEnemy
	if b.enemyFirst {
		first, second = TeamEnemy, TeamAlly
	}
	a, c := b.living(first), b.living(second)
	out := make([]*unit, 0, len(a)+len(c))
	for i := 0; i < len(a) || i < len(c); i++ {
		if i < len(a) {
			out = append(out, a[i])
		}
		if i < len(c) {
			out = append(out, c[i])
		}
	}
	return out
}
