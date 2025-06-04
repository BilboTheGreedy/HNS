DO $$
DECLARE
    v_template_id INTEGER;
    v_epiroc_template_id INTEGER;
BEGIN
    -- Handle Server Template
    SELECT id INTO v_template_id FROM templates WHERE name = 'Server Template';

    IF v_template_id IS NOT NULL THEN
        -- Delete template groups
        DELETE FROM template_groups WHERE template_id = v_template_id;
    END IF;

    -- Handle Epiroc VM Standard template
    SELECT id INTO v_epiroc_template_id FROM templates WHERE name = 'Epiroc VM Standard';

    IF v_epiroc_template_id IS NOT NULL THEN
        -- Delete Epiroc template groups
        DELETE FROM template_groups WHERE template_id = v_epiroc_template_id;
        -- Delete the template
        DELETE FROM templates WHERE id = v_epiroc_template_id;
    END IF;
END $$;