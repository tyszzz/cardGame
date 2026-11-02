# Project Context

## 專案目標

這是一個使用 Go 後端與網頁測試前端製作的卡牌放置遊戲，玩法方向類似市面上的卡牌放置遊戲。正式前端尚未以 Cocos 或美術客戶端為目標，目前網頁 client 主要用來測試後端流程。

## 技術棧

- Backend：Go 1.23，toolchain Go 1.24.5
- Network：Pitaya v2，WebSocket acceptor
- Database：PostgreSQL + GORM
- Migration：SQL migration，現階段可使用 `psql` 執行；原本的 `postgrator` binary 在 Alpine 上有 glibc/musl 相容性問題
- Cache：未來可能加入 Redis
- Frontend test client：`web/` 純 HTML/JS/CSS，瀏覽器原生 WebSocket

## 後端分層

```text
Pitaya
  -> handler
  -> services
  -> repository
  -> database / GORM / Redis

services
  -> feature
```

責任分工：

- `modules`：ModulesManager，建立共用依賴、注入服務、管理啟動與關閉生命週期。
- `handler`：Pitaya 對外入口，解析 request、取得 session、呼叫 service、回傳 response。
- `services`：業務驗證、流程協調、transaction 邊界與資料轉換。
- `repository`：資料庫與 Redis 的查詢實作。
- `database`：PostgreSQL/GORM 與未來 Redis client 的初始化、connection pool 與關閉。
- `dto`：儲存層資料結構。`dto/postgres` 可一個檔案對應一張 PostgreSQL 表；`dto/redis` 依 key schema 或 cache 用途拆分。
- `feature`：純遊戲邏輯，不直接依賴 GORM、Redis 或 Pitaya session。
- `model/packet`：前後端封包資料結構。
- `utils`：共用工具、Pitaya 初始化、logger 與環境設定。

## 目前功能

### Pitaya JS client

`web/js/pitaya_client.js` 已支援：

- WebSocket 連線（Promise/async API）
- Pitaya handshake 與 handshake ACK
- binary request/response 與 request ID
- heartbeat 回應
- route listener：`registerRoute()` / `unregisterRoute()`
- client notify/publish：`publish(route, body)`
- `login()`、`createAccount()`、`summon()`、`getCurrencies()`

`web/js/pitaya_protocol.js` 已處理：

- packet header
- request、notify、response、push message
- Pitaya message flag
- zlib compressed response/push（`DecompressionStream`）
- JSON body

### 網頁前端

`web/js/app.js` 目前包含：

- 登入與建立帳戶畫面（含格式驗證與錯誤顯示）
- 貨幣顯示
- 多卡池單抽／十連與保底資訊
- 依稀有度上色的抽卡結果卡片
- 登出返回登入畫面

前端輸入規則：

- Account ID：英文字母與數字
- Password：最多 16 字元，只允許英數與目前 regex 定義的 ASCII 特殊字元

前端驗證只改善 UX，後端必須重新驗證。

## Account API

目前 handler route：

```text
AccountHandler.CreateNewAccount
AccountHandler.Login
```

response 使用共用結構：

```go
type BaseResult struct {
    ResultCode int32 `json:"resultCode"`
}
```

具體 response 可匿名嵌入 `BaseResult`，JSON 會展平成：

```json
{
  "resultCode": 0,
  "accountId": "test"
}
```

## PostgreSQL 與 GORM

- `database/postgres/postgres.go` 負責 GORM 初始化、Ping 與 connection pool。
- `dto/postgres/account.go` 是 Account model。
- `repository/postgres/account_repository.go` 負責建立與查詢帳戶。
- 目前建議透過 migration 管理 schema；開發初期可暫時 AutoMigrate，但正式環境不要由 server 啟動時自行改表。
- 新帳戶重複判斷應依賴 DB primary key/unique constraint，service 捕捉 duplicate key 後轉成業務錯誤。
- 密碼只能保存 bcrypt/Argon2id hash，不可保存明文。
- `CreatedAt` 只在建立時設定，`UpdatedAt` 建立與更新時由 GORM 更新。
- 查單一玩家卡片可直接 `WHERE player_id = ?`；N+1 是對 N 筆主資料逐筆執行關聯查詢。

