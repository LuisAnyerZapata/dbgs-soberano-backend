package postgres

import (
	"context"
	"database/sql"
	"errors"

	"DBGS_SOBERANO_BACKEND/internal/domain/entity"
	"DBGS_SOBERANO_BACKEND/internal/domain/repository"
)

type catalogoPostgresRepository struct {
	db *sql.DB
}

// NewCatalogoPostgresRepository instancia el adaptador de almacenamiento de catálogos
func NewCatalogoPostgresRepository(db *sql.DB) repository.CatalogoRepository {
	return &catalogoPostgresRepository{db: db}
}

func (r *catalogoPostgresRepository) ObtenerPorID(ctx context.Context, id string) (*entity.Catalogo, error) {
	query := `
		SELECT id, codigo, nombre, descripcion, estado, created_at, created_by, updated_at, updated_by
		FROM dbgs_schema.catalogos
		WHERE id = $1
	`
	row := r.db.QueryRowContext(ctx, query, id)

	var c entity.Catalogo
	err := row.Scan(&c.ID, &c.Codigo, &c.Nombre, &c.Descripcion, &c.Estado, &c.CreatedAt, &c.CreatedBy, &c.UpdatedAt, &c.UpdatedBy)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, entity.ErrEntidadNoEncontrada
		}
		return nil, entity.ErrErrorInterno
	}

	return &c, nil
}

func (r *catalogoPostgresRepository) ObtenerPorCodigo(ctx context.Context, codigo string) (*entity.Catalogo, error) {
	query := `
		SELECT id, codigo, nombre, descripcion, estado, created_at, created_by, updated_at, updated_by
		FROM dbgs_schema.catalogos
		WHERE codigo = $1
	`
	row := r.db.QueryRowContext(ctx, query, codigo)

	var c entity.Catalogo
	err := row.Scan(&c.ID, &c.Codigo, &c.Nombre, &c.Descripcion, &c.Estado, &c.CreatedAt, &c.CreatedBy, &c.UpdatedAt, &c.UpdatedBy)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, entity.ErrEntidadNoEncontrada
		}
		return nil, entity.ErrErrorInterno
	}

	return &c, nil
}

func (r *catalogoPostgresRepository) Listar(ctx context.Context, soloActivos bool, limite, offset int) ([]entity.Catalogo, int64, error) {
	countQuery := `SELECT COUNT(*) FROM dbgs_schema.catalogos WHERE ($1 = false OR estado = true)`
	var total int64
	if err := r.db.QueryRowContext(ctx, countQuery, soloActivos).Scan(&total); err != nil {
		return nil, 0, entity.ErrErrorInterno
	}

	query := `
		SELECT id, codigo, nombre, descripcion, estado, created_at, created_by, updated_at, updated_by
		FROM dbgs_schema.catalogos
		WHERE ($1 = false OR estado = true)
		ORDER BY codigo ASC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.QueryContext(ctx, query, soloActivos, limite, offset)
	if err != nil {
		return nil, 0, entity.ErrErrorInterno
	}
	defer rows.Close()

	var result []entity.Catalogo
	for rows.Next() {
		var c entity.Catalogo
		if err := rows.Scan(&c.ID, &c.Codigo, &c.Nombre, &c.Descripcion, &c.Estado, &c.CreatedAt, &c.CreatedBy, &c.UpdatedAt, &c.UpdatedBy); err != nil {
			return nil, 0, entity.ErrErrorInterno
		}
		result = append(result, c)
	}

	return result, total, nil
}

func (r *catalogoPostgresRepository) Guardar(ctx context.Context, catalogo *entity.Catalogo) error {
	query := `
		INSERT INTO dbgs_schema.catalogos (id, codigo, nombre, descripcion, estado, created_at, created_by, updated_at, updated_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	_, err := r.db.ExecContext(ctx, query,
		catalogo.ID, catalogo.Codigo, catalogo.Nombre, catalogo.Descripcion,
		catalogo.Estado, catalogo.CreatedAt, catalogo.CreatedBy, catalogo.UpdatedAt, catalogo.UpdatedBy,
	)
	if err != nil {
		return entity.ErrErrorInterno
	}
	return nil
}

func (r *catalogoPostgresRepository) ActualizarEstado(ctx context.Context, id string, estado bool, usuarioModificador string) error {
	query := `
		UPDATE dbgs_schema.catalogos
		SET estado = $1, updated_by = $2, updated_at = NOW()
		WHERE id = $3
	`
	res, err := r.db.ExecContext(ctx, query, estado, usuarioModificador, id)
	if err != nil {
		return entity.ErrErrorInterno
	}

	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		return entity.ErrEntidadNoEncontrada
	}

	return nil
}