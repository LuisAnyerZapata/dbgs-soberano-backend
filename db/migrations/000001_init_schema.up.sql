-- 1. Crear Esquema Principal
CREATE SCHEMA IF NOT EXISTS dbgs_schema;

-- Extensiones necesarias
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- 2. Tabla de Catalogos
CREATE TABLE IF NOT EXISTS dbgs_schema.catalogos (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    codigo VARCHAR(50) NOT NULL UNIQUE,
    nombre VARCHAR(150) NOT NULL,
    descripcion TEXT,
    estado BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(100) NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_by VARCHAR(100)
);

-- 3. Fuentes de Datos
CREATE TABLE IF NOT EXISTS dbgs_schema.fuentes_datos (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    nombre VARCHAR(150) NOT NULL,
    descripcion TEXT,
    estado BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(100) NOT NULL
);

-- 4. Conjuntos de Datos (Datasets)
CREATE TABLE IF NOT EXISTS dbgs_schema.conjuntos_datos (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    fuente_dato_id UUID NOT NULL REFERENCES dbgs_schema.fuentes_datos(id),
    nombre VARCHAR(150) NOT NULL,
    proposito TEXT NOT NULL,
    propietario_dato VARCHAR(150) NOT NULL,
    clasificacion VARCHAR(50) NOT NULL CHECK (clasificacion IN ('PUBLICO', 'RESTRINGIDO', 'CONFIDENCIAL')),
    estado BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(100) NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_by VARCHAR(100)
);

-- 5. Roles y Permisos (RBAC)
CREATE TABLE IF NOT EXISTS dbgs_schema.roles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    nombre VARCHAR(50) NOT NULL UNIQUE,
    descripcion TEXT,
    estado BOOLEAN NOT NULL DEFAULT TRUE
);

CREATE TABLE IF NOT EXISTS dbgs_schema.permisos (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    codigo VARCHAR(100) NOT NULL UNIQUE,
    descripcion TEXT
);

CREATE TABLE IF NOT EXISTS dbgs_schema.roles_permisos (
    rol_id UUID NOT NULL REFERENCES dbgs_schema.roles(id) ON DELETE CASCADE,
    permiso_id UUID NOT NULL REFERENCES dbgs_schema.permisos(id) ON DELETE CASCADE,
    PRIMARY KEY (rol_id, permiso_id)
);

-- 6. Usuarios
CREATE TABLE IF NOT EXISTS dbgs_schema.usuarios (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    username VARCHAR(100) NOT NULL UNIQUE,
    email VARCHAR(150) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    rol_id UUID NOT NULL REFERENCES dbgs_schema.roles(id),
    es_tecnico BOOLEAN NOT NULL DEFAULT FALSE,
    estado BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 7. Auditoria de Eventos (Inmutable)
CREATE TABLE IF NOT EXISTS dbgs_schema.auditoria_eventos (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    usuario_id VARCHAR(100),
    username VARCHAR(100) NOT NULL,
    operacion VARCHAR(100) NOT NULL,
    recurso VARCHAR(150) NOT NULL,
    detalles TEXT,
    resultado VARCHAR(20) NOT NULL CHECK (resultado IN ('EXITOSO', 'FALLIDO', 'DENEGADO')),
    ip_origen VARCHAR(45) NOT NULL,
    fecha_creacion TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Índices de Rendimiento
CREATE INDEX IF NOT EXISTS idx_catalogos_codigo ON dbgs_schema.catalogos(codigo);
CREATE INDEX IF NOT EXISTS idx_conjuntos_datos_clasificacion ON dbgs_schema.conjuntos_datos(clasificacion);
CREATE INDEX IF NOT EXISTS idx_auditoria_fecha ON dbgs_schema.auditoria_eventos(fecha_creacion DESC);