目前 Account model 方向：

```go
type Account struct {
    AccountId    string     `gorm:"primaryKey;column:account_id"`
    PasswordHash string     `gorm:"column:password_hash"`
    UID          string     `gorm:"column:uid"`
    Status       string     `gorm:"size:20;not null;default:active"`
    CreatedAt    time.Time  `gorm:"autoCreateTime"`
    LastLoginAt  *time.Time `gorm:"column:last_login_at"`
    UpdatedAt    time.Time  `gorm:"autoUpdateTime"`
}
```

## Migration

目前 migration 位置：

```text
database/migration/
```

目前可使用：

```bash
sudo apk add --no-cache postgresql-client
PGPASSWORD=... psql -h 127.0.0.1 -p 5433 -U tyszzz -d tyszzz -f migration/000001_create_accounts.up.sql
```

原本的 `database/postgrator` 是 glibc binary，在 Alpine musl 執行時出現 relocation errors，因此不建議繼續使用。Docker migration 也受限於目前環境沒有 Docker daemon。

## 卡片、抽卡與怒氣規則

- 卡片定義以 JSON 維護於 `database/seed/<version>/cards/*.json`（可用 `tools/card-editor.html` 產生），再以 `go run ./cmd/cardseed <version>`（在 `Server/` 下）轉成 `cards.sql`，由 `database/seed.sh` 匯入。`cardseed` 會驗證欄位、稀有度與技能規則。
- 技能：每張卡都有普攻、主動、被動；天賦僅 SSR（3 個）與 UR（4 個）有。普攻與主動技能不填 `trigger`，被動與天賦才由 `trigger` 決定發動時機。
- 怒氣（所有卡共用，常數見 `model/constants/rage.go`）：初始 50，滿 100 才施放主動技能，施放後扣 100 並保留剩餘；怒氣上限與累積方式（普攻、受傷等）待戰鬥系統設計時決定。效果類型 `Rage` 可增減怒氣，`value` 為固定點數（可為負），其餘效果的 `value` 為百分比。
- Buff 可勾選 `permanent`（永久，此時不填 `duration`）與 `nonRemovable`（不會被驅散或淨化，預設可移除）；Status 也可設定 `nonRemovable`。`Dispel`（驅散增益）只能對敵方，移除正值 Buff 與護盾；`Cleanse`（淨化減益）只能對我方或自己，移除負值 Buff 與所有 Status；兩者以選填的 `count` 指定移除數量，未填代表全部。移除哪幾個由戰鬥系統決定。
- Buff 與 Status 效果都必填 `name`（顯示名稱，最多 20 個字），只用來顯示，不參與戰鬥邏輯比對；`status`（中毒、燃燒、暈眩）是類別，同類別可有不同名稱與數值（例如 A 毒、B 毒）。buff/status 的疊加以「來源技能與效果」區分，不同來源各自獨立。
- 護盾（效果類型 `Shield`）：只能對我方或自己；必填 `name`、`scaleStat`、`value`，額度為施放者 `scaleStat` 的 `value`%；`duration` 與 `permanent` 二擇一，可設 `nonRemovable`，不開放 `maxStack`。護盾算增益，會被 `Dispel` 移除（`nonRemovable` 除外），不受 `Cleanse` 影響。護盾吸收所有傷害，**真實傷害除外**；多個護盾的消耗順序與同來源重複施放的處理由戰鬥系統決定。
- 抽卡以卡池（`summon_pools`）為單位，請求帶 `poolId` 與 `times`（1 或 10），費用、貨幣、機率都取自資料庫的卡池定義，不信任前端。常駐卡池由 `database/seed/v01/pools.sql` 提供：`normal`（Coin 5000，R:SR = 70:30）、`friend`（FriendPoint 20，70:30）、`diamond`（Diamond 200，R/SR/SSR/UR = 50:30:18:2，十連 1800 且至少一張 SSR 以上）；十連不打折（Diamond 除外）。同稀有度內機率均等。
- 卡池內容：`include_standard_cards` 為真時包含所有啟用且非限定（`cards.is_limited = false`）的卡，再疊加 `summon_pool_cards` 指定的卡（可標記 `is_up`）。限定卡只有被明確列入某個卡池才抽得到。`start_at`、`end_at` 為空代表不限時間，開放時間以伺服器時間判斷。
- UP 與保底（只有設定的卡池才有，常駐池沒有）：`up_rates` 為各稀有度「出貨時為 UP 卡的百分比」，多張 UP 卡平分；`pity_threshold`、`pity_min_rarity` 為累積保底（自上次出保底稀有度以上後，第 N 抽必出，出了就歸零）；`up_guarantee` 為大保底（出了非 UP 的保底稀有度以上後，下一張必為 UP；只看有 UP 卡的稀有度，其他稀有度不影響狀態）。保底與十連保底並存，玩家狀態存於 `player_summon_pity`。建立 UP 卡池的 SQL 範例在 `pools.sql` 註解中。
- 抽卡流程：service 驗證卡池後，repository 在單一 transaction 內鎖定玩家 → 扣款 → 讀取已擁有的卡與保底狀態 → 抽卡與判定 → 寫入卡片、碎片與保底；任何一步失敗整筆回滾。抽卡在 transaction 內執行，同一玩家的所有抽卡（不論貨幣與卡池）都會排隊。
- 玩家對每張卡最多持有一張（`player_cards` 有 `UNIQUE (uid, card_id)`）；抽到重複的卡改給碎片放入背包：R 30、SR 30、SSR 10、UR 5。無卡片容量限制。
- 卡池與卡片資料於 server 載入記憶體，每 5 分鐘重新載入（失敗時沿用舊資料並每 30 秒重試）；定義不合法的卡池（例如某稀有度沒有卡）會被略過並記錄警告，抽該池回傳 code 9。部署新版 server 前須先套用 migration 與 seed。

