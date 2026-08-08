# DBGS Soberano Backend

## Visión general
DBGS Soberano Backend es una aplicación de servicio gRPC desarrollada en Go para la gestión de auditoría, catálogos, conjuntos de datos, seguridad, integración y respaldos. El proyecto está organizado con una arquitectura hexagonal que separa claramente la lógica de negocio de los adaptadores de entrada y salida.

## Arquitectura
La solución adopta una arquitectura hexagonal con las siguientes responsabilidades:
- `internal/domain`: entidades de dominio, errores y repositorios abstractos.
- `internal/application`: casos de uso y lógica de aplicación.
- `internal/adapter/handler/grpc`: controladores gRPC que actúan como adaptadores de entrada.
- `internal/adapter/repository/postgres`: adaptadores de salida para PostgreSQL.
- `api/proto/v1`: definiciones de Protobuf y servicios gRPC.
- `config`: carga y validación de configuración.

## Prerrequisitos
Antes de instalar y ejecutar el proyecto, asegúrese de contar con lo siguiente:

- Go 1.26 o posterior.
- PostgreSQL 12 o posterior.
- `protoc` (Protocol Buffers compiler).
- `make` para automatizar tareas comunes.
- Acceso a la base de datos PostgreSQL con credenciales válidas.

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
Para iniciar el servidor gRPC, ejecute los siguientes pasos en el orden indicado:

1. Asegúrese de que PostgreSQL esté activo y accesible.
2. Aplique las migraciones de la base de datos:

   ```bash
   make migrate-up
   ```

3. Cargue los datos iniciales y las semillas:

   ```bash
   make seed
   ```

4. Compile el proyecto y ejecute el servidor:

   ```bash
   make run
   ```

El servidor gRPC quedará disponible en el puerto configurado, por defecto `50051`.

## Configuración
La aplicación carga la configuración desde uno de los siguientes orígenes:

- `config/config.json`
- variables de entorno con el prefijo `DBGS_`
- valores por defecto embebidos

Las variables más relevantes son:

- `DBGS_SERVER_HOST`
- `DBGS_SERVER_PORT`
- `DBGS_DATABASE_HOST`
- `DBGS_DATABASE_PORT`
- `DBGS_DATABASE_USER`
- `DBGS_DATABASE_PASSWORD`
- `DBGS_DATABASE_NAME`
- `DBGS_DATABASE_SSL_MODE`
- `DBGS_SECURITY_JWT_SECRET`
- `DBGS_SECURITY_TOKEN_TTL_MINUTES`

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

## Estructura del proyecto
A continuación se detalla la estructura principal del repositorio. Esta organización refleja la arquitectura hexagonal y facilita la separación de responsabilidades.

```text
DBGS_SOBERANO_BACKEND/
├── api/
│   └── proto/
│       └── v1/
│           ├── auditoria_servicio.proto
│           ├── catalogos_servicio.proto
│           ├── datasets_servicio.proto
│           └── seguridad_servicio.proto
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
│   │   └── restore_dbgs.sh
│   ├── migrations/
│   │   ├── 000001_init_schema.up.sql
│   │   ├── 000002_add_rbac_roles.sql
│   │   ├── 000003_audit_triggers.sql
│   │   ├── 000004_add_password_hash_to_usuarios.down.sql
│   │   └── 000004_add_password_hash_to_usuarios.up.sql
│   ├── seeds/
│   │   ├── 01_catalogos_referencia.sql
│   │   ├── 02_datos_prueba_sinteticos.sql
│   │   ├── initial_data.sql
│   │   └── seed_users.sql
│   └── security/
│       └── roles_permisos.sql
├── internal/
│   ├── adapter/
│   │   ├── handler/
│   │   │   └── grpc/
│   │   │       ├── auditoria_handler.go
│   │   │       ├── catalogos_handler.go
│   │   │       ├── datasets_handler.go
│   │   │       ├── seguridad_handler.go
│   │   │       ├── server.go
│   │   │       ├── errors.go
│   │   │       ├── interceptors/
│   │   │       │   ├── auth_interceptor.go
│   │   │       │   └── integracion_interceptor.go
│   │   │       ├── integracion/
│   │   │       │   └── handler.go
│   │   │       └── respaldo/
│   │   │           └── handler.go
│   │   └── repository/
│   │       └── postgres/
│   │           ├── auditoria_postgres.go
│   │           ├── catalogo_postgres.go
│   │           ├── connection.go
│   │           ├── dataset_postgres.go
│   │           └── seguridad_postgres.go
│   ├── application/
│   │   ├── port/
│   │   │   ├── auditoria_port.go
│   │   │   ├── catalogo_port.go
│   │   │   ├── dataset_port.go
│   │   │   ├── integracion_port.go
│   │   │   ├── respaldo_port.go
│   │   │   └── seguridad_port.go
│   │   └── usecase/
│   │       ├── auditoria_usecase.go
│   │       ├── catalogo_usecase.go
│   │       ├── dataset_usecase.go
│   │       ├── integracion_usecase.go
│   │       ├── respaldo_usecase.go
│   │       └── seguridad_usecase.go
│   └── domain/
│       ├── entity/
│       │   ├── auditoria_evento.go
│       │   ├── catalogo.go
│       │   ├── fuente_dato.go
│       │   ├── integracion.go
│       │   ├── respaldo.go
│       │   ├── usuario_rol.go
│       │   └── errors.go
│       ├── errors.go
│       └── repository/
│           ├── auditoria_repository.go
│           ├── catalogo_repository.go
│           ├── dataset_repository.go
│           ├── integracion_repository.go
│           ├── respaldo_repository.go
│           └── seguridad_repository.go
├── Makefile
├── go.mod
└── go.sum
```

> Nota: Los archivos `*.pb.go` y `*_grpc.pb.go` se generan durante la compilación de los archivos `.proto`.

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

