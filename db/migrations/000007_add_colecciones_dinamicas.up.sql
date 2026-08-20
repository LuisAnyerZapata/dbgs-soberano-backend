-- Migración 000007: Tabla para registrar los metadatos de las tablas dinámicas creadas por el motor
CREATE TABLE IF NOT EXISTS dbgs_schema.colecciones_dinamicas (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    nombre_logico VARCHAR(100) NOT NULL UNIQUE,
    nombre_fisico VARCHAR(100) NOT NULL UNIQUE,
    descripcion TEXT,
    institucion_id UUID, -- Al omitir NOT NULL, PostgreSQL asume que permite nulos (NULLABLE)
    estructura JSONB NOT NULL,
    esta_activa BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(100) NOT NULL
);