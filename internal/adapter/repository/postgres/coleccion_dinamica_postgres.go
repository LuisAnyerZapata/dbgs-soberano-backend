package postgres

import (
    "context"
    "database/sql"
    "log"

    "DBGS_SOBERANO_BACKEND/internal/domain/entity"
    "DBGS_SOBERANO_BACKEND/internal/domain/repository"
)

type coleccionDinamicaPostgresRepository struct {
    db *sql.DB
}

func NewColeccionDinamicaPostgresRepository(db *sql.DB) repository.ColeccionDinamicaRepository {
    return &coleccionDinamicaPostgresRepository{db: db}
}

// EjecutarDDL corre código DDL dinámico (CREATE TABLE, CREATE TRIGGER, etc.)
func (r *coleccionDinamicaPostgresRepository) EjecutarDDL(ctx context.Context, ddl string) error {
    // Se usa Exec porque los comandos DDL no devuelven filas
    _, err := r.db.ExecContext(ctx, ddl)
    if err != nil {
        log.Printf("ERROR CRÍTICO EN DDL DINÁMICO: %v\nSQL Fallido: %s", err, ddl)
        return entity.ErrErrorInterno
    }
    return nil
}

// GuardarMetadatos persiste la definición de la tabla en el diccionario de datos
func (r *coleccionDinamicaPostgresRepository) GuardarMetadatos(ctx context.Context, reg *entity.ColeccionRegistro) error {
    query := `
        INSERT INTO dbgs_schema.colecciones_dinamicas 
        (id, nombre_logico, nombre_fisico, descripcion, institucion_id, estructura, esta_activa, created_at, created_by)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
    `
    
    _, err := r.db.ExecContext(ctx, query,
        reg.ID,
        reg.NombreLogico,
        reg.NombreFisico,
        reg.Descripcion,
        reg.InstitucionID,
        reg.EstructuraJSON,
        reg.EstaActiva,
        reg.CreatedAt,
        reg.CreatedBy,
    )
    if err != nil {
        log.Printf("ERROR EN BD (ColeccionDinamica.GuardarMetadatos): %v", err)
        return entity.ErrErrorInterno
    }
    return nil
}

// ListarMetadatos obtiene las colecciones registradas en el diccionario
func (r *coleccionDinamicaPostgresRepository) ListarMetadatos(ctx context.Context, limite, offset int) ([]entity.ColeccionRegistro, int64, error) {
    countQuery := `SELECT COUNT(*) FROM dbgs_schema.colecciones_dinamicas`
    var total int64
    if err := r.db.QueryRowContext(ctx, countQuery).Scan(&total); err != nil {
        return nil, 0, entity.ErrErrorInterno
    }

    query := `
        SELECT id, nombre_logico, nombre_fisico, descripcion, estructura, esta_activa, created_at
        FROM dbgs_schema.colecciones_dinamicas
        ORDER BY created_at DESC
        LIMIT $1 OFFSET $2
    `
    rows, err := r.db.QueryContext(ctx, query, limite, offset)
    if err != nil {
        return nil, 0, entity.ErrErrorInterno
    }
    defer rows.Close()

    var colecciones []entity.ColeccionRegistro
    for rows.Next() {
        var reg entity.ColeccionRegistro
        if err := rows.Scan(&reg.ID, &reg.NombreLogico, &reg.NombreFisico, &reg.Descripcion, &reg.EstructuraJSON, &reg.EstaActiva, &reg.CreatedAt); err != nil {
            return nil, 0, entity.ErrErrorInterno
        }
        colecciones = append(colecciones, reg)
    }
    return colecciones, total, nil
}