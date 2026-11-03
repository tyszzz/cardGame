package services

import (
	"cardGame/feature/level"
	"context"
	"errors"
	"testing"
	"time"
)

type fakeLevelRepo struct {
	rows  []level.Row
	err   error
	calls int
}

func (f *fakeLevelRepo) ListLevelExp(context.Context) ([]level.Row, error) {
	f.calls++
	return f.rows, f.err
}

func newTestLevelService(repo *fakeLevelRepo, now *time.Time) (*LevelService, *[]string) {
	var warnings []string
	s := NewLevelService(repo)
	s.now = func() time.Time { return *now }
	s.warn = func(format string, _ ...any) { warnings = append(warnings, format) }
	return s, &warnings
}

func cardRows(needs ...int64) []level.Row {
	rows := make([]level.Row, 0, len(needs))
	for i, n := range needs {
		rows = append(rows, level.Row{Kind: level.KindCard, Level: i + 1, ExpRequired: n})
	}
	return rows
}

func TestLevelServiceCardTable(t *testing.T) {
	now := time.Unix(0, 0)
	repo := &fakeLevelRepo{rows: cardRows(100, 120)}
	s, _ := newTestLevelService(repo, &now)

	table, err := s.CardTable(context.Background())
	if err != nil || table.MaxLevel() != 3 {
		t.Fatalf("table = %+v, err = %v", table, err)
	}
	if _, err := s.CardTable(context.Background()); err != nil || repo.calls != 1 {
		t.Errorf("second call must use the cache, calls = %d, err = %v", repo.calls, err)
	}
}

func TestLevelServiceReloadsAfterTTL(t *testing.T) {
	now := time.Unix(0, 0)
	repo := &fakeLevelRepo{rows: cardRows(100, 120)}
	s, _ := newTestLevelService(repo, &now)
	s.CardTable(context.Background())

	repo.rows = cardRows(100, 120, 144)
	now = now.Add(levelCacheTTL)
	table, err := s.CardTable(context.Background())
	if err != nil || table.MaxLevel() != 4 {
		t.Fatalf("expected reloaded table, got %+v, err %v", table, err)
	}
}

func TestLevelServiceKeepsStaleOnReloadFailure(t *testing.T) {
	now := time.Unix(0, 0)
	repo := &fakeLevelRepo{rows: cardRows(100, 120)}
	s, warnings := newTestLevelService(repo, &now)
	s.CardTable(context.Background())

	repo.err = errors.New("db down")
	now = now.Add(levelCacheTTL)
	table, err := s.CardTable(context.Background())
	if err != nil || table.MaxLevel() != 3 {
		t.Fatalf("expected stale table, got %+v, err %v", table, err)
	}
	if len(*warnings) == 0 {
		t.Error("a warning must be logged when falling back to stale data")
	}
	calls := repo.calls
	now = now.Add(levelReloadRetryWait - time.Second)
	s.CardTable(context.Background())
	if repo.calls != calls {
		t.Error("must not retry before the retry wait has passed")
	}
	now = now.Add(2 * time.Second)
	s.CardTable(context.Background())
	if repo.calls != calls+1 {
		t.Error("must retry after the retry wait")
	}
}

func TestLevelServiceFirstLoadFailure(t *testing.T) {
	now := time.Unix(0, 0)
	boom := errors.New("db down")
	s, _ := newTestLevelService(&fakeLevelRepo{err: boom}, &now)
	if _, err := s.CardTable(context.Background()); !errors.Is(err, boom) {
		t.Errorf("err = %v, want %v", err, boom)
	}
}

func TestLevelServiceInvalidOrMissingKind(t *testing.T) {
	now := time.Unix(0, 0)
	s, warnings := newTestLevelService(&fakeLevelRepo{rows: cardRows(100, 100)}, &now)
	if _, err := s.CardTable(context.Background()); !errors.Is(err, level.ErrTableUnavailable) {
		t.Errorf("invalid table err = %v, want ErrTableUnavailable", err)
	}
	if len(*warnings) == 0 {
		t.Error("an invalid table must be reported")
	}

	s2, _ := newTestLevelService(&fakeLevelRepo{}, &now)
	if _, err := s2.CardTable(context.Background()); !errors.Is(err, level.ErrTableUnavailable) {
		t.Errorf("missing table err = %v, want ErrTableUnavailable", err)
	}
}
