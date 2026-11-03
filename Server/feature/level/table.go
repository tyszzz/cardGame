package level

import (
	"errors"
	"fmt"
	"sort"
)

type (
	// Row 為 level_exp 的一筆資料：Level 為目前等級，ExpRequired 為升到下一級所需經驗。
	Row struct {
		Kind        string
		Level       int
		ExpRequired int64
	}

	// Table 為單一種類（卡片或玩家）的升級經驗表，等級上限為最大 Level + 1。
	Table struct {
		need []int64 // need[i] 為 i+1 級升到 i+2 級所需經驗
	}
)

const (
	KindCard   = "card"
	KindPlayer = "player"
)

var (
	ErrInvalidTable     = errors.New("invalid level exp table")
	ErrTableUnavailable = errors.New("level exp table unavailable")
)

// BuildTable 驗證並建立單一種類的經驗表：等級須從 1 連續、經驗大於 0 且嚴格遞增。
func BuildTable(kind string, rows []Row) (*Table, error) {
	if len(rows) == 0 {
		return nil, fmt.Errorf("%w: %s has no rows", ErrInvalidTable, kind)
	}
	sorted := append([]Row(nil), rows...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].Level < sorted[j].Level })

	need := make([]int64, 0, len(sorted))
	for i, row := range sorted {
		switch {
		case row.Level != i+1:
			return nil, fmt.Errorf("%w: %s expects level %d but found %d", ErrInvalidTable, kind, i+1, row.Level)
		case row.ExpRequired <= 0:
			return nil, fmt.Errorf("%w: %s level %d needs a positive exp", ErrInvalidTable, kind, row.Level)
		case i > 0 && row.ExpRequired <= need[i-1]:
			return nil, fmt.Errorf("%w: %s level %d exp %d must exceed level %d exp %d",
				ErrInvalidTable, kind, row.Level, row.ExpRequired, row.Level-1, need[i-1])
		}
		need = append(need, row.ExpRequired)
	}
	return &Table{need: need}, nil
}

// BuildTables 依 Kind 分組建立經驗表；不合法的種類會被略過，並在回傳的 error 中列出原因。
func BuildTables(rows []Row) (map[string]*Table, error) {
	byKind := make(map[string][]Row)
	for _, row := range rows {
		byKind[row.Kind] = append(byKind[row.Kind], row)
	}
	tables := make(map[string]*Table, len(byKind))
	var errs []error
	for kind, kindRows := range byKind {
		table, err := BuildTable(kind, kindRows)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		tables[kind] = table
	}
	return tables, errors.Join(errs...)
}

// MaxLevel 為等級上限。
func (t *Table) MaxLevel() int { return len(t.need) + 1 }

// ExpToNext 回傳指定等級升到下一級所需經驗；已達上限或等級不合法時回傳 0。
func (t *Table) ExpToNext(level int) int64 {
	if level < 1 || level > len(t.need) {
		return 0
	}
	return t.need[level-1]
}
