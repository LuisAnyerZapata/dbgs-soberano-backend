# DBGS Soberano Backend

## Visión general
DBGS Soberano Backend es una aplicación de servicio desarrollada en Go para la gestión de auditoría, catálogos, conjuntos de datos, seguridad, integración, colecciones dinámicas y respaldos. Expone simultáneamente dos protocolos sobre los mismos contratos:

- **gRPC nativo** en el puerto `50051` (comunicación interna, móvil y escritorio).
- **REST/JSON** mediante grpc-gateway en el puerto `8080` (aplicaciones web), con CORS habilitado.

Su rasgo distintivo es el **motor soberano de datos dinámicos**: la UI puede crear estructuras de datos nuevas (`Colecciones`) y consumirlas de inmediato (`DatosDinámicos`) sin reiniciar el servidor ni compilar código nuevo, al estilo PocketBase pero con auditoría forense a nivel de motor.

El proyecto está organizado con una arquitectura hexagonal que separa claramente la lógica de negocio de los adaptadores de entrada y salida.

## Arquitectura
La solución adopta una arquitectura hexagonal con las siguientes responsabilidades:
- `internal/domain`: entidades de dominio, errores y repositorios abstractos.
- `internal/application`: casos de uso y lógica de aplicación.
- `internal/adapter/handler/grpc`: controladores gRPC que actúan como adaptadores de entrada.
- `internal/adapter/repository/postgres`: adaptadores de salida para PostgreSQL.
- `internal/adapter/executor`: adaptador que ejecuta procesos del sistema operativo (scripts de respaldo con pg_dump/pg_restore).
- `internal/adapter/handler/grpc/interceptors`: interceptor unificado de autenticación (JWT para humanos + API Keys para máquinas).
- `api/proto/v1`: definiciones de Protobuf, servicios gRPC y anotaciones REST.
- `config`: carga y validación de configuración.

## Prerrequisitos
Antes de instalar y ejecutar el proyecto, asegúrese de contar con lo siguiente:

- Go 1.26 o posterior.
- PostgreSQL 12 o posterior.
- Herramientas cliente de PostgreSQL: `psql`, `pg_dump` y `pg_restore` (requeridas por el dominio de respaldos).
- `protoc` (Protocol Buffers compiler).
- `make` para automatizar tareas comunes.
- Acceso a la base de datos PostgreSQL con credenciales válidas.

> **Windows**: usa la consola **Git Bash** (o MSYS2) para ejecutar los comandos `make`, ya que incluye `make`, `bash`, `sed`, `rm` y `curl`. Asegúrate de que `pg_dump`, `pg_restore` y `psql` estén en el `PATH`. El Makefile detecta Windows y genera el binario `dbgs_soberano_backend.exe`; el motor de respaldos invoca `pg_dump`/`pg_restore` directamente (sin depender de bash) para generar y restaurar los `.dump`.

## Instalación
1. Clone el repositorio y acceda al directorio del proyecto:

   ```bash
   git clone <url-del-repositorio>
   cd DBGS_SOBERANO_BACKEND
   ```

2. Copie el archivo de ejemplo de configuración y actualice los valores según su entorno:

   ```bash
   cp config/config.example.json config/config.json
   ```

3. Configure los valores de base de datos y seguridad en `config/config.json` o mediante variables de entorno con prefijo `DBGS_`.

4. Instale las dependencias de Go:

   ```bash
   go mod download
   go mod tidy
   ```

5. Instale las herramientas de Protobuf requeridas:

   ```bash
   make setup
   ```

## Inicio rápido
Para iniciar los servidores, ejecute los siguientes pasos en el orden indicado:

1. Asegúrese de que PostgreSQL esté activo y accesible.
2. Aplique las migraciones de la base de datos:

   ```bash
   make migrate-up
   ```

3. Cargue los datos iniciales y las semillas (opcional: el sistema puede autoinicializarse mediante el flujo de Bootstrapping):

   ```bash
   make seed
   ```

4. Compile el proyecto y ejecute el servidor:

   ```bash
   make run
   ```

El servidor gRPC quedará disponible en el puerto configurado, por defecto `50051`, y el Gateway REST en el puerto `8080`.

### Primer arranque sin seeds (Bootstrapping)
Si no cargó semillas, el sistema crea su propio superadministrador la primera vez:

1. Consulte si el sistema ya fue inicializado: `GET /v1/auth/setup-status`
2. Cree el usuario raíz (solo funciona una única vez):

   ```bash
   curl -X POST http://localhost:8080/v1/auth/setup \
     -H "Content-Type: application/json" \
     -d '{"username": "superadmin", "password": "SuClaveSegura", "email": "admin@institucion.gob"}'
   ```

La respuesta entrega un JWT listo para consumir el resto de la API.

## Configuración
La aplicación carga la configuración desde uno de los siguientes orígenes:

- `config/config.json`
- variables de entorno con el prefijo `DBGS_`
- valores por defecto embebidos

Las variables más relevantes son:

