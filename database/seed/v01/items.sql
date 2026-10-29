-- item_type: 1 = Shard（碎片）；rarity 與 cards.rarity 同值（1=R, 2=SR, 3=SSR, 4=UR）
INSERT INTO items (item_id, item_name, item_type, rarity, config, is_enabled)
VALUES
    ('shard_r',   'R 碎片',   1, 1, '{}'::jsonb, TRUE),
    ('shard_sr',  'SR 碎片',  1, 2, '{}'::jsonb, TRUE),
    ('shard_ssr', 'SSR 碎片', 1, 3, '{}'::jsonb, TRUE),
    ('shard_ur',  'UR 碎片',  1, 4, '{}'::jsonb, TRUE)
ON CONFLICT (item_id) DO UPDATE SET
    item_name = EXCLUDED.item_name,
    item_type = EXCLUDED.item_type,
    rarity = EXCLUDED.rarity,
    config = EXCLUDED.config,
    is_enabled = EXCLUDED.is_enabled,
    updated_at = CURRENT_TIMESTAMP;
