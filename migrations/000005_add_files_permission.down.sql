DELETE FROM role_permissions
WHERE permission_id IN (SELECT id FROM permissions WHERE resource = 'files' AND action = 'create');

DELETE FROM permissions WHERE resource = 'files' AND action = 'create';
