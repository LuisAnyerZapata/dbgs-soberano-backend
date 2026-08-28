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
- `domain`: sistema profesional de validación de datos y manejo de excepciones:
  - Tipo `AppError` estructurado con código, campo, mensaje interno y mensaje seguro para el cliente.
  - Códigos de error tipados (`ErrorCode`) con mapeo a HTTP y gRPC.
  - Validadores reutilizables: `ValidateRequired`, `ValidateEmail`, `ValidateUUID`, `ValidatePassword`, `ValidateEnum`, `ValidateRange`, `ValidateCodigo`, `ValidateName`, `ValidateMinLength`, `ValidateMaxLength`, `ValidateNoSpaces`.
  - `ValidationError` para acumulación de múltiples errores de validación.
  - `PostgresErrorClassifier` para detectar códigos PostgreSQL (23505, 23503, 23502, 22P02, etc.) y mapearlos a errores de dominio.
  - Handler `mapDomainErrorToGRPC` actualizado para usar `errors.As()` y sanitizar mensajes (nunca expone detalles internos).
  - JWT claims parsing con type assertions seguras (previene panic).
  - Context keys tipados para evitar colisiones.

### Fixed
- `explorador` (datos dinámicos): el método `Actualizar` del repositorio solo construía un `SELECT` y nunca ejecutaba el cambio en base de datos, por lo que la edición de registros desde el Explorador de Datos (`PUT /v1/data/{tabla}/{id}`) devolvía siempre «Error interno del servidor» y no persistía. Ahora ejecuta un `UPDATE` real (incluye `updated_by` y `updated_at = now()`) y recupera el registro actualizado con un `SELECT row_to_json` (mismo patrón que `ObtenerPorID`).
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
- Soporte para ejecutar el backend en Windows: el Makefile es ahora portable (detección de `python3`/`python` y extensión `.exe` del binario en Windows; se ejecuta dentro de Git Bash).
- El motor de respaldos es multiplataforma: en Unix sigue delegando en los scripts bash de `db/backup`, y en Windows invoca `pg_dump`/`pg_restore` directamente (sin depender de bash), replicando los mismos flags de los scripts.
- Los scripts `backup_dbgs.sh` / `restore_dbgs.sh` ahora usan `python3` con fallback a `python` para leer `config.json`.
- Estructura inicial de proyecto y configuración base.

### Initial release
- Inicialización del proyecto DBGS Soberano Backend con arquitectura de carpetas y módulos base.
