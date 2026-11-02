-- 限定卡預設不進入常駐卡池，只有被明確列入某個卡池才抽得到。
ALTER TABLE cards ADD COLUMN IF NOT EXISTS is_limited BOOLEAN NOT NULL DEFAULT FALSE;

-- 卡池定義。costs 為 {"1": 單抽費用, "10": 十連費用}，rarity_weights 為 {"R": 權重, ...}，
-- up_rates 為 {"SSR": 50} 表示出該稀有度時有 50% 為 UP 卡。start_at、end_at 為空代表不限時間。
CREATE TABLE IF NOT EXISTS summon_pools (
    pool_id VARCHAR(64) PRIMARY KEY,
    pool_name VARCHAR(100) NOT NULL,
    currency_type SMALLINT NOT NULL,
    costs JSONB NOT NULL,
    rarity_weights JSONB NOT NULL,
    ten_pull_min_rarity SMALLINT CHECK (ten_pull_min_rarity BETWEEN 1 AND 4),
    up_rates JSONB NOT NULL DEFAULT '{}'::jsonb,
    pity_threshold INT CHECK (pity_threshold > 0),
    pity_min_rarity SMALLINT CHECK (pity_min_rarity BETWEEN 1 AND 4),
    up_guarantee BOOLEAN NOT NULL DEFAULT FALSE,
    include_standard_cards BOOLEAN NOT NULL DEFAULT TRUE,
    start_at TIMESTAMPTZ,
    end_at TIMESTAMPTZ,
    is_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CHECK (start_at IS NULL OR end_at IS NULL OR end_at > start_at),
    CHECK ((pity_threshold IS NULL) = (pity_min_rarity IS NULL))
);

-- 卡池額外指定的卡片；include_standard_cards 為真時，會與所有啟用且非限定的卡合併。
CREATE TABLE IF NOT EXISTS summon_pool_cards (
    pool_id VARCHAR(64) NOT NULL REFERENCES summon_pools(pool_id) ON DELETE CASCADE,
    card_id VARCHAR(64) NOT NULL REFERENCES cards(card_id),
    is_up BOOLEAN NOT NULL DEFAULT FALSE,
    PRIMARY KEY (pool_id, card_id)
);

-- 玩家在每個有保底的卡池的累積狀態。
CREATE TABLE IF NOT EXISTS player_summon_pity (
    uid UUID NOT NULL REFERENCES players(uid) ON DELETE CASCADE,
    pool_id VARCHAR(64) NOT NULL REFERENCES summon_pools(pool_id),
    pulls_since_high INT NOT NULL DEFAULT 0 CHECK (pulls_since_high >= 0),
    guaranteed_up BOOLEAN NOT NULL DEFAULT FALSE,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (uid, pool_id)
);
