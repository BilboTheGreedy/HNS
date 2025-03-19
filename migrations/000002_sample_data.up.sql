-- Migration: sample_data

-- Add initial admin user (password: admin123)
INSERT INTO users (
    username, email, password_hash, first_name, last_name, 
    role, is_active, created_at, updated_at
) VALUES (
    'admin', 'admin@example.com', 
    '$2a$10$8KxO8t8eFJF0TvGQ9Jsj3OaLEhE6hYDE3eCWf/hX.CHQYnMxBJXOK', 
    'Admin', 'User', 'admin', TRUE, NOW(), NOW()
) ON CONFLICT (username) DO NOTHING;

-- Add sample template
INSERT INTO templates (
    name, description, max_length, sequence_start, 
    sequence_length, sequence_padding, sequence_increment,
    created_by, created_at, updated_at, is_active
) VALUES (
    'Server Template', 'Standard server naming template', 
    15, 1, 3, TRUE, 1, 'admin', NOW(), NOW(), TRUE
) ON CONFLICT (name) DO NOTHING;

-- Get the template ID and add template groups
DO $$
DECLARE
    template_id INTEGER;
BEGIN
    SELECT id INTO template_id FROM templates WHERE name = 'Server Template';

    IF template_id IS NOT NULL THEN
        -- Add template groups
        INSERT INTO template_groups (
            template_id, name, length, position, 
            is_required, validation_type, validation_value
        ) VALUES
        (template_id, 'location', 2, 1, TRUE, 'list', 'SF,NY,LA,CH,AU'),
        (template_id, 'type', 2, 2, TRUE, 'list', 'DB,WS,AP,DC'),
        (template_id, 'environment', 2, 3, TRUE, 'list', 'DV,QA,ST,PR'),
        (template_id, 'sequence', 3, 4, TRUE, 'sequence', '')
        ON CONFLICT (template_id, position) DO NOTHING;
    END IF;
END $$;