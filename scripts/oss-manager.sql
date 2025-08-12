/*
 Navicat Premium Dump SQL

 Source Server         : pg-ems-plus-uat
 Source Server Type    : PostgreSQL
 Source Server Version : 160004 (160004)
 Source Host           : 121.37.156.167:15423
 Source Catalog        : oss_manager
 Source Schema         : public

 Target Server Type    : PostgreSQL
 Target Server Version : 160004 (160004)
 File Encoding         : 65001

 Date: 12/08/2025 10:21:33
*/


-- ----------------------------
-- Sequence structure for audit_logs_id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."audit_logs_id_seq";
CREATE SEQUENCE "public"."audit_logs_id_seq" 
INCREMENT 1
MINVALUE  1
MAXVALUE 9223372036854775807
START 1
CACHE 1;
-- 

-- ----------------------------
-- Sequence structure for oss_configs_id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."oss_configs_id_seq";
CREATE SEQUENCE "public"."oss_configs_id_seq" 
INCREMENT 1
MINVALUE  1
MAXVALUE 9223372036854775807
START 1
CACHE 1;


-- ----------------------------
-- Sequence structure for oss_files_id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."oss_files_id_seq";
CREATE SEQUENCE "public"."oss_files_id_seq" 
INCREMENT 1
MINVALUE  1
MAXVALUE 9223372036854775807
START 1
CACHE 1;


-- ----------------------------
-- Sequence structure for permissions_id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."permissions_id_seq";
CREATE SEQUENCE "public"."permissions_id_seq" 
INCREMENT 1
MINVALUE  1
MAXVALUE 9223372036854775807
START 1
CACHE 1;


-- ----------------------------
-- Sequence structure for region_bucket_mapping_id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."region_bucket_mapping_id_seq";
CREATE SEQUENCE "public"."region_bucket_mapping_id_seq" 
INCREMENT 1
MINVALUE  1
MAXVALUE 2147483647
START 1
CACHE 1;


-- ----------------------------
-- Sequence structure for role_region_bucket_access_id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."role_region_bucket_access_id_seq";
CREATE SEQUENCE "public"."role_region_bucket_access_id_seq" 
INCREMENT 1
MINVALUE  1
MAXVALUE 2147483647
START 1
CACHE 1;


-- ----------------------------
-- Sequence structure for roles_id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."roles_id_seq";
CREATE SEQUENCE "public"."roles_id_seq" 
INCREMENT 1
MINVALUE  1
MAXVALUE 9223372036854775807
START 1
CACHE 1;


-- ----------------------------
-- Sequence structure for users_id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."users_id_seq";
CREATE SEQUENCE "public"."users_id_seq" 
INCREMENT 1
MINVALUE  1
MAXVALUE 9223372036854775807
START 1
CACHE 1;


-- ----------------------------
-- Table structure for audit_logs
-- ----------------------------
DROP TABLE IF EXISTS "public"."audit_logs";
CREATE TABLE "public"."audit_logs" (
  "id" int8 NOT NULL DEFAULT nextval('audit_logs_id_seq'::regclass),
  "created_at" timestamptz(6),
  "updated_at" timestamptz(6),
  "deleted_at" timestamptz(6),
  "user_id" int8,
  "username" varchar(50) COLLATE "pg_catalog"."default",
  "action" varchar(50) COLLATE "pg_catalog"."default" NOT NULL,
  "resource_type" varchar(50) COLLATE "pg_catalog"."default",
  "resource_id" varchar(100) COLLATE "pg_catalog"."default",
  "details" jsonb,
  "ip_address" varchar(50) COLLATE "pg_catalog"."default",
  "user_agent" text COLLATE "pg_catalog"."default",
  "status" varchar(20) COLLATE "pg_catalog"."default" NOT NULL
)
;


