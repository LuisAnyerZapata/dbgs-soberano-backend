-- Seed: 01_catalogos_referencia.sql
-- Descripción: Carga de catálogos maestro iniciales y valores de referencia obligatorios.

SET search_path TO dbgs_schema, public;

-- 1. Catálogos Base de Dominio
INSERT INTO dbgs_schema.catalogos (id, codigo, nombre, descripcion, estado, created_by) VALUES
('c1000000-0000-0000-0000-000000000001', 'CAT_CLASIFICACION_SEG', 'Clasificación de Seguridad de Datos', 'Niveles de confidencialidad y restricciones de acceso a datos', true, 'sistema_seed'),
('c1000000-0000-0000-0000-000000000002', 'CAT_TIPO_FUENTE', 'Tipos de Fuentes de Datos', 'Formatos y orígenes de almacenamiento de información', true, 'sistema_seed'),
('c1000000-0000-0000-0000-000000000003', 'CAT_ESTADO_OPERATIVO', 'Estados Operativos de Nodos', 'Condición de conectividad e integración de servicios', true, 'sistema_seed'),
('c1000000-0000-0000-0000-000000000004', 'CAT_ORGANISMOS_ESTADO', 'Organismos de la Administración Pública', 'Entidades públicas integradas en la red de datos', true, 'sistema_seed')
ON CONFLICT (codigo) DO UPDATE 
SET nombre = EXCLUDED.nombre, 
    descripcion = EXCLUDED.descripcion, 
    updated_at = CURRENT_TIMESTAMP;

-- 2. Fuentes de Datos Primarias del Sistema
INSERT INTO dbgs_schema.fuentes_datos (id, nombre, descripcion, estado, created_by) VALUES
('f1000000-0000-0000-0000-000000000001', 'Registro Civil y Cédulas', 'Base de datos de identificación ciudadana', true, 'sistema_seed'),
('f1000000-0000-0000-0000-000000000002', 'Sistema Integrado de Salud Pública', 'Consolidado nacional de atenciones asistenciales', true, 'sistema_seed'),
('f1000000-0000-0000-0000-000000000003', 'Catastro y Bienes Nacionales', 'Registro público patrimonial e inmobiliario', true, 'sistema_seed')
ON CONFLICT (id) DO NOTHING;