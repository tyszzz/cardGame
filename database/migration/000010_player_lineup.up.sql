-- 玩家的戰鬥陣容：位置 1~3 為前排、4~6 為後排；同一張卡只能放在一個位置。
CREATE TABLE IF NOT EXISTS player_lineup (
    uid UUID NOT NULL REFERENCES players(uid) ON DELETE CASCADE,
    position SMALLINT NOT NULL CHECK (position BETWEEN 1 AND 6),
    card_id VARCHAR(64) NOT NULL REFERENCES cards(card_id),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (uid, position),
    UNIQUE (uid, card_id)
);
