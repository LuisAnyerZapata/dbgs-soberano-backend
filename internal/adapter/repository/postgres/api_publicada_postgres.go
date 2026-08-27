package postgres

import (
	"context"
	"database/sql"
	"errors"
	"log"

	"DBGS_SOBERANO_BACKEND/internal/domain/entity"
	"DBGS_SOBERANO_BACKEND/internal/domain/repository"
)

type apiPublicadaPostgresRepository struct {
	db *sql.DB
}

func NewApiPublicadaPostgresRepository(db *sql.DB) repository.ApiPublicadaRepository {
	return &apiPublicadaPostgresRepository{db: db}
}

const apiCols = `id, nombre, descripcion, slug, conexion_id, conexion_nombre, esquema, tabla, max_filas, activa, api_key, endpoint, created_at, created_by`

func scanApi(row scanner) (*entity.ApiPublicada, error) {
	var a entity.ApiPublicada
	if err := row.Scan(&a.ID, &a.Name, &a.Description, &a.Slug, &a.ConnectionID, &a.ConnectionName, &a.Schema, &a.Table, &a.MaxRows, &a.Active, &a.APIKey, &a.Endpoint, &a.CreatedAt, &a.CreatedBy); err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *apiPublicadaPostgresRepository) Guardar(ctx context.Context, a *entity.ApiPublicada) error {
	query := `INSERT INTO dbgs_schema.apis_publicadas (id, nombre, descripcion, slug, conexion_id, conexion_nombre, esquema, tabla, max_filas, activa, api_key, endpoint, created_at, created_by)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)`
	if _, err := r.db.ExecContext(ctx, query, a.ID, a.Name, a.Description, a.Slug, a.ConnectionID, a.ConnectionName, a.Schema, a.Table, a.MaxRows, a.Active, a.APIKey, a.Endpoint, a.CreatedAt, a.CreatedBy); err != nil {
		log.Printf("apiPublicadaPostgresRepository.Guardar exec error: %v", err)
		if isUniqueViolation(err) {
			return entity.ErrCodigoDuplicado
		}
		return entity.ErrErrorInterno
	}
	return nil
}

func (r *apiPublicadaPostgresRepository) ObtenerPorID(ctx context.Context, id string) (*entity.ApiPublicada, error) {
	row := r.db.QueryRowContext(ctx, `SELECT `+apiCols+` FROM dbgs_schema.apis_publicadas WHERE id=$1`, id)
	return getApiOrNotFound(row)
}

func (r *apiPublicadaPostgresRepository) Listar(ctx context.Context, limite, offset int) ([]entity.ApiPublicada, int64, error) {
	var total int64
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM dbgs_schema.apis_publicadas`).Scan(&total); err != nil {
		log.Printf("apiPublicadaPostgresRepository.Listar count error: %v", err)
		return nil, 0, entity.ErrErrorInterno
	}
	rows, err := r.db.QueryContext(ctx, `SELECT `+apiCols+` FROM dbgs_schema.apis_publicadas ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limite, offset)
	if err != nil {
		log.Printf("apiPublicadaPostgresRepository.Listar query error: %v", err)
		return nil, 0, entity.ErrErrorInterno
	}
	defer rows.Close()
	var list []entity.ApiPublicada
	for rows.Next() {
		a, err := scanApi(rows)
		if err != nil {
			log.Printf("apiPublicadaPostgresRepository.Listar scan error: %v", err)
			return nil, 0, entity.ErrErrorInterno
		}
		list = append(list, *a)
	}
	return list, total, nil
}

func (r *apiPublicadaPostgresRepository) Actualizar(ctx context.Context, a *entity.ApiPublicada) (*entity.ApiPublicada, error) {
	row := r.db.QueryRowContext(ctx, `UPDATE dbgs_schema.apis_publicadas SET nombre=$1, descripcion=$2, slug=$3, conexion_id=$4, conexion_nombre=$5, esquema=$6, tabla=$7, max_filas=$8, activa=$9, api_key=$10, endpoint=$11
		WHERE id=$12 RETURNING `+apiCols,
		a.Name, a.Description, a.Slug, a.ConnectionID, a.ConnectionName, a.Schema, a.Table, a.MaxRows, a.Active, a.APIKey, a.Endpoint, a.ID)
	return getApiOrNotFound(row)
}

func (r *apiPublicadaPostgresRepository) Eliminar(ctx context.Context, id string) error {
	if _, err := r.db.ExecContext(ctx, `DELETE FROM dbgs_schema.apis_publicadas WHERE id=$1`, id); err != nil {
		log.Printf("apiPublicadaPostgresRepository.Eliminar exec error: %v", err)
		return entity.ErrErrorInterno
	}
	return nil
}

func (r *apiPublicadaPostgresRepository) SlugEnUso(ctx context.Context, slug, excluirID string) (bool, error) {
	var existe bool
	query := `SELECT EXISTS(SELECT 1 FROM dbgs_schema.apis_publicadas WHERE slug=$1 AND ($2='' OR id::text<>$2))`
	if err := r.db.QueryRowContext(ctx, query, slug, excluirID).Scan(&existe); err != nil {
		log.Printf("apiPublicadaPostgresRepository.SlugEnUso query error: %v", err)
		return false, entity.ErrErrorInterno
	}
	return existe, nil
}

func getApiOrNotFound(row *sql.Row) (*entity.ApiPublicada, error) {
	a, err := scanApi(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, entity.ErrEntidadNoEncontrada
		}
		log.Printf("apiPublicadaPostgresRepository scan error: %v", err)
		return nil, entity.ErrErrorInterno
	}
	return a, nil
}
