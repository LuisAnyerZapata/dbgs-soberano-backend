package usecase

import (
	"context"

	"DBGS_SOBERANO_BACKEND/internal/application/port"
	"DBGS_SOBERANO_BACKEND/internal/domain"
	"DBGS_SOBERANO_BACKEND/internal/domain/entity"
	"DBGS_SOBERANO_BACKEND/internal/domain/repository"
)

type seguridadUseCase struct {
	seguridadRepo repository.SeguridadRepository
}

func NewSeguridadUseCase(repo repository.SeguridadRepository) port.SeguridadUseCasePort {
	return &seguridadUseCase{
		seguridadRepo: repo,
	}
}

func (u *seguridadUseCase) ObtenerPerfilUsuario(ctx context.Context, username string) (*entity.Usuario, error) {
	if username == "" {
		return nil, domain.ErrDatosInvalidos
	}

	usuario, err := u.seguridadRepo.ObtenerUsuarioPorUsername(ctx, username)
	if err != nil {
		return nil, err
	}

	if !usuario.EstaActivo() {
		return nil, domain.ErrRegistroInactivo
	}

	return usuario, nil
}

func (u *seguridadUseCase) ValidarAcceso(ctx context.Context, input port.ValidarAccesoInput) (bool, error) {
	if input.Username == "" || input.Permiso == "" {
		return false, domain.ErrDatosInvalidos
	}

	usuario, err := u.seguridadRepo.ObtenerUsuarioPorUsername(ctx, input.Username)
	if err != nil {
		return false, err
	}

	if !usuario.EstaActivo() {
		return false, domain.ErrRegistroInactivo
	}

	return u.seguridadRepo.ValidarPermiso(ctx, usuario.RolID, input.Permiso)
}