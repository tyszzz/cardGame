-- 卡片升階、升星。star 為升星數，stage 為進階數（從 0 開始）。
ALTER TABLE player_cards
    ADD COLUMN IF NOT EXISTS star SMALLINT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS stage SMALLINT NOT NULL DEFAULT 0;

ALTER TABLE player_cards
    ADD CONSTRAINT chk_player_cards_star CHECK (star >= 0),
    ADD CONSTRAINT chk_player_cards_stage CHECK (stage >= 0);

-- 既有卡片依等級回填階與星，避免資料違反規則（階：30/50/80/100 為各階上限；星：0 星 100 級、1~10 星每星 +10、11~15 星每星 +20）。
UPDATE player_cards SET
    stage = CASE WHEN level <= 30 THEN 0 WHEN level <= 50 THEN 1 WHEN level <= 80 THEN 2 WHEN level <= 100 THEN 3 ELSE 4 END,
    star = CASE WHEN level <= 100 THEN 0
                WHEN level <= 200 THEN CEIL((level - 100) / 10.0)
                ELSE 10 + CEIL((level - 200) / 20.0) END;

-- 星數對應的等級上限。
CREATE TABLE IF NOT EXISTS card_star_level (
    star SMALLINT PRIMARY KEY,
    max_level INT NOT NULL,
    CONSTRAINT chk_card_star_level_star CHECK (star >= 0),
    CONSTRAINT chk_card_star_level_max CHECK (max_level > 0)
);

-- 各稀有度的星數上限（rarity 數值：1=R, 2=SR, 3=SSR, 4=UR）。
CREATE TABLE IF NOT EXISTS card_rarity_limit (
    rarity SMALLINT PRIMARY KEY,
    max_star SMALLINT NOT NULL,
    CONSTRAINT chk_card_rarity_limit_star CHECK (max_star >= 1)
);

-- 進階設定。max_level 為該階的等級上限（NULL 代表不限，只受星數限制）；
-- ascend_* 為「從上一階進入本階」所需的進階道具與離開那一階的經驗比例，第 0 階不需要。
CREATE TABLE IF NOT EXISTS card_stage (
    stage SMALLINT PRIMARY KEY,
    max_level INT,
    ascend_item_id VARCHAR(64) REFERENCES items (item_id),
    ascend_item_qty BIGINT,
    ascend_exp_percent INT,
    CONSTRAINT chk_card_stage_stage CHECK (stage >= 0),
    CONSTRAINT chk_card_stage_max CHECK (max_level IS NULL OR max_level > 0),
    CONSTRAINT chk_card_stage_qty CHECK (ascend_item_qty IS NULL OR ascend_item_qty > 0),
    CONSTRAINT chk_card_stage_percent CHECK (ascend_exp_percent IS NULL OR (ascend_exp_percent BETWEEN 0 AND 100))
);

-- 升星材料：該稀有度升到第 star 星所需的各項道具。
CREATE TABLE IF NOT EXISTS card_star_cost (
    rarity SMALLINT NOT NULL,
    star SMALLINT NOT NULL,
    item_id VARCHAR(64) NOT NULL REFERENCES items (item_id),
    quantity BIGINT NOT NULL,
    PRIMARY KEY (rarity, star, item_id),
    CONSTRAINT chk_card_star_cost_star CHECK (star >= 1),
    CONSTRAINT chk_card_star_cost_qty CHECK (quantity > 0)
);
