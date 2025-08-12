-- 创建 WebDAV 令牌表
CREATE TABLE IF NOT EXISTS webdav_tokens (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE NOT NULL,
    bucket VARCHAR(100) NOT NULL,
    token_hash VARCHAR(255) NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_webdav_tokens_user_id ON webdav_tokens(user_id);
CREATE INDEX IF NOT EXISTS idx_webdav_tokens_token_hash ON webdav_tokens(token_hash);
CREATE INDEX IF NOT EXISTS idx_webdav_tokens_bucket ON webdav_tokens(bucket);
CREATE INDEX IF NOT EXISTS idx_webdav_tokens_expires_at ON webdav_tokens(expires_at);

-- 添加 WebDAV 相关权限
INSERT INTO permissions (name, description, resource, action) VALUES 
('webdav_read', 'WebDAV 读取权限', 'webdav', 'read'),
('webdav_write', 'WebDAV 写入权限', 'webdav', 'write'),
('webdav_delete', 'WebDAV 删除权限', 'webdav', 'delete'),
('webdav_manage', 'WebDAV 管理权限', 'webdav', 'manage')
ON CONFLICT (name) DO NOTHING;

-- 为管理员角色分配 WebDAV 权限
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.name = 'admin' AND p.resource = 'webdav'
ON CONFLICT DO NOTHING;
