-- Create database
CREATE DATABASE servername;

-- Connect to database
\c servername;

-- Create users table
CREATE TABLE users (
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
CREATE TABLE api_keys (
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
CREATE TABLE templates (
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
CREATE TABLE template_groups (
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
CREATE TABLE hostnames (
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
CREATE INDEX idx_hostnames_template_id ON hostnames(template_id);
CREATE INDEX idx_hostnames_status ON hostnames(status);
CREATE INDEX idx_hostnames_sequence_num ON hostnames(template_id, sequence_num);

-- Create index on template groups
CREATE INDEX idx_template_groups_template_id ON template_groups(template_id);

-- Add initial admin user (password: admin123)
INSERT INTO users (
    username, email, password_hash, first_name, last_name, 
    role, is_active, created_at, updated_at
) VALUES (
    'admin', 'admin@example.com', 
    '$2a$10$8KxO8t8eFJF0TvGQ9Jsj3OaLEhE6hYDE3eCWf/hX.CHQYnMxBJXOK', 
    'Admin', 'User', 'admin', TRUE, NOW(), NOW()
);

-- Add sample template
INSERT INTO templates (
    name, description, max_length, sequence_start, 
    sequence_length, sequence_padding, sequence_increment,
    created_by, created_at, updated_at, is_active
) VALUES (
    'Server Template', 'Standard server naming template', 
    15, 1, 3, TRUE, 1, 'admin', NOW(), NOW(), TRUE
);

-- Get the template ID
DO $$
DECLARE
    template_id INTEGER;
BEGIN
    SELECT id INTO template_id FROM templates WHERE name = 'Server Template';

    -- Add template groups
    INSERT INTO template_groups (
        template_id, name, length, position, 
        is_required, validation_type, validation_value
    ) VALUES
    (template_id, 'location', 2, 1, TRUE, 'list', 'SF,NY,LA,CH,AU'),
    (template_id, 'type', 2, 2, TRUE, 'list', 'DB,WS,AP,DC'),
    (template_id, 'environment', 2, 3, TRUE, 'list', 'DV,QA,ST,PR'),
    (template_id, 'sequence', 3, 4, TRUE, 'sequence', '');
END $$;