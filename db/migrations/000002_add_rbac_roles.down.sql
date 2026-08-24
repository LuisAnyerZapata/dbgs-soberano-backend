-- Reversión de la migración 000002: elimina las estructuras RBAC complementarias.
DROP VIEW IF EXISTS dbgs_schema.v_usuarios_permisos;
DROP TABLE IF EXISTS dbgs_schema.rol_permisos;
DROP TABLE IF EXISTS dbgs_schema.usuario_roles;
DROP TABLE IF EXISTS dbgs_schema.roles_rbac;