## Logger

目前 logger 特性：

- Logrus caller 欄位：`package.Struct.Function`
- stdout 使用精簡 `ConsoleFormatter`
- file 使用完整 JSON formatter
- `logger.log-file` 控制是否寫檔
- `logger.log-stdout` 控制是否輸出到終端機
- SQL/GORM logger 若接入，應透過 GORM `logger.Interface` adapter 呼叫現有 `utils.Log()`
- 不要把 password、password hash 或 DB secret 寫入 log

## 跨電腦恢復

1. 在新電腦使用同一個 Git 帳號 clone repository。
2. 開啟 `.github/copilot-instructions.md`，讓 AI 讀取合作規則。
3. 開啟本文件 `docs/project-context.md`，讓 AI 恢復專案背景與目前進度。
4. 若有使用 codebase-memory MCP，重新 index repository 或載入可攜式 index。
5. 依 README 重建 Go 環境，不要提交密碼或 secrets。
6. 在新的 Copilot Chat 中說：

```text
請先讀取 .github/copilot-instructions.md 與 docs/project-context.md，依照其中的合作規則與專案架構繼續工作。先不要修改程式碼，先回報你理解的目前狀態。
```

## 待辦與注意事項

- 完成 AccountService 的實際 bcrypt 驗證與建立帳戶 transaction。
- 確認 Account UID 使用 UUID 還是一般字串，並與 migration/DTO 統一。
- 補齊 PostgreSQL migration 的 down migration 與版本管理工具。
- 將登入成功後的 session UID 與真正 player UID 對齊，不要依賴前端傳入 UID。
- 若啟用 GORM SQL logging，使用 parameterized queries 並避免敏感資料進 log。
- 未來加入 Redis 時，讓 database/redis 初始化與 repository 位於 feature 外部。
