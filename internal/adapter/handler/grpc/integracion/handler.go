package integracion

import (
	"context"

	"DBGS_SOBERANO_BACKEND/internal/application/port"
	"DBGS_SOBERANO_BACKEND/internal/domain/entity"
)

type Handler struct {
	useCase port.IntegracionPort
}

func NewHandler(useCase port.IntegracionPort) *Handler {
	return &Handler{useCase: useCase}
}

func (h *Handler) RegistrarCliente(ctx context.Context, input port.RegistrarClienteInput) (*entity.ClienteIntegracion, error) {
	return h.useCase.RegistrarCliente(ctx, input)
}

func (h *Handler) ValidarAcceso(ctx context.Context, input port.ValidarAccesoIntegracionInput) (*entity.ClienteIntegracion, error) {
	return h.useCase.ValidarAcceso(ctx, input)
}

func (h *Handler) RegistrarSolicitud(ctx context.Context, input port.RegistrarSolicitudInput) error {
	return h.useCase.RegistrarSolicitud(ctx, input)
}
