# Matriz RBAC — Roles, Permisos y API

## Roles del Sistema

| Rol | Descripción | Permisos Asignados |
|-----|-------------|-------------------|
| `ADMIN_PLATFORM` | Superadministrador con acceso total | Todos (12) |
| `DBA` | Administrador de base de datos | `respaldo:ejecutar`, `restauracion:ejecutar`, `usuarios:leer`, `auditoria:leer` |
| `DEVELOPER` | Desarrollador institucional | `catalogos:leer`, `catalogos:escribir`, `datasets:leer`, `colecciones:crear`, `colecciones:leer`, `colecciones:actualizar` |
| `AUDITOR` | Responsable de seguridad y auditoría | `auditoria:leer`, `catalogos:leer` |
| `ANALYST` | Analista autorizado | `catalogos:leer`, `datasets:leer`, `colecciones:leer` |
| `SERVICE_ACCOUNT` | Cuenta de servicio | `catalogos:leer`, `datasets:leer` |

---

## Matriz de Permisos por Operación

| Operación | Permiso Requerido | Rol que lo Posee |
|-----------|-------------------|-------------------|
| Crear catálogo | `catalogos:escribir` | ADMIN_PLATFORM, DEVELOPER |
| Leer catálogos | `catalogos:leer` | ADMIN_PLATFORM, DEVELOPER, AUDITOR, ANALYST, SERVICE_ACCOUNT |
| Leer datasets | `datasets:leer` | ADMIN_PLATFORM, DEVELOPER, ANALYST, SERVICE_ACCOUNT |
| Crear colección | `colecciones:crear` | ADMIN_PLATFORM, DEVELOPER |
| Leer colecciones | `colecciones:leer` | ADMIN_PLATFORM, DEVELOPER, AUDITOR, ANALYST |
| Actualizar colección | `colecciones:actualizar` | ADMIN_PLATFORM, DEVELOPER |
| Eliminar colección | `colecciones:eliminar` | ADMIN_PLATFORM |
| Registrar evento auditoría | `auditoria:leer` | ADMIN_PLATFORM, DBA, AUDITOR |
| Ejecutar respaldo | `respaldo:ejecutar` | ADMIN_PLATFORM, DBA |
| Ejecutar restauración | `restauracion:ejecutar` | ADMIN_PLATFORM, DBA |
| Listar usuarios | `usuarios:leer` | ADMIN_PLATFORM, DBA |
| Admin usuarios/roles | `usuarios:admin` | ADMIN_PLATFORM |

---

## Códigos de Permisos Disponibles

| Código | Descripción |
|--------|-------------|
| `catalogos:leer` | Consultar catálogos |
| `catalogos:escribir` | Crear/modificar catálogos |
| `datasets:leer` | Consultar conjuntos de datos |
| `colecciones:crear` | Crear colecciones dinámicas |
| `colecciones:leer` | Consultar colecciones |
| `colecciones:actualizar` | Modificar estructura de colecciones |
| `colecciones:eliminar` | Eliminar colecciones |
| `auditoria:leer` | Consultar bitácora de auditoría |
| `respaldo:ejecutar` | Crear respaldos |
| `restauracion:ejecutar` | Restaurar desde respaldo |
| `usuarios:leer` | Consultar usuarios |
| `usuarios:admin` | Administrar usuarios y roles |

---

## Autenticación

Todos los endpoints (excepto `/v1/auth/setup-status`, `/v1/auth/setup` y `/v1/seguridad/login`) requieren:

```
Authorization: Bearer <access_token>
```

---

## API — Endpoints de Seguridad

### Autenticación

