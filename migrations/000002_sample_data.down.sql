DO $$
DECLARE
    v_template_id INTEGER;
BEGIN
    SELECT id INTO v_template_id FROM templates WHERE name = 'Server Template';

    IF v_template_id IS NOT NULL THEN
        -- Delete template groups
        DELETE FROM template_groups WHERE template_id = v_template_id;
    END IF;
END $$;