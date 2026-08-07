package repository

import (
	"context"

	"DBGS_SOBERANO_BACKEND/internal/domain/entity"
)

// IntegracionRepository define los puertos para clientes y solicitudes de integración.
type IntegracionRepository interface {
	GuardarCliente(ctx context.Context, cliente *entity.ClienteIntegracion) error
	ObtenerClientePorID(ctx context.Context, id string) (*entity.ClienteIntegracion, error)
	GuardarSolicitud(ctx context.Context, solicitud *entity.SolicitudIntegracion) error
}