-- ----------------------------
-- Table structure for oss_configs
-- ----------------------------
DROP TABLE IF EXISTS "public"."oss_configs";
CREATE TABLE "public"."oss_configs" (
  "id" int8 NOT NULL DEFAULT nextval('oss_configs_id_seq'::regclass),
  "created_at" timestamptz(6),
  "updated_at" timestamptz(6),
  "deleted_at" timestamptz(6),
  "name" varchar(100) COLLATE "pg_catalog"."default" NOT NULL,
  "storage_type" varchar(20) COLLATE "pg_catalog"."default" NOT NULL,
  "access_key" varchar(255) COLLATE "pg_catalog"."default" NOT NULL,
  "secret_key" varchar(255) COLLATE "pg_catalog"."default" NOT NULL,
  "endpoint" varchar(255) COLLATE "pg_catalog"."default" NOT NULL,
  "bucket" varchar(100) COLLATE "pg_catalog"."default" NOT NULL,
  "region" varchar(50) COLLATE "pg_catalog"."default",
  "is_default" bool DEFAULT false,
  "url_expire_time" int4 DEFAULT 86400
)
;


-- ----------------------------
-- Table structure for oss_files
-- ----------------------------
DROP TABLE IF EXISTS "public"."oss_files";
CREATE TABLE "public"."oss_files" (
  "id" int8 NOT NULL DEFAULT nextval('oss_files_id_seq'::regclass),
  "created_at" timestamptz(6),
  "updated_at" timestamptz(6),
  "deleted_at" timestamptz(6),
  "filename" varchar(255) COLLATE "pg_catalog"."default" NOT NULL,
  "original_filename" varchar(255) COLLATE "pg_catalog"."default" NOT NULL,
  "file_size" int8 NOT NULL,
  "md5" varchar(32) COLLATE "pg_catalog"."default",
  "md5_status" varchar(20) COLLATE "pg_catalog"."default" DEFAULT 'PENDING'::character varying,
  "storage_type" varchar(20) COLLATE "pg_catalog"."default" NOT NULL,
  "bucket" varchar(100) COLLATE "pg_catalog"."default" NOT NULL,
  "object_key" varchar(255) COLLATE "pg_catalog"."default" NOT NULL,
  "download_url" text COLLATE "pg_catalog"."default",
  "expires_at" timestamptz(6),
  "uploader_id" int8 NOT NULL,
  "upload_ip" varchar(50) COLLATE "pg_catalog"."default",
  "status" varchar(20) COLLATE "pg_catalog"."default" DEFAULT 'ACTIVE'::character varying,
  "config_id" int8 NOT NULL
)
;


-- ----------------------------
-- Table structure for permissions
-- ----------------------------
DROP TABLE IF EXISTS "public"."permissions";
CREATE TABLE "public"."permissions" (
  "id" int8 NOT NULL DEFAULT nextval('permissions_id_seq'::regclass),
  "created_at" timestamptz(6),
  "updated_at" timestamptz(6),
  "deleted_at" timestamptz(6),
  "name" varchar(100) COLLATE "pg_catalog"."default" NOT NULL,
  "description" text COLLATE "pg_catalog"."default",
  "resource" varchar(100) COLLATE "pg_catalog"."default" NOT NULL,
  "action" varchar(50) COLLATE "pg_catalog"."default" NOT NULL
)
;


-- ----------------------------
-- Table structure for region_bucket_mapping
-- ----------------------------
DROP TABLE IF EXISTS "public"."region_bucket_mapping";
CREATE TABLE "public"."region_bucket_mapping" (
  "id" int4 NOT NULL DEFAULT nextval('region_bucket_mapping_id_seq'::regclass),
  "region_code" varchar(50) COLLATE "pg_catalog"."default" NOT NULL,
  "bucket_name" varchar(255) COLLATE "pg_catalog"."default" NOT NULL,
  "created_at" timestamp(6) DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamp(6) DEFAULT CURRENT_TIMESTAMP,
  "deleted_at" timestamp(6) DEFAULT NULL::timestamp without time zone
)
;


-- ----------------------------
-- Table structure for role_permissions
-- ----------------------------
DROP TABLE IF EXISTS "public"."role_permissions";
CREATE TABLE "public"."role_permissions" (
  "permission_id" int8 NOT NULL,
  "role_id" int8 NOT NULL
)
;


-- ----------------------------
-- Table structure for role_region_bucket_access
-- ----------------------------
DROP TABLE IF EXISTS "public"."role_region_bucket_access";
CREATE TABLE "public"."role_region_bucket_access" (
  "id" int4 NOT NULL DEFAULT nextval('role_region_bucket_access_id_seq'::regclass),
  "role_id" int4,
  "region_bucket_mapping_id" int4,
  "created_at" timestamp(6) DEFAULT CURRENT_TIMESTAMP
)
;


