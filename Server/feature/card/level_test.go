package card

import "testing"

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
