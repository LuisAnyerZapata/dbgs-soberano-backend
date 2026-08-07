package port

import (
	"context"

	"DBGS_SOBERANO_BACKEND/internal/domain/entity"
)

type RegistrarClienteInput struct {
	Nombre          string
	Tipo            string
	VersionContrato string
	Scopes          []string
}

type ValidarAccesoIntegracionInput struct {
	ClienteID       string
	Token           string
	VersionContrato string
	Scope           string
}

type RegistrarSolicitudInput struct {
	ClienteID string
	Metodo    string
	Recurso   string
	Estado    string
	Detalles  string
}

type IntegracionPort interface {
	RegistrarCliente(ctx context.Context, input RegistrarClienteInput) (*entity.ClienteIntegracion, error)
	ValidarAcceso(ctx context.Context, input ValidarAccesoIntegracionInput) (*entity.ClienteIntegracion, error)
	RegistrarSolicitud(ctx context.Context, input RegistrarSolicitudInput) error
}
