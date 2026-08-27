-- Migración 000011: Tabla de APIs Públicas Publicadas
-- Registra las APIs de solo lectura publicadas sobre tablas de conexiones externas.

CREATE TABLE IF NOT EXISTS dbgs_schema.apis_publicadas (
    id UUID PRIMARY KEY,
    nombre VARCHAR(150) NOT NULL,
    descripcion TEXT NOT NULL DEFAULT '',
    slug VARCHAR(120) NOT NULL,
    conexion_id UUID NOT NULL REFERENCES dbgs_schema.conexiones(id) ON DELETE CASCADE,
    conexion_nombre VARCHAR(150) NOT NULL DEFAULT '',
    esquema VARCHAR(120) NOT NULL DEFAULT 'public',
    tabla VARCHAR(150) NOT NULL,
    max_filas INT NOT NULL DEFAULT 500,
    activa BOOLEAN NOT NULL DEFAULT TRUE,
    api_key VARCHAR(255) NOT NULL,
    endpoint VARCHAR(255) NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(100) NOT NULL DEFAULT 'system'
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_apis_publicadas_slug ON dbgs_schema.apis_publicadas(slug);
