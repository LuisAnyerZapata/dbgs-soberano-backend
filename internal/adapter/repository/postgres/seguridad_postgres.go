package postgres

import (
	"context"
	"database/sql"
	"errors"

	"DBGS_SOBERANO_BACKEND/internal/domain/entity"
	"DBGS_SOBERANO_BACKEND/internal/domain/repository"
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