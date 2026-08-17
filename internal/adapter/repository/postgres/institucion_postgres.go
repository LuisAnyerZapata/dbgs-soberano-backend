package postgres

import (
    "context"
    "database/sql"
    "errors"

    "DBGS_SOBERANO_BACKEND/internal/domain/entity"
    "DBGS_SOBERANO_BACKEND/internal/domain/repository"
)

type institucionPostgresRepository struct {
    db *sql.DB
}

// NewInstitucionPostgresRepository crea una nueva instancia del adaptador de instituciones
func NewInstitucionPostgresRepository(db *sql.DB) repository.InstitucionRepository {
    return &institucionPostgresRepository{db: db}
}

// ObtenerPorID busca una institución por su identificador único
func (r *institucionPostgresRepository) ObtenerPorID(ctx context.Context, id string) (*entity.Institucion, error) {
    query := `
        SELECT id, codigo_institucional, nombre_institucion, siglas, estatus, created_at, created_by, updated_at, updated_by
        FROM dbgs_schema.instituciones WHERE id = $1
    `
    row := r.db.QueryRowContext(ctx, query, id)

    var inst entity.Institucion
    var createdBy, updatedBy sql.NullString
    err := row.Scan(&inst.ID, &inst.CodigoInstitucional, &inst.NombreInstitucion, &inst.Siglas, &inst.Estatus, &inst.CreatedAt, &createdBy, &inst.UpdatedAt, &updatedBy)
    
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return nil, entity.ErrEntidadNoEncontrada
        }
        return nil, entity.ErrErrorInterno
    }
    
    inst.CreatedBy = nullableString(createdBy)
    inst.UpdatedBy = nullableString(updatedBy)
    return &inst, nil
}

// Listar obtiene todas las instituciones, con opción de filtrar solo las activas
func (r *institucionPostgresRepository) Listar(ctx context.Context, soloActivos bool) ([]entity.Institucion, error) {
    query := `
        SELECT id, codigo_institucional, nombre_institucion, siglas, estatus, created_at, created_by, updated_at, updated_by
        FROM dbgs_schema.instituciones
        WHERE ($1 = false OR estatus = true)
        ORDER BY nombre_institucion ASC
    `
    rows, err := r.db.QueryContext(ctx, query, soloActivos)
    if err != nil {
        return nil, entity.ErrErrorInterno
    }
    defer rows.Close()

    var instituciones []entity.Institucion
    for rows.Next() {
        var inst entity.Institucion
        var createdBy, updatedBy sql.NullString
        if err := rows.Scan(&inst.ID, &inst.CodigoInstitucional, &inst.NombreInstitucion, &inst.Siglas, &inst.Estatus, &inst.CreatedAt, &createdBy, &inst.UpdatedAt, &updatedBy); err != nil {
            return nil, entity.ErrErrorInterno
        }
        inst.CreatedBy = nullableString(createdBy)
        inst.UpdatedBy = nullableString(updatedBy)
        instituciones = append(instituciones, inst)
    }
    return instituciones, nil
}