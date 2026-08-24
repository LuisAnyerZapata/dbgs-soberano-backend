-- Reversión 000009: elimina las tablas del dominio de Respaldos.

DROP TABLE IF EXISTS dbgs_schema.metricas_sistema;
DROP TABLE IF EXISTS dbgs_schema.logs_operativos;
DROP TABLE IF EXISTS dbgs_schema.restauraciones;
DROP TABLE IF EXISTS dbgs_schema.respaldos;
