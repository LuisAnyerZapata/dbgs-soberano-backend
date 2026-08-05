-- Script de triggers de auditoría 000003_audit_triggers.sql

CREATE OR REPLACE FUNCTION fn_auditar_cambios()
RETURNS TRIGGER AS $$
DECLARE
    v_usuario_id UUID;
    v_ip VARCHAR(45);
BEGIN
    -- Intenta obtener las variables de sesión del contexto actual de la aplicación
    BEGIN
        v_usuario_id := NULLIF(current_setting('app.current_user_id', true), '')::UUID;
    EXCEPTION WHEN OTHERS THEN
        v_usuario_id := NULL;
    END;

    -- CORRECCIÓN: Garantiza que la IP no sea NULL asignando fallback local mediante COALESCE
    v_ip := COALESCE(inet_client_addr()::VARCHAR, '127.0.0.1');

    IF (TG_OP = 'DELETE') THEN
        INSERT INTO auditoria_eventos (
            id, usuario_id, accion, tabla_afectada, registro_id, datos_anteriores, ip_origen, fecha_evento
        ) VALUES (
            gen_random_uuid(), v_usuario_id, 'DELETE', TG_TABLE_NAME, OLD.id, row_to_json(OLD)::jsonb, v_ip, CURRENT_TIMESTAMP
        );
        RETURN OLD;
    ELSIF (TG_OP = 'UPDATE') THEN
        INSERT INTO auditoria_eventos (
            id, usuario_id, accion, tabla_afectada, registro_id, datos_anteriores, datos_nuevos, ip_origen, fecha_evento
        ) VALUES (
            gen_random_uuid(), v_usuario_id, 'UPDATE', TG_TABLE_NAME, NEW.id, row_to_json(OLD)::jsonb, row_to_json(NEW)::jsonb, v_ip, CURRENT_TIMESTAMP
        );
        RETURN NEW;
    ELSIF (TG_OP = 'INSERT') THEN
        INSERT INTO auditoria_eventos (
            id, usuario_id, accion, tabla_afectada, registro_id, datos_nuevos, ip_origen, fecha_evento
        ) VALUES (
            gen_random_uuid(), v_usuario_id, 'INSERT', TG_TABLE_NAME, NEW.id, row_to_json(NEW)::jsonb, v_ip, CURRENT_TIMESTAMP
        );
        RETURN NEW;
    END IF;

    RETURN NULL;
END;
$$ LANGUAGE plpgsql;