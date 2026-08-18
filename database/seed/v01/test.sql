INSERT INTO cards (
    card_id, 
    card_name, 
    rarity, 
    race, 
    base_config, 
    skill_config, 
    is_enabled
)
VALUES (
    'card_001', 
    'Test Card', 
    1, 
    1, 
    '{}'::jsonb, 
    '{}'::jsonb, 
    TRUE
)
ON CONFLICT (card_id) DO NOTHING;