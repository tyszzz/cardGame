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

啟用成功後，終端機提示字元通常會出現 `(.venv)`。

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
/home/tyszzz/project/cardGame/.venv/bin/python
```

### Alpine Linux fallback

如果 `python3 -m venv .venv` 顯示 `ensurepip is not available`，可以改用 Alpine 的虛擬環境工具：

```bash
apk add --no-cache py3-virtualenv
virtualenv .venv
source .venv/bin/activate
python -m pip install websocket-client
```

不建議使用 `--break-system-packages`，以免破壞由系統套件管理的 Python 環境。