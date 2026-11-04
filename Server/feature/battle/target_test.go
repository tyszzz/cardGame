package battle

import (
	"reflect"
	"testing"

	"cardGame/model/constants"
)

func fullBattle(t *testing.T) (*battle, *unit) {
	t.Helper()
	var ally, enemy []UnitInput
	for i := 1; i <= 6; i++ {
		ally = append(ally, uin("a", i, 1, 0, 10))
		enemy = append(enemy, uin("e", i, 1, 0, 10))
	}
	b, err := newBattle(Input{Ally: ally, Enemy: enemy, Seed: 1})
	if err != nil {
		t.Fatal(err)
	}
	return b, b.units[TeamAlly][1] // 我方 2 號
}

func positions(us []*unit) []int {
	var out []int
	for _, u := range us {
		out = append(out, u.ref.Position)
	}
	return out
}

func TestResolveTargets(t *testing.T) {
	b, caster := fullBattle(t)
	enemy := constants.SideEnemy
	cases := []struct {
		name  string
		side  constants.Side
		shape constants.Target
		want  []int
	}{
		{"single", enemy, constants.TargetSingle, []int{1}},
		{"all", enemy, constants.TargetAll, []int{1, 2, 3, 4, 5, 6}},
		{"front", enemy, constants.TargetFrontRow, []int{1, 2, 3}},
		{"back", enemy, constants.TargetBackRow, []int{4, 5, 6}},
		{"column", enemy, constants.TargetColumn, []int{2, 5}},
		{"opposite", enemy, constants.TargetOpposite, []int{2}},
		{"ally all", constants.SideAlly, constants.TargetAll, []int{1, 2, 3, 4, 5, 6}},
		{"self", constants.SideSelf, constants.TargetAll, []int{2}},
	}
	for _, c := range cases {
		if got := positions(b.resolveTargets(caster, c.side, c.shape)); !reflect.DeepEqual(got, c.want) {
			t.Errorf("%s = %v, want %v", c.name, got, c.want)
		}
	}
}

func TestResolveRandom3(t *testing.T) {
	b, caster := fullBattle(t)
	got := b.resolveTargets(caster, constants.SideEnemy, constants.TargetRandom3)
	if len(got) != 3 {
		t.Fatalf("random3 len = %d", len(got))
	}
	seen := map[int]bool{}
	for _, u := range got {
		if seen[u.ref.Position] {
			t.Fatal("duplicate target")
		}
		seen[u.ref.Position] = true
	}
	for _, u := range b.units[TeamEnemy][:4] {
		u.health = 0
	}
	if n := len(b.resolveTargets(caster, constants.SideEnemy, constants.TargetRandom3)); n != 2 {
		t.Fatalf("random3 with 2 alive = %d", n)
	}
}

func TestResolveFallbacks(t *testing.T) {
	b, caster := fullBattle(t)
	enemy := b.units[TeamEnemy]

	// 對位 2 號死亡 → 同縱排後排 5 號。
	enemy[1].health = 0
	if got := positions(b.resolveTargets(caster, constants.SideEnemy, constants.TargetOpposite)); !reflect.DeepEqual(got, []int{5}) {
		t.Fatalf("opposite fallback = %v", got)
	}
	// 同縱排全死 → 取第一個存活單位。
	enemy[4].health = 0
	if got := positions(b.resolveTargets(caster, constants.SideEnemy, constants.TargetOpposite)); !reflect.DeepEqual(got, []int{1}) {
		t.Fatalf("opposite last resort = %v", got)
	}
	// 縱排同列沒人 → 用第一個存活單位的縱排（1 與 4）。
	if got := positions(b.resolveTargets(caster, constants.SideEnemy, constants.TargetColumn)); !reflect.DeepEqual(got, []int{1, 4}) {
		t.Fatalf("column fallback = %v", got)
	}
	// 前排死光 → 前排類目標打後排。
	for _, i := range []int{0, 2} {
		enemy[i].health = 0
	}
	if got := positions(b.resolveTargets(caster, constants.SideEnemy, constants.TargetFrontRow)); !reflect.DeepEqual(got, []int{4, 6}) {
		t.Fatalf("front row fallback = %v", got)
	}
	// 後排死光 → 後排類目標改打前排（此時前排只剩 2 號）。
	for _, i := range []int{3, 5} {
		enemy[i].health = 0
	}
	enemy[1].health = 5
	if got := positions(b.resolveTargets(caster, constants.SideEnemy, constants.TargetBackRow)); !reflect.DeepEqual(got, []int{2}) {
		t.Fatalf("back row fallback = %v", got)
	}
}

func TestResolveSingleLowestHealthRatio(t *testing.T) {
	b, caster := fullBattle(t)
	enemy := b.units[TeamEnemy]
	// 全滿血 → 位置最小。
	if got := positions(b.resolveTargets(caster, constants.SideEnemy, constants.TargetSingle)); !reflect.DeepEqual(got, []int{1}) {
		t.Fatalf("full health = %v", got)
	}
	enemy[3].health = 5 // 50%
	enemy[5].health = 5 // 50%，同百分比取位置小的
	enemy[2].health = 6 // 60%
	if got := positions(b.resolveTargets(caster, constants.SideEnemy, constants.TargetSingle)); !reflect.DeepEqual(got, []int{4}) {
		t.Fatalf("lowest ratio = %v", got)
	}
	// 百分比而非絕對值：maxHealth 不同。
	enemy[0].maxHealth, enemy[0].health = 100, 40 // 40%
	if got := positions(b.resolveTargets(caster, constants.SideEnemy, constants.TargetSingle)); !reflect.DeepEqual(got, []int{1}) {
		t.Fatalf("ratio not absolute = %v", got)
	}
	// 我方單體（補血）同樣選最低百分比。
	b.units[TeamAlly][4].health = 1
	if got := positions(b.resolveTargets(caster, constants.SideAlly, constants.TargetSingle)); !reflect.DeepEqual(got, []int{5}) {
		t.Fatalf("ally single = %v", got)
	}
}
