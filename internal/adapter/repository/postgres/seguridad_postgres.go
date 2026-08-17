package postgres

import (
	"context"
	"database/sql"
	"errors"

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

	query := `
		SELECT id, username, email, password_hash, rol_id, es_tecnico, estado, created_at, updated_at
		FROM dbgs_schema.usuarios
		WHERE username = $1
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