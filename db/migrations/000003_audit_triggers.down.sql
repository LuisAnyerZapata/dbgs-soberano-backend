-- Reversión de la migración 000003: elimina la función de auditoría forense.
DROP FUNCTION IF EXISTS dbgs_schema.fn_auditar_cambios();
