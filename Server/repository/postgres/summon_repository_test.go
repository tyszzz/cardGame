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
	"reflect"
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
		`INSERT INTO cards (card_id, card_name, rarity, race, is_limited) VALUES ('it_r1','IT R1',1,1,FALSE),('it_sr1','IT SR1',2,1,FALSE),('it_limited','IT Limited',3,1,TRUE) ON CONFLICT (card_id) DO NOTHING`,
		`INSERT INTO summon_pools (pool_id, pool_name, currency_type, costs, rarity_weights) VALUES ('it_pool','IT Pool',1,'{"1":5000}','{"R":100}') ON CONFLICT (pool_id) DO NOTHING`,
		`INSERT INTO summon_pools (pool_id, pool_name, currency_type, costs, rarity_weights, ten_pull_min_rarity, up_rates, pity_threshold, pity_min_rarity, up_guarantee, start_at, end_at)
		 VALUES ('it_pity','IT Pity',2,'{"1":200,"10":1800}','{"R":50,"SR":30,"SSR":20}',3,'{"SSR":50}',20,3,TRUE,'2026-01-01T00:00:00Z','2027-01-01T00:00:00Z') ON CONFLICT (pool_id) DO NOTHING`,
		`INSERT INTO summon_pool_cards (pool_id, card_id, is_up) VALUES ('it_pity','it_limited',TRUE) ON CONFLICT DO NOTHING`,
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
		db.Exec("DELETE FROM accounts WHERE uid = ?", uid)                         // 連帶刪除 players、貨幣、卡片、道具、保底狀態
		db.Exec("DELETE FROM summon_pools WHERE pool_id IN ('it_pool','it_pity')") // 連帶刪除 summon_pool_cards
		db.Exec("DELETE FROM cards WHERE card_id IN ('it_r1','it_sr1','it_limited')")
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
	cardR  = summon.PoolCard{Card: card.Card{CardId: "it_r1", CardName: "IT R1", Rarity: constants.RarityR}}
	cardSR = summon.PoolCard{Card: card.Card{CardId: "it_sr1", CardName: "IT SR1", Rarity: constants.RaritySR}}
)

func resolveOf(drawn ...summon.PoolCard) summon.ResolveFunc {
	return func(owned map[string]struct{}, _ summon.PityState) (summon.Outcome, error) {
		return summon.Resolve(drawn, owned)
	}
}

// apply 以不追蹤保底的測試卡池執行 ApplySummon。
func apply(repo *SummonRepo, uid string, cost int64, resolve summon.ResolveFunc) (*summon.Outcome, error) {
	return repo.ApplySummon(context.Background(), uid, "it_pool", constants.CurrencyCoin, cost, false, resolve)
}

