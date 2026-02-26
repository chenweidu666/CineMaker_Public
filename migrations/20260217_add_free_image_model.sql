-- 先取消原来的默认模型
UPDATE ai_service_configs SET is_default = 0 WHERE service_type = 'image';

-- 插入新的免费图片模型（使用端点ID）
INSERT INTO ai_service_configs (
    service_type,
    provider,
    name,
    base_url,
    api_key,
    model,
    priority,
    is_default,
    is_active,
    created_at,
    updated_at
) VALUES (
    'image',
    'volcengine',
    '火山引擎 Seedream 4.0 (免费20张)',
    'https://ark.cn-beijing.volces.com',
    'YOUR_API_KEY_HERE',
    '["doubao-seedream-4-0-250828"]',
    10,
    1,
    1,
    datetime('now'),
    datetime('now')
);
