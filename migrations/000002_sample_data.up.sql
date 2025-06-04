DO $$
DECLARE
    v_template_id INTEGER;
    v_epiroc_template_id INTEGER;
BEGIN
    -- Handle existing Server Template
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

    -- Create Epiroc VM Standard template
    INSERT INTO templates (
        name, description, max_length, sequence_start, sequence_length, 
        sequence_padding, sequence_increment, sequence_position, 
        created_by, created_at, updated_at, is_active
    ) VALUES (
        'Epiroc VM Standard',
        'Epiroc 15-character VM naming standard: [Unit(3)][Type(1)][Provider(1)][Region(4)][Env(1)][Function(2)][Seq(3)]',
        15, 1, 3, TRUE, 1, 7, 'system', NOW(), NOW(), TRUE
    ) ON CONFLICT (name) DO NOTHING
    RETURNING id INTO v_epiroc_template_id;

    -- Get the ID if it already exists
    IF v_epiroc_template_id IS NULL THEN
        SELECT id INTO v_epiroc_template_id FROM templates WHERE name = 'Epiroc VM Standard';
    END IF;

    -- Add Epiroc template groups
    IF v_epiroc_template_id IS NOT NULL THEN
        INSERT INTO template_groups (
            template_id, name, length, position, 
            is_required, validation_type, validation_value
        ) VALUES
        (v_epiroc_template_id, 'unit_code', 3, 1, TRUE, 'list', 'SFD,SFS,DAL,COP,JAC,RDT,AVT,RIG,FRD,EST,FAG,CER,ALL,HPI,NJB,OUL,ARE,CHR,CLK,FRB,GAR,GHA,KAL,LAN,NAC,OEB,SKE,UDD,ZED,GDE,JHA,NAN,ROS,ITS,LEG,CHI,BRA,USA,PER,MEX,CAN,AUS,IND,AFR,RUS,MON,KAZ,EUR,GLO,MAL,THA,KOR,JAP,NIG,GEO,FIN'),
        (v_epiroc_template_id, 'type', 1, 2, TRUE, 'fixed', 'S'),
        (v_epiroc_template_id, 'provider', 1, 3, TRUE, 'list', 'E,M,A,G'),
        (v_epiroc_template_id, 'region', 4, 4, TRUE, 'list', 'OCAU,ASCN,ASIN,ASSG,EUSE,AFZA,NAUS,ACL1,ACL2,AUEA,SEAU,INCE,SHA1,BJB1,EAAS,JPEA,JPWE,KRCE,KRSO,INSO,SEAS,INWE,FRCE,DEWC,ISCE,ITNO,NOEU,NWEA,PLCE,QACE,SANO,SAWE,SWCE,SZNO,UANO,UKSO,UKWE,WEEU,BRSO,CCAN,ECAN,CEUS,CLCE,EUS1,EUS2,NCUS,SCUS,WCUS,WUS1,WUS2,WUS3'),
        (v_epiroc_template_id, 'environment', 1, 5, TRUE, 'list', 'P,T,D'),
        (v_epiroc_template_id, 'function', 2, 6, TRUE, 'list', 'AS,BU,DB,DC,FS,IS,MG,PO,WS,PV,CC,UC,OT'),
        (v_epiroc_template_id, 'sequence', 3, 7, TRUE, 'sequence', '')
        ON CONFLICT (template_id, position) DO NOTHING;
    END IF;
END $$;