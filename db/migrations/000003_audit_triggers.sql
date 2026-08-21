-- Script de triggers de auditoría 000003_audit_triggers.sql

CREATE OR REPLACE FUNCTION dbgs_schema.fn_auditar_cambios()
RETURNS TRIGGER AS $$ DECLARE
    v_usuario_id VARCHAR(100);
    v_ip VARCHAR(45);
    v_detalles TEXT;
BEGIN
    -- Intenta obtener las variables de sesión del contexto actual de la aplicación
    BEGIN
        v_usuario_id := NULLIF(current_setting('app.current_user_id', true), '');
    EXCEPTION WHEN OTHERS THEN
        v_usuario_id := NULL;
    END;

    -- Garantiza que la IP no sea NULL asignando fallback local
    v_ip := COALESCE(inet_client_addr()::VARCHAR, '127.0.0.1');

    IF (TG_OP = 'DELETE') THEN
        v_detalles := 'Datos eliminados: ' || row_to_json(OLD)::TEXT;
        
        INSERT INTO dbgs_schema.auditoria_eventos (
            id, usuario_id, username, operacion, recurso, detalles, resultado, ip_origen, fecha_creacion
        ) VALUES (
            gen_random_uuid(), v_usuario_id, current_user, 'DELETE', TG_TABLE_NAME, v_detalles, 'EXITOSO', v_ip, CURRENT_TIMESTAMP
        );
        RETURN OLD;
        
    ELSIF (TG_OP = 'UPDATE') THEN
        v_detalles := 'Antes: ' || row_to_json(OLD)::TEXT || ' | Nuevo: ' || row_to_json(NEW)::TEXT;
        
        INSERT INTO dbgs_schema.auditoria_eventos (
            id, usuario_id, username, operacion, recurso, detalles, resultado, ip_origen, fecha_creacion
        ) VALUES (
            gen_random_uuid(), v_usuario_id, current_user, 'UPDATE', TG_TABLE_NAME, v_detalles, 'EXITOSO', v_ip, CURRENT_TIMESTAMP
        );
        RETURN NEW;
        
    ELSIF (TG_OP = 'INSERT') THEN
        v_detalles := 'Datos insertados: ' || row_to_json(NEW)::TEXT;
        
        INSERT INTO dbgs_schema.auditoria_eventos (
            id, usuario_id, username, operacion, recurso, detalles, resultado, ip_origen, fecha_creacion
        ) VALUES (
            gen_random_uuid(), v_usuario_id, current_user, 'INSERT', TG_TABLE_NAME, v_detalles, 'EXITOSO', v_ip, CURRENT_TIMESTAMP
        );
        RETURN NEW;
    END IF;

    RETURN NULL;
END;
 $$ LANGUAGE plpgsql;