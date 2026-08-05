-- Crear Rol Aplicativo (Sin acceso superusuario)
DO $$
BEGIN
    IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = 'dbgs_app_role') THEN
        CREATE ROLE dbgs_app_role NOLOGIN;
    END IF;
END
$$;

-- Permisos sobre el esquema
GRANT USAGE ON SCHEMA dbgs_schema TO dbgs_app_role;

-- Permisos CRUD sobre las tablas existentes
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA dbgs_schema TO dbgs_app_role;
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA dbgs_schema TO dbgs_app_role;

-- Asegurar permisos para futuras tablas
ALTER DEFAULT PRIVILEGES IN SCHEMA dbgs_schema GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO dbgs_app_role;

-- Restricción inmutable en Auditoría: Solo INSERT y SELECT (No UPDATE ni DELETE)
REVOKE UPDATE, DELETE ON dbgs_schema.auditoria_eventos FROM dbgs_app_role;