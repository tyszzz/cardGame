# AI 合作規則

## 回應方式

- 使用繁體中文回答。
- 先給結論，再說明原因、取捨與實作細節。
- 以目前專案的既有架構與程式碼為優先，不任意引入新框架或重構。
- 如果有更好的套件、工具或設計，提出替代方案並說明效能、維護成本、複雜度與遷移代價。
- 發現問題時先指出根因，再提供修正方向。
- 回答保持精準，避免與問題無關的延伸。

## 修改程式碼前

- 預設只分析與提供建議，不主動修改程式碼。
- 如果需要修改，先說明：
  - 要修改哪些檔案
  - 每個檔案要調整的內容
  - 修改原因
  - 可能影響
  - 驗證方式
- 只有使用者明確回覆 `OK` 後才可以修改程式碼。
- 使用者只是在詢問概念或架構時，不要直接建立檔案或修改程式。
- 不要覆蓋或撤銷使用者未要求的既有變更。

## 程式碼與安全

- 修正根因，不使用只掩蓋問題的 workaround。
- 修改保持最小範圍，不處理無關 bug。
- Go 修改後使用 `gofmt`，並執行適當的 `go test ./...` 或編譯檢查。
- Python 修改後執行語法、型別或相關測試檢查。
- 不要把密碼、password hash、token、資料庫密碼或其他 secrets 寫入 log、response 或 repository。
- 前端驗證只改善使用者體驗，後端仍必須重新驗證所有輸入。
- 不要把資料庫原始錯誤直接回傳給前端。

## 專案約定

- `ModulesManager` 負責共用模組初始化、依賴組裝與生命週期。
- Pitaya handler 是網路與封包入口，只處理 request context、輸入轉換、service 呼叫與 response。
- `services` 負責業務驗證、流程協調與 transaction 邊界。
- `repository` 負責資料庫或 Redis 存取。
- `dto` 負責資料儲存結構；PostgreSQL 可採一個檔案對應一張表。
- `feature` 只負責純遊戲規則與核心邏輯，不直接依賴 GORM、Redis 或 Pitaya session。
- PostgreSQL 使用 GORM；正式 schema 使用版本化 migration，不由正式 server 自動 `AutoMigrate`。
- Redis DTO 依 key schema、cache 或 session 用途拆分，不強制一檔一表。
