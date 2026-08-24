-- Migración 000008: Inmutabilidad forense de la bitácora de auditoría.
-- Crea el trigger trg_prohibir_modificacion que bloquea UPDATE, DELETE y TRUNCATE
-- sobre auditoria_eventos a nivel de motor. Solo se permite INSERT (vía fn_auditar_cambios).

CREATE OR REPLACE FUNCTION dbgs_schema.fn_prohibir_modificacion()
RETURNS TRIGGER AS $$
BEGIN
    RAISE EXCEPTION 'AUDITORÍA INMUTABLE: la operación % está prohibida sobre la tabla %.%',
        TG_OP, TG_TABLE_SCHEMA, TG_TABLE_NAME
        USING ERRCODE = 'insufficient_privilege';
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_prohibir_modificacion ON dbgs_schema.auditoria_eventos;

CREATE TRIGGER trg_prohibir_modificacion
    BEFORE UPDATE OR DELETE OR TRUNCATE ON dbgs_schema.auditoria_eventos
    FOR EACH STATEMENT EXECUTE FUNCTION dbgs_schema.fn_prohibir_modificacion();
