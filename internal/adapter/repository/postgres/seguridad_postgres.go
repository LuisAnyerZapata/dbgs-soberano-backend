package postgres

import (
	"context"
	"database/sql"
	"errors"
	"log"

	"DBGS_SOBERANO_BACKEND/internal/domain/entity"
	"DBGS_SOBERANO_BACKEND/internal/domain/repository"
	"golang.org/x/crypto/bcrypt"
)

type seguridadPostgresRepository struct {
	db *sql.DB
}

func NewSeguridadPostgresRepository(db *sql.DB) repository.SeguridadRepository {
	return &seguridadPostgresRepository{db: db}
}

func (r *seguridadPostgresRepository) ObtenerUsuarioPorUsername(ctx context.Context, username string) (*entity.Usuario, error) {
	query := `
		SELECT id, username, email, rol_id, es_tecnico, estado, created_at, updated_at
		FROM dbgs_schema.usuarios
		WHERE username = $1
	`
	row := r.db.QueryRowContext(ctx, query, username)

	var u entity.Usuario
	err := row.Scan(&u.ID, &u.Username, &u.Email, &u.RolID, &u.EsTecnico, &u.Estado, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, entity.ErrEntidadNoEncontrada
		}
		return nil, entity.ErrErrorInterno
	}
	return &u, nil
}

func verificarCredencial(passwordHash, password string) bool {
	if passwordHash == "" || password == "" {
		return false
	}
	return bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password)) == nil
}

func (r *seguridadPostgresRepository) AutenticarUsuario(ctx context.Context, username, password string) (*entity.Usuario, error) {
	if username == "" || password == "" {
		return nil, entity.ErrDatosInvalidos
	}

	// El login se realiza únicamente por correo electrónico (insensible a mayúsculas).
	query := `
		SELECT id, username, email, password_hash, rol_id, es_tecnico, estado, created_at, updated_at
		FROM dbgs_schema.usuarios
		WHERE LOWER(email) = LOWER($1)
	`
	row := r.db.QueryRowContext(ctx, query, username)

	var u entity.Usuario
	var passwordHash string
	err := row.Scan(&u.ID, &u.Username, &u.Email, &passwordHash, &u.RolID, &u.EsTecnico, &u.Estado, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, entity.ErrEntidadNoEncontrada
		}
		return nil, entity.ErrErrorInterno
	}

	if !verificarCredencial(passwordHash, password) {
		return nil, entity.ErrAccesoNoAutorizado
	}

	return &u, nil
}

func (r *seguridadPostgresRepository) ObtenerRolPorID(ctx context.Context, rolID string) (*entity.Rol, error) {
	query := `SELECT id, nombre, descripcion, estado FROM dbgs_schema.roles WHERE id = $1`
	row := r.db.QueryRowContext(ctx, query, rolID)

	var rol entity.Rol
	if err := row.Scan(&rol.ID, &rol.Nombre, &rol.Descripcion, &rol.Estado); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, entity.ErrEntidadNoEncontrada
		}
		return nil, entity.ErrErrorInterno
	}
	return &rol, nil
}

func (r *seguridadPostgresRepository) ValidarPermiso(ctx context.Context, rolID, permiso string) (bool, error) {
	query := `
		SELECT EXISTS (
			SELECT 1 FROM dbgs_schema.roles_permisos rp
			JOIN dbgs_schema.permisos p ON rp.permiso_id = p.id
			WHERE rp.rol_id = $1 AND p.codigo = $2
		)
	`
	var tienePermiso bool
	err := r.db.QueryRowContext(ctx, query, rolID, permiso).Scan(&tienePermiso)
	if err != nil {
		return false, entity.ErrErrorInterno
	}
	return tienePermiso, nil
}

// ContarSuperAdmins ejecuta la consulta para determinar si el sistema ya fue configurado.
// NOTA: Se asume que el rol de superadministrador en los seeds se llama 'ADMIN_PLATFORM'.
func (r *seguridadPostgresRepository) ContarSuperAdmins(ctx context.Context) (int64, error) {
    query := `
        SELECT COUNT(*) 
        FROM dbgs_schema.usuarios u 
        JOIN dbgs_schema.roles r ON u.rol_id = r.id 
        WHERE r.nombre = 'ADMIN_PLATFORM' AND u.estado = true
    `
    var count int64
    // Se usa QueryRowContext porque esperamos un solo valor escalar
    if err := r.db.QueryRowContext(ctx, query).Scan(&count); err != nil {
        return 0, entity.ErrErrorInterno
    }
    return count, nil
}

