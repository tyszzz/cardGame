package card

import (
	"cardGame/model/constants"
	"testing"
)

func TestExpToNext(t *testing.T) {
	want := map[int]int64{1: 100, 2: 120, 3: 144, 4: 173, 5: 208, 6: 250, 7: 300}
	for level, exp := range want {
		if got := ExpToNext(level); got != exp {
			t.Errorf("ExpToNext(%d) = %d, want %d", level, got, exp)
		}
	}
	for _, level := range []int{0, -1, constants.CardMaxLevel, constants.CardMaxLevel + 1} {
		if got := ExpToNext(level); got != 0 {
			t.Errorf("ExpToNext(%d) = %d, want 0", level, got)
		}
	}
	if ExpToNext(constants.CardMaxLevel-1) <= 0 {
		t.Error("the level below the cap must have a positive cost")
	}
}

func TestStatAtLevel(t *testing.T) {
	cases := []struct {
		base  int
		ratio float64
		level int
		want  int
	}{
		{500, 1.2, 1, 500},
		{500, 1.2, 2, 1100},
		{500, 1.2, 10, 5900},
		{8000, 0.8, 2, 14400},
		{100, 1, 3, 300},
		{101, 0.5, 2, 152}, // 50.5 向上取整為 51
		{500, 1.2, 0, 500}, // 不合法等級視為 1
		{500, 0, 50, 500},  // ratio 為 0 時不成長
	}
	for _, c := range cases {
		if got := StatAtLevel(c.base, c.ratio, c.level); got != c.want {
			t.Errorf("StatAtLevel(%d, %v, %d) = %d, want %d", c.base, c.ratio, c.level, got, c.want)
		}
	}
}