- `DBGS_SERVER_HOST`
- `DBGS_SERVER_PORT`
- `DBGS_SERVER_ENABLE_REFLECTION`
- `DBGS_DATABASE_HOST`
- `DBGS_DATABASE_PORT`
- `DBGS_DATABASE_USER`
- `DBGS_DATABASE_PASSWORD`
- `DBGS_DATABASE_NAME`
- `DBGS_DATABASE_SSL_MODE`
- `DBGS_SECURITY_JWT_SECRET`
- `DBGS_SECURITY_TOKEN_TTL_MINUTES`
- `DBGS_BACKUP_SCRIPTS_DIR` (ubicación de los scripts de respaldo)
- `DBGS_BACKUP_DUMPS_DIR` (repositorio físico de archivos .dump)
- `DBGS_BACKUP_TIMEOUT_MINUTOS` (presupuesto de tiempo de pg_dump/pg_restore)

## Operaciones comunes
- Compilar el binario:

  ```bash
  make build
  ```

- Ejecutar en modo desarrollo:

  ```bash
  make dev
  ```

- Ejecutar pruebas:

  ```bash
  make test
  ```

- Formatear el código:

  ```bash
  make fmt
  ```

- Generar reporte de cobertura:

  ```bash
  make test-coverage
  ```

## Servicios y endpoints REST
Todos los servicios se consumen indistintamente por gRPC (`:50051`) o REST (`:8080`). Salvo los marcados como públicos, requieren cabecera `Authorization: Bearer <jwt>` o API Key (`x-client-id` + `x-api-token`).

| Servicio | Endpoints REST | Autenticación |
|---|---|---|
| SistemaService | `GET /v1/system/health`, `GET /v1/system/version` | Público |
| SeguridadService | `POST /v1/seguridad/login`, `POST /v1/seguridad/validar-token`, `POST /v1/seguridad/verificar-permiso` | Público (login/validar) |
| Setup (bootstrapping) | `GET /v1/auth/setup-status`, `POST /v1/auth/setup` | Público (solo primer arranque) |
| CatalogosService | CRUD completo en `/v1/catalogos` | JWT / API Key |
| DatasetsService | CRUD en `/v1/datasets` | JWT / API Key |
| AuditoriaService | `GET /v1/auditoria/eventos` (solo lectura) | JWT / API Key |
| ColeccionesService | `POST /v1/colecciones`, `GET /v1/colecciones` (motor DDL dinámico) | JWT / API Key |
| DatosDinamicosService | CRUD genérico en `/v1/data/{nombre_tabla}` (motor soberano) | JWT / API Key |
| RespaldoService | `POST /v1/backups`, `GET /v1/backups[/{id}][/contenido]`, `POST /v1/backups/{id}/restaurar`, `POST /v1/backups/retencion` | JWT / API Key + permisos `respaldo:ejecutar` / `restauracion:ejecutar` |

## Estructura del proyecto
A continuación se detalla la estructura principal del repositorio. Esta organización refleja la arquitectura hexagonal y facilita la separación de responsabilidades.