| Operación | Método | Ruta | Body |
|-----------|--------|------|------|
| Estado setup | `GET` | `/v1/auth/setup-status` | — |
| Setup inicial | `POST` | `/v1/auth/setup` | `{"username":"admin","password":"Admin123","email":"admin@gob.ve"}` |
| Login | `POST` | `/v1/seguridad/login` | `{"username":"admin","password":"Admin123"}` |
| Validar token | `POST` | `/v1/seguridad/validar-token` | `{"token":"eyJhbG..."}` |
| Verificar permiso | `POST` | `/v1/seguridad/verificar-permiso` | `{"usuario_id":"UUID","recurso":"catalogos","accion":"leer"}` |

### CRUD de Roles

| Operación | Método | Ruta | Body |
|-----------|--------|------|------|
| Crear rol | `POST` | `/v1/seguridad/roles` | `{"nombre":"DBA","description":"Administrador de BD"}` |
| Listar roles | `GET` | `/v1/seguridad/roles` | — |
| Obtener rol | `GET` | `/v1/seguridad/roles/{id}` | — |
| Actualizar rol | `PUT` | `/v1/seguridad/roles/{id}` | `{"nombre":"DBASenior","description":"DBA Senior"}` |
| Eliminar rol | `DELETE` | `/v1/seguridad/roles/{id}` | — |

### Permisos de Roles

| Operación | Método | Ruta | Body |
|-----------|--------|------|------|
| Vincular permisos | `POST` | `/v1/seguridad/roles/{rol_id}/permisos` | `{"permisos":["respaldo:ejecutar","auditoria:leer"]}` |
| Listar permisos | `GET` | `/v1/seguridad/roles/{rol_id}/permisos` | — |
| Desvincular permiso | `DELETE` | `/v1/seguridad/roles/{rol_id}/permisos/{permiso_codigo}` | — |

### CRUD de Usuarios

| Operación | Método | Ruta | Body |
|-----------|--------|------|------|
| Crear usuario | `POST` | `/v1/seguridad/usuarios` | `{"username":"dev","email":"dev@test.com","password":"Dev12345","rol_id":"UUID","es_tecnico":false}` |
| Listar usuarios | `GET` | `/v1/seguridad/usuarios` | — |
| Obtener usuario | `GET` | `/v1/seguridad/usuarios/{id}` | — |
| Actualizar usuario | `PUT` | `/v1/seguridad/usuarios/{id}` | `{"email":"nuevo@test.com","rol_id":"UUID","es_tecnico":false,"estado":true}` |
| Eliminar usuario | `DELETE` | `/v1/seguridad/usuarios/{id}` | — |

---

## Matriz de Usuarios

| Username | Email | Rol | Tipo | Estado |
|----------|-------|-----|------|--------|
| `superadmin` | `admin@gob.ve` | ADMIN_PLATFORM | Técnico | Activo |
| *(crear vía API)* | *(definir)* | *(asignar)* | *(definir)* | Activo |

### Ejemplo de creación de usuario completo

```bash
# 1. Login
curl -X POST http://localhost:8080/v1/seguridad/login \
  -H 'Content-Type: application/json' \
  -d '{"username":"superadmin","password":"Secret123"}'

# 2. Obtener rol_id
curl http://localhost:8080/v1/seguridad/roles \
  -H "Authorization: Bearer <token>"

# 3. Crear usuario
curl -X POST http://localhost:8080/v1/seguridad/usuarios \
  -H "Authorization: Bearer <token>" \
  -H 'Content-Type: application/json' \
  -d '{"username":"dev_user","email":"dev@test.com","password":"Dev12345","rol_id":"<rol_id>","es_tecnico":false}'

# 4. Vincular permisos al rol
curl -X POST http://localhost:8080/v1/seguridad/roles/<rol_id>/permisos \
  -H "Authorization: Bearer <token>" \
  -H 'Content-Type: application/json' \
  -d '{"permisos":["catalogos:leer","datasets:leer"]}'

# 5. Login como el nuevo usuario
curl -X POST http://localhost:8080/v1/seguridad/login \
  -H 'Content-Type: application/json' \
  -d '{"username":"dev_user","password":"Dev12345"}'
```
