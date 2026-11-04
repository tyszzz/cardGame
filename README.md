# gameProject
self create game

## target
- use interface to create different card
- have shop, summon, card list, battle with ai
- level up system
- save data in db or redis

## repos
### server
- database
    - 資料套件模組
- dto
    - 資料庫表格結構
- feature
    - 各個模組的最小功能
    - 提供介面
    - 核心處理邏輯
- global
    - 環境參數結構
- handler
    - Pitaya 對外入口
    - request/response流程
- model
    - 跨層共用的資料模型與協定型別
- modules
    - 建立 database、repository、service、handler
    - 注入依賴
- repository
    - 實作feature的interface
    - dto和feature entity的轉換
    - 操作資料庫
- services
    - 功能流程協調/驗證
- utils
    - 全專案共用的基礎工具
    - logger，環境變數，Pitaya 初始化與共用 helper

### web
- 網頁測試 client（純 HTML/JS/CSS，無建置步驟）
    - `js/pitaya_protocol.js`：Pitaya 封包編解碼
    - `js/pitaya_client.js`：WebSocket 連線、request/response、push
    - `js/app.js`：登入、註冊、召喚畫面

### tools
- card-editor
    - 輸出卡片json

### database
- migration
    - PostgreSQL schema migration
- seed
    - 各版本的卡片定義

## 網頁 client

網頁使用瀏覽器原生 WebSocket 連到 Pitaya，不需要安裝任何套件。ES module 無法從 `file://` 載入，請用靜態伺服器開啟：

```bash
sh web/start.sh
```

開啟 `http://localhost:5173`。後端位址在 `web/js/app.js` 的 `SERVER_URL`，預設為 `ws://<目前主機>:8080/`。
