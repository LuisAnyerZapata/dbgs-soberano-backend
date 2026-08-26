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

// ObtenerMetadatosPorNombre localiza una colección activa por su nombre lógico
func (r *coleccionDinamicaPostgresRepository) ObtenerMetadatosPorNombre(ctx context.Context, nombreLogico string) (*entity.ColeccionRegistro, error) {
    query := `
        SELECT id, nombre_logico, nombre_fisico, descripcion, estructura, esta_activa, created_at
        FROM dbgs_schema.colecciones_dinamicas
        WHERE LOWER(nombre_logico) = LOWER($1)
    `
    row := r.db.QueryRowContext(ctx, query, nombreLogico)

    var reg entity.ColeccionRegistro
    var descripcion sql.NullString
    if err := row.Scan(&reg.ID, &reg.NombreLogico, &reg.NombreFisico, &descripcion, &reg.EstructuraJSON, &reg.EstaActiva, &reg.CreatedAt); err != nil {
        if err == sql.ErrNoRows {
            return nil, entity.ErrEntidadNoEncontrada
        }
        log.Printf("ERROR EN BD (ColeccionDinamica.ObtenerMetadatosPorNombre '%s'): %v", nombreLogico, err)
        return nil, entity.ErrErrorInterno
    }
    reg.Descripcion = descripcion.String

    return &reg, nil
}

// ListarMetadatos obtiene las colecciones registradas en el diccionario
func (r *coleccionDinamicaPostgresRepository) ListarMetadatos(ctx context.Context, limite, offset int) ([]entity.ColeccionRegistro, int64, error) {
    countQuery := `SELECT COUNT(*) FROM dbgs_schema.colecciones_dinamicas`
    var total int64
    if err := r.db.QueryRowContext(ctx, countQuery).Scan(&total); err != nil {
        log.Printf("ERROR EN BD (ColeccionDinamica.ListarMetadatos COUNT): %v", err)
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
        log.Printf("ERROR EN BD (ColeccionDinamica.ListarMetadatos): %v", err)
        return nil, 0, entity.ErrErrorInterno
    }
    defer rows.Close()

    var colecciones []entity.ColeccionRegistro
    for rows.Next() {
        var reg entity.ColeccionRegistro
        // La columna descripcion es NULLABLE: se escanea en NullString para evitar errores
        var descripcion sql.NullString
        if err := rows.Scan(&reg.ID, &reg.NombreLogico, &reg.NombreFisico, &descripcion, &reg.EstructuraJSON, &reg.EstaActiva, &reg.CreatedAt); err != nil {
            log.Printf("ERROR EN BD (ColeccionDinamica.ListarMetadatos scan): %v", err)
            return nil, 0, entity.ErrErrorInterno
        }
        reg.Descripcion = descripcion.String
        colecciones = append(colecciones, reg)
    }
    return colecciones, total, nil
}

// ActualizarMetadatos sobrescribe la estructura JSON de una colección existente
func (r *coleccionDinamicaPostgresRepository) ActualizarMetadatos(ctx context.Context, nombreLogico string, estructura []byte) error {
    query := `
        UPDATE dbgs_schema.colecciones_dinamicas
        SET estructura = $1
        WHERE LOWER(nombre_logico) = LOWER($2)
    `
    result, err := r.db.ExecContext(ctx, query, estructura, nombreLogico)
    if err != nil {
        log.Printf("ERROR EN BD (ColeccionDinamica.ActualizarMetadatos '%s'): %v", nombreLogico, err)
        return entity.ErrErrorInterno
    }
    filas, _ := result.RowsAffected()
    if filas == 0 {
        return entity.ErrEntidadNoEncontrada
    }
    return nil
}

// DesactivarMetadatos marca una colección como inactiva (soft delete)
func (r *coleccionDinamicaPostgresRepository) DesactivarMetadatos(ctx context.Context, nombreLogico string) error {
    query := `
        UPDATE dbgs_schema.colecciones_dinamicas
        SET esta_activa = false
        WHERE LOWER(nombre_logico) = LOWER($1)
    `
    result, err := r.db.ExecContext(ctx, query, nombreLogico)
    if err != nil {
        log.Printf("ERROR EN BD (ColeccionDinamica.DesactivarMetadatos '%s'): %v", nombreLogico, err)
        return entity.ErrErrorInterno
    }
    filas, _ := result.RowsAffected()
    if filas == 0 {
        return entity.ErrEntidadNoEncontrada
    }
    return nil
}

// RenombrarMetadatos actualiza el nombre lógico y físico de una colección
func (r *coleccionDinamicaPostgresRepository) RenombrarMetadatos(ctx context.Context, nombreActual, nombreNuevo, fisicoNuevo string) error {
    query := `
        UPDATE dbgs_schema.colecciones_dinamicas
        SET nombre_logico = $1, nombre_fisico = $2
        WHERE LOWER(nombre_logico) = LOWER($3)
    `
    result, err := r.db.ExecContext(ctx, query, nombreNuevo, fisicoNuevo, nombreActual)
    if err != nil {
        log.Printf("ERROR EN BD (ColeccionDinamica.RenombrarMetadatos '%s' → '%s'): %v", nombreActual, nombreNuevo, err)
        return entity.ErrErrorInterno
    }
    filas, _ := result.RowsAffected()
    if filas == 0 {
        return entity.ErrEntidadNoEncontrada
    }
    return nil
}

// EliminarMetadatos elimina el registro del diccionario de datos
func (r *coleccionDinamicaPostgresRepository) EliminarMetadatos(ctx context.Context, nombreLogico string) error {
    query := `
        DELETE FROM dbgs_schema.colecciones_dinamicas
        WHERE LOWER(nombre_logico) = LOWER($1)
    `
    result, err := r.db.ExecContext(ctx, query, nombreLogico)
    if err != nil {
        log.Printf("ERROR EN BD (ColeccionDinamica.EliminarMetadatos '%s'): %v", nombreLogico, err)
        return entity.ErrErrorInterno
    }
    filas, _ := result.RowsAffected()
    if filas == 0 {
        return entity.ErrEntidadNoEncontrada
    }
    return nil
}