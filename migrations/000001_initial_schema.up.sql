-- Migration: initial_schema

-- Create users table
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    role VARCHAR(20) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    last_login TIMESTAMP,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);

-- Create API keys table
CREATE TABLE IF NOT EXISTS api_keys (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    key VARCHAR(255) UNIQUE NOT NULL,
    scope VARCHAR(255) NOT NULL,
    last_used TIMESTAMP,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL,
    UNIQUE (user_id, name)
);

-- Create templates table
CREATE TABLE IF NOT EXISTS templates (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) UNIQUE NOT NULL,
    description TEXT,
    max_length INTEGER NOT NULL,
    sequence_start INTEGER NOT NULL DEFAULT 1,
    sequence_length INTEGER NOT NULL,
    sequence_padding BOOLEAN NOT NULL DEFAULT TRUE,
    sequence_increment INTEGER NOT NULL DEFAULT 1,
    sequence_position INTEGER,
    created_by VARCHAR(100) NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE
);

-- Create template groups table
CREATE TABLE IF NOT EXISTS template_groups (
    id SERIAL PRIMARY KEY,
    template_id INTEGER NOT NULL REFERENCES templates(id) ON DELETE CASCADE,
    name VARCHAR(50) NOT NULL,
    length INTEGER NOT NULL,
    position INTEGER NOT NULL,
    is_required BOOLEAN NOT NULL DEFAULT TRUE,
    validation_type VARCHAR(20),
    validation_value TEXT,
    UNIQUE (template_id, position)
);

-- Create hostnames table
CREATE TABLE IF NOT EXISTS hostnames (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) UNIQUE NOT NULL,
    template_id INTEGER NOT NULL REFERENCES templates(id),
    status VARCHAR(20) NOT NULL,
    sequence_num INTEGER NOT NULL,
    reserved_by VARCHAR(100) NOT NULL,
    reserved_at TIMESTAMP NOT NULL,
    committed_by VARCHAR(100),
    committed_at TIMESTAMP,
    released_by VARCHAR(100),
    released_at TIMESTAMP,
    dns_verified BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);

-- Create index on hostnames
CREATE INDEX IF NOT EXISTS idx_hostnames_template_id ON hostnames(template_id);
CREATE INDEX IF NOT EXISTS idx_hostnames_status ON hostnames(status);
CREATE INDEX IF NOT EXISTS idx_hostnames_sequence_num ON hostnames(template_id, sequence_num);

-- Create index on template groups
CREATE INDEX IF NOT EXISTS idx_template_groups_template_id ON template_groups(template_id);

-- Add schema version tracking table for migrations
CREATE TABLE IF NOT EXISTS schema_migrations (
    version bigint NOT NULL,
    dirty boolean NOT NULL,
    applied_at timestamp with time zone DEFAULT now() NOT NULL
);