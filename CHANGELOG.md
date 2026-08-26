# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

## [Unreleased]

### Added
- `seguridad`: CRUD completo de usuarios vía API REST/gRPC: `CrearUsuario`, `ListarUsuarios`, `ObtenerUsuario`, `ActualizarUsuario`, `EliminarUsuario`. Los usuarios se asocian a roles existentes.
- `seguridad`: CRUD completo de roles vía API REST/gRPC: `CrearRol`, `ListarRoles`, `ObtenerRol`, `ActualizarRol`, `EliminarRol`, `VincularPermisoRol`, `DesvincularPermisoRol`, `ListarPermisosRol`.
- `seguridad`: los roles operativos (DBA, DEVELOPER, AUDITOR, ANALYST, SERVICE_ACCOUNT) ya no se definen en el repositorio; el Administrador de Plataforma los crea a través de la API (`POST /v1/seguridad/roles` + `POST /v1/seguridad/roles/{id}/permisos`).
- `colecciones`: edicion completa de tablas dinamicas con 5 operaciones: agregar, renombrar, cambiar tipo, eliminar columnas y renombrar tabla.
- `colecciones`: mensajes `RenombrarColumnaProto` y `CambiarTipoColumnaProto` en el proto para las operaciones de edición.
- `colecciones`: función `GenerarSQLRenombrarColumna` en el ddl generator.
- `colecciones`: función `GenerarSQLCambiarTipoColumna` en el ddl generator.
- `colecciones`: función `GenerarSQUEliminarColumna` en el ddl generator.
- `colecciones`: función `GenerarSQLRenombrarTabla` en el ddl generator.
- `colecciones`: método `RenombrarMetadatos` en el repositorio para actualizar nombre lógico y físico en el diccionario.
- `colecciones`: RBAC con permisos `colecciones:actualizar` y `colecciones:eliminar` a nivel de use case; solo `ADMIN_PLATFORM` posee estos permisos.
- `colecciones`: reactivada la vinculación automática del trigger forense de auditoría en toda tabla dinámica creada por el motor DDL.
- `auditoria`: migración `000008` con el trigger `trg_prohibir_modificacion` que bloquea `UPDATE`, `DELETE` y `TRUNCATE` sobre la bitácora a nivel de motor.
- `docs`: actualización de README con estructura, servicios, endpoints y notas de los ocho dominios.

### Fixed
- `seguridad`: campo `description` en JSON de roles ahora acepta el nombre en inglés (`description` en vez de `descripcion`); alinear el proto con la API REST.
- `seguridad`: endpoint `DELETE /v1/seguridad/roles/{rol_id}/permisos/{permiso_codigo}` ahora acepta el código del permiso (ej. `"datasets:leer"`) en vez del UUID interno; consistente con `VincularPermisoRol`.
- `colecciones`: corregido `DROP TABLE` en `EliminarColeccion` — `nombre_fisico` ya contiene el schema, se eliminó la concatenación redundante que generaba un nombre inválido.
- `auditoria`: eliminado el endpoint público `RegistrarEvento` (`POST /v1/auditoria/eventos`) para impedir la falsificación de trazas; la bitácora ahora es de solo lectura desde la API.
- `auditoria`: corregido `ListarEventos` — cast explícito `::timestamptz` en parámetros de fecha (PostgreSQL no infería el tipo de un NULL enviado por lib/pq) y `COALESCE` para columnas nulas (`usuario_id`, `detalles`); el endpoint `GET /v1/auditoria/eventos` nunca había funcionado.
- `db`: migraciones `000002` y `000003` renombradas con sufijo `.up.sql`; golang-migrate las ignoraba silenciosamente, por lo que la función `fn_auditar_cambios()` (base de toda la auditoría forense) no existía en entornos migrados.
- `tests`: actualizados los stubs de seguridad, integración y datasets desalineados con los contratos vigentes; `make test` y `go vet` vuelven a pasar.
- `respaldo`: implementación de gestión de backups, observabilidad y pruebas.
- `integracion`: servicio de clientes, interceptor gRPC e implementación de tests.
- `auditoria`: pruebas unitarias para el uso de eventos de auditoría.
- `auditoria`: mejoras en el filtrado y la persistencia de eventos de auditoría.
- `catalogos/datasets`: casos de uso, repositorios PostgreSQL y tests unitarios.
- `seguridad`: autenticación JWT, hashing BCrypt y repositorio de seguridad.

### Changed
- Estructura inicial de proyecto y configuración base.

### Initial release
- Inicialización del proyecto DBGS Soberano Backend con arquitectura de carpetas y módulos base.
