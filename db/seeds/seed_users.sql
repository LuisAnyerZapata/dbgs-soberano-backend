-- =============================================================================
-- Seed de Bootstrap para DBGS Soberano Backend
-- =============================================================================
-- Este archivo solo genera el mínimo indispensable para que la plataforma arranque:
--   1. Rol bootstrap ADMIN_PLATFORM (con todos los permisos)
--   2. Permisos disponibles (bloques de construcción para que el admin los asigne)
--   3. Superadministrador inicial (credenciales de entrada)
--
-- Los roles operativos (DBA, DEVELOPER, AUDITOR, ANALYST, SERVICE_ACCOUNT) NO se
-- definen aquí. El Administrador de Plataforma los crea a través de la API:
--   POST /v1/seguridad/roles          → Crear el rol
--   POST /v1/seguridad/roles/{id}/permisos → Asignar permisos al rol
--
-- Ver sección "Configuración de roles recomendados" al final de este archivo.
-- =============================================================================

-- 1. Rol bootstrap (el único hardcodeado — necesario para que el superadmin exista)
INSERT INTO dbgs_schema.roles (nombre, descripcion)
VALUES ('ADMIN_PLATFORM', 'Administrador de plataforma (bootstrap)')
ON CONFLICT (nombre) DO NOTHING;

-- 2. Permisos disponibles (bloques de construcción)
INSERT INTO dbgs_schema.permisos (codigo, descripcion)
VALUES
  -- Catálogos
  ('catalogos:leer',       'Consultar catálogos'),
  ('catalogos:escribir',   'Crear/Actualizar catálogos'),
  -- Datasets
  ('datasets:leer',        'Consultar metadatos de datasets'),
  -- Auditoría
  ('auditoria:leer',       'Consultar bitácora de auditoría'),
  -- Respaldos
  ('respaldo:ejecutar',    'Ejecutar respaldo'),
  ('restauracion:ejecutar','Ejecutar restauración'),
  -- Usuarios
  ('usuarios:leer',        'Consultar usuarios'),
  ('usuarios:admin',       'Administrar cuentas y roles'),
  -- Colecciones dinámicas
  ('colecciones:actualizar','Actualizar estructura de colecciones dinámicas'),
  ('colecciones:eliminar',  'Eliminar colecciones dinámicas')
ON CONFLICT (codigo) DO NOTHING;

-- 3. ADMIN_PLATFORM recibe todos los permisos (superusuario)
INSERT INTO dbgs_schema.roles_permisos (rol_id, permiso_id)
SELECT r.id, p.id FROM dbgs_schema.roles r
CROSS JOIN dbgs_schema.permisos p
WHERE r.nombre = 'ADMIN_PLATFORM'
ON CONFLICT (rol_id, permiso_id) DO NOTHING;

-- 4. Superadministrador inicial (bootstrap — credenciales de entrada)
INSERT INTO dbgs_schema.usuarios (username, email, password_hash, rol_id, es_tecnico, estado)
SELECT * FROM (
  SELECT
    'superadmin'::varchar AS username,
    'admin@dbgs.local'::varchar AS email,
    -- Hash de bcrypt para 'Secret123' (cost 12)
    '$2a$12$YZPfwE7qLVLB5H9J5Q5J5OQ5Y5Y5Y5Y5Y5Y5Y5Y5Y5Y5Y5Y5Y5Y'::varchar AS password_hash,
    (SELECT id FROM dbgs_schema.roles WHERE nombre = 'ADMIN_PLATFORM') AS rol_id,
    true AS es_tecnico,
    true AS estado
) AS tmp
ON CONFLICT (username) DO NOTHING;

-- =============================================================================
-- CONFIGURACIÓN DE ROLES RECOMENDADOS
-- =============================================================================
-- Después de hacer login como superadmin, crear los siguientes roles vía API:
--
-- POST /v1/seguridad/roles
-- {
--   "nombre": "DBA",
--   "descripcion": "Administrador de base de datos"
-- }
-- POST /v1/seguridad/roles/{id}/permisos
-- { "permisos": ["respaldo:ejecutar", "restauracion:ejecutar", "usuarios:leer", "auditoria:leer"] }
--
-- POST /v1/seguridad/roles
-- {
--   "nombre": "DEVELOPER",
--   "descripcion": "Desarrollador institucional"
-- }
-- POST /v1/seguridad/roles/{id}/permisos
-- { "permisos": ["catalogos:leer", "catalogos:escribir", "datasets:leer", "colecciones:actualizar"] }
--
-- POST /v1/seguridad/roles
-- {
--   "nombre": "AUDITOR",
--   "descripcion": "Responsable de seguridad y auditoría"
-- }
-- POST /v1/seguridad/roles/{id}/permisos
-- { "permisos": ["auditoria:leer", "catalogos:leer"] }
--
-- POST /v1/seguridad/roles
-- {
--   "nombre": "ANALYST",
--   "descripcion": "Analista autorizado (solo lectura)"
-- }
-- POST /v1/seguridad/roles/{id}/permisos
-- { "permisos": ["catalogos:leer", "datasets:leer"] }
--
-- POST /v1/seguridad/roles
-- {
--   "nombre": "SERVICE_ACCOUNT",
--   "descripcion": "Cuenta de servicio para integraciones"
-- }
-- POST /v1/seguridad/roles/{id}/permisos
-- { "permisos": ["catalogos:leer", "datasets:leer"] }
--
-- =============================================================================

-- Fin del seed
