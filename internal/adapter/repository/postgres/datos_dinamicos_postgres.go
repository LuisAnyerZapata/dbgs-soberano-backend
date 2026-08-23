package postgres

import (
    "context"
    "database/sql"
    "encoding/json"
    "errors"
    "fmt"
    "log"
    "strings"

    "DBGS_SOBERANO_BACKEND/internal/domain/entity"
    "DBGS_SOBERANO_BACKEND/internal/domain/repository"

    "github.com/lib/pq"
)

// codigoViolacionUnica corresponde al SQLSTATE de PostgreSQL para UNIQUE violations
const codigoViolacionUnica = "23505"

// esErrorUnico determina si el error devuelto por PostgreSQL es una violación de unicidad
func esErrorUnico(err error) bool {
    var pqErr *pq.Error
    return errors.As(err, &pqErr) && pqErr.Code == codigoViolacionUnica
}

// Constante local para evitar hardcodeos y errores tipográficos
const DBGS_SCHEMA = "dbgs_schema"

type datosDinamicosPostgresRepository struct {
    db *sql.DB
}

func NewDatosDinamicosPostgresRepository(db *sql.DB) repository.DatosDinamicosRepository {
    return &datosDinamicosPostgresRepository{db: db}
}

// camposSistema define qué campos maneja PostgreSQL automáticamente y no deben ser inyectados por el usuario en INSERT/UPDATE
var camposSistema = map[string]bool{
    "id": true, "created_at": true, "updated_at": true,
}

func (r *datosDinamicosPostgresRepository) Listar(ctx context.Context, nombreFisico string, limite, offset int) ([]map[string]interface{}, int64, error) {
    // 1. Contar total
    countQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s.%s", DBGS_SCHEMA, pq.QuoteIdentifier(nombreFisico))
    var total int64
    if err := r.db.QueryRowContext(ctx, countQuery).Scan(&total); err != nil {
        log.Printf("ERROR EN BD (DatosDinamicos.Listar COUNT en %s): %v", nombreFisico, err)
        return nil, 0, entity.ErrErrorInterno
    }

    // 2. Ejecutar consulta paginada usando row_to_json para conversión dinámica segura
    query := fmt.Sprintf(`
        SELECT row_to_json(t)::json 
        FROM (SELECT * FROM %s.%s ORDER BY created_at DESC LIMIT $1 OFFSET $2) t`,
        DBGS_SCHEMA, pq.QuoteIdentifier(nombreFisico),
    )

    rows, err := r.db.QueryContext(ctx, query, limite, offset)
    if err != nil {
        log.Printf("ERROR EN BD (DatosDinamicos.Listar en %s): %v", nombreFisico, err)
        return nil, 0, entity.ErrErrorInterno
    }
    defer rows.Close()

    var registros []map[string]interface{}
    for rows.Next() {
        var jsonData []byte
        if err := rows.Scan(&jsonData); err != nil {
            log.Printf("ERROR EN BD (DatosDinamicos.Listar scan en %s): %v", nombreFisico, err)
            return nil, 0, entity.ErrErrorInterno
        }
        
        var registro map[string]interface{}
        if err := json.Unmarshal(jsonData, &registro); err != nil {
            log.Printf("ERROR EN BD (DatosDinamicos.Listar unmarshal en %s): %v", nombreFisico, err)
            return nil, 0, entity.ErrErrorInterno
        }
        registros = append(registros, registro)
    }
    return registros, total, nil
}

func (r *datosDinamicosPostgresRepository) ObtenerPorID(ctx context.Context, nombreFisico string, id string) (map[string]interface{}, error) {
    query := fmt.Sprintf(`
        SELECT row_to_json(t)::json 
        FROM (SELECT * FROM %s.%s WHERE id = $1) t`,
        DBGS_SCHEMA, pq.QuoteIdentifier(nombreFisico),
    )

    row := r.db.QueryRowContext(ctx, query, id)
    var jsonData []byte
    if err := row.Scan(&jsonData); err != nil {
        if err == sql.ErrNoRows {
            return nil, entity.ErrEntidadNoEncontrada
        }
        log.Printf("ERROR EN BD (DatosDinamicos.ObtenerPorID en %s): %v", nombreFisico, err)
        return nil, entity.ErrErrorInterno
    }

    var resultado map[string]interface{}
    if err := json.Unmarshal(jsonData, &resultado); err != nil {
        log.Printf("ERROR EN BD (DatosDinamicos.ObtenerPorID unmarshal en %s): %v", nombreFisico, err)
        return nil, entity.ErrErrorInterno
    }
    return resultado, nil
}

func (r *datosDinamicosPostgresRepository) Insertar(ctx context.Context, nombreFisico string, datos map[string]interface{}, createdBy string) (string, error) {
    var columnas []string
    var placeholders []string
    var valores []interface{}

    // Iteramos el mapa dinámico ignorando campos del sistema y de auditoría
    // (created_by/updated_by se inyectan de forma controlada más abajo)
    for clave, valor := range datos {
        if camposSistema[clave] || clave == "created_by" || clave == "updated_by" {
            continue
        }
        columnas = append(columnas, pq.QuoteIdentifier(clave))
        placeholders = append(placeholders, fmt.Sprintf("$%d", len(valores)+1))
        valores = append(valores, valor)
    }

    // El auditor obligatorio se agrega AL FINAL para que el placeholder
    // quede emparejado con la posición exacta de su columna
    columnas = append(columnas, "created_by")
    placeholders = append(placeholders, fmt.Sprintf("$%d", len(valores)+1))
    valores = append(valores, createdBy)

    query := fmt.Sprintf(
        "INSERT INTO %s.%s (%s) VALUES (%s) RETURNING id",
        DBGS_SCHEMA,
        pq.QuoteIdentifier(nombreFisico),
        strings.Join(columnas, ", "),
        strings.Join(placeholders, ", "),
    )

    var nuevoID string
    err := r.db.QueryRowContext(ctx, query, valores...).Scan(&nuevoID)
    if err != nil {
        log.Printf("ERROR EN BD (DatosDinamicos.Insertar en %s): %v\nSQL: %s", nombreFisico, err, query)
        if esErrorUnico(err) {
            return "", entity.ErrCodigoDuplicado
        }
        return "", entity.ErrErrorInterno
    }
    return nuevoID, nil
}

func (r *datosDinamicosPostgresRepository) Actualizar(ctx context.Context, nombreFisico string, id string, datos map[string]interface{}, updatedBy string) (map[string]interface{}, error) {
    var sets []string
    var valores []interface{}

    // Inyectamos el auditor obligatorio ($1)
    valores = append(valores, updatedBy)
    sets = append(sets, "updated_by = $1")

    // Iteramos el mapa dinámico ignorando campos del sistema y el ID
    for clave, valor := range datos {
        if clave == "id" || camposSistema[clave] {
            continue
        }
        valores = append(valores, valor)
        sets = append(sets, fmt.Sprintf("%s = $%d", pq.QuoteIdentifier(clave), len(valores)))
    }

    if len(sets) == 1 {
        return nil, entity.ErrDatosInvalidos // No hay campos válidos para actualizar
    }

    // El ID del registro a actualizar va al final
    valores = append(valores, id)
    idPosicion := len(valores)

    query := fmt.Sprintf(`
        SELECT row_to_json(t)::json 
        FROM (SELECT * FROM %s.%s WHERE id = $%d) t`,
        DBGS_SCHEMA,
        pq.QuoteIdentifier(nombreFisico),
        idPosicion,
    )

    row := r.db.QueryRowContext(ctx, query, valores...)
    var jsonData []byte
    if err := row.Scan(&jsonData); err != nil {
        if err == sql.ErrNoRows {
            return nil, entity.ErrEntidadNoEncontrada
        }
        log.Printf("ERROR EN BD (DatosDinamicos.Actualizar en %s): %v\nSQL: %s", nombreFisico, err, query)
        if esErrorUnico(err) {
            return nil, entity.ErrCodigoDuplicado
        }
        return nil, entity.ErrErrorInterno
    }

    var resultado map[string]interface{}
    if err := json.Unmarshal(jsonData, &resultado); err != nil {
        log.Printf("ERROR EN BD (DatosDinamicos.Actualizar unmarshal en %s): %v", nombreFisico, err)
        return nil, entity.ErrErrorInterno
    }
    return resultado, nil
}

func (r *datosDinamicosPostgresRepository) Eliminar(ctx context.Context, nombreFisico string, id string) error {
    query := fmt.Sprintf("DELETE FROM %s.%s WHERE id = $1", DBGS_SCHEMA, pq.QuoteIdentifier(nombreFisico))
    
    res, err := r.db.ExecContext(ctx, query, id)
    if err != nil {
        log.Printf("ERROR EN BD (DatosDinamicos.Eliminar en %s): %v", nombreFisico, err)
        return entity.ErrErrorInterno
    }
    
    filasAfectadas, _ := res.RowsAffected()
    if filasAfectadas == 0 {
        return entity.ErrEntidadNoEncontrada
    }
    return nil
}