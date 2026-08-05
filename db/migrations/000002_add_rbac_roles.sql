-- Script de migración 000002_add_rbac_roles.up.sql

CREATE TABLE IF NOT EXISTS roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    nombre VARCHAR(50) UNIQUE NOT NULL,
    descripcion TEXT,
    creado_en TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS usuario_roles (
    usuario_id UUID NOT NULL,
    rol_id UUID NOT NULL,
    PRIMARY KEY (usuario_id, rol_id),
    FOREIGN KEY (usuario_id) REFERENCES usuarios(id) ON DELETE CASCADE,
    FOREIGN KEY (rol_id) REFERENCES roles(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS rol_permisos (
    rol_id UUID NOT NULL,
    permiso_id UUID NOT NULL,
    PRIMARY KEY (rol_id, permiso_id),
    FOREIGN KEY (rol_id) REFERENCES roles(id) ON DELETE CASCADE,
    FOREIGN KEY (permiso_id) REFERENCES permisos(id) ON DELETE CASCADE
);

CREATE OR REPLACE VIEW v_usuarios_permisos AS
SELECT 
    ur.usuario_id,
    r.nombre AS rol_nombre,
    p.nombre AS permiso_nombre,
    p.recurso,
    p.accion
FROM usuario_roles ur
INNER JOIN roles r ON ur.rol_id = r.id
INNER JOIN rol_permisos rp ON r.id = rp.rol_id
INNER JOIN permisos p ON rp.permiso_id = p.id;