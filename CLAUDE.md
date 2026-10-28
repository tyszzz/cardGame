# CLAUDE.md

卡牌放置遊戲：Go 後端 + Python tkinter 測試前端。詳細背景與進度見 @docs/project-context.md，目錄說明見 @README.md。

## 合作規則

- 使用繁體中文回答，先給結論，再說明原因、取捨與實作細節。
- 預設只分析與提供建議，不主動修改程式碼。
- 需要修改時，先說明：要改哪些檔案、每個檔案的調整內容、修改原因、可能影響、驗證方式。**使用者明確回覆 `OK` 後才能修改。**
- 使用者只是詢問概念或架構時，不要建立檔案或修改程式。
- 以既有架構與程式碼為優先，不任意引入新框架或重構；有更好的方案時提出替代方案，並說明效能、維護成本、複雜度與遷移代價。
- 先指出問題根因再給修正方向；修正根因，不用只掩蓋問題的 workaround；修改保持最小範圍，不處理無關 bug。
- 不要覆蓋或撤銷使用者未要求的既有變更（工作區可能有未提交的修改）。

## 技術棧

- Backend：Go 1.23（toolchain 1.24.5），Pitaya v2 + WebSocket，PostgreSQL + GORM，未來可能加入 Redis
- Test client：Python tkinter + `websocket-client`（使用專案根目錄 `.venv`）
- 環境以 Alpine Linux 為準（套件用 `apk`，libc 為 musl）。glibc binary 在此環境會有相容性問題（例如舊的 `postgrator`）

## 專案結構

Go 程式位於 `Server/`，分層依賴：

```text
Pitaya -> handler -> services -> repository -> database / GORM / Redis
services -> feature
```

- `modules`：ModulesManager，建立共用依賴、注入服務、管理啟動與關閉生命週期
- `handler`：Pitaya 入口，只處理 request context、輸入轉換、service 呼叫與 response
- `services`：業務驗證、流程協調、transaction 邊界
- `repository`：資料庫/Redis 存取；實作 `feature` 定義的 interface，負責 dto 與 feature entity 轉換
- `database`：PostgreSQL/GORM（及未來 Redis）初始化、connection pool、關閉
- `dto`：儲存層結構。`dto/postgres` 一檔一表；`dto/redis` 依 key schema 或用途拆分，不強制一檔一表
- `feature`：純遊戲邏輯與 interface，**不得依賴 GORM、Redis 或 Pitaya session**
- `model`：`packet`（前後端封包）、`constants`、`cards` 等跨層共用型別
- `global`：環境參數結構；`utils`：logger、環境設定、Pitaya 初始化、helper
- 設定檔：`Server/.config/game.yaml`

其他目錄：

- `clients/`：`main.py`（tkinter UI）、`pitaya_client.py`、`pitaya_protocol.py`
- `database/`：`migration/`（版本化 SQL schema）、`seed/<version>/`（卡片定義）、`migration.sh`、`seed.sh`
- `tools/card-editor.html`：產生卡片 JSON 的網頁
- `docs/project-context.md`：專案背景與進度

## 常用指令

後端（在 `Server/` 下；VS Code 的 launch 設定也是以 `Server` 為 program）：

```bash
cd Server && go run .
cd Server && go build ./... && go test ./...
```

測試 client：

```bash
source .venv/bin/activate && python clients/main.py
```

資料庫（兩支腳本皆以 `DATABASE_URL` 環境變數覆寫連線，預設 `postgres://tyszzz:tyszzz@127.0.0.1:5432/tyszzz?sslmode=disable`）：

```bash
database/migration.sh        # 執行 database/migration 下的 up migration（使用 database/migrate）
database/seed.sh v01         # 套用 database/seed/<version>/*.sql，需指定版本，使用 psql
```

## 程式碼規範

- Go 修改後執行 `gofmt`，並跑 `go test ./...` 或至少編譯檢查。
- Python 修改後做語法、型別或相關測試檢查。
- PostgreSQL schema 只透過版本化 migration 管理，正式 server **不可** `AutoMigrate`。
- 重複帳戶判斷依賴 DB primary key/unique constraint，service 捕捉 duplicate key 後轉成業務錯誤。
- Pitaya response 以匿名嵌入 `BaseResult`（`resultCode`）共用結構。
- 前端驗證只改善 UX，後端必須重新驗證所有輸入。

## 安全

- 密碼只存 bcrypt/Argon2id hash，不存明文。
- 不要把密碼、password hash、token、資料庫密碼或其他 secrets 寫入 log、response 或 repository。
- 不要把資料庫原始錯誤直接回傳給前端。
- 不要提交 `.venv`、`__pycache__`、secrets。
- 若啟用 GORM SQL logging，透過 GORM `logger.Interface` adapter 呼叫 `utils.Log()`，使用 parameterized queries，避免敏感資料進 log。

## Logger

- Logrus caller 欄位格式：`package.Struct.Function`
- stdout 使用精簡 `ConsoleFormatter`，file 使用完整 JSON formatter
- `logger.log-file` / `logger.log-stdout` 控制輸出目的地
