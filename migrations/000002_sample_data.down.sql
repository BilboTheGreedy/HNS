-- Revert: sample_data

-- Get the template ID and remove template groups
DO $$
DECLARE
    template_id INTEGER;
BEGIN
    SELECT id INTO template_id FROM templates WHERE name = 'Server Template';

    IF template_id IS NOT NULL THEN
        -- Delete template groups
        DELETE FROM template_groups WHERE template_id = template_id;
    END IF;
END $$;

-- Remove sample template
DELETE FROM templates WHERE name = 'Server Template';

-- Remove admin user
DELETE FROM users WHERE username = 'admin';