-- Seed de roles, permisos y usuarios para pruebas DBGS
-- Idempotente: puede ejecutarse varias veces con pgAdmin4

-- Roles de negocio
INSERT INTO dbgs_schema.roles (nombre, descripcion)
VALUES
  ('ADMIN_PLATFORM', 'Administrador de plataforma'),
  ('DBA', 'Administrador de base de datos'),
  ('DEVELOPER', 'Desarrollador institucional'),
  ('AUDITOR', 'Responsable de auditoría'),
  ('SERVICE_ACCOUNT', 'Cuenta de servicio para integraciones')
ON CONFLICT (nombre) DO NOTHING;

-- Permisos (códigos usados por la aplicación: recurso:accion)
INSERT INTO dbgs_schema.permisos (codigo, descripcion)
VALUES
  ('catalogos:leer', 'Consultar catálogos'),
  ('catalogos:escribir', 'Crear/Actualizar catálogos'),
  ('datasets:leer', 'Consultar metadatos de datasets'),
  ('auditoria:leer', 'Consultar bitácora de auditoría'),
  ('respaldo:ejecutar', 'Ejecutar respaldo'),
  ('restauracion:ejecutar', 'Ejecutar restauración'),
  ('usuarios:leer', 'Consultar usuarios'),
  ('usuarios:admin', 'Administrar cuentas y roles'),
  ('colecciones:actualizar', 'Actualizar estructura de colecciones dinámicas'),
  ('colecciones:eliminar', 'Eliminar colecciones dinámicas')
ON CONFLICT (codigo) DO NOTHING;

-- Asociar permisos a roles (idempotente)
INSERT INTO dbgs_schema.roles_permisos (rol_id, permiso_id)
SELECT r.id, p.id FROM dbgs_schema.roles r
JOIN dbgs_schema.permisos p ON p.codigo IN (
  'catalogos:leer', 'catalogos:escribir', 'datasets:leer', 'auditoria:leer',
  'respaldo:ejecutar', 'restauracion:ejecutar', 'usuarios:leer', 'usuarios:admin',
  'colecciones:actualizar', 'colecciones:eliminar'
)
WHERE r.nombre = 'ADMIN_PLATFORM'
ON CONFLICT (rol_id, permiso_id) DO NOTHING;

-- DBA: copia de seguridad y restauración, consulta usuarios
INSERT INTO dbgs_schema.roles_permisos (rol_id, permiso_id)
SELECT r.id, p.id FROM dbgs_schema.roles r
JOIN dbgs_schema.permisos p ON p.codigo IN ('respaldo:ejecutar','restauracion:ejecutar','usuarios:leer')
WHERE r.nombre = 'DBA'
ON CONFLICT (rol_id, permiso_id) DO NOTHING;

-- DEVELOPER: lectura/escritura sobre catálogos y datasets lectura
INSERT INTO dbgs_schema.roles_permisos (rol_id, permiso_id)
SELECT r.id, p.id FROM dbgs_schema.roles r
JOIN dbgs_schema.permisos p ON p.codigo IN ('catalogos:leer','catalogos:escribir','datasets:leer')
WHERE r.nombre = 'DEVELOPER'
ON CONFLICT (rol_id, permiso_id) DO NOTHING;

-- AUDITOR: solo lectura de auditoría y catálogos
INSERT INTO dbgs_schema.roles_permisos (rol_id, permiso_id)
SELECT r.id, p.id FROM dbgs_schema.roles r
JOIN dbgs_schema.permisos p ON p.codigo IN ('auditoria:leer','catalogos:leer')
WHERE r.nombre = 'AUDITOR'
ON CONFLICT (rol_id, permiso_id) DO NOTHING;

-- SERVICE_ACCOUNT: permisos limitados (lectura)
INSERT INTO dbgs_schema.roles_permisos (rol_id, permiso_id)
SELECT r.id, p.id FROM dbgs_schema.roles r
JOIN dbgs_schema.permisos p ON p.codigo IN ('datasets:leer','catalogos:leer')
WHERE r.nombre = 'SERVICE_ACCOUNT'
ON CONFLICT (rol_id, permiso_id) DO NOTHING;

-- Usuarios de prueba (no almacenan contraseña en esta tabla; autenticación de demo usa username 'admin')
INSERT INTO dbgs_schema.usuarios (username, email, password_hash, rol_id, es_tecnico, estado)
SELECT * FROM (
  SELECT 'admin'::varchar AS username, 'admin@example.local'::varchar AS email, '$2a$10$ml5UjdnV8l12NY8ccG4qXexZ10hgmnEI041G0lGQ.MOaBqfrF.SAK'::varchar AS password_hash, (SELECT id FROM dbgs_schema.roles WHERE nombre='ADMIN_PLATFORM') AS rol_id, true AS es_tecnico, true AS estado
) AS tmp
ON CONFLICT (username) DO NOTHING;

INSERT INTO dbgs_schema.usuarios (username, email, password_hash, rol_id, es_tecnico, estado)
SELECT * FROM (
  SELECT 'dba'::varchar, 'dba@example.local'::varchar, '$2a$10$liR7AbaN5dtBpROpoh5kHercsCZbhYiz.wic4tB8txOAj0OOkxF2.'::varchar, (SELECT id FROM dbgs_schema.roles WHERE nombre='DBA')::uuid, true, true
) AS tmp
ON CONFLICT (username) DO NOTHING;

INSERT INTO dbgs_schema.usuarios (username, email, password_hash, rol_id, es_tecnico, estado)
SELECT * FROM (
  SELECT 'dev'::varchar, 'dev@example.local'::varchar, '$2a$10$LxWusq0ZswovCRRzZ/Y5FOcXi1C9q7542vHlywlZMXO55enCWvF4S'::varchar, (SELECT id FROM dbgs_schema.roles WHERE nombre='DEVELOPER')::uuid, false, true
) AS tmp
ON CONFLICT (username) DO NOTHING;

INSERT INTO dbgs_schema.usuarios (username, email, password_hash, rol_id, es_tecnico, estado)
SELECT * FROM (
  SELECT 'auditor'::varchar, 'auditor@example.local'::varchar, '$2a$10$arx2Qq.1mIJEJyNb3KuGiOGncuIywUUfCO1W71ULnSV2xyp30Ejcy'::varchar, (SELECT id FROM dbgs_schema.roles WHERE nombre='AUDITOR')::uuid, false, true
) AS tmp
ON CONFLICT (username) DO NOTHING;

INSERT INTO dbgs_schema.usuarios (username, email, password_hash, rol_id, es_tecnico, estado)
SELECT * FROM (
  SELECT 'service_api'::varchar, 'service@example.local'::varchar, '$2a$10$.7QESBL97RRTPxz/am09P.wjRKt6CJQYAD0VbehtyUwd3ekCwz0pC'::varchar, (SELECT id FROM dbgs_schema.roles WHERE nombre='SERVICE_ACCOUNT')::uuid, true, true
) AS tmp
ON CONFLICT (username) DO NOTHING;

-- Nota: La autenticación real por contraseña todavía depende del repositorio/implementación.
-- En el flujo de pruebas actuales la cuenta 'admin' acepta la contraseña definida en las pruebas (hash demo).

-- Fin del seed
