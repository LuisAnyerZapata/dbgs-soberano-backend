package port

import (
	"context"

	"DBGS_SOBERANO_BACKEND/internal/domain/entity"
)

// MotorRespaldo abstrae la ejecución del motor físico de respaldos
// (los scripts bash de db/backup que invocan pg_dump/pg_restore).
// Permite que la lógica de negocio sea testeable sin depender del sistema operativo.
type MotorRespaldo interface {
	// EjecutarCreacion genera un respaldo dentro de destinoDir y retorna la ruta del archivo creado.
	EjecutarCreacion(ctx context.Context, destinoDir string) (string, error)
	// EjecutarRestauracion restaura la base de datos desde rutaArchivo.
	EjecutarRestauracion(ctx context.Context, rutaArchivo string) error
}

type CrearRespaldoInput struct {
	Tipo          string
	RetencionDias int
	Detalles      string
	Usuario       string
}

type RestaurarBackupInput struct {
	BackupID  string
	Confirmar bool
	Usuario   string
}

type DescargarRespaldoInput struct {
	ID      string
	Usuario string
}

type RetencionInput struct {
	DiasRetencion int
	MaximoBackups int
}

type RegistrarLogInput struct {
	Nivel   string
	Modulo  string
	Mensaje string
}

type RegistrarMetricaInput struct {
	Nombre string
	Valor  float64
	Unidad string
}

type HealthCheckInput struct {
	Componente string
}

type CrearRespaldoOutput struct {
	Respaldo *entity.RespaldoOperacion
}

type ObtenerRespaldoOutput struct {
	Respaldo *entity.RespaldoOperacion
}

type ListarRespaldosOutput struct {
	Respaldos []entity.RespaldoOperacion
	Total     int32
}

type DescargarRespaldoOutput struct {
	ID            string
	NombreArchivo string
	Contenido     []byte
}

type RestaurarBackupOutput struct {
	Restauracion *entity.Restauracion
}

type AplicarRetencionOutput struct {
	IDsEliminados []string
}

type RegistroLogOutput struct {
	ID string
}

type RegistroMetricaOutput struct {
	ID string
}

type HealthCheckOutput struct {
	Estado  string
	Mensaje string
}

type RespaldoPort interface {
	CrearRespaldo(ctx context.Context, input CrearRespaldoInput) (*CrearRespaldoOutput, error)
	ObtenerRespaldo(ctx context.Context, id string) (*ObtenerRespaldoOutput, error)
	ListarRespaldos(ctx context.Context, limite int) (*ListarRespaldosOutput, error)
	DescargarRespaldo(ctx context.Context, input DescargarRespaldoInput) (*DescargarRespaldoOutput, error)
	RestaurarRespaldo(ctx context.Context, input RestaurarBackupInput) (*RestaurarBackupOutput, error)
	AplicarRetencion(ctx context.Context, input RetencionInput) (*AplicarRetencionOutput, error)
	RegistrarLog(ctx context.Context, input RegistrarLogInput) (*RegistroLogOutput, error)
	RegistrarMetrica(ctx context.Context, input RegistrarMetricaInput) (*RegistroMetricaOutput, error)
	EjecutarHealthCheck(ctx context.Context, input HealthCheckInput) (*HealthCheckOutput, error)
}
