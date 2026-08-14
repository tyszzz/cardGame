# gameProject
self create game

## target
- use interface to create different card
- have shop, summon, car list, battle with ai
- level up system
- save data in db or redis

## repos
- feature
    - 各種功能
- global
    - 環境參數結構
- handler
    - 與客端溝通的註冊封包
- model
    - 全域通用的結構
- modules
    - 模組總管
- services
    - 各個功能的主要邏輯
    - redis db等連線請求也在這
- utils
    - 全專案共用的工具，不引用其他專案的package
    - 控制初始化

## Python virtual environment

前端測試工具使用 Python 虛擬環境，避免直接修改系統 Python。以下指令請在專案根目錄執行：

### 建立虛擬環境

```bash
cd /cardGame
python3 -m venv .venv
```

### 啟用虛擬環境

```bash
cd /cardGame
source .venv/bin/activate
```

啟用成功後，終端機提示字元會出現 `(.venv)`。

### 安裝前端測試套件

```bash
python -m pip install --upgrade pip
python -m pip install websocket-client
```

確認套件是否安裝成功：

```bash
python -c "import websocket; print(websocket.__version__)"
```

套件名稱是 `websocket-client`，但 Python 程式中的 import 名稱是 `websocket`。

### 安裝 tkinter 與系統字型

`tkinter` 和系統字型不是使用 `pip` 安裝，而是使用 Alpine Linux 的套件管理工具：

```bash
sudo apk add --no-cache python3-tkinter fontconfig ttf-dejavu ttf-liberation
fc-cache -f -v
```

確認 tkinter 與系統字型是否可用：

```bash
python -c "import tkinter; print(tkinter.TkVersion)"
fc-list | head
```

如果 `python3 -m venv .venv` 顯示 `ensurepip is not available`，可以改用 Alpine 的虛擬環境工具：

```bash
sudo apk add --no-cache py3-virtualenv
virtualenv .venv
source .venv/bin/activate
python -m pip install websocket-client
```

### 重新啟用既有虛擬環境

重新開啟終端機或 VS Code 後，不需要重新建立虛擬環境，只需要再次執行：

```bash
cd /cardGame
source .venv/bin/activate
```

### 離開虛擬環境

```bash
deactivate
```

### VS Code interpreter

在 VS Code 執行 `Python: Select Interpreter`，選擇：

```text
/cardGame/.venv/bin/python
```

## Python 前端打包

Python 前端可以使用 PyInstaller 打包成執行檔，使用者不需要另外安裝 Python、tkinter 等套件。

Windows 的 `.exe` 建議在 Windows 環境中打包。Linux 不能直接產生可在 Windows 執行的 `.exe`。

### Windows 安裝打包工具

在 Windows 開啟 PowerShell，進入 `clients` 目錄或專案目錄後執行：

```powershell
python -m venv .venv
.venv\Scripts\activate
python -m pip install --upgrade pip
python -m pip install websocket-client pyinstaller
```

### 建立測試版

建議先使用 `--onedir` 測試，輸出內容較容易檢查：

```powershell
cd clients
pyinstaller --clean --onedir --name CardGameClient main.py
```

輸出位置：

```text
clients\dist\CardGameClient\
```

測試版若發生錯誤，可以直接在 PowerShell 執行，查看錯誤訊息。

### 建立正式 GUI 版

確認測試版可以正常登入與操作後，使用 `--onefile --windowed` 打包成單一 GUI 執行檔：

```powershell
cd clients
pyinstaller --clean --onefile --windowed --name CardGameClient main.py
```

輸出檔案：

```text
clients\dist\CardGameClient.exe
```

`--onefile` 會輸出單一 exe；`--windowed` 執行時不會開啟命令列視窗，適合 tkinter GUI。

### 除錯版本

如果正式 GUI 版沒有顯示錯誤，可以先移除 `--windowed`：

```powershell
cd clients
pyinstaller --clean --onefile --name CardGameClient main.py
```

執行時保留 PowerShell 視窗，可以查看：

- WebSocket 連線錯誤
- Pitaya handshake 或封包解析錯誤
- tkinter 啟動錯誤
- 後端未啟動或無法連線

### 前端檔案

PyInstaller 會依照 `main.py` 的 import 自動尋找同目錄的模組：

```text
clients/
    main.py
    pitaya_client.py
    pitaya_protocol.py
```

如果三個檔案位於同一個 `clients` 目錄，通常不需要額外設定 spec 檔案。


不建議使用 `--break-system-packages`，以免破壞由系統套件管理的 Python 環境。