// CrearUsuarioAdmin realiza la inserción del primer usuario con privilegios máximos.
// Recibe el hash ya procesado por la capa de aplicación para mantener la responsabilidad de cifrado fuera del adaptador.
func (r *seguridadPostgresRepository) CrearUsuarioAdmin(ctx context.Context, username, email, passwordHash, rolID string) (*entity.Usuario, error) {
    query := `
        INSERT INTO dbgs_schema.usuarios (username, email, password_hash, rol_id, es_tecnico, estado)
        VALUES ($1, $2, $3, $4, true, true)
        RETURNING id, username, email, rol_id, es_tecnico, estado, created_at, updated_at
    `
	
    var u entity.Usuario
    err := r.db.QueryRowContext(ctx, query, username, email, passwordHash, rolID).Scan(
        &u.ID, &u.Username, &u.Email, &u.RolID, &u.EsTecnico, &u.Estado, &u.CreatedAt, &u.UpdatedAt,
    )

    if err != nil {
        // Evitamos exponer el error exacto de base de datos por seguridad (ej. violación de unique)
        return nil, entity.ErrErrorInterno
    }

    return &u, nil
}

// AsegurarRolSuperAdmin utiliza un UPSERT (INSERT ... ON CONFLICT DO NOTHING) para garantizar
// que el rol principal exista sin lanzar errores de duplicidad. Retorna el ID del rol.
func (r *seguridadPostgresRepository) AsegurarRolSuperAdmin(ctx context.Context) (string, error) {
    const rolID = "11111111-1111-1111-1111-111111111111" // UUID soberano fijo para el rol principal
    const rolNombre = "ADMIN_PLATFORM"

    query := `
        INSERT INTO dbgs_schema.roles (id, nombre, descripcion, estado) 
        VALUES ($1, $2, 'Rol de superadministrador con acceso total al sistema', true)
        ON CONFLICT (id) DO NOTHING 
        RETURNING id
    `

    var id string
    err := r.db.QueryRowContext(ctx, query, rolID, rolNombre).Scan(&id)
    
    // Si el ON CONFLICT entró en acción (DO NOTHING), el RETURNING no devuelve filas, 
    // lo que genera un error sql.ErrNoRows. En ese caso, significa que el rol ya existe,
    // así que simplemente retornamos el ID que ya conocemos.
    if errors.Is(err, sql.ErrNoRows) {
        return rolID, nil
    }

    if err != nil {
        return "", entity.ErrErrorInterno
    }

    return id, nil
}

// CrearRol inserta un nuevo rol en la base de datos
func (r *seguridadPostgresRepository) CrearRol(ctx context.Context, nombre, descripcion string) (*entity.Rol, error) {
    query := `
        INSERT INTO dbgs_schema.roles (nombre, descripcion)
        VALUES ($1, $2)
        RETURNING id, nombre, descripcion
    `
    var rol entity.Rol
    if err := r.db.QueryRowContext(ctx, query, nombre, descripcion).Scan(&rol.ID, &rol.Nombre, &rol.Descripcion); err != nil {
        log.Printf("ERROR EN BD (Seguridad.CrearRol): %v", err)
        return nil, entity.ErrErrorInterno
    }
    return &rol, nil
}

// ListarRoles obtiene todos los roles registrados
func (r *seguridadPostgresRepository) ListarRoles(ctx context.Context) ([]entity.Rol, error) {
    query := `SELECT id, nombre, descripcion FROM dbgs_schema.roles ORDER BY nombre`
    rows, err := r.db.QueryContext(ctx, query)
    if err != nil {
        log.Printf("ERROR EN BD (Seguridad.ListarRoles): %v", err)
        return nil, entity.ErrErrorInterno
    }
    defer rows.Close()

    var roles []entity.Rol
    for rows.Next() {
        var rol entity.Rol
        if err := rows.Scan(&rol.ID, &rol.Nombre, &rol.Descripcion); err != nil {
            log.Printf("ERROR EN BD (Seguridad.ListarRoles scan): %v", err)
            return nil, entity.ErrErrorInterno
        }
        roles = append(roles, rol)
    }
    return roles, nil
}

