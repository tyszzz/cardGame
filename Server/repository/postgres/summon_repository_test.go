package postgres

import (
	"cardGame/feature/account"
	"cardGame/feature/card"
	"cardGame/feature/player"
	"cardGame/feature/summon"
	"cardGame/model/constants"
	"context"
	"errors"
	"os"
	"sync"
	"testing"

	"github.com/google/uuid"
	gormpg "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// 整合測試：需要已套用 migration 與 seed 的 PostgreSQL，
// 以 TEST_DATABASE_URL 指定連線，未設定時略過。
func newSummonTestEnv(t *testing.T) (*gorm.DB, *SummonRepo, string) {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL not set")
	}
	db, err := gorm.Open(gormpg.Open(dsn), &gorm.Config{TranslateError: true})
	if err != nil {
		t.Fatal(err)
	}

	for _, stmt := range []string{
		`INSERT INTO cards (card_id, card_name, rarity, race) VALUES ('it_r1','IT R1',1,1),('it_sr1','IT SR1',2,1) ON CONFLICT (card_id) DO NOTHING`,
		`INSERT INTO items (item_id, item_name, item_type, rarity) VALUES ('shard_r','R',1,1),('shard_sr','SR',1,2) ON CONFLICT (item_id) DO NOTHING`,
	} {
		if err := db.Exec(stmt).Error; err != nil {
			t.Fatal(err)
		}
	}

	uid := uuid.NewString()
	p := &player.Player{UID: uid, Nickname: "it" + uid[:8], Level: 1, Currencies: map[constants.CurrencyType]int64{constants.CurrencyCoin: 0}}
	acc := &account.Account{AccountId: "it" + uid[:8], PasswordHash: "x", UID: uid, Status: "active"}
	if err := InitAccountRepo(db).CreateNewAccount(context.Background(), acc, p); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		db.Exec("DELETE FROM accounts WHERE uid = ?", uid) // 連帶刪除 players、貨幣、卡片、道具
		db.Exec("DELETE FROM cards WHERE card_id IN ('it_r1','it_sr1')")
	})
	return db, InitSummonRepo(db), uid
}

func setCoin(t *testing.T, db *gorm.DB, uid string, amount int64) {
	t.Helper()
	if err := db.Exec("UPDATE player_currencies SET amount = ? WHERE uid = ? AND currency_type = ?", amount, uid, constants.CurrencyCoin).Error; err != nil {
		t.Fatal(err)
	}
}

func scalar[T any](t *testing.T, db *gorm.DB, query string, args ...any) T {
	t.Helper()
	var v T
	if err := db.Raw(query, args...).Scan(&v).Error; err != nil {
		t.Fatal(err)
	}
	return v
}

var (
	cardR  = card.Card{CardId: "it_r1", CardName: "IT R1", Rarity: constants.RarityR}
	cardSR = card.Card{CardId: "it_sr1", CardName: "IT SR1", Rarity: constants.RaritySR}
)

func resolveOf(drawn ...card.Card) summon.ResolveFunc {
	return func(owned map[string]struct{}) (summon.Outcome, error) { return summon.Resolve(drawn, owned) }
}

func TestApplySummonNotEnoughCurrency(t *testing.T) {
	db, repo, uid := newSummonTestEnv(t)
	setCoin(t, db, uid, 100)

	_, err := repo.ApplySummon(context.Background(), uid, constants.CurrencyCoin, 5000, resolveOf(cardR))
	if !errors.Is(err, summon.ErrNotEnoughCurrency) {
		t.Fatalf("err = %v, want ErrNotEnoughCurrency", err)
	}
	if got := scalar[int64](t, db, "SELECT amount FROM player_currencies WHERE uid = ? AND currency_type = ?", uid, constants.CurrencyCoin); got != 100 {
		t.Errorf("amount = %d, want 100", got)
	}
	if n := scalar[int64](t, db, "SELECT COUNT(*) FROM player_cards WHERE uid = ?", uid); n != 0 {
		t.Errorf("player_cards = %d, want 0", n)
	}
}

