DO $$
DECLARE
    v_template_id INTEGER;
BEGIN
    SELECT id INTO v_template_id FROM templates WHERE name = 'Server Template';

    IF v_template_id IS NOT NULL THEN
        -- Add template groups
        INSERT INTO template_groups (
            template_id, name, length, position, 
            is_required, validation_type, validation_value
        ) VALUES
        (v_template_id, 'location', 2, 1, TRUE, 'list', 'SF,NY,LA,CH,AU'),
        (v_template_id, 'type', 2, 2, TRUE, 'list', 'DB,WS,AP,DC'),
        (v_template_id, 'environment', 2, 3, TRUE, 'list', 'DV,QA,ST,PR'),
        (v_template_id, 'sequence', 3, 4, TRUE, 'sequence', '')
        ON CONFLICT (template_id, position) DO NOTHING;
    END IF;
END $$;