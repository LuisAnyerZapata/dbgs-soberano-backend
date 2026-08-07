package respaldo

import (
	"context"

	"DBGS_SOBERANO_BACKEND/internal/application/port"
)

type Handler struct {
	useCase port.RespaldoPort
}

func NewHandler(useCase port.RespaldoPort) *Handler {
	return &Handler{useCase: useCase}
}

func (h *Handler) CrearRespaldo(ctx context.Context, input port.CrearRespaldoInput) (*port.CrearRespaldoOutput, error) {
	return h.useCase.CrearRespaldo(ctx, input)
}

func (h *Handler) ValidarRestauracion(ctx context.Context, input port.RestaurarBackupInput) (*port.RestaurarBackupOutput, error) {
	return h.useCase.ValidarRestauracion(ctx, input)
}

func (h *Handler) AplicarRetencion(ctx context.Context, input port.RetencionInput) ([]string, error) {
	return h.useCase.AplicarRetencion(ctx, input)
}

func (h *Handler) RegistrarLog(ctx context.Context, input port.RegistrarLogInput) (*port.RegistroLogOutput, error) {
	return h.useCase.RegistrarLog(ctx, input)
}

func (h *Handler) RegistrarMetrica(ctx context.Context, input port.RegistrarMetricaInput) (*port.RegistroMetricaOutput, error) {
	return h.useCase.RegistrarMetrica(ctx, input)
}

func (h *Handler) EjecutarHealthCheck(ctx context.Context, input port.HealthCheckInput) (*port.HealthCheckOutput, error) {
	return h.useCase.EjecutarHealthCheck(ctx, input)
}
