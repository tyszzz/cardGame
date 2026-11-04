package battle

import (
	"sort"

	"cardGame/model/constants"
)

// resolveTargets 將「陣營 + 目標形狀」轉成存活的單位清單（依位置排序）。
// 新增目標形狀只需修改這個函式。
func (b *battle) resolveTargets(caster *unit, side constants.Side, shape constants.Target) []*unit {
	if side == constants.SideSelf {
		return []*unit{caster}
	}
	team := caster.ref.Team
	if side == constants.SideEnemy {
		team = opponent(team)
	}
	pool := b.living(team)
	if len(pool) == 0 {
		return nil
	}

	var out []*unit
	switch shape {
	case constants.TargetSingle:
		out = []*unit{lowestHealthRatio(pool)}
	case constants.TargetRandom3:
		shuffled := append([]*unit(nil), pool...)
		b.rng.Shuffle(len(shuffled), func(i, j int) { shuffled[i], shuffled[j] = shuffled[j], shuffled[i] })
		if len(shuffled) > random3Number {
			shuffled = shuffled[:random3Number]
		}
		out = shuffled
	case constants.TargetAll:
		out = pool
	case constants.TargetFrontRow:
		out = pick(pool, func(u *unit) bool { return u.frontRow() })
		if len(out) == 0 { // 前排死光，打後排
			out = pool
		}
	case constants.TargetBackRow:
		out = pick(pool, func(u *unit) bool { return !u.frontRow() })
		if len(out) == 0 { // 後排死光，打前排
			out = pool
		}
	case constants.TargetColumn:
		col := caster.column()
		out = pick(pool, func(u *unit) bool { return u.column() == col })
		if len(out) == 0 { // 同列沒人，改以第一個存活單位的縱排
			col = pool[0].column()
			out = pick(pool, func(u *unit) bool { return u.column() == col })
		}
	case constants.TargetOpposite:
		out = b.opposite(caster, pool)
	}
	sorted := append([]*unit(nil), out...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].ref.Position < sorted[j].ref.Position })
	return sorted
}

// inner function

// lowestHealthRatio 回傳血量百分比最低者，同百分比取位置較小的（pool 已依位置排序）。
func lowestHealthRatio(pool []*unit) *unit {
	best := pool[0]
	for _, u := range pool[1:] {
		// u.health/u.maxHealth < best.health/best.maxHealth，以交叉相乘避免浮點誤差
		if u.health*best.maxHealth < best.health*u.maxHealth {
			best = u
		}
	}
	return best
}

func pick(pool []*unit, keep func(*unit) bool) []*unit {
	var out []*unit
	for _, u := range pool {
		if keep(u) {
			out = append(out, u)
		}
	}
	return out
}

// opposite 回傳對位：同位置的敵人；已死亡則找同縱排另一排；仍沒有則取第一個存活單位。
func (b *battle) opposite(caster *unit, pool []*unit) []*unit {
	pos := caster.ref.Position
	other := pos + frontRowLast
	if pos > frontRowLast {
		other = pos - frontRowLast
	}
	for _, want := range []int{pos, other} {
		for _, u := range pool {
			if u.ref.Position == want {
				return []*unit{u}
			}
		}
	}
	return pool[:1]
}
