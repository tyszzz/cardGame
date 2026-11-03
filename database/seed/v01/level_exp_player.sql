-- 玩家升級所需經驗：最高 200 級（level 1~199 各一列），目前暫以等級數字當作所需經驗，之後自行調整。
-- 等級上限為最大 level + 1；經驗需求必須嚴格遞增，否則該 kind 在 server 載入時會被略過。
INSERT INTO level_exp (kind, level, exp_required)
SELECT 'player', lv, lv FROM generate_series(1, 199) AS lv
ON CONFLICT (kind, level) DO UPDATE SET exp_required = EXCLUDED.exp_required;
