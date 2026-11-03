package level

import "context"

type LevelExpRepository interface {
	// ListLevelExp 回傳所有種類的經驗表資料。
	ListLevelExp(ctx context.Context) ([]Row, error)
}
