-- Migración 000010: Tabla de Conexiones a Bases de Datos Externas
-- Almacena los metadatos de conexión. La contraseña viaja cifrada (AES-GCM) con
-- el secreto del servidor; nunca se expone hacia el cliente.

CREATE TABLE IF NOT EXISTS dbgs_schema.conexiones (
    id UUID PRIMARY KEY,
    nombre VARCHAR(150) NOT NULL,
    motor VARCHAR(20) NOT NULL DEFAULT 'postgresql', -- postgresql | mysql
    host VARCHAR(255) NOT NULL,
    puerto INT NOT NULL,
    usuario VARCHAR(100) NOT NULL,
    password_cifrado TEXT NOT NULL DEFAULT '',       -- AES-GCM, base64
    base_datos VARCHAR(150) NOT NULL,
    ssl_mode VARCHAR(20) NOT NULL DEFAULT 'disable',
    solo_lectura BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(100) NOT NULL DEFAULT 'system'
);

CREATE INDEX IF NOT EXISTS idx_conexiones_created ON dbgs_schema.conexiones(created_at DESC);
