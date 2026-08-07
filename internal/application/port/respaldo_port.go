package port

import "context"

type CrearRespaldoInput struct {
	Tipo      string
	Ruta      string
	Detalles  string
	Usuario   string
	Retencion int
}

type RestaurarBackupInput struct {
	BackupID string
	Validar  bool
	Usuario  string
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
	ID     string
	Estado string
}

type RestaurarBackupOutput struct {
	Validado bool
	Estado   string
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
	ValidarRestauracion(ctx context.Context, input RestaurarBackupInput) (*RestaurarBackupOutput, error)
	AplicarRetencion(ctx context.Context, input RetencionInput) ([]string, error)
	RegistrarLog(ctx context.Context, input RegistrarLogInput) (*RegistroLogOutput, error)
	RegistrarMetrica(ctx context.Context, input RegistrarMetricaInput) (*RegistroMetricaOutput, error)
	EjecutarHealthCheck(ctx context.Context, input HealthCheckInput) (*HealthCheckOutput, error)
}
