# CLAUDE.md

卡牌放置遊戲：Go 後端 + 網頁測試前端。詳細背景與進度見 @docs/project-context.md，目錄說明見 @README.md。

## 技術棧

- Backend：Go 1.25（以 `Server/go.mod` 為準），Pitaya v2 + WebSocket，PostgreSQL + GORM，未來可能加入 Redis
- Test client：純 HTML/JS/CSS 網頁（`web/`），瀏覽器原生 WebSocket，無建置步驟

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

- `web/`：`index.html`、`app.css`、`js/app.js`（UI）、`js/pitaya_client.js`、`js/pitaya_protocol.js`
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
python3 -m http.server -d web 5173   # 開啟 http://localhost:5173
```

資料庫（兩支腳本皆以 `DATABASE_URL` 環境變數指定連線）：

```bash
database/migration.sh        # 執行 database/migration 下的 up migration（使用 database/migrate）
database/seed.sh v01         # 套用 database/seed/<version>/*.sql，需指定版本，使用 psql
```

## 程式碼規範

- Go 修改後執行 `gofmt`，並跑 `go test ./...` 或至少編譯檢查。
- JS 修改後至少用 `node --check` 檢查語法；協議變動需與後端實測。
- PostgreSQL schema 只透過版本化 migration 管理，正式 server **不可** `AutoMigrate`。
- 重複帳戶判斷依賴 DB primary key/unique constraint，service 捕捉 duplicate key 後轉成業務錯誤。
- Pitaya response 以匿名嵌入 `BaseResult`（`resultCode`）共用結構。
- 前端驗證只改善 UX，後端必須重新驗證所有輸入。

## 安全

- 密碼只存 bcrypt/Argon2id hash，不存明文。
- 不要把密碼、password hash、token、資料庫密碼或其他 secrets 寫入 log、response 或 repository。
- 不要把資料庫原始錯誤直接回傳給前端。
- 不要提交 secrets。
- 若啟用 GORM SQL logging，透過 GORM `logger.Interface` adapter 呼叫 `utils.Log()`，使用 parameterized queries，避免敏感資料進 log。

## Logger

- Logrus caller 欄位格式：`package.Struct.Function`
- stdout 使用精簡 `ConsoleFormatter`，file 使用完整 JSON formatter
- `logger.log-file` / `logger.log-stdout` 控制輸出目的地
