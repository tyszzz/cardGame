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
    - 主要的邏輯控制
- utils
    - 全專案共用的工具，不引用其他專案的package
    - 控制初始化