-- 每位玩家對同一張卡最多持有一張，重複抽到的卡轉換為碎片。
CREATE UNIQUE INDEX IF NOT EXISTS uq_player_cards_uid_card_id ON player_cards (uid, card_id);
