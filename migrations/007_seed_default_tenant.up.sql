-- Default tenant
INSERT INTO tenants (id, name, unique_code, status) VALUES
    ('00000000-0000-0000-0000-000000000001', 'Default Tenant', 'default', 'active')
ON CONFLICT (unique_code) DO NOTHING;

-- Default admin user (password: Admin@123, bcrypt cost 12)
INSERT INTO users (id, tenant_id, email, password_hash, name, status, email_verified) VALUES
    ('00000000-0000-0000-0000-000000000001',
     '00000000-0000-0000-0000-000000000001',
     'admin@iam.local',
     '$2a$12$eua2PeLJG3k7hENuSEeuZu8oApXN0W2R..qMkqIUN9khWe9Sj2lI6',
     'System Admin',
     'active',
     true)
ON CONFLICT (tenant_id, email) DO NOTHING;

-- Default admin role
INSERT INTO roles (id, tenant_id, name, description, is_system) VALUES
    ('00000000-0000-0000-0000-000000000001',
     '00000000-0000-0000-0000-000000000001',
     'admin',
     'System administrator',
     true)
ON CONFLICT (tenant_id, name) DO NOTHING;

-- Assign admin role to admin user
INSERT INTO user_roles (user_id, role_id, tenant_id) VALUES
    ('00000000-0000-0000-0000-000000000001',
     '00000000-0000-0000-0000-000000000001',
     '00000000-0000-0000-0000-000000000001')
ON CONFLICT (user_id, role_id) DO NOTHING;
