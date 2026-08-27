package port

import (
	"context"

	"DBGS_SOBERANO_BACKEND/internal/domain/entity"
)

// ConexionMetadata es la información de credenciales resuelta para abrir
// una conexión real a la base externa. El password se descifra en el adaptador
// de salida y se pasa únicamente al driver externo.
type ConexionCredenciales struct {
	Engine       string
	Host         string
	Port         int
	User         string
	Password     string
	Database     string
	SSLMode      string
}

// TablaExterna describe una columna descubierta dentro de una tabla externa.
type TablaExterna struct {
	Columns []ColumnaExterna
	Rows    []map[string]interface{}
	Total   int64
}

// ColumnaExterna describe el esquema de una columna.
type ColumnaExterna struct {
	Name      string
	DataType  string
	Nullable  bool
	PrimaryKey bool
}

// PruebaConexionResult encapsula el resultado de una prueba de conectividad.
type PruebaConexionResult struct {
	OK        bool
	Message   string
	LatencyMS float64
}

// ConexionExternaPort define el contrato para operar sobre bases de datos
// externas (probar, listar esquemas, listar tablas y explorar datos).
type ConexionExternaPort interface {
	Probar(ctx context.Context, creds ConexionCredenciales) (*PruebaConexionResult, error)
	ListarEsquemas(ctx context.Context, creds ConexionCredenciales) ([]string, error)
	ListarTablas(ctx context.Context, creds ConexionCredenciales, schema string) ([]string, error)
	ExplorarDatos(ctx context.Context, creds ConexionCredenciales, schema, tabla string, limite, offset int) (*TablaExterna, error)
}

// ConexionPort define los casos de uso del dominio de conexiones.
type ConexionPort interface {
	ProbarConexion(ctx context.Context, c *entity.Conexion, password string) (*PruebaConexionResult, error)
	CrearConexion(ctx context.Context, c *entity.Conexion, password string) (*entity.Conexion, error)
	ListarConexiones(ctx context.Context, limite, offset int) ([]entity.Conexion, int64, error)
	ObtenerConexion(ctx context.Context, id string) (*entity.Conexion, error)
	ActualizarConexion(ctx context.Context, c *entity.Conexion, password string) (*entity.Conexion, error)
	EliminarConexion(ctx context.Context, id string) error
	ListarEsquemas(ctx context.Context, id string) ([]string, error)
	ListarTablas(ctx context.Context, id, schema string) ([]string, error)
	ExplorarDatos(ctx context.Context, id, schema, tabla string, limite, offset int) (*TablaExterna, error)
}

// ApiPublicadaPort define los casos de uso del dominio de APIs públicas.
type ApiPublicadaPort interface {
	CrearApi(ctx context.Context, a *entity.ApiPublicada) (*entity.ApiPublicada, error)
	ListarApis(ctx context.Context, limite, offset int) ([]entity.ApiPublicada, int64, error)
	ObtenerApi(ctx context.Context, id string) (*entity.ApiPublicada, error)
	CambiarEstadoApi(ctx context.Context, id string, active bool) (*entity.ApiPublicada, error)
	EliminarApi(ctx context.Context, id string) error
}
