-- 常駐卡池。rarity 數值：1=R, 2=SR, 3=SSR, 4=UR；currency_type：1=Coin, 2=Diamond, 3=FriendPoint。
INSERT INTO summon_pools (pool_id, pool_name, currency_type, costs, rarity_weights, ten_pull_min_rarity)
VALUES
    ('normal',  '普通召喚', 1, '{"1":5000,"10":50000}', '{"R":70,"SR":30}', NULL),
    ('friend',  '友情召喚', 3, '{"1":20,"10":200}',     '{"R":70,"SR":30}', NULL),
    ('diamond', '鑽石召喚', 2, '{"1":200,"10":1800}',   '{"R":50,"SR":30,"SSR":18,"UR":2}', 3)
ON CONFLICT (pool_id) DO UPDATE SET
    pool_name = EXCLUDED.pool_name,
    currency_type = EXCLUDED.currency_type,
    costs = EXCLUDED.costs,
    rarity_weights = EXCLUDED.rarity_weights,
    ten_pull_min_rarity = EXCLUDED.ten_pull_min_rarity,
    updated_at = CURRENT_TIMESTAMP;


-- UP 卡池：40 抽必出 UR，歪了（出非 UP 的 UR）下一張 UR 必為 UP。
-- 卡池 = 全部啟用的非限定卡 + summon_pool_cards 指定的卡（限定卡只有被列入才抽得到）。
INSERT INTO summon_pools (pool_id, pool_name, currency_type, costs, rarity_weights, ten_pull_min_rarity,
                          up_rates, pity_threshold, pity_min_rarity, up_guarantee, start_at, end_at)
VALUES
    ('up_mo',        '魔族UR UP召喚', 2, '{"1":200,"10":1800}', '{"R":50,"SR":30,"SSR":18,"UR":2}', 3, '{"UR":50}', 40, 4, TRUE, NULL, NULL),
    ('limited_xian', '仙族限定召喚',  2, '{"1":200,"10":1800}', '{"R":50,"SR":30,"SSR":18,"UR":2}', 3, '{"UR":50}', 40, 4, TRUE, NULL, NULL)
ON CONFLICT (pool_id) DO UPDATE SET
    pool_name = EXCLUDED.pool_name,
    currency_type = EXCLUDED.currency_type,
    costs = EXCLUDED.costs,
    rarity_weights = EXCLUDED.rarity_weights,
    ten_pull_min_rarity = EXCLUDED.ten_pull_min_rarity,
    up_rates = EXCLUDED.up_rates,
    pity_threshold = EXCLUDED.pity_threshold,
    pity_min_rarity = EXCLUDED.pity_min_rarity,
    up_guarantee = EXCLUDED.up_guarantee,
    start_at = EXCLUDED.start_at,
    end_at = EXCLUDED.end_at,
    updated_at = CURRENT_TIMESTAMP;

INSERT INTO summon_pool_cards (pool_id, card_id, is_up)
VALUES
    ('up_mo',        'mo_mage_ur',   TRUE),
    ('limited_xian', 'xian_ur_test', TRUE)
ON CONFLICT (pool_id, card_id) DO UPDATE SET is_up = EXCLUDED.is_up;
