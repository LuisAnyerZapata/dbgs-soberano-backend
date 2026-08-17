package postgres

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"strings"

	"DBGS_SOBERANO_BACKEND/internal/domain/entity"
	"DBGS_SOBERANO_BACKEND/internal/domain/repository"
)

type datasetPostgresRepository struct {
	db *sql.DB
}

func NewDatasetPostgresRepository(db *sql.DB) repository.DatasetRepository {
	return &datasetPostgresRepository{db: db}
}

func (r *datasetPostgresRepository) ObtenerFuentePorID(ctx context.Context, id string) (*entity.FuenteDato, error) {
	query := `SELECT id, nombre, descripcion, estado, created_at, created_by FROM dbgs_schema.fuentes_datos WHERE id = $1`
	row := r.db.QueryRowContext(ctx, query, id)

	var f entity.FuenteDato
	if err := row.Scan(&f.ID, &f.Nombre, &f.Descripcion, &f.Estado, &f.CreatedAt, &f.CreatedBy); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, entity.ErrEntidadNoEncontrada
		}
		log.Printf("datasetPostgresRepository.ObtenerFuentePorID scan error: %v", err)
		return nil, entity.ErrErrorInterno
	}
	return &f, nil
}

func (r *datasetPostgresRepository) ListarFuentes(ctx context.Context) ([]entity.FuenteDato, error) {
	query := `SELECT id, nombre, descripcion, estado, created_at, created_by FROM dbgs_schema.fuentes_datos WHERE estado = true`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		log.Printf("datasetPostgresRepository.ListarFuentes query error: %v", err)
		return nil, entity.ErrErrorInterno
	}
	defer rows.Close()

	var fuentes []entity.FuenteDato
	for rows.Next() {
		var f entity.FuenteDato
		if err := rows.Scan(&f.ID, &f.Nombre, &f.Descripcion, &f.Estado, &f.CreatedAt, &f.CreatedBy); err != nil {
			log.Printf("datasetPostgresRepository.ListarFuentes scan error: %v", err)
			return nil, entity.ErrErrorInterno
		}
		fuentes = append(fuentes, f)
	}
	return fuentes, nil
}

func (r *datasetPostgresRepository) ObtenerDatasetPorID(ctx context.Context, id string) (*entity.ConjuntoDato, error) {
	query := `
		SELECT id, fuente_dato_id, nombre, proposito, propietario_dato, clasificacion, estado, created_at, created_by, updated_at, updated_by
		FROM dbgs_schema.conjuntos_datos WHERE id = $1
	`
	row := r.db.QueryRowContext(ctx, query, id)

	var cd entity.ConjuntoDato
	var createdBy, updatedBy sql.NullString
	err := row.Scan(&cd.ID, &cd.FuenteDatoID, &cd.Nombre, &cd.Proposito, &cd.PropietarioDato, &cd.Clasificacion, &cd.Estado, &cd.CreatedAt, &createdBy, &cd.UpdatedAt, &updatedBy)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, entity.ErrEntidadNoEncontrada
		}
		return nil, entity.ErrErrorInterno
	}
	cd.CreatedBy = nullableString(createdBy)
	cd.UpdatedBy = nullableString(updatedBy)
	return &cd, nil
}

func (r *datasetPostgresRepository) ListarDatasets(ctx context.Context, clasificacion, propietario string, limite, offset int) ([]entity.ConjuntoDato, int64, error) {
	countQuery := `
		SELECT COUNT(*) FROM dbgs_schema.conjuntos_datos
		WHERE ($1 = '' OR clasificacion = $1) AND ($2 = '' OR propietario_dato = $2)
	`
	var total int64
	if err := r.db.QueryRowContext(ctx, countQuery, clasificacion, propietario).Scan(&total); err != nil {
		log.Printf("datasetPostgresRepository.ListarDatasets count query error: %v", err)
		return nil, 0, entity.ErrErrorInterno
	}

	query := `
		SELECT id, fuente_dato_id, nombre, proposito, propietario_dato, clasificacion, estado, created_at, created_by, updated_at, updated_by
		FROM dbgs_schema.conjuntos_datos
		WHERE ($1 = '' OR clasificacion = $1) AND ($2 = '' OR propietario_dato = $2)
		ORDER BY nombre ASC
		LIMIT $3 OFFSET $4
	`
	rows, err := r.db.QueryContext(ctx, query, clasificacion, propietario, limite, offset)
	if err != nil {
		log.Printf("datasetPostgresRepository.ListarDatasets query error: %v", err)
		return nil, 0, entity.ErrErrorInterno
	}
	defer rows.Close()

	var datasets []entity.ConjuntoDato
	for rows.Next() {
		var cd entity.ConjuntoDato
		var createdBy, updatedBy sql.NullString
		if err := rows.Scan(&cd.ID, &cd.FuenteDatoID, &cd.Nombre, &cd.Proposito, &cd.PropietarioDato, &cd.Clasificacion, &cd.Estado, &cd.CreatedAt, &createdBy, &cd.UpdatedAt, &updatedBy); err != nil {
			log.Printf("datasetPostgresRepository.ListarDatasets scan error: %v", err)
			return nil, 0, entity.ErrErrorInterno
		}
		cd.CreatedBy = nullableString(createdBy)
		cd.UpdatedBy = nullableString(updatedBy)
		datasets = append(datasets, cd)
	}
	return datasets, total, nil
}

