package port

import (
	"context"

	"DBGS_SOBERANO_BACKEND/internal/domain/entity"
)

type ValidarAccesoInput struct {
	Username string
	Permiso  string
}

type SeguridadUseCasePort interface {
	ObtenerPerfilUsuario(ctx context.Context, username string) (*entity.Usuario, error)
	ValidarAcceso(ctx context.Context, input ValidarAccesoInput) (bool, error)
}