// ObtenerRolPorNombre busca un rol por su nombre
func (r *seguridadPostgresRepository) ObtenerRolPorNombre(ctx context.Context, nombre string) (*entity.Rol, error) {
    query := `SELECT id, nombre, descripcion FROM dbgs_schema.roles WHERE nombre = $1`
    var rol entity.Rol
    if err := r.db.QueryRowContext(ctx, query, nombre).Scan(&rol.ID, &rol.Nombre, &rol.Descripcion); err != nil {
        if err == sql.ErrNoRows {
            return nil, entity.ErrEntidadNoEncontrada
        }
        log.Printf("ERROR EN BD (Seguridad.ObtenerRolPorNombre '%s'): %v", nombre, err)
        return nil, entity.ErrErrorInterno
    }
    return &rol, nil
}

// ActualizarRol modifica nombre y descripción de un rol
func (r *seguridadPostgresRepository) ActualizarRol(ctx context.Context, id, nombre, descripcion string) error {
    query := `
        UPDATE dbgs_schema.roles
        SET nombre = $1, descripcion = $2
        WHERE id = $3
    `
    result, err := r.db.ExecContext(ctx, query, nombre, descripcion, id)
    if err != nil {
        log.Printf("ERROR EN BD (Seguridad.ActualizarRol '%s'): %v", id, err)
        return entity.ErrErrorInterno
    }
    filas, _ := result.RowsAffected()
    if filas == 0 {
        return entity.ErrEntidadNoEncontrada
    }
    return nil
}

// EliminarRol elimina un rol de la base de datos
func (r *seguridadPostgresRepository) EliminarRol(ctx context.Context, id string) error {
    // Primero desvincular todos los permisos
    _, _ = r.db.ExecContext(ctx, `DELETE FROM dbgs_schema.roles_permisos WHERE rol_id = $1`, id)
    // Luego eliminar el rol
    query := `DELETE FROM dbgs_schema.roles WHERE id = $1`
    result, err := r.db.ExecContext(ctx, query, id)
    if err != nil {
        log.Printf("ERROR EN BD (Seguridad.EliminarRol '%s'): %v", id, err)
        return entity.ErrErrorInterno
    }
    filas, _ := result.RowsAffected()
    if filas == 0 {
        return entity.ErrEntidadNoEncontrada
    }
    return nil
}

// VincularPermisos vincula múltiples permisos a un rol
func (r *seguridadPostgresRepository) VincularPermisos(ctx context.Context, rolID string, codigos []string) (int64, error) {
    query := `
        INSERT INTO dbgs_schema.roles_permisos (rol_id, permiso_id)
        SELECT $1, p.id FROM dbgs_schema.permisos p WHERE p.codigo = $2
        ON CONFLICT (rol_id, permiso_id) DO NOTHING
    `
    var vinculados int64
    for _, codigo := range codigos {
        result, err := r.db.ExecContext(ctx, query, rolID, codigo)
        if err != nil {
            log.Printf("ERROR EN BD (Seguridad.VincularPermisos rol=%s permiso=%s): %v", rolID, codigo, err)
            continue
        }
        filas, _ := result.RowsAffected()
        vinculados += filas
    }
    return vinculados, nil
}

// DesvincularPermiso elimina un permiso de un rol por código
func (r *seguridadPostgresRepository) DesvincularPermiso(ctx context.Context, rolID, permisoCodigo string) error {
    query := `
        DELETE FROM dbgs_schema.roles_permisos
        WHERE rol_id = $1
          AND permiso_id = (SELECT id FROM dbgs_schema.permisos WHERE codigo = $2)
    `
    _, err := r.db.ExecContext(ctx, query, rolID, permisoCodigo)
    if err != nil {
        log.Printf("ERROR EN BD (Seguridad.DesvincularPermiso rol=%s codigo=%s): %v", rolID, permisoCodigo, err)
        return entity.ErrErrorInterno
    }
    return nil
}

// ListarPermisosRol obtiene los códigos de permisos de un rol
func (r *seguridadPostgresRepository) ListarPermisosRol(ctx context.Context, rolID string) ([]string, error) {
    query := `
        SELECT p.codigo FROM dbgs_schema.permisos p
        JOIN dbgs_schema.roles_permisos rp ON rp.permiso_id = p.id
        WHERE rp.rol_id = $1
        ORDER BY p.codigo
    `
    rows, err := r.db.QueryContext(ctx, query, rolID)
    if err != nil {
        log.Printf("ERROR EN BD (Seguridad.ListarPermisosRol): %v", err)
        return nil, entity.ErrErrorInterno
    }
    defer rows.Close()

    var permisos []string
    for rows.Next() {
        var codigo string
        if err := rows.Scan(&codigo); err != nil {
            return nil, entity.ErrErrorInterno
        }
        permisos = append(permisos, codigo)
    }
    return permisos, nil
}

// =========================================================================================================
// CRUD DE USUARIOS
// =========================================================================================================

func (r *seguridadPostgresRepository) CrearUsuario(ctx context.Context, username, email, passwordHash, rolID string, esTecnico bool) (*entity.Usuario, error) {
    query := `
        INSERT INTO dbgs_schema.usuarios (username, email, password_hash, rol_id, es_tecnico, estado)
        VALUES ($1, $2, $3, $4, $5, true)
        RETURNING id, username, email, rol_id, es_tecnico, estado, created_at, updated_at
    `
    var u entity.Usuario
    err := r.db.QueryRowContext(ctx, query, username, email, passwordHash, rolID, esTecnico).
        Scan(&u.ID, &u.Username, &u.Email, &u.RolID, &u.EsTecnico, &u.Estado, &u.CreatedAt, &u.UpdatedAt)
    if err != nil {
        log.Printf("ERROR EN BD (Seguridad.CrearUsuario): %v", err)
        return nil, entity.ErrErrorInterno
    }
    return &u, nil
}

func (r *seguridadPostgresRepository) ListarUsuarios(ctx context.Context) ([]entity.Usuario, error) {
    query := `
        SELECT u.id, u.username, u.email, u.rol_id, COALESCE(r.nombre,'') AS rol_nombre,
               u.es_tecnico, u.estado, u.created_at, u.updated_at
        FROM dbgs_schema.usuarios u
        LEFT JOIN dbgs_schema.roles r ON r.id = u.rol_id
        ORDER BY u.created_at DESC
    `
    rows, err := r.db.QueryContext(ctx, query)
    if err != nil {
        log.Printf("ERROR EN BD (Seguridad.ListarUsuarios): %v", err)
        return nil, entity.ErrErrorInterno
    }
    defer rows.Close()

    var usuarios []entity.Usuario
    for rows.Next() {
        var u entity.Usuario
        if err := rows.Scan(&u.ID, &u.Username, &u.Email, &u.RolID, &u.Rol.Nombre,
            &u.EsTecnico, &u.Estado, &u.CreatedAt, &u.UpdatedAt); err != nil {
            return nil, entity.ErrErrorInterno
        }
        usuarios = append(usuarios, u)
    }
    return usuarios, nil
}

func (r *seguridadPostgresRepository) ObtenerUsuarioPorID(ctx context.Context, id string) (*entity.Usuario, error) {
    query := `
        SELECT u.id, u.username, u.email, u.rol_id, COALESCE(r.nombre,'') AS rol_nombre,
               u.es_tecnico, u.estado, u.created_at, u.updated_at
        FROM dbgs_schema.usuarios u
        LEFT JOIN dbgs_schema.roles r ON r.id = u.rol_id
        WHERE u.id = $1
    `
    var u entity.Usuario
    err := r.db.QueryRowContext(ctx, query, id).
        Scan(&u.ID, &u.Username, &u.Email, &u.RolID, &u.Rol.Nombre,
            &u.EsTecnico, &u.Estado, &u.CreatedAt, &u.UpdatedAt)
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, entity.ErrEntidadNoEncontrada
        }
        log.Printf("ERROR EN BD (Seguridad.ObtenerUsuarioPorID): %v", err)
        return nil, entity.ErrErrorInterno
    }
    return &u, nil
}

func (r *seguridadPostgresRepository) ActualizarUsuario(ctx context.Context, id, email, rolID string, esTecnico, estado bool) error {
    query := `
        UPDATE dbgs_schema.usuarios
        SET email = $1, rol_id = $2, es_tecnico = $3, estado = $4, updated_at = NOW()
        WHERE id = $5
    `
    _, err := r.db.ExecContext(ctx, query, email, rolID, esTecnico, estado, id)
    if err != nil {
        log.Printf("ERROR EN BD (Seguridad.ActualizarUsuario): %v", err)
        return entity.ErrErrorInterno
    }
    return nil
}

func (r *seguridadPostgresRepository) EliminarUsuario(ctx context.Context, id string) error {
    query := `DELETE FROM dbgs_schema.usuarios WHERE id = $1`
    _, err := r.db.ExecContext(ctx, query, id)
    if err != nil {
        log.Printf("ERROR EN BD (Seguridad.EliminarUsuario): %v", err)
        return entity.ErrErrorInterno
    }
    return nil
}