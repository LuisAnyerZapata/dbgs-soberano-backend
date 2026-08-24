# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

## [Unreleased]

### Added
- `auditoria`: migración `000008` con el trigger `trg_prohibir_modificacion` que bloquea `UPDATE`, `DELETE` y `TRUNCATE` sobre la bitácora a nivel de motor.
- `colecciones`: reactivada la vinculación automática del trigger forense de auditoría en toda tabla dinámica creada por el motor DDL.

### Removed
- `auditoria`: eliminado el endpoint público `RegistrarEvento` (`POST /v1/auditoria/eventos`) para impedir la falsificación de trazas; la bitácora ahora es de solo lectura desde la API.

### Fixed
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
