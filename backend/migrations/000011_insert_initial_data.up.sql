-- 创建默认管理员账户 (密码: admin123)
INSERT INTO users (id, username, password_hash, user_type, license_plate, battery_capacity)
VALUES (
    '00000000-0000-0000-0000-000000000000',
    'admin',
    '$2a$10$nEFHmT4WLwKa8DB5sEQgv.3fU/4gr.931HiWw6JlnhiaSYBIuH/My',
    'admin',
    'ADMIN-001',
    100.0
) ON CONFLICT (id) DO NOTHING;

-- 创建初始用户账户 (京AUV001-023)
INSERT INTO users (id, username, password_hash, user_type, license_plate, battery_capacity)
VALUES 
    ('11111111-1111-1111-1111-111111111111', 'V1', '$2a$10$nEFHmT4WLwKa8DB5sEQgv.3fU/4gr.931HiWw6JlnhiaSYBIuH/My', 'user', 'V1', 60.0),
    ('22222222-2222-2222-2222-222222222222', 'V2', '$2a$10$nEFHmT4WLwKa8DB5sEQgv.3fU/4gr.931HiWw6JlnhiaSYBIuH/My', 'user', 'V2', 65.0),
    ('33333333-3333-3333-3333-333333333333', 'V3', '$2a$10$nEFHmT4WLwKa8DB5sEQgv.3fU/4gr.931HiWw6JlnhiaSYBIuH/My', 'user', 'V3', 70.0),
    ('44444444-4444-4444-4444-444444444444', 'V4', '$2a$10$nEFHmT4WLwKa8DB5sEQgv.3fU/4gr.931HiWw6JlnhiaSYBIuH/My', 'user', 'V4', 120.0),
    ('55555555-5555-5555-5555-555555555555', 'V5', '$2a$10$nEFHmT4WLwKa8DB5sEQgv.3fU/4gr.931HiWw6JlnhiaSYBIuH/My', 'user', 'V5', 80.0),
    ('66666666-6666-6666-6666-666666666666', 'V6', '$2a$10$nEFHmT4WLwKa8DB5sEQgv.3fU/4gr.931HiWw6JlnhiaSYBIuH/My', 'user', 'V6', 85.0),
    ('77777777-7777-7777-7777-777777777777', 'V7', '$2a$10$nEFHmT4WLwKa8DB5sEQgv.3fU/4gr.931HiWw6JlnhiaSYBIuH/My', 'user', 'V7', 90.0),
    ('88888888-8888-8888-8888-888888888888', 'V8', '$2a$10$nEFHmT4WLwKa8DB5sEQgv.3fU/4gr.931HiWw6JlnhiaSYBIuH/My', 'user', 'V8', 95.0),
    ('99999999-9999-9999-9999-999999999999', 'V9', '$2a$10$nEFHmT4WLwKa8DB5sEQgv.3fU/4gr.931HiWw6JlnhiaSYBIuH/My', 'user', 'V9', 100.0),
    ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'V10', '$2a$10$nEFHmT4WLwKa8DB5sEQgv.3fU/4gr.931HiWw6JlnhiaSYBIuH/My', 'user', 'V10', 110.0),
    ('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', 'V11', '$2a$10$nEFHmT4WLwKa8DB5sEQgv.3fU/4gr.931HiWw6JlnhiaSYBIuH/My', 'user', 'V11', 115.0),
    ('cccccccc-cccc-cccc-cccc-cccccccccccc', 'V12', '$2a$10$nEFHmT4WLwKa8DB5sEQgv.3fU/4gr.931HiWw6JlnhiaSYBIuH/My', 'user', 'V12', 75.0),
    ('dddddddd-dddd-dddd-dddd-dddddddddddd', 'V13', '$2a$10$nEFHmT4WLwKa8DB5sEQgv.3fU/4gr.931HiWw6JlnhiaSYBIuH/My', 'user', 'V13', 85.0),
    ('eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee', 'V14', '$2a$10$nEFHmT4WLwKa8DB5sEQgv.3fU/4gr.931HiWw6JlnhiaSYBIuH/My', 'user', 'V14', 90.0),
    ('ffffffff-ffff-ffff-ffff-ffffffffffff', 'V15', '$2a$10$nEFHmT4WLwKa8DB5sEQgv.3fU/4gr.931HiWw6JlnhiaSYBIuH/My', 'user', 'V15', 95.0),
    ('10101010-1010-1010-1010-101010101010', 'V16', '$2a$10$nEFHmT4WLwKa8DB5sEQgv.3fU/4gr.931HiWw6JlnhiaSYBIuH/My', 'user', 'V16', 100.0),
    ('20202020-2020-2020-2020-202020202020', 'V17', '$2a$10$nEFHmT4WLwKa8DB5sEQgv.3fU/4gr.931HiWw6JlnhiaSYBIuH/My', 'user', 'V17', 105.0),
    ('30303030-3030-3030-3030-303030303030', 'V18', '$2a$10$nEFHmT4WLwKa8DB5sEQgv.3fU/4gr.931HiWw6JlnhiaSYBIuH/My', 'user', 'V18', 110.0),
    ('40404040-4040-4040-4040-404040404040', 'V19', '$2a$10$nEFHmT4WLwKa8DB5sEQgv.3fU/4gr.931HiWw6JlnhiaSYBIuH/My', 'user', 'V19', 80.0),
    ('50505050-5050-5050-5050-505050505050', 'V20', '$2a$10$nEFHmT4WLwKa8DB5sEQgv.3fU/4gr.931HiWw6JlnhiaSYBIuH/My', 'user', 'V20', 85.0),
    ('60606060-6060-6060-6060-606060606060', 'V21', '$2a$10$nEFHmT4WLwKa8DB5sEQgv.3fU/4gr.931HiWw6JlnhiaSYBIuH/My', 'user', 'V21', 90.0),
    ('70707070-7070-7070-7070-707070707070', 'V22', '$2a$10$nEFHmT4WLwKa8DB5sEQgv.3fU/4gr.931HiWw6JlnhiaSYBIuH/My', 'user', 'V22', 125.0),
    ('80808080-8080-8080-8080-808080808080', 'V23', '$2a$10$nEFHmT4WLwKa8DB5sEQgv.3fU/4gr.931HiWw6JlnhiaSYBIuH/My', 'user', 'V23', 130.0)
ON CONFLICT (id) DO NOTHING;
