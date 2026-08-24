-- Reversión de la migración 000008: elimina la inmutabilidad de la bitácora.
DROP TRIGGER IF EXISTS trg_prohibir_modificacion ON dbgs_schema.auditoria_eventos;
DROP FUNCTION IF EXISTS dbgs_schema.fn_prohibir_modificacion();