-- ----------------------------
-- Table structure for roles
-- ----------------------------
DROP TABLE IF EXISTS "public"."roles";
CREATE TABLE "public"."roles" (
  "id" int8 NOT NULL DEFAULT nextval('roles_id_seq'::regclass),
  "created_at" timestamptz(6),
  "updated_at" timestamptz(6),
  "deleted_at" timestamptz(6),
  "name" varchar(50) COLLATE "pg_catalog"."default" NOT NULL,
  "description" text COLLATE "pg_catalog"."default"
)
;


-- ----------------------------
-- Table structure for user_roles
-- ----------------------------
DROP TABLE IF EXISTS "public"."user_roles";
CREATE TABLE "public"."user_roles" (
  "role_id" int8 NOT NULL,
  "user_id" int8 NOT NULL,
  "created_at" timestamp(6) DEFAULT CURRENT_TIMESTAMP
)
;


-- ----------------------------
-- Table structure for users
-- ----------------------------
DROP TABLE IF EXISTS "public"."users";
CREATE TABLE "public"."users" (
  "id" int8 NOT NULL DEFAULT nextval('users_id_seq'::regclass),
  "created_at" timestamptz(6),
  "updated_at" timestamptz(6),
  "deleted_at" timestamptz(6),
  "username" varchar(50) COLLATE "pg_catalog"."default" NOT NULL,
  "password" varchar(255) COLLATE "pg_catalog"."default" NOT NULL,
  "email" varchar(100) COLLATE "pg_catalog"."default" NOT NULL,
  "real_name" varchar(100) COLLATE "pg_catalog"."default",
  "status" bool DEFAULT true
)
;


-- ----------------------------
-- Alter sequences owned by
-- ----------------------------
ALTER SEQUENCE "public"."audit_logs_id_seq"
OWNED BY "public"."audit_logs"."id";
SELECT setval('"public"."audit_logs_id_seq"', 456, true);

-- ----------------------------
-- Alter sequences owned by
-- ----------------------------
ALTER SEQUENCE "public"."oss_configs_id_seq"
OWNED BY "public"."oss_configs"."id";
SELECT setval('"public"."oss_configs_id_seq"', 2, true);

-- ----------------------------
-- Alter sequences owned by
-- ----------------------------
ALTER SEQUENCE "public"."oss_files_id_seq"
OWNED BY "public"."oss_files"."id";
SELECT setval('"public"."oss_files_id_seq"', 179, true);

-- ----------------------------
-- Alter sequences owned by
-- ----------------------------
ALTER SEQUENCE "public"."permissions_id_seq"
OWNED BY "public"."permissions"."id";
SELECT setval('"public"."permissions_id_seq"', 26, true);

-- ----------------------------
-- Alter sequences owned by
-- ----------------------------
ALTER SEQUENCE "public"."region_bucket_mapping_id_seq"
OWNED BY "public"."region_bucket_mapping"."id";
SELECT setval('"public"."region_bucket_mapping_id_seq"', 773, true);

-- ----------------------------
-- Alter sequences owned by
-- ----------------------------
ALTER SEQUENCE "public"."role_region_bucket_access_id_seq"
OWNED BY "public"."role_region_bucket_access"."id";
SELECT setval('"public"."role_region_bucket_access_id_seq"', 531, true);

-- ----------------------------
-- Alter sequences owned by
-- ----------------------------
ALTER SEQUENCE "public"."roles_id_seq"
OWNED BY "public"."roles"."id";
SELECT setval('"public"."roles_id_seq"', 8, true);

-- ----------------------------
-- Alter sequences owned by
-- ----------------------------
ALTER SEQUENCE "public"."users_id_seq"
OWNED BY "public"."users"."id";
SELECT setval('"public"."users_id_seq"', 24, true);

-- ----------------------------
-- Indexes structure for table audit_logs
-- ----------------------------
CREATE INDEX "idx_audit_logs_action" ON "public"."audit_logs" USING btree (
  "action" COLLATE "pg_catalog"."default" "pg_catalog"."text_ops" ASC NULLS LAST
);
CREATE INDEX "idx_audit_logs_created_at" ON "public"."audit_logs" USING btree (
  "created_at" "pg_catalog"."timestamptz_ops" ASC NULLS LAST
);
CREATE INDEX "idx_audit_logs_deleted_at" ON "public"."audit_logs" USING btree (
  "deleted_at" "pg_catalog"."timestamptz_ops" ASC NULLS LAST
);
CREATE INDEX "idx_audit_logs_user_id" ON "public"."audit_logs" USING btree (
  "user_id" "pg_catalog"."int8_ops" ASC NULLS LAST
);

-- ----------------------------
-- Primary Key structure for table audit_logs
-- ----------------------------
ALTER TABLE "public"."audit_logs" ADD CONSTRAINT "audit_logs_pkey" PRIMARY KEY ("id");

-- ----------------------------
-- Indexes structure for table oss_configs
-- ----------------------------
CREATE INDEX "idx_oss_configs_deleted_at" ON "public"."oss_configs" USING btree (
  "deleted_at" "pg_catalog"."timestamptz_ops" ASC NULLS LAST
);

-- ----------------------------
-- Primary Key structure for table oss_configs
-- ----------------------------
ALTER TABLE "public"."oss_configs" ADD CONSTRAINT "oss_configs_pkey" PRIMARY KEY ("id");

-- ----------------------------
-- Indexes structure for table oss_files
-- ----------------------------
CREATE INDEX "idx_oss_files_created_at" ON "public"."oss_files" USING btree (
  "created_at" "pg_catalog"."timestamptz_ops" ASC NULLS LAST
);
CREATE INDEX "idx_oss_files_deleted_at" ON "public"."oss_files" USING btree (
  "deleted_at" "pg_catalog"."timestamptz_ops" ASC NULLS LAST
);
CREATE INDEX "idx_oss_files_filename" ON "public"."oss_files" USING btree (
  "filename" COLLATE "pg_catalog"."default" "pg_catalog"."text_ops" ASC NULLS LAST
);
CREATE INDEX "idx_oss_files_md5" ON "public"."oss_files" USING btree (
  "md5" COLLATE "pg_catalog"."default" "pg_catalog"."text_ops" ASC NULLS LAST
);
CREATE INDEX "idx_oss_files_uploader_id" ON "public"."oss_files" USING btree (
  "uploader_id" "pg_catalog"."int8_ops" ASC NULLS LAST
);

-- ----------------------------
-- Primary Key structure for table oss_files
-- ----------------------------
ALTER TABLE "public"."oss_files" ADD CONSTRAINT "oss_files_pkey" PRIMARY KEY ("id");

-- ----------------------------
-- Indexes structure for table permissions
-- ----------------------------
CREATE INDEX "idx_permissions_deleted_at" ON "public"."permissions" USING btree (
  "deleted_at" "pg_catalog"."timestamptz_ops" ASC NULLS LAST
);
CREATE UNIQUE INDEX "idx_permissions_name" ON "public"."permissions" USING btree (
  "name" COLLATE "pg_catalog"."default" "pg_catalog"."text_ops" ASC NULLS LAST
);

-- ----------------------------
-- Primary Key structure for table permissions
-- ----------------------------
ALTER TABLE "public"."permissions" ADD CONSTRAINT "permissions_pkey" PRIMARY KEY ("id");

-- ----------------------------
-- Indexes structure for table region_bucket_mapping
-- ----------------------------
CREATE INDEX "idx_region_bucket_mapping_bucket_name" ON "public"."region_bucket_mapping" USING btree (
  "bucket_name" COLLATE "pg_catalog"."default" "pg_catalog"."text_ops" ASC NULLS LAST
);
CREATE INDEX "idx_region_bucket_mapping_region_code" ON "public"."region_bucket_mapping" USING btree (
  "region_code" COLLATE "pg_catalog"."default" "pg_catalog"."text_ops" ASC NULLS LAST
);

-- ----------------------------
-- Uniques structure for table region_bucket_mapping
-- ----------------------------
ALTER TABLE "public"."region_bucket_mapping" ADD CONSTRAINT "region_bucket_mapping_region_code_bucket_name_key" UNIQUE ("region_code", "bucket_name");

-- ----------------------------
-- Primary Key structure for table region_bucket_mapping
-- ----------------------------
ALTER TABLE "public"."region_bucket_mapping" ADD CONSTRAINT "region_bucket_mapping_pkey" PRIMARY KEY ("id");

-- ----------------------------
-- Primary Key structure for table role_permissions
-- ----------------------------
ALTER TABLE "public"."role_permissions" ADD CONSTRAINT "role_permissions_pkey" PRIMARY KEY ("permission_id", "role_id");

-- ----------------------------
-- Indexes structure for table role_region_bucket_access
-- ----------------------------
CREATE INDEX "idx_role_region_bucket_access_rbm_id" ON "public"."role_region_bucket_access" USING btree (
  "region_bucket_mapping_id" "pg_catalog"."int4_ops" ASC NULLS LAST
);
CREATE INDEX "idx_role_region_bucket_access_role_id" ON "public"."role_region_bucket_access" USING btree (
  "role_id" "pg_catalog"."int4_ops" ASC NULLS LAST
);

-- ----------------------------
-- Uniques structure for table role_region_bucket_access
-- ----------------------------
ALTER TABLE "public"."role_region_bucket_access" ADD CONSTRAINT "role_region_bucket_access_role_id_region_bucket_mapping_id_key" UNIQUE ("role_id", "region_bucket_mapping_id");

-- ----------------------------
-- Primary Key structure for table role_region_bucket_access
-- ----------------------------
ALTER TABLE "public"."role_region_bucket_access" ADD CONSTRAINT "role_region_bucket_access_pkey" PRIMARY KEY ("id");

-- ----------------------------
-- Indexes structure for table roles
-- ----------------------------
CREATE INDEX "idx_roles_deleted_at" ON "public"."roles" USING btree (
  "deleted_at" "pg_catalog"."timestamptz_ops" ASC NULLS LAST
);
CREATE UNIQUE INDEX "idx_roles_name" ON "public"."roles" USING btree (
  "name" COLLATE "pg_catalog"."default" "pg_catalog"."text_ops" ASC NULLS LAST
);

-- ----------------------------
-- Primary Key structure for table roles
-- ----------------------------
ALTER TABLE "public"."roles" ADD CONSTRAINT "roles_pkey" PRIMARY KEY ("id");

-- ----------------------------
-- Primary Key structure for table user_roles
-- ----------------------------
ALTER TABLE "public"."user_roles" ADD CONSTRAINT "user_roles_pkey" PRIMARY KEY ("role_id", "user_id");

-- ----------------------------
-- Indexes structure for table users
-- ----------------------------
CREATE INDEX "idx_users_deleted_at" ON "public"."users" USING btree (
  "deleted_at" "pg_catalog"."timestamptz_ops" ASC NULLS LAST
);
CREATE UNIQUE INDEX "idx_users_email" ON "public"."users" USING btree (
  "email" COLLATE "pg_catalog"."default" "pg_catalog"."text_ops" ASC NULLS LAST
);
CREATE UNIQUE INDEX "idx_users_username" ON "public"."users" USING btree (
  "username" COLLATE "pg_catalog"."default" "pg_catalog"."text_ops" ASC NULLS LAST
);

-- ----------------------------
-- Primary Key structure for table users
-- ----------------------------
ALTER TABLE "public"."users" ADD CONSTRAINT "users_pkey" PRIMARY KEY ("id");

-- ----------------------------
-- Foreign Keys structure for table oss_files
-- ----------------------------
ALTER TABLE "public"."oss_files" ADD CONSTRAINT "fk_oss_files_uploader" FOREIGN KEY ("uploader_id") REFERENCES "public"."users" ("id") ON DELETE NO ACTION ON UPDATE NO ACTION;

-- ----------------------------
-- Foreign Keys structure for table role_permissions
-- ----------------------------
ALTER TABLE "public"."role_permissions" ADD CONSTRAINT "fk_role_permissions_permission" FOREIGN KEY ("permission_id") REFERENCES "public"."permissions" ("id") ON DELETE NO ACTION ON UPDATE NO ACTION;
ALTER TABLE "public"."role_permissions" ADD CONSTRAINT "fk_role_permissions_role" FOREIGN KEY ("role_id") REFERENCES "public"."roles" ("id") ON DELETE NO ACTION ON UPDATE NO ACTION;

-- ----------------------------
-- Foreign Keys structure for table role_region_bucket_access
-- ----------------------------
ALTER TABLE "public"."role_region_bucket_access" ADD CONSTRAINT "role_region_bucket_access_region_bucket_mapping_id_fkey" FOREIGN KEY ("region_bucket_mapping_id") REFERENCES "public"."region_bucket_mapping" ("id") ON DELETE CASCADE ON UPDATE NO ACTION;
ALTER TABLE "public"."role_region_bucket_access" ADD CONSTRAINT "role_region_bucket_access_role_id_fkey" FOREIGN KEY ("role_id") REFERENCES "public"."roles" ("id") ON DELETE CASCADE ON UPDATE NO ACTION;

-- ----------------------------
-- Foreign Keys structure for table user_roles
-- ----------------------------
ALTER TABLE "public"."user_roles" ADD CONSTRAINT "fk_user_roles_role" FOREIGN KEY ("role_id") REFERENCES "public"."roles" ("id") ON DELETE NO ACTION ON UPDATE NO ACTION;
ALTER TABLE "public"."user_roles" ADD CONSTRAINT "fk_user_roles_user" FOREIGN KEY ("user_id") REFERENCES "public"."users" ("id") ON DELETE NO ACTION ON UPDATE NO ACTION;

-- ----------------------------
-- Initial data setup
-- ----------------------------

-- Insert default admin role
INSERT INTO "public"."roles" ("id", "name", "description", "created_at", "updated_at") 
VALUES (1, 'admin', 'Administrator with full access', NOW(), NOW())
ON CONFLICT ("name") DO NOTHING;

-- Insert default user role
INSERT INTO "public"."roles" ("id", "name", "description", "created_at", "updated_at") 
VALUES (2, 'user', 'Regular user with limited access', NOW(), NOW())
ON CONFLICT ("name") DO NOTHING;

-- Insert basic permissions
INSERT INTO "public"."permissions" ("id", "name", "description", "resource", "action", "created_at", "updated_at") VALUES
(1, 'MANAGE_USERS', 'Manage user accounts', 'USER', 'MANAGE', NOW(), NOW()),
(2, 'MANAGE_ROLES', 'Manage roles and permissions', 'ROLE', 'MANAGE', NOW(), NOW()),
(3, 'MANAGE_OSS', 'Manage OSS configurations', 'OSS', 'MANAGE', NOW(), NOW()),
(4, 'UPLOAD_FILES', 'Upload files', 'FILE', 'UPLOAD', NOW(), NOW()),
(5, 'DOWNLOAD_FILES', 'Download files', 'FILE', 'DOWNLOAD', NOW(), NOW()),
(6, 'VIEW_FILES', 'View file list', 'FILE', 'VIEW', NOW(), NOW()),
(7, 'DELETE_FILES', 'Delete files', 'FILE', 'DELETE', NOW(), NOW()),
(8, 'VIEW_AUDIT_LOGS', 'View audit logs', 'AUDIT', 'VIEW', NOW(), NOW())
ON CONFLICT ("name") DO NOTHING;

-- Assign all permissions to admin role
INSERT INTO "public"."role_permissions" ("role_id", "permission_id") VALUES
(1, 1), (1, 2), (1, 3), (1, 4), (1, 5), (1, 6), (1, 7), (1, 8)
ON CONFLICT ("role_id", "permission_id") DO NOTHING;

-- Assign basic permissions to user role
INSERT INTO "public"."role_permissions" ("role_id", "permission_id") VALUES
(2, 4), (2, 5), (2, 6)
ON CONFLICT ("role_id", "permission_id") DO NOTHING;

-- Insert default admin user (password: admin123)
INSERT INTO "public"."users" ("id", "username", "password", "email", "real_name", "status", "created_at", "updated_at")
VALUES (1, 'admin', '$2a$10$N9qo8uLOickgx2ZMRZoMye3FGHaOMr4fGf.HDY./WdFo8qZpMV7JG', 'admin@ossmanager.local', 'System Administrator', true, NOW(), NOW())
ON CONFLICT ("username") DO UPDATE SET
  "password" = EXCLUDED."password",
  "email" = EXCLUDED."email",
  "updated_at" = NOW();

-- Assign admin role to admin user
INSERT INTO "public"."user_roles" ("user_id", "role_id", "created_at")
VALUES (1, 1, NOW())
ON CONFLICT ("user_id", "role_id") DO NOTHING;

-- Update sequence values to avoid conflicts
SELECT setval('"public"."roles_id_seq"', 10, true);
SELECT setval('"public"."users_id_seq"', 10, true);
SELECT setval('"public"."permissions_id_seq"', 50, true);
