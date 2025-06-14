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
    ('11111111-1111-1111-1111-111111111111', 'u1', '$2a$10$nEFHmT4WLwKa8DB5sEQgv.3fU/4gr.931HiWw6JlnhiaSYBIuH/My', 'user', '京AUV001', 60.0),
    ('22222222-2222-2222-2222-222222222222', 'u2', '$2a$10$nEFHmT4WLwKa8DB5sEQgv.3fU/4gr.931HiWw6JlnhiaSYBIuH/My', 'user', '京AUV002', 65.0),
    ('33333333-3333-3333-3333-333333333333', 'u3', '$2a$10$nEFHmT4WLwKa8DB5sEQgv.3fU/4gr.931HiWw6JlnhiaSYBIuH/My', 'user', '京AUV003', 70.0),
    ('44444444-4444-4444-4444-444444444444', 'u4', '$2a$10$nEFHmT4WLwKa8DB5sEQgv.3fU/4gr.931HiWw6JlnhiaSYBIuH/My', 'user', '京AUV004', 120.0),
    ('55555555-5555-5555-5555-555555555555', 'u5', '$2a$10$nEFHmT4WLwKa8DB5sEQgv.3fU/4gr.931HiWw6JlnhiaSYBIuH/My', 'user', '京AUV005', 80.0),
    ('66666666-6666-6666-6666-666666666666', 'u6', '$2a$10$nEFHmT4WLwKa8DB5sEQgv.3fU/4gr.931HiWw6JlnhiaSYBIuH/My', 'user', '京AUV006', 85.0),
    ('77777777-7777-7777-7777-777777777777', 'u7', '$2a$10$nEFHmT4WLwKa8DB5sEQgv.3fU/4gr.931HiWw6JlnhiaSYBIuH/My', 'user', '京AUV007', 90.0),
    ('88888888-8888-8888-8888-888888888888', 'u8', '$2a$10$nEFHmT4WLwKa8DB5sEQgv.3fU/4gr.931HiWw6JlnhiaSYBIuH/My', 'user', '京AUV008', 95.0),
    ('99999999-9999-9999-9999-999999999999', 'u9', '$2a$10$nEFHmT4WLwKa8DB5sEQgv.3fU/4gr.931HiWw6JlnhiaSYBIuH/My', 'user', '京AUV009', 100.0),
    ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'u10', '$2a$10$nEFHmT4WLwKa8DB5sEQgv.3fU/4gr.931HiWw6JlnhiaSYBIuH/My', 'user', '京AUV010', 110.0),
    ('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', 'u11', '$2a$10$nEFHmT4WLwKa8DB5sEQgv.3fU/4gr.931HiWw6JlnhiaSYBIuH/My', 'user', '京AUV011', 115.0),
    ('cccccccc-cccc-cccc-cccc-cccccccccccc', 'u12', '$2a$10$nEFHmT4WLwKa8DB5sEQgv.3fU/4gr.931HiWw6JlnhiaSYBIuH/My', 'user', '京AUV012', 75.0),
    ('dddddddd-dddd-dddd-dddd-dddddddddddd', 'u13', '$2a$10$nEFHmT4WLwKa8DB5sEQgv.3fU/4gr.931HiWw6JlnhiaSYBIuH/My', 'user', '京AUV013', 85.0),
    ('eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee', 'u14', '$2a$10$nEFHmT4WLwKa8DB5sEQgv.3fU/4gr.931HiWw6JlnhiaSYBIuH/My', 'user', '京AUV014', 90.0),
    ('ffffffff-ffff-ffff-ffff-ffffffffffff', 'u15', '$2a$10$nEFHmT4WLwKa8DB5sEQgv.3fU/4gr.931HiWw6JlnhiaSYBIuH/My', 'user', '京AUV015', 95.0),
    ('10101010-1010-1010-1010-101010101010', 'u16', '$2a$10$nEFHmT4WLwKa8DB5sEQgv.3fU/4gr.931HiWw6JlnhiaSYBIuH/My', 'user', '京AUV016', 100.0),
    ('20202020-2020-2020-2020-202020202020', 'u17', '$2a$10$nEFHmT4WLwKa8DB5sEQgv.3fU/4gr.931HiWw6JlnhiaSYBIuH/My', 'user', '京AUV017', 105.0),
    ('30303030-3030-3030-3030-303030303030', 'u18', '$2a$10$nEFHmT4WLwKa8DB5sEQgv.3fU/4gr.931HiWw6JlnhiaSYBIuH/My', 'user', '京AUV018', 110.0),
    ('40404040-4040-4040-4040-404040404040', 'u19', '$2a$10$nEFHmT4WLwKa8DB5sEQgv.3fU/4gr.931HiWw6JlnhiaSYBIuH/My', 'user', '京AUV019', 80.0),
    ('50505050-5050-5050-5050-505050505050', 'u20', '$2a$10$nEFHmT4WLwKa8DB5sEQgv.3fU/4gr.931HiWw6JlnhiaSYBIuH/My', 'user', '京AUV020', 85.0),
    ('60606060-6060-6060-6060-606060606060', 'u21', '$2a$10$nEFHmT4WLwKa8DB5sEQgv.3fU/4gr.931HiWw6JlnhiaSYBIuH/My', 'user', '京AUV021', 90.0),
    ('70707070-7070-7070-7070-707070707070', 'u22', '$2a$10$nEFHmT4WLwKa8DB5sEQgv.3fU/4gr.931HiWw6JlnhiaSYBIuH/My', 'user', '京AUV022', 125.0),
    ('80808080-8080-8080-8080-808080808080', 'u23', '$2a$10$nEFHmT4WLwKa8DB5sEQgv.3fU/4gr.931HiWw6JlnhiaSYBIuH/My', 'user', '京AUV023', 130.0)
ON CONFLICT (id) DO NOTHING;
