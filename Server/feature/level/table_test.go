package level

import (
	"errors"
	"testing"
)

func rows(kind string, needs ...int64) []Row {
	out := make([]Row, 0, len(needs))
	for i, n := range needs {
		out = append(out, Row{Kind: kind, Level: i + 1, ExpRequired: n})
	}
	return out
}

func TestBuildTable(t *testing.T) {
	table, err := BuildTable(KindCard, rows(KindCard, 100, 120, 144))
	if err != nil {
		t.Fatal(err)
	}
	if table.MaxLevel() != 4 {
		t.Errorf("MaxLevel = %d, want 4", table.MaxLevel())
	}
	for level, want := range map[int]int64{1: 100, 2: 120, 3: 144, 0: 0, -1: 0, 4: 0, 5: 0} {
		if got := table.ExpToNext(level); got != want {
			t.Errorf("ExpToNext(%d) = %d, want %d", level, got, want)
		}
	}
}

func TestBuildTableSortsRows(t *testing.T) {
	unsorted := []Row{{KindCard, 3, 144}, {KindCard, 1, 100}, {KindCard, 2, 120}}
	table, err := BuildTable(KindCard, unsorted)
	if err != nil || table.ExpToNext(2) != 120 {
		t.Fatalf("table = %+v, err = %v", table, err)
	}
	if unsorted[0].Level != 3 {
		t.Error("BuildTable must not reorder the caller's slice")
	}
}

func TestBuildTableRejectsInvalid(t *testing.T) {
	cases := map[string][]Row{
		"empty":      nil,
		"gap":        {{KindCard, 1, 100}, {KindCard, 3, 120}},
		"not from 1": {{KindCard, 2, 100}},
		"duplicate":  {{KindCard, 1, 100}, {KindCard, 1, 120}},
		"zero exp":   rows(KindCard, 0, 10),
		"equal exp":  rows(KindCard, 100, 100),
		"decreasing": rows(KindCard, 100, 90),
	}
	for name, in := range cases {
		if _, err := BuildTable(KindCard, in); !errors.Is(err, ErrInvalidTable) {
			t.Errorf("%s: err = %v, want ErrInvalidTable", name, err)
		}
	}
}

func TestBuildTablesSkipsInvalidKinds(t *testing.T) {
	all := append(rows(KindCard, 100, 120), rows(KindPlayer, 5, 5)...)
	tables, err := BuildTables(all)
	if !errors.Is(err, ErrInvalidTable) {
		t.Errorf("err = %v, want ErrInvalidTable for the broken kind", err)
	}
	if tables[KindCard] == nil || tables[KindPlayer] != nil {
		t.Errorf("tables = %v; only the valid kind must remain", tables)
	}
}