func TestApplySummonNewAndDuplicate(t *testing.T) {
	db, repo, uid := newSummonTestEnv(t)
	setCoin(t, db, uid, 60000)

	out, err := repo.ApplySummon(context.Background(), uid, constants.CurrencyCoin, 50000, resolveOf(cardR, cardSR, cardR))
	if err != nil {
		t.Fatal(err)
	}
	if len(out.NewCardIds) != 2 {
		t.Errorf("NewCardIds = %v", out.NewCardIds)
	}
	if got := scalar[int64](t, db, "SELECT amount FROM player_currencies WHERE uid = ? AND currency_type = ?", uid, constants.CurrencyCoin); got != 10000 {
		t.Errorf("amount = %d, want 10000", got)
	}
	if n := scalar[int64](t, db, "SELECT COUNT(*) FROM player_cards WHERE uid = ?", uid); n != 2 {
		t.Errorf("player_cards = %d, want 2", n)
	}
	if q := scalar[int64](t, db, "SELECT quantity FROM player_items WHERE uid = ? AND item_id = 'shard_r'", uid); q != 30 {
		t.Errorf("shard_r = %d, want 30", q)
	}

	// 再抽到已擁有的卡：不新增卡片，碎片累加
	if _, err := repo.ApplySummon(context.Background(), uid, constants.CurrencyCoin, 5000, resolveOf(cardR)); err != nil {
		t.Fatal(err)
	}
	if n := scalar[int64](t, db, "SELECT COUNT(*) FROM player_cards WHERE uid = ?", uid); n != 2 {
		t.Errorf("player_cards = %d, want still 2", n)
	}
	if q := scalar[int64](t, db, "SELECT quantity FROM player_items WHERE uid = ? AND item_id = 'shard_r'", uid); q != 60 {
		t.Errorf("shard_r = %d, want 60", q)
	}
}

func TestApplySummonRollsBackOnResolveError(t *testing.T) {
	db, repo, uid := newSummonTestEnv(t)
	setCoin(t, db, uid, 5000)

	boom := errors.New("boom")
	_, err := repo.ApplySummon(context.Background(), uid, constants.CurrencyCoin, 5000, func(map[string]struct{}) (summon.Outcome, error) {
		return summon.Outcome{}, boom
	})
	if !errors.Is(err, boom) {
		t.Fatalf("err = %v, want boom", err)
	}
	if got := scalar[int64](t, db, "SELECT amount FROM player_currencies WHERE uid = ? AND currency_type = ?", uid, constants.CurrencyCoin); got != 5000 {
		t.Errorf("amount = %d, want 5000 (rolled back)", got)
	}
}

func TestApplySummonConcurrent(t *testing.T) {
	db, repo, uid := newSummonTestEnv(t)
	setCoin(t, db, uid, 50000)

	const workers = 20
	var (
		wg      sync.WaitGroup
		mu      sync.Mutex
		success int
		poor    int
	)
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := repo.ApplySummon(context.Background(), uid, constants.CurrencyCoin, 5000, resolveOf(cardSR))
			mu.Lock()
			defer mu.Unlock()
			switch {
			case err == nil:
				success++
			case errors.Is(err, summon.ErrNotEnoughCurrency):
				poor++
			default:
				t.Errorf("unexpected error: %v", err)
			}
		}()
	}
	wg.Wait()

	if success != 10 || poor != 10 {
		t.Errorf("success=%d poor=%d, want 10/10", success, poor)
	}
	if got := scalar[int64](t, db, "SELECT amount FROM player_currencies WHERE uid = ? AND currency_type = ?", uid, constants.CurrencyCoin); got != 0 {
		t.Errorf("amount = %d, want 0", got)
	}
	if n := scalar[int64](t, db, "SELECT COUNT(*) FROM player_cards WHERE uid = ? AND card_id = 'it_sr1'", uid); n != 1 {
		t.Errorf("it_sr1 rows = %d, want exactly 1", n)
	}
	// 10 次成功：1 次新卡，9 次重複各 30 碎片
	if q := scalar[int64](t, db, "SELECT quantity FROM player_items WHERE uid = ? AND item_id = 'shard_sr'", uid); q != 270 {
		t.Errorf("shard_sr = %d, want 270", q)
	}
}

func TestListEnabledCards(t *testing.T) {
	db, repo, _ := newSummonTestEnv(t)
	cards, err := repo.ListEnabledCards(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	found := map[string]bool{}
	for _, c := range cards {
		found[c.CardId] = true
	}
	if !found["it_r1"] || !found["it_sr1"] {
		t.Errorf("test cards missing: %v", found)
	}
	db.Exec("UPDATE cards SET is_enabled = FALSE WHERE card_id = 'it_r1'")
	cards, _ = repo.ListEnabledCards(context.Background())
	for _, c := range cards {
		if c.CardId == "it_r1" {
			t.Error("disabled card must not be listed")
		}
	}
}