```text
DBGS_SOBERANO_BACKEND/
├── api/
│   ├── proto/
│   │   ├── third_party/
│   │   │   └── google/
│   │   │       └── api/
│   │   │           ├── annotations.proto
│   │   │           └── http.proto
│   │   └── v1/
│   │       ├── auditoria_servicio.proto
│   │       ├── catalogos_servicio.proto
│   │       ├── colecciones_servicio.proto
│   │       ├── datos_dinamicos_servicio.proto
│   │       ├── datasets_servicio.proto
│   │       ├── respaldo_servicio.proto
│   │       ├── seguridad_servicio.proto
│   │       └── sistema_servicio.proto
├── cmd/
│   └── server/
│       └── main.go
├── config/
│   ├── config.example.json
│   ├── config.json
│   └── config.go
├── db/
│   ├── backup/
│   │   ├── backup_dbgs.sh
│   │   ├── dumps/                      (repositorio de .dump generados en runtime)
│   │   └── restore_dbgs.sh
│   ├── migrations/
│   │   ├── 000001_init_schema.up.sql
│   │   ├── 000002_add_rbac_roles.up.sql
│   │   ├── 000002_add_rbac_roles.down.sql
│   │   ├── 000003_audit_triggers.up.sql
│   │   ├── 000003_audit_triggers.down.sql
│   │   ├── 000004_add_password_hash_to_usuarios.up.sql
│   │   ├── 000004_add_password_hash_to_usuarios.down.sql
│   │   ├── 000005_add_instituciones_table.up.sql
│   │   ├── 000005_add_instituciones_table.down.sql
│   │   ├── 000006_add_integracion_tables.up.sql
│   │   ├── 000006_add_integracion_tables.down.sql
│   │   ├── 000007_add_colecciones_dinamicas.up.sql
│   │   ├── 000008_auditoria_inmutable.up.sql
│   │   ├── 000008_auditoria_inmutable.down.sql
│   │   ├── 000009_add_respaldo_tables.up.sql
│   │   └── 000009_add_respaldo_tables.down.sql
│   ├── seeds/
│   │   ├── 01_catalogos_referencia.sql
│   │   ├── 02_datos_prueba_sinteticos.sql
│   │   ├── initial_data.sql
│   │   └── seed_users.sql
│   └── security/
│       └── roles_permisos.sql
├── internal/
│   ├── adapter/
│   │   ├── executor/
│   │   │   └── respaldo_scripts.go
│   │   ├── handler/
│   │   │   └── grpc/
│   │   │       ├── auditoria_handler.go
│   │   │       ├── catalogos_handler.go
│   │   │       ├── colecciones_handler.go
│   │   │       ├── datos_dinamicos_handler.go
│   │   │       ├── datasets_handler.go
│   │   │       ├── errors.go
│   │   │       ├── respaldo_handler.go
│   │   │       ├── seguridad_handler.go
│   │   │       ├── server.go
│   │   │       ├── sistema_handler.go
│   │   │       ├── interceptors/
│   │   │       │   ├── auth_interceptor.go
│   │   │       │   └── integracion_interceptor.go
│   │   │       └── integracion/
│   │   │           └── handler.go
│   │   └── repository/
│   │       └── postgres/
│   │           ├── auditoria_postgres.go
│   │           ├── catalogo_postgres.go
│   │           ├── coleccion_dinamica_postgres.go
│   │           ├── connection.go
│   │           ├── datos_dinamicos_postgres.go
│   │           ├── dataset_postgres.go
│   │           ├── integracion_postgres.go
│   │           ├── institucion_postgres.go
│   │           ├── respaldo_postgres.go
│   │           └── seguridad_postgres.go
│   ├── application/
│   │   ├── port/
│   │   │   ├── auditoria_port.go
│   │   │   ├── catalogo_port.go
│   │   │   ├── coleccion_port.go
│   │   │   ├── datos_dinamicos_port.go
│   │   │   ├── dataset_port.go
│   │   │   ├── integracion_port.go
│   │   │   ├── respaldo_port.go
│   │   │   ├── seguridad_port.go
│   │   │   └── sistema_port.go
│   │   └── usecase/
│   │       ├── auditoria_usecase.go
│   │       ├── catalogo_usecase.go
│   │       ├── coleccion_usecase.go
│   │       ├── datos_dinamicos_usecase.go
│   │       ├── dataset_usecase.go
│   │       ├── ddl_generator.go
│   │       ├── integracion_usecase.go
│   │       ├── respaldo_usecase.go
│   │       ├── seguridad_usecase.go
│   │       └── sistema_usecase.go
│   └── domain/
│       ├── entity/
│       │   ├── auditoria_evento.go
│       │   ├── catalogo.go
│       │   ├── coleccion_dinamica.go
│       │   ├── fuente_dato.go
│       │   ├── institucion.go
│       │   ├── integracion.go
│       │   ├── respaldo.go
│       │   ├── sistema.go
│       │   ├── usuario_rol.go
│       │   └── errors.go
│       ├── errors.go
│       └── repository/
│           ├── auditoria_repository.go
│           ├── catalogo_repository.go
│           ├── coleccion_dinamica_repository.go
│           ├── datos_dinamicos_repository.go
│           ├── dataset_repository.go
│           ├── integracion_repository.go
│           ├── institucion_repository.go
│           ├── respaldo_repository.go
│           └── seguridad_repository.go
├── Makefile
├── go.mod
└── go.sum
```

> Nota: Los archivos `*.pb.go`, `*_grpc.pb.go` y `*.pb.gw.go` se generan durante la compilación de los archivos `.proto` mediante `make proto`. Las dependencias de terceros (`third_party`) se descargan automáticamente la primera vez.

### Descripción de las capas
- `api/proto/v1`: definición de contratos gRPC y mensajes de intercambio.
- `cmd/server`: punto de entrada del servicio.
- `config`: archivos y lógica para carga de parámetros de configuración.
- `internal/adapter`: adaptadores de entrada/salida, incluyendo gRPC y persistencia.
- `internal/application`: casos de uso que orquestan la lógica del dominio.
- `internal/domain`: modelo del dominio, entidades y repositorios abstractos.
- `db`: scripts de migración, respaldo y datos iniciales.

## Notas adicionales
- El proyecto utiliza gRPC Reflection para facilitar la inspección de servicios durante el desarrollo.
- La configuración sensible, como `jwt_secret`, debe manejarse con cuidado en entornos de producción.
- Las migraciones de seguridad incluyen el almacenamiento de hashes BCrypt para usuarios, y las semillas iniciales pueden cargar usuarios base mediante `db/seeds/seed_users.sql`.
- Si utiliza `make restore`, indique el archivo de respaldo con `BACKUP_FILE`.
- Los respaldos también pueden gestionarse por API (dominio de Respaldos): creación asíncrona con pg_dump, descarga del `.dump`, restauración confirmada vía pg_restore y política de retención. Cada operación queda registrada en la bitácora de la base de datos.
- La bitácora de auditoría es inmutable a nivel de motor: un trigger bloquea `UPDATE`, `DELETE` y `TRUNCATE` sobre `auditoria_eventos`; solo se permite insertar eventos.
- Toda tabla dinámica creada mediante `ColeccionesService` nace con columnas soberanas de trazabilidad (`created_by`, `created_at`, etc.) y con su trigger de auditoría ya vinculado.

