-- Create test user with correct schema
INSERT INTO users (
    id,
    phone,
    name,
    trusted_contacts,
    settings,
    created_at,
    updated_at
) VALUES (
    '550e8400-e29b-41d4-a716-446655440000',
    '+2348012345678',
    'Test User',
    '[]'::jsonb,
    '{}'::jsonb,
    NOW(),
    NOW()
)
ON CONFLICT (id) DO UPDATE SET
    updated_at = NOW();

-- Verify user was created
SELECT id, phone, name, created_at FROM users WHERE phone = '+2348012345678';
