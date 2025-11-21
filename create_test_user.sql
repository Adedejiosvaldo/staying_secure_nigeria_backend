-- Create test user for SafeTrace mobile app
-- This matches the user_id hardcoded in the mobile app

INSERT INTO users (
    id,
    phone_number,
    country_code,
    device_id,
    platform,
    app_version,
    hmac_secret,
    status,
    created_at,
    updated_at
) VALUES (
    '550e8400-e29b-41d4-a716-446655440000',
    '+2348012345678',
    '+234',
    'test-device-123',
    'android',
    '1.0.0',
    'your-hmac-secret-here',
    'active',
    NOW(),
    NOW()
)
ON CONFLICT (id) DO UPDATE SET
    updated_at = NOW(),
    status = 'active';

-- Verify user was created
SELECT id, phone_number, status, created_at FROM users
WHERE id = '550e8400-e29b-41d4-a716-446655440000';

