-- Seed: 02_datos_prueba_sinteticos.sql
-- Descripción: Datos sintéticos ficticios para pruebas funcionales, filtros y pruebas de carga.

SET search_path TO dbgs_schema, public;

-- 1. Conjuntos de Datos (Datasets) Sintéticos
INSERT INTO dbgs_schema.conjuntos_datos (id, fuente_dato_id, nombre, proposito, propietario_dato, clasificacion, estado, created_by) VALUES
('d2000000-0000-0000-0000-000000000001', 'f1000000-0000-0000-0000-000000000001', 'Directorio de Cedulados V2', 'Verificación de identidad para trámites institucionales', 'Ministerio de Interiores', 'RESTRINGIDO', true, 'admin_dbgs'),
('d2000000-0000-0000-0000-000000000002', 'f1000000-0000-0000-0000-000000000002', 'Estadísticas Epidemiológicas Anuales', 'Consolidado público para estudios de salud pública', 'Ministerio de Salud', 'PUBLICO', true, 'admin_dbgs'),
('d2000000-0000-0000-0000-000000000003', 'f1000000-0000-0000-0000-000000000003', 'Registro de Propiedad Inmobiliaria', 'Análisis fiscal y planificación territorial', 'Catastro Nacional', 'CONFIDENCIAL', true, 'admin_dbgs'),
('d2000000-0000-0000-0000-000000000004', 'f1000000-0000-0000-0000-000000000001', 'Histórico de Trámites Consulares', 'Control migratorio e historial de expedientes', 'Cancillería', 'CONFIDENCIAL', false, 'admin_dbgs')
ON CONFLICT (id) DO NOTHING;

-- 2. Usuarios Sintéticos de Prueba (Diferentes roles)
INSERT INTO dbgs_schema.usuarios (id, username, email, password_hash, rol_id, es_tecnico, estado) VALUES
('02000000-0000-0000-0000-000000000001', 'operador_demo', 'operador@dbgs.gob.ve', '$2a$12$LJ3m4ys3Hz0JEh2l5KIX9.FGHvhMxOqJqKJv5h6kzT7nR8mGsUCvG', '22222222-2222-2222-2222-222222222222', false, true),
('02000000-0000-0000-0000-000000000002', 'auditor_demo', 'auditor@dbgs.gob.ve', '$2a$12$LJ3m4ys3Hz0JEh2l5KIX9.FGHvhMxOqJqKJv5h6kzT7nR8mGsUCvG', '33333333-3333-3333-3333-333333333333', false, true),
('02000000-0000-0000-0000-000000000003', 'tecnico_soporte', 'soporte@dbgs.gob.ve', '$2a$12$LJ3m4ys3Hz0JEh2l5KIX9.FGHvhMxOqJqKJv5h6kzT7nR8mGsUCvG', '22222222-2222-2222-2222-222222222222', true, true)
ON CONFLICT (username) DO NOTHING;

-- 3. Eventos de Auditoría Sintéticos para Pruebas de Trazabilidad
INSERT INTO dbgs_schema.auditoria_eventos (id, usuario_id, username, operacion, recurso, detalles, resultado, ip_origen, fecha_creacion) VALUES
('a9000000-0000-0000-0000-000000000001', '99999999-9999-9999-9999-999999999999', 'admin_dbgs', 'LOGIN', 'auth_service', 'Inicio de sesión exitoso desde consola admin', 'EXITOSO', '192.168.1.50', NOW() - INTERVAL '2 hours'),
('a9000000-0000-0000-0000-000000000002', 'u2000000-0000-0000-0000-000000000001', 'operador_demo', 'SELECT', 'dbgs_schema.conjuntos_datos', 'Consulta masiva sobre datasets confidenciales', 'DENEGADO', '10.0.0.15', NOW() - INTERVAL '1 hour'),
('a9000000-0000-0000-0000-000000000003', 'u2000000-0000-0000-0000-000000000003', 'tecnico_soporte', 'UPDATE', 'dbgs_schema.catalogos', 'Actualización de parámetros operativos', 'EXITOSO', '10.0.0.88', NOW() - INTERVAL '10 minutes')
ON CONFLICT (id) DO NOTHING;