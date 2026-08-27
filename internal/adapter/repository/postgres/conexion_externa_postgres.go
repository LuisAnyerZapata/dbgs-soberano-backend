package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	"DBGS_SOBERANO_BACKEND/internal/application/port"
	"DBGS_SOBERANO_BACKEND/internal/domain"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
)

// conexionExternaAdapter implementa port.ConexionExternaPort para operar sobre
// bases de datos externas (PostgreSQL / MySQL), aislada de la BD del sistema.
type conexionExternaAdapter struct{}

// NewConexionExternaAdapter crea el adaptador de conexiones externas.
func NewConexionExternaAdapter() port.ConexionExternaPort {
	return &conexionExternaAdapter{}
}

func dsnParaPostgres(c port.ConexionCredenciales) string {
	ssl := c.SSLMode
	if ssl == "" {
		ssl = "disable"
	}
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s connect_timeout=5",
		c.Host, c.Port, c.User, c.Password, c.Database, ssl)
}

func dsnParaMysql(c port.ConexionCredenciales) string {
	// ssl-mode: prefer | required | disable
	params := "parseTime=true&timeout=5s"
	sslmode := strings.ToLower(c.SSLMode)
	switch sslmode {
	case "require", "required", "verify-ca", "verify-full":
		params += "&tls=true"
	case "disable", "":
		params += "&tls=false"
	default:
		params += "&tls=preferred"
	}
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?%s", c.User, c.Password, c.Host, c.Port, c.Database, params)
}

func abrirConexionExterna(c port.ConexionCredenciales) (*sql.DB, error) {
	switch strings.ToLower(c.Engine) {
	case "mysql":
		return sql.Open("mysql", dsnParaMysql(c))
	case "postgres", "postgresql", "pgsql":
		return sql.Open("postgres", dsnParaPostgres(c))
	default:
		return nil, fmt.Errorf("%w: motor de base de datos no soportado '%s'", domain.InvalidArgument, c.Engine)
	}
}

func (a *conexionExternaAdapter) Probar(ctx context.Context, c port.ConexionCredenciales) (*port.PruebaConexionResult, error) {
	db, err := abrirConexionExterna(c)
	if err != nil {
		return &port.PruebaConexionResult{OK: false, Message: err.Error()}, nil
	}
	defer db.Close()
	start := time.Now()
	if err := db.PingContext(ctx); err != nil {
		return &port.PruebaConexionResult{OK: false, Message: err.Error()}, nil
	}
	return &port.PruebaConexionResult{
		OK:        true,
		Message:   "Conexión correcta.",
		LatencyMS: float64(time.Since(start).Microseconds()) / 1000.0,
	}, nil
}

func (a *conexionExternaAdapter) ListarEsquemas(ctx context.Context, c port.ConexionCredenciales) ([]string, error) {
	db, err := abrirConexionExterna(c)
	if err != nil {
		return nil, err
	}
	defer db.Close()
	if err := db.PingContext(ctx); err != nil {
		return nil, domain.InternalError.WithMessage("no se pudo conectar a la base externa")
	}
	var query string
	if strings.EqualFold(c.Engine, "mysql") {
		query = "SELECT SCHEMA_NAME FROM INFORMATION_SCHEMA.SCHEMATA ORDER BY SCHEMA_NAME"
	} else {
		query = "SELECT nspname FROM pg_catalog.pg_namespace WHERE nspname NOT LIKE 'pg_%' AND nspname <> 'information_schema' ORDER BY nspname"
	}
	return listarStrings(ctx, db, query)
}

func (a *conexionExternaAdapter) ListarTablas(ctx context.Context, c port.ConexionCredenciales, schema string) ([]string, error) {
	db, err := abrirConexionExterna(c)
	if err != nil {
		return nil, err
	}
	defer db.Close()
	if err := db.PingContext(ctx); err != nil {
		return nil, domain.InternalError.WithMessage("no se pudo conectar a la base externa")
	}
	var query string
	var args []interface{}
	if strings.EqualFold(c.Engine, "mysql") {
		query = "SELECT TABLE_NAME FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_TYPE='BASE TABLE' AND TABLE_SCHEMA=? ORDER BY TABLE_NAME"
		args = append(args, schema)
	} else {
		if schema == "" {
			schema = "public"
		}
		query = "SELECT tablename FROM pg_catalog.pg_tables WHERE schemaname=$1 ORDER BY tablename"
		args = append(args, schema)
	}
	return listarStrings(ctx, db, query, args...)
}

