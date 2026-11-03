package main

import (
	"math"
	"strings"
	"testing"
)

func sumRows(need []int64, from, to int) int64 { // 第 from ~ to 列（含），列編號從 1 開始
	var total int64
	for l := from; l <= to; l++ {
		total += need[l-1]
	}
	return total
}

func within(got int64, want, tolerance float64) bool {
	return math.Abs(float64(got)-want)/want <= tolerance
}

func TestCardCurveShape(t *testing.T) {
	need := cardCurve()
	if len(need) != maxLevel-1 {
		t.Fatalf("rows = %d, want %d", len(need), maxLevel-1)
	}
	if need[0] != startExp {
		t.Errorf("level 1 = %d, want %d", need[0], startExp)
	}
	for i := 1; i < len(need); i++ {
		if need[i] <= need[i-1] {
			t.Fatalf("level %d (%d) must exceed level %d (%d)", i+1, need[i], i, need[i-1])
		}
	}
}

func TestCardCurveTotals(t *testing.T) {
	need := cardCurve()
	if got := sumRows(need, earlyRows-rowsPerBand+1, earlyRows); !within(got, earlyBandTotal, 0.0001) {
		t.Errorf("91~100 total = %d, want ~%d", got, earlyBandTotal)
	}
	if got := sumRows(need, 1, len(need)); !within(got, totalExp, 0.00001) {
		t.Errorf("grand total = %d, want ~%d", got, int64(totalExp))
	}
	// 101~110 約 180 萬
	if got := sumRows(need, earlyRows+1, earlyRows+rowsPerBand); got < 1_750_000 || got > 1_950_000 {
		t.Errorf("101~110 total = %d, want about 1.8 million", got)
	}
}

func TestCardCurveBandsAfterEarlyGrowByConstantRatio(t *testing.T) {
	need := cardCurve()
	var first float64
	for from := earlyRows + 1; from+2*rowsPerBand-1 <= len(need); from += rowsPerBand {
		prev := sumRows(need, from, from+rowsPerBand-1)
		cur := sumRows(need, from+rowsPerBand, from+2*rowsPerBand-1)
		ratio := float64(cur) / float64(prev)
		if first == 0 {
			first = ratio
		}
		if math.Abs(ratio-first) > 0.002 {
			t.Errorf("band at %d: ratio %.4f differs from first ratio %.4f", from, ratio, first)
		}
	}
}

func TestCardCurveEarlyGrowthSlowsDownPerSegment(t *testing.T) {
	need := cardCurve()
	rate := func(from, to int) float64 { // 第 from 列到第 to 列的平均每列成長率
		return math.Pow(float64(need[to-1])/float64(need[from-1]), 1.0/float64(to-from))
	}
	segments := [][2]int{{1, 30}, {30, 50}, {50, 80}, {80, 100}, {101, 200}}
	prev := math.Inf(1)
	for _, s := range segments {
		r := rate(s[0], s[1])
		if r >= prev {
			t.Errorf("segment %v growth %.4f must be below the previous segment %.4f", s, r, prev)
		}
		prev = r
	}
}

func TestCardCurveSmoothAtBoundary(t *testing.T) {
	need := cardCurve()
	if ratio := float64(need[earlyRows]) / float64(need[earlyRows-1]); ratio > 1.03 {
		t.Errorf("level %d vs %d grows %.3fx; the curve must stay smooth", earlyRows+1, earlyRows, ratio)
	}
}

func TestRender(t *testing.T) {
	out := render([]int64{100, 108, 120})
	for _, want := range []string{"DO NOT EDIT", "('card', 1, 100),", "('card', 3, 120)\n", "ON CONFLICT (kind, level) DO UPDATE"} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q:\n%s", want, out)
		}
	}
}