func normalizeDatasetAudit(createdBy, updatedBy string) (string, string) {
	createdBy = strings.TrimSpace(createdBy)
	updatedBy = strings.TrimSpace(updatedBy)
	if createdBy == "" {
		createdBy = "system"
	}
	if updatedBy == "" {
		updatedBy = createdBy
	}
	return createdBy, updatedBy
}

// GuardarDataset inserta un nuevo conjunto de datos en la base de datos.
func (r *datasetPostgresRepository) GuardarDataset(ctx context.Context, dataset *entity.ConjuntoDato) error {
    createdBy, updatedBy := normalizeDatasetAudit(dataset.CreatedBy, dataset.UpdatedBy)
    dataset.CreatedBy = createdBy
    dataset.UpdatedBy = updatedBy

    query := `
        INSERT INTO dbgs_schema.conjuntos_datos (id, fuente_dato_id, nombre, proposito, propietario_dato, clasificacion, estado, created_at, created_by, updated_at, updated_by)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
    `

    _, err := r.db.ExecContext(ctx, query,
        dataset.ID,                 // $1 (NULL o UUID)
        dataset.FuenteDatoID,       // $2
        dataset.Nombre,             // $3
        dataset.Proposito,          // $4
        dataset.PropietarioDato,    // $5
        dataset.Clasificacion,      // $6
        dataset.Estado,             // $7
        dataset.CreatedAt,          // $8
        dataset.CreatedBy,          // $9
        dataset.UpdatedAt,          // $10
        dataset.UpdatedBy,          // $11
    )
    if err != nil {
        log.Printf("datasetPostgresRepository.GuardarDataset exec error: %v", err)
        return entity.ErrErrorInterno
    }
    return nil
}

func (r *datasetPostgresRepository) ActualizarDataset(ctx context.Context, dataset *entity.ConjuntoDato) (*entity.ConjuntoDato, error) {
	createdBy, updatedBy := normalizeDatasetAudit(dataset.CreatedBy, dataset.UpdatedBy)
	dataset.CreatedBy = createdBy
	dataset.UpdatedBy = updatedBy

	query := `
		UPDATE dbgs_schema.conjuntos_datos
		SET fuente_dato_id = $1, nombre = $2, proposito = $3, propietario_dato = $4, clasificacion = $5, estado = $6, updated_at = $7, updated_by = $8
		WHERE id = $9
		RETURNING id, fuente_dato_id, nombre, proposito, propietario_dato, clasificacion, estado, created_at, created_by, updated_at, updated_by
	`

	row := r.db.QueryRowContext(ctx, query,
		dataset.FuenteDatoID, dataset.Nombre, dataset.Proposito,
		dataset.PropietarioDato, dataset.Clasificacion, dataset.Estado,
		dataset.UpdatedAt, dataset.UpdatedBy, dataset.ID,
	)

	var updated entity.ConjuntoDato
	if err := row.Scan(&updated.ID, &updated.FuenteDatoID, &updated.Nombre, &updated.Proposito, &updated.PropietarioDato, &updated.Clasificacion, &updated.Estado, &updated.CreatedAt, &updated.CreatedBy, &updated.UpdatedAt, &updated.UpdatedBy); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, entity.ErrEntidadNoEncontrada
		}
		log.Printf("datasetPostgresRepository.ActualizarDataset scan/return error: %v", err)
		return nil, entity.ErrErrorInterno
	}

	return &updated, nil
}
