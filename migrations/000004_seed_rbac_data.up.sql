INSERT INTO roles (name, description) VALUES
    ('admin', 'Full access to all resources'),
    ('manager', 'Manage users and orders'),
    ('viewer', 'Read-only access');

INSERT INTO permissions (resource, action) VALUES
    ('users', 'create'),
    ('users', 'read'),
    ('users', 'update'),
    ('users', 'delete'),
    ('orders', 'create'),
    ('orders', 'read'),
    ('orders', 'update'),
    ('orders', 'delete'),
    ('settings', 'read'),
    ('settings', 'update');

INSERT INTO role_permissions (role_id, permission_id)
SELECT roles.id, permissions.id
FROM roles, permissions
WHERE roles.name = 'admin';

INSERT INTO role_permissions (role_id, permission_id)
SELECT roles.id, permissions.id
FROM roles, permissions
WHERE roles.name = 'manager' AND permissions.resource IN ('users', 'orders');

INSERT INTO role_permissions (role_id, permission_id)
SELECT roles.id, permissions.id
FROM roles, permissions
WHERE roles.name = 'viewer' AND permissions.action = 'read';