func listarStrings(ctx context.Context, db *sql.DB, query string, args ...interface{}) ([]string, error) {
	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		log.Printf("conexionExternaAdapter query error: %v", err)
		return nil, domain.InternalError.WithMessage("error consultando el catálogo externo")
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var s string
		if err := rows.Scan(&s); err != nil {
			return nil, domain.InternalError
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

func (a *conexionExternaAdapter) ExplorarDatos(ctx context.Context, c port.ConexionCredenciales, schema, tabla string, limite, offset int) (*port.TablaExterna, error) {
	db, err := abrirConexionExterna(c)
	if err != nil {
		return nil, err
	}
	defer db.Close()
	if err := db.PingContext(ctx); err != nil {
		return nil, domain.InternalError.WithMessage("no se pudo conectar a la base externa")
	}
	if strings.EqualFold(c.Engine, "mysql") {
		return explorarMysql(ctx, db, schema, tabla, limite, offset)
	}
	if schema == "" {
		schema = "public"
	}
	if limite <= 0 {
		limite = 50
	}
	if offset < 0 {
		offset = 0
	}
	return explorarPostgres(ctx, db, schema, tabla, limite, offset)
}

func explorarPostgres(ctx context.Context, db *sql.DB, schema, tabla string, limite, offset int) (*port.TablaExterna, error) {
	// Descubrir columnas
	colQuery := `SELECT c.column_name, c.data_type, c.is_nullable, COALESCE((SELECT true FROM information_schema.table_constraints tc JOIN information_schema.key_column_usage kcu ON tc.constraint_name = kcu.constraint_name AND tc.constraint_schema = kcu.constraint_schema WHERE tc.constraint_type='PRIMARY KEY' AND tc.table_schema=c.table_schema AND tc.table_name=c.table_name AND kcu.column_name=c.column_name), false)
		FROM information_schema.columns c WHERE c.table_schema=$1 AND c.table_name=$2 ORDER BY c.ordinal_position`
	colRows, err := db.QueryContext(ctx, colQuery, schema, tabla)
	if err != nil {
		return nil, domain.InternalError.WithMessage("error consultando columnas externas")
	}
	defer colRows.Close()
	var cols []port.ColumnaExterna
	for colRows.Next() {
		var c port.ColumnaExterna
		var nullable string
		var pk bool
		if err := colRows.Scan(&c.Name, &c.DataType, &nullable, &pk); err != nil {
			return nil, domain.InternalError
		}
		c.Nullable = strings.EqualFold(nullable, "YES")
		c.PrimaryKey = pk
		cols = append(cols, c)
	}
	if err := colRows.Err(); err != nil {
		return nil, domain.InternalError
	}
	if len(cols) == 0 {
		return nil, domain.NotFoundError("tabla")
	}

	// Contar registros
	var total int64
	ident := fmt.Sprintf("%s.%s", schema, tabla)
	if err := db.QueryRowContext(ctx, fmt.Sprintf("SELECT COUNT(*) FROM %s", ident)).Scan(&total); err != nil {
		return nil, domain.InternalError
	}

	// Leer filas
	selRows, err := db.QueryContext(ctx, fmt.Sprintf("SELECT * FROM %s ORDER BY 1 LIMIT %d OFFSET %d", ident, limite, offset))
	if err != nil {
		return nil, domain.InternalError
	}
	defer selRows.Close()
	rowsMap, err := scanFilas(selRows)
	if err != nil {
		return nil, err
	}
	return &port.TablaExterna{Columns: cols, Rows: rowsMap, Total: total}, nil
}

func explorarMysql(ctx context.Context, db *sql.DB, schema, tabla string, limite, offset int) (*port.TablaExterna, error) {
	colQuery := `SELECT COLUMN_NAME, COLUMN_TYPE, IS_NULLABLE, COLUMN_KEY='PRI' FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA=? AND TABLE_NAME=? ORDER BY ORDINAL_POSITION`
	colRows, err := db.QueryContext(ctx, colQuery, schema, tabla)
	if err != nil {
		return nil, domain.InternalError.WithMessage("error consultando columnas externas")
	}
	defer colRows.Close()
	var cols []port.ColumnaExterna
	for colRows.Next() {
		var c port.ColumnaExterna
		var nullable string
		var pk bool
		if err := colRows.Scan(&c.Name, &c.DataType, &nullable, &pk); err != nil {
			return nil, domain.InternalError
		}
		c.Nullable = strings.EqualFold(nullable, "YES")
		c.PrimaryKey = pk
		cols = append(cols, c)
	}
	if err := colRows.Err(); err != nil {
		return nil, domain.InternalError
	}
	if len(cols) == 0 {
		return nil, domain.NotFoundError("tabla")
	}

	ident := schema + "." + tabla
	var total int64
	if err := db.QueryRowContext(ctx, fmt.Sprintf("SELECT COUNT(*) FROM %s", ident)).Scan(&total); err != nil {
		return nil, domain.InternalError
	}
	selRows, err := db.QueryContext(ctx, fmt.Sprintf("SELECT * FROM %s ORDER BY 1 LIMIT %d OFFSET %d", ident, limite, offset))
	if err != nil {
		return nil, domain.InternalError
	}
	defer selRows.Close()
	rowsMap, err := scanFilas(selRows)
	if err != nil {
		return nil, err
	}
	return &port.TablaExterna{Columns: cols, Rows: rowsMap, Total: total}, nil
}

// scanFilas convierte las filas de un *sql.Rows en []map[string]interface{}.
func scanFilas(rows *sql.Rows) ([]map[string]interface{}, error) {
	cols, err := rows.Columns()
	if err != nil {
		return nil, domain.InternalError
	}
	var out []map[string]interface{}
	for rows.Next() {
		vals := make([]interface{}, len(cols))
		ptrs := make([]interface{}, len(cols))
		for i := range vals {
			ptrs[i] = &vals[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			return nil, domain.InternalError
		}
		row := make(map[string]interface{}, len(cols))
		for i, name := range cols {
			row[name] = normalizeValor(vals[i])
		}
		out = append(out, row)
	}
	return out, rows.Err()
}

// normalizeValor convierte tipos no-JSON ([]byte, time.Time) a tipos serializables.
func normalizeValor(v interface{}) interface{} {
	switch t := v.(type) {
	case []byte:
		return string(t)
	case [16]byte:
		return fmt.Sprintf("%x-%x-%x-%x-%x", t[0:4], t[4:6], t[6:8], t[8:10], t[10:16])
	case time.Time:
		return t.Format(time.RFC3339)
	case nil:
		return nil
	default:
		return v
	}
}
