-- Insertar Roles por Defecto
INSERT INTO dbgs_schema.roles (id, nombre, descripcion, estado) VALUES
('11111111-1111-1111-1111-111111111111', 'ADMINISTRADOR', 'Administrador del sistema', true),
('22222222-2222-2222-2222-222222222222', 'OPERADOR', 'Operador funcional de datos', true),
('33333333-3333-3333-3333-333333333333', 'AUDITOR', 'Auditor de seguridad e historial', true)
ON CONFLICT (nombre) DO NOTHING;

-- Insertar Permisos Basicos
INSERT INTO dbgs_schema.permisos (id, codigo, descripcion) VALUES
('a1111111-1111-1111-1111-111111111111', 'CATALOGO_LEER', 'Permite consultar los catálogos'),
('a2222222-2222-2222-2222-222222222222', 'CATALOGO_CREAR', 'Permite registrar nuevos catálogos'),
('a3333333-3333-3333-3333-333333333333', 'DATASET_LEER', 'Permite consultar conjuntos de datos'),
('a4444444-4444-4444-4444-444444444444', 'AUDITORIA_LEER', 'Permite consultar eventos de auditoría')
ON CONFLICT (codigo) DO NOTHING;

-- Asignar Permisos a Roles
INSERT INTO dbgs_schema.roles_permisos (rol_id, permiso_id) VALUES
('11111111-1111-1111-1111-111111111111', 'a1111111-1111-1111-1111-111111111111'),
('11111111-1111-1111-1111-111111111111', 'a2222222-2222-2222-2222-222222222222'),
('11111111-1111-1111-1111-111111111111', 'a3333333-3333-3333-3333-333333333333'),
('11111111-1111-1111-1111-111111111111', 'a4444444-4444-4444-4444-444444444444')
ON CONFLICT DO NOTHING;

-- Insertar Usuario Administrador Inicial
INSERT INTO dbgs_schema.usuarios (id, username, email, password_hash, rol_id, es_tecnico, estado) VALUES
('99999999-9999-9999-9999-999999999999', 'admin_dbgs', 'admin@dbgs.gob.ve', '$2a$12$LJ3m4ys3Hz0JEh2l5KIX9.FGHvhMxOqJqKJv5h6kzT7nR8mGsUCvG', '11111111-1111-1111-1111-111111111111', true, true)
ON CONFLICT (username) DO NOTHING;

-- Insertar Fuentes de Datos de Ejemplo
INSERT INTO dbgs_schema.fuentes_datos (id, nombre, descripcion, estado, created_by) VALUES
('f1111111-1111-1111-1111-111111111111', 'Sistema Central de Identificación', 'Base de datos principal de registros', true, 'admin_dbgs')
ON CONFLICT DO NOTHING;