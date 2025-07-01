-- 1. 插入权限（保持不变）
INSERT INTO permissions (name, description, resource, action) VALUES
('文件上传', '允许上传文件到指定存储桶', 'FILE', 'UPLOAD'),
('文件下载', '允许从指定存储桶下载文件', 'FILE', 'DOWNLOAD'),
('文件查询', '允许查询指定存储桶中的文件', 'FILE', 'QUERY'),
('文件删除', '允许删除指定存储桶中的文件', 'FILE', 'DELETE'),
('存储桶配置', '允许配置和管理存储桶', 'BUCKET', 'CONFIGURE');

-- 2. 插入角色（保持不变）
INSERT INTO roles (name, description) VALUES
('国际一部', '国际一部成员，主要负责国际业务相关文件管理'),
('软件二组', '软件二组开发人员，负责项目相关文件管理'),
('IT流程', 'IT流程管理人员，负责流程文档管理'),
('运维组', '运维团队成员，负责系统运维相关文件管理');

-- 3. 插入存储桶映射（增加更多存储桶）
INSERT INTO region_bucket_mapping (region_code, bucket_name) VALUES
('cn-hangzhou', 'international-files'),     -- 国际业务文件
('cn-hangzhou', 'international-shared'),    -- 国际共享文件
('cn-shanghai', 'software-dev-files'),      -- 软件开发文件
('cn-shanghai', 'software-test-files'),     -- 软件测试文件
('cn-beijing', 'it-process-files'),         -- IT流程文件
('cn-beijing', 'it-docs-files'),            -- IT文档文件
('cn-shenzhen', 'ops-files'),               -- 运维文件
('cn-shenzhen', 'ops-backup-files');        -- 运维备份文件

-- 4. 为角色分配权限（保持不变）
-- 国际一部：所有权限
INSERT INTO role_permissions (role_id, permission_id)
SELECT 
    (SELECT id FROM roles WHERE name = '国际一部'),
    id
FROM permissions;

-- 软件二组：上传、下载、查询权限
INSERT INTO role_permissions (role_id, permission_id)
SELECT 
    (SELECT id FROM roles WHERE name = '软件二组'),
    id
FROM permissions 
WHERE action IN ('UPLOAD', 'DOWNLOAD', 'QUERY');

-- IT流程：上传、下载、查询、删除权限
INSERT INTO role_permissions (role_id, permission_id)
SELECT 
    (SELECT id FROM roles WHERE name = 'IT流程'),
    id
FROM permissions 
WHERE action IN ('UPLOAD', 'DOWNLOAD', 'QUERY', 'DELETE');

-- 运维组：所有权限
INSERT INTO role_permissions (role_id, permission_id)
SELECT 
    (SELECT id FROM roles WHERE name = '运维组'),
    id
FROM permissions;

-- 5. 为角色分配多个存储桶访问权限
-- 国际一部：访问国际相关存储桶
INSERT INTO role_region_bucket_access (role_id, region_bucket_mapping_id)
SELECT 
    (SELECT id FROM roles WHERE name = '国际一部'),
    id
FROM region_bucket_mapping 
WHERE bucket_name IN ('international-files', 'international-shared');

-- 软件二组：访问开发和测试存储桶
INSERT INTO role_region_bucket_access (role_id, region_bucket_mapping_id)
SELECT 
    (SELECT id FROM roles WHERE name = '软件二组'),
    id
FROM region_bucket_mapping 
WHERE bucket_name IN ('software-dev-files', 'software-test-files');

-- IT流程：访问IT相关存储桶
INSERT INTO role_region_bucket_access (role_id, region_bucket_mapping_id)
SELECT 
    (SELECT id FROM roles WHERE name = 'IT流程'),
    id
FROM region_bucket_mapping 
WHERE bucket_name IN ('it-process-files', 'it-docs-files');

-- 运维组：访问运维相关存储桶
INSERT INTO role_region_bucket_access (role_id, region_bucket_mapping_id)
SELECT 
    (SELECT id FROM roles WHERE name = '运维组'),
    id
FROM region_bucket_mapping 
WHERE bucket_name IN ('ops-files', 'ops-backup-files');