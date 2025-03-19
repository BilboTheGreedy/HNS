-- Revert: initial_schema

-- Drop indexes
DROP INDEX IF EXISTS idx_template_groups_template_id;
DROP INDEX IF EXISTS idx_hostnames_sequence_num;
DROP INDEX IF EXISTS idx_hostnames_status;
DROP INDEX IF EXISTS idx_hostnames_template_id;

-- Drop tables in reverse order of dependencies
DROP TABLE IF EXISTS hostnames;
DROP TABLE IF EXISTS template_groups;
DROP TABLE IF EXISTS templates;
DROP TABLE IF EXISTS api_keys;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS schema_migrations;