-- Migración 000009: Tablas del Dominio de Respaldos y Operación.
-- Registra cada respaldo/restauración ejecutada por el motor (pg_dump/pg_restore),
-- además de logs operativos y métricas básicas para el panel administrativo.

CREATE TABLE IF NOT EXISTS dbgs_schema.respaldos (
    id UUID PRIMARY KEY,
    tipo VARCHAR(30) NOT NULL DEFAULT 'FULL',
    estado VARCHAR(20) NOT NULL DEFAULT 'EN_PROGRESO', -- EN_PROGRESO | COMPLETADO | FALLIDO
    ruta_archivo TEXT NOT NULL DEFAULT '',
    tamano_bytes BIGINT NOT NULL DEFAULT 0,
    detalles TEXT NOT NULL DEFAULT '',
    retencion_dias INT NOT NULL DEFAULT 30,
    usuario_creador VARCHAR(100) NOT NULL,
    fecha_creacion TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    fecha_finalizacion TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_respaldos_fecha ON dbgs_schema.respaldos(fecha_creacion DESC);
CREATE INDEX IF NOT EXISTS idx_respaldos_estado ON dbgs_schema.respaldos(estado);

CREATE TABLE IF NOT EXISTS dbgs_schema.restauraciones (
    id UUID PRIMARY KEY,
    backup_id UUID NOT NULL REFERENCES dbgs_schema.respaldos(id) ON DELETE CASCADE,
    usuario VARCHAR(100) NOT NULL,
    estado VARCHAR(20) NOT NULL DEFAULT 'EN_PROGRESO', -- EN_PROGRESO | COMPLETADO | FALLIDO
    validado BOOLEAN NOT NULL DEFAULT FALSE,
    observaciones TEXT NOT NULL DEFAULT '',
    fecha_creacion TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_restauraciones_backup ON dbgs_schema.restauraciones(backup_id);

CREATE TABLE IF NOT EXISTS dbgs_schema.logs_operativos (
    id UUID PRIMARY KEY,
    nivel VARCHAR(10) NOT NULL, -- INFO | WARN | ERROR
    modulo VARCHAR(50) NOT NULL,
    mensaje TEXT NOT NULL,
    fecha_creacion TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS dbgs_schema.metricas_sistema (
    id UUID PRIMARY KEY,
    nombre VARCHAR(100) NOT NULL,
    valor NUMERIC(15,4) NOT NULL,
    unidad VARCHAR(20) NOT NULL DEFAULT '',
    fecha_creacion TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
