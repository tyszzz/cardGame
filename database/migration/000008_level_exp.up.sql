-- 升級所需經驗表。level 為目前等級，exp_required 為升到下一級所需經驗；
-- 等級上限由該 kind 的最大 level + 1 決定。kind：card（卡片）、player（玩家）。
CREATE TABLE IF NOT EXISTS level_exp (
    kind VARCHAR(16) NOT NULL,
    level INT NOT NULL,
    exp_required BIGINT NOT NULL,
    PRIMARY KEY (kind, level),
    CONSTRAINT chk_level_exp_level CHECK (level >= 1),
    CONSTRAINT chk_level_exp_required CHECK (exp_required > 0)
);
