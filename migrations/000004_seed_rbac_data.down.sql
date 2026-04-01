DELETE FROM role_permissions
WHERE role_id IN (SELECT id FROM roles WHERE name IN ('admin', 'manager', 'viewer'));

DELETE FROM permissions
WHERE (resource, action) IN (
    ('users', 'create'),
    ('users', 'read'),
    ('users', 'update'),
    ('users', 'delete'),
    ('orders', 'create'),
    ('orders', 'read'),
    ('orders', 'update'),
    ('orders', 'delete'),
    ('settings', 'read'),
    ('settings', 'update')
);

DELETE FROM roles
WHERE name IN ('admin', 'manager', 'viewer');
