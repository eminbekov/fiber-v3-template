INSERT INTO permissions (resource, action) VALUES ('files', 'create');

INSERT INTO role_permissions (role_id, permission_id)
SELECT roles.id, permissions.id
FROM roles, permissions
WHERE roles.name = 'manager' AND permissions.resource = 'files' AND permissions.action = 'create';
