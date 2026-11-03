-- 卡片升階與升星設定；rarity 數值：1=R, 2=SR, 3=SSR, 4=UR。要調整數值直接改這個檔案（或資料庫）。
-- 重新執行 seed 會以本檔覆蓋資料庫中相同主鍵的列。

-- 星數對應的等級上限：0 星 100 級，1~10 星每星 +10，11~15 星每星 +20。
INSERT INTO card_star_level (star, max_level)
VALUES
    (0, 100), (1, 110), (2, 120), (3, 130), (4, 140), (5, 150),
    (6, 160), (7, 170), (8, 180), (9, 190), (10, 200),
    (11, 220), (12, 240), (13, 260), (14, 280), (15, 300)
ON CONFLICT (star) DO UPDATE SET max_level = EXCLUDED.max_level;

-- 各稀有度的星數上限（R 5 星 150 級、SR 10 星 200 級、SSR 與 UR 15 星 300 級）。
INSERT INTO card_rarity_limit (rarity, max_star)
VALUES (1, 5), (2, 10), (3, 15), (4, 15)
ON CONFLICT (rarity) DO UPDATE SET max_star = EXCLUDED.max_star;

-- 進階：0 階 1~30 級、1 階 30~50 級、2 階 50~80 級、3 階 80~100 級、4 階不限（只受星數限制）。
-- 進階需要等級到達當前階的上限；消耗進階丹與「離開那一階」經驗總和的 ascend_exp_percent%。
INSERT INTO card_stage (stage, max_level, ascend_item_id, ascend_item_qty, ascend_exp_percent)
VALUES
    (0, 30,   NULL,          NULL, NULL),
    (1, 50,   'ascend_pill', 500,  50),
    (2, 80,   'ascend_pill', 1000, 50),
    (3, 100,  'ascend_pill', 2000, 50),
    (4, NULL, 'ascend_pill', 4000, 50)
ON CONFLICT (stage) DO UPDATE SET
    max_level = EXCLUDED.max_level,
    ascend_item_id = EXCLUDED.ascend_item_id,
    ascend_item_qty = EXCLUDED.ascend_item_qty,
    ascend_exp_percent = EXCLUDED.ascend_exp_percent;

-- 升星材料：升到第 s 星時，自己稀有度的碎片 50×s，每低一階的碎片再加 50（50×(s+j)，j 為低幾階）。
-- 例：SSR 升 1 星 = SSR 碎片 50、SR 碎片 100、R 碎片 150；升 2 星 = 100、150、200。
INSERT INTO card_star_cost (rarity, star, item_id, quantity)
SELECT r.rarity, s, 'shard_' || (ARRAY['r', 'sr', 'ssr', 'ur'])[r.rarity - j], 50 * (s + j)
FROM card_rarity_limit r,
     generate_series(1, r.max_star) AS s,
     generate_series(0, r.rarity - 1) AS j
ON CONFLICT (rarity, star, item_id) DO UPDATE SET quantity = EXCLUDED.quantity;
