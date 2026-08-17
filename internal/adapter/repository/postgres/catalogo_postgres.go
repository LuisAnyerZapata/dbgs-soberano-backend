package postgres

import (
	"context"
	"database/sql"
	"errors"
	"log"

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

func nullableString(value sql.NullString) string {
	if value.Valid {
		return value.String
	}
	return ""
}

func (r *catalogoPostgresRepository) ObtenerPorID(ctx context.Context, id string) (*entity.Catalogo, error) {
	query := `
		SELECT id, codigo, nombre, descripcion, estado, created_at, created_by, updated_at, updated_by
		FROM dbgs_schema.catalogos
		WHERE id = $1
	`
	row := r.db.QueryRowContext(ctx, query, id)

	var c entity.Catalogo
	var createdBy, updatedBy sql.NullString
	err := row.Scan(&c.ID, &c.Codigo, &c.Nombre, &c.Descripcion, &c.Estado, &c.CreatedAt, &createdBy, &c.UpdatedAt, &updatedBy)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, entity.ErrEntidadNoEncontrada
		}
		return nil, entity.ErrErrorInterno
	}

	c.CreatedBy = nullableString(createdBy)
	c.UpdatedBy = nullableString(updatedBy)

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
	var createdBy, updatedBy sql.NullString
	err := row.Scan(&c.ID, &c.Codigo, &c.Nombre, &c.Descripcion, &c.Estado, &c.CreatedAt, &createdBy, &c.UpdatedAt, &updatedBy)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, entity.ErrEntidadNoEncontrada
		}
		return nil, entity.ErrErrorInterno
	}

	c.CreatedBy = nullableString(createdBy)
	c.UpdatedBy = nullableString(updatedBy)

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
		var createdBy, updatedBy sql.NullString
		if err := rows.Scan(&c.ID, &c.Codigo, &c.Nombre, &c.Descripcion, &c.Estado, &c.CreatedAt, &createdBy, &c.UpdatedAt, &updatedBy); err != nil {
			return nil, 0, entity.ErrErrorInterno
		}
		c.CreatedBy = nullableString(createdBy)
		c.UpdatedBy = nullableString(updatedBy)
		result = append(result, c)
	}

	return result, total, nil
}

// Guardar inserta un nuevo catálogo en la base de datos.
// Si no se proporciona un ID, se envía NULL para que PostgreSQL ejecute la función DEFAULT (uuid_generate_v4())
func (r *catalogoPostgresRepository) Guardar(ctx context.Context, catalogo *entity.Catalogo) error {
    query := `
        INSERT INTO dbgs_schema.catalogos (id, codigo, nombre, descripcion, estado, created_at, created_by, updated_at, updated_by)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
    `
    
    _, err := r.db.ExecContext(ctx, query,
        catalogo.ID,           // $1 (Siempre vendrá con UUID desde el Use Case)
        catalogo.Codigo,        // $2
        catalogo.Nombre,        // $3
        catalogo.Descripcion,   // $4
        catalogo.Estado,        // $5
        catalogo.CreatedAt,     // $6
        catalogo.CreatedBy,     // $7
        catalogo.UpdatedAt,     // $8
        catalogo.UpdatedBy,     // $9
    )
    if err != nil {
        log.Printf("ERROR CRÍTICO EN BD (Catalogo.Guardar): %v", err) 
        return entity.ErrErrorInterno
    }
    return nil
}

func (r *catalogoPostgresRepository) Actualizar(ctx context.Context, catalogo *entity.Catalogo) (*entity.Catalogo, error) {
	query := `
		UPDATE dbgs_schema.catalogos
		SET codigo = $1, nombre = $2, descripcion = $3, estado = $4, updated_at = $5, updated_by = $6
		WHERE id = $7
		RETURNING id, codigo, nombre, descripcion, estado, created_at, created_by, updated_at, updated_by
	`

	row := r.db.QueryRowContext(ctx, query,
		catalogo.Codigo, catalogo.Nombre, catalogo.Descripcion,
		catalogo.Estado, catalogo.UpdatedAt, catalogo.UpdatedBy, catalogo.ID,
	)

	var updated entity.Catalogo
	if err := row.Scan(&updated.ID, &updated.Codigo, &updated.Nombre, &updated.Descripcion, &updated.Estado, &updated.CreatedAt, &updated.CreatedBy, &updated.UpdatedAt, &updated.UpdatedBy); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, entity.ErrEntidadNoEncontrada
		}
		return nil, entity.ErrErrorInterno
	}

	return &updated, nil
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
