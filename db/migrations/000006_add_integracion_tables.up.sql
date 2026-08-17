-- Migración 000006: Tabla para la gestión de Integración (Clientes API)
-- Permite que otros sistemas gubernamentales se autentiquen sin necesidad de un usuario humano.

CREATE TABLE IF NOT EXISTS dbgs_schema.clientes_api (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    nombre_cliente VARCHAR(150) NOT NULL,
    tipo_cliente VARCHAR(30) NOT NULL DEFAULT 'INTERNAL_SERVICE',
    token_hash VARCHAR(255) NOT NULL, -- Hash SHA-256 de la API Key
    version_contrato VARCHAR(20) DEFAULT 'v1',
    estatus BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(100) NOT NULL
);

-- Índice único para buscar rápidamente por el hash durante la autenticación
CREATE UNIQUE INDEX IF NOT EXISTS idx_clientes_api_hash ON dbgs_schema.clientes_api(token_hash);