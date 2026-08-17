# Project Context

## 專案目標

這是一個使用 Go 後端與 Python tkinter 測試前端製作的卡牌放置遊戲，玩法方向類似市面上的卡牌放置遊戲。正式前端尚未以 Cocos 或美術客戶端為目標，目前 tkinter 主要用來測試後端流程。

## 技術棧

- Backend：Go 1.23，toolchain Go 1.24.5
- Network：Pitaya v2，WebSocket acceptor
- Database：PostgreSQL + GORM
- Migration：SQL migration，現階段可使用 `psql` 執行；原本的 `postgrator` binary 在 Alpine 上有 glibc/musl 相容性問題
- Cache：未來可能加入 Redis
- Frontend test client：Python tkinter + `websocket-client`
- Python packaging：Windows 使用 PyInstaller 打包 `.exe`

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

### Pitaya Python client

`clients/pitaya_client.py` 已支援：

- WebSocket 連線
- Pitaya handshake 與 handshake ACK
- binary request/response
- request ID 與 receiver thread
- heartbeat 回應
- route listener：`register_route()` / `unregister_route()`
- client notify/publish：`publish(route, body)`
- `login()`
- `create_account()`，route 為 `AccountHandler.CreateNewAccount`
- `summon()`

`clients/pitaya_protocol.py` 已處理：

- packet header
- request、notify、response、push message
- Pitaya message flag
- zlib compressed response/push
- JSON body

### tkinter 前端

`clients/main.py` 目前包含：

- 登入畫面
- 帳戶 ID 與密碼格式驗證
- 建立帳戶畫面
- 建立帳戶成功返回登入畫面
- 建立帳戶失敗留在註冊畫面並顯示錯誤
- 登入後遊戲畫面
- 抽卡按鈕
- 登出返回登入畫面
- 結果文字區

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
5. 依 README 重建 Go/Python 環境，不要提交 `.venv`、`__pycache__`、密碼或 secrets。
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