func TestApplySummonNotEnoughCurrency(t *testing.T) {
	db, repo, uid := newSummonTestEnv(t)
	setCoin(t, db, uid, 100)

	_, err := apply(repo, uid, 5000, resolveOf(cardR))
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

	out, err := apply(repo, uid, 50000, resolveOf(cardR, cardSR, cardR))
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
	if _, err := apply(repo, uid, 5000, resolveOf(cardR)); err != nil {
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
	_, err := apply(repo, uid, 5000, func(map[string]struct{}, summon.PityState) (summon.Outcome, error) {
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
			_, err := apply(repo, uid, 5000, resolveOf(cardSR))
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

func TestListEnabledCardsMarksLimited(t *testing.T) {
	_, repo, _ := newSummonTestEnv(t)
	cards, err := repo.ListEnabledCards(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	limited := map[string]bool{}
	for _, c := range cards {
		limited[c.CardId] = c.IsLimited
	}
	if !limited["it_limited"] || limited["it_r1"] {
		t.Errorf("IsLimited wrong: it_limited=%v it_r1=%v", limited["it_limited"], limited["it_r1"])
	}
}

func TestListPools(t *testing.T) {
	_, repo, _ := newSummonTestEnv(t)
	defs, err := repo.ListPools(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	byId := map[string]summon.PoolDefinition{}
	for _, d := range defs {
		byId[d.PoolId] = d
	}

	if normal, ok := byId["normal"]; ok { // 常駐卡池來自 seed
		wantWeights := []summon.RarityWeight{{Rarity: constants.RarityR, Weight: 70}, {Rarity: constants.RaritySR, Weight: 30}}
		if normal.Currency != constants.CurrencyCoin || normal.Costs[1] != 5000 || normal.Costs[10] != 50000 ||
			!reflect.DeepEqual(normal.Weights, wantWeights) || normal.Pity != nil || !normal.IncludeStandardCards {
			t.Errorf("normal pool = %+v", normal)
		}
	} else {
		t.Log("seed pool 'normal' not found; run database/seed.sh to cover it")
	}

	up, ok := byId["it_pity"]
	if !ok {
		t.Fatal("it_pity not listed")
	}
	if up.Currency != constants.CurrencyDiamond || up.TenPullMinRarity != constants.RaritySSR ||
		up.UpRates[constants.RaritySSR] != 50 || up.Pity == nil ||
		*up.Pity != (summon.PityRule{Threshold: 20, MinRarity: constants.RaritySSR, UpGuarantee: true}) ||
		up.StartAt == nil || up.EndAt == nil || up.Costs[10] != 1800 || !up.IsEnabled {
		t.Errorf("it_pity = %+v", up)
	}
	if len(up.Weights) != 3 || up.Weights[0].Rarity != constants.RarityR || up.Weights[2].Rarity != constants.RaritySSR {
		t.Errorf("weights not sorted by rarity: %+v", up.Weights)
	}

	refs, err := repo.ListPoolCards(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if want := []summon.PoolCardRef{{CardId: "it_limited", IsUp: true}}; !reflect.DeepEqual(refs["it_pity"], want) {
		t.Errorf("pool cards = %+v, want %+v", refs["it_pity"], want)
	}
}

func pityRow(t *testing.T, db *gorm.DB, uid string) (pulls int64, guaranteed bool, found bool) {
	t.Helper()
	type row struct {
		PullsSinceHigh int64
		GuaranteedUp   bool
	}
	var rows []row
	if err := db.Raw("SELECT pulls_since_high, guaranteed_up FROM player_summon_pity WHERE uid = ? AND pool_id = 'it_pity'", uid).Scan(&rows).Error; err != nil {
		t.Fatal(err)
	}
	if len(rows) == 0 {
		return 0, false, false
	}
	return rows[0].PullsSinceHigh, rows[0].GuaranteedUp, true
}

func applyPity(repo *SummonRepo, uid string, cost int64, resolve summon.ResolveFunc) (*summon.Outcome, error) {
	return repo.ApplySummon(context.Background(), uid, "it_pity", constants.CurrencyCoin, cost, true, resolve)
}

func TestApplySummonPityPersistence(t *testing.T) {
	db, repo, uid := newSummonTestEnv(t)
	setCoin(t, db, uid, 50000)

	var seen []summon.PityState
	step := func(next summon.PityState) summon.ResolveFunc {
		return func(owned map[string]struct{}, pity summon.PityState) (summon.Outcome, error) {
			seen = append(seen, pity)
			out, err := summon.Resolve([]summon.PoolCard{cardR}, owned)
			out.Pity = next
			return out, err
		}
	}

	if _, err := applyPity(repo, uid, 5000, step(summon.PityState{PullsSinceHigh: 7, GuaranteedUp: true})); err != nil {
		t.Fatal(err)
	}
	if _, err := applyPity(repo, uid, 5000, step(summon.PityState{PullsSinceHigh: 8})); err != nil {
		t.Fatal(err)
	}
	wantSeen := []summon.PityState{{}, {PullsSinceHigh: 7, GuaranteedUp: true}}
	if !reflect.DeepEqual(seen, wantSeen) {
		t.Errorf("resolve saw %+v, want %+v", seen, wantSeen)
	}
	if pulls, up, found := pityRow(t, db, uid); !found || pulls != 8 || up {
		t.Errorf("pity row = %d %v %v, want 8 false true", pulls, up, found)
	}

	// 餘額不足：保底狀態不變
	setCoin(t, db, uid, 0)
	if _, err := applyPity(repo, uid, 5000, step(summon.PityState{PullsSinceHigh: 99})); !errors.Is(err, summon.ErrNotEnoughCurrency) {
		t.Fatalf("err = %v, want ErrNotEnoughCurrency", err)
	}
	// resolve 失敗：整筆回滾，餘額與保底都不變
	setCoin(t, db, uid, 5000)
	boom := errors.New("boom")
	if _, err := applyPity(repo, uid, 5000, func(map[string]struct{}, summon.PityState) (summon.Outcome, error) {
		return summon.Outcome{Pity: summon.PityState{PullsSinceHigh: 99}}, boom
	}); !errors.Is(err, boom) {
		t.Fatalf("err = %v, want boom", err)
	}
	if pulls, _, _ := pityRow(t, db, uid); pulls != 8 {
		t.Errorf("pity changed to %d after failed summons, want 8", pulls)
	}
	if got := scalar[int64](t, db, "SELECT amount FROM player_currencies WHERE uid = ? AND currency_type = ?", uid, constants.CurrencyCoin); got != 5000 {
		t.Errorf("amount = %d, want 5000", got)
	}

	// 不追蹤保底的卡池不寫入保底狀態
	if _, err := apply(repo, uid, 5000, resolveOf(cardR)); err != nil {
		t.Fatal(err)
	}
	if n := scalar[int64](t, db, "SELECT COUNT(*) FROM player_summon_pity WHERE uid = ? AND pool_id = 'it_pool'", uid); n != 0 {
		t.Errorf("pity rows for untracked pool = %d, want 0", n)
	}
}

func TestApplySummonPityConcurrent(t *testing.T) {
	db, repo, uid := newSummonTestEnv(t)
	setCoin(t, db, uid, 50000)

	const workers = 20
	var (
		wg      sync.WaitGroup
		mu      sync.Mutex
		success int
	)
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := applyPity(repo, uid, 5000, func(owned map[string]struct{}, pity summon.PityState) (summon.Outcome, error) {
				out, err := summon.Resolve([]summon.PoolCard{cardR}, owned)
				out.Pity = summon.PityState{PullsSinceHigh: pity.PullsSinceHigh + 1}
				return out, err
			})
			mu.Lock()
			defer mu.Unlock()
			switch {
			case err == nil:
				success++
			case errors.Is(err, summon.ErrNotEnoughCurrency):
			default:
				t.Errorf("unexpected error: %v", err)
			}
		}()
	}
	wg.Wait()

	// 玩家被鎖住序列化：每次成功都讀到前一次寫入的保底，最後累積數等於成功次數
	pulls, _, _ := pityRow(t, db, uid)
	if success != 10 || pulls != int64(success) {
		t.Errorf("success=%d pulls=%d, want 10/10", success, pulls)
	}
}

func TestPlayerRepoGetCurrencies(t *testing.T) {
	db, _, uid := newSummonTestEnv(t)
	setCoin(t, db, uid, 1234)
	got, err := InitPlayerRepo(db).GetCurrencies(context.Background(), uid)
	if err != nil {
		t.Fatal(err)
	}
	if got[constants.CurrencyCoin] != 1234 {
		t.Errorf("coin = %d, want 1234", got[constants.CurrencyCoin])
	}
	empty, err := InitPlayerRepo(db).GetCurrencies(context.Background(), uuid.NewString())
	if err != nil || len(empty) != 0 {
		t.Errorf("unknown uid: got %v, err %v", empty, err)
	}
}
