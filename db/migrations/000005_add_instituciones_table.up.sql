-- Migración 000005: Creación de la tabla maestra de Instituciones.
-- Es vital para el Multi-Tenancy y la trazabilidad de a quién pertenecen los datos o datasets.

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS dbgs_schema.instituciones (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    codigo_institucional VARCHAR(50) NOT NULL UNIQUE, -- Ej: GOB-MONAGAS-DCTI
    nombre_institucion VARCHAR(200) NOT NULL,
    siglas VARCHAR(30),
    estatus BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(100) NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_by VARCHAR(100)
);

-- Índice para búsquedas rápidas por estatus activo
CREATE INDEX IF NOT EXISTS idx_instituciones_estatus ON dbgs_schema.instituciones(estatus);