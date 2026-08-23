package postgres

import (
    "context"
    "database/sql"
    "encoding/json"
    "fmt"
    "strings"

    "DBGS_SOBERANO_BACKEND/internal/domain/entity"
    "DBGS_SOBERANO_BACKEND/internal/domain/repository"

    "github.com/lib/pq"
)

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
        return nil, 0, entity.ErrErrorInterno
    }
    defer rows.Close()

    var registros []map[string]interface{}
    for rows.Next() {
        var jsonData []byte
        if err := rows.Scan(&jsonData); err != nil {
            return nil, 0, entity.ErrErrorInterno
        }
        
        var registro map[string]interface{}
        if err := json.Unmarshal(jsonData, &registro); err != nil {
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
        return nil, entity.ErrErrorInterno
    }

    var resultado map[string]interface{}
    if err := json.Unmarshal(jsonData, &resultado); err != nil {
        return nil, entity.ErrErrorInterno
    }
    return resultado, nil
}

func (r *datosDinamicosPostgresRepository) Insertar(ctx context.Context, nombreFisico string, datos map[string]interface{}, createdBy string) (string, error) {
    var columnas []string
    var placeholders []string
    var valores []interface{}

    // Inyectamos el auditor obligatorio ($1)
    valores = append(valores, createdBy)
    placeholders = append(placeholders, "$1")

    // Iteramos el mapa dinámico ignorando campos del sistema
    for clave, valor := range datos {
        if camposSistema[clave] {
            continue // No permitimos sobreescribir ID o fechas del sistema
        }
        columnas = append(columnas, pq.QuoteIdentifier(clave))
        placeholders = append(placeholders, fmt.Sprintf("$%d", len(valores)+1))
        valores = append(valores, valor)
    }

    query := fmt.Sprintf(
        "INSERT INTO %s.%s (%s, created_by) VALUES (%s) RETURNING id",
        DBGS_SCHEMA,
        pq.QuoteIdentifier(nombreFisico),
        strings.Join(columnas, ", "),
        strings.Join(placeholders, ", "),
    )

    var nuevoID string
    err := r.db.QueryRowContext(ctx, query, valores...).Scan(&nuevoID)
    if err != nil {
        if err == sql.ErrNoRows {
            return "", entity.ErrErrorInterno
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
        return nil, entity.ErrErrorInterno
    }

    var resultado map[string]interface{}
    if err := json.Unmarshal(jsonData, &resultado); err != nil {
        return nil, entity.ErrErrorInterno
    }
    return resultado, nil
}

func (r *datosDinamicosPostgresRepository) Eliminar(ctx context.Context, nombreFisico string, id string) error {
    query := fmt.Sprintf("DELETE FROM %s.%s WHERE id = $1", DBGS_SCHEMA, pq.QuoteIdentifier(nombreFisico))
    
    res, err := r.db.ExecContext(ctx, query, id)
    if err != nil {
        return entity.ErrErrorInterno
    }
    
    filasAfectadas, _ := res.RowsAffected()
    if filasAfectadas == 0 {
        return entity.ErrEntidadNoEncontrada
    }
    return nil
}