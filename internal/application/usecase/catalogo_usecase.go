package usecase

import (
	"context"
	"time"

	"DBGS_SOBERANO_BACKEND/internal/application/port"
	"DBGS_SOBERANO_BACKEND/internal/domain"
	"DBGS_SOBERANO_BACKEND/internal/domain/entity"
	"DBGS_SOBERANO_BACKEND/internal/domain/repository"
)

type catalogoUseCase struct {
	catalogoRepo repository.CatalogoRepository
}

func NewCatalogoUseCase(repo repository.CatalogoRepository) port.CatalogoUseCasePort {
	return &catalogoUseCase{
		catalogoRepo: repo,
	}
}

func (u *catalogoUseCase) ObtenerPorID(ctx context.Context, id string) (*entity.Catalogo, error) {
	if id == "" {
		return nil, domain.ErrDatosInvalidos
	}
	return u.catalogoRepo.ObtenerPorID(ctx, id)
}

func (u *catalogoUseCase) ObtenerPorCodigo(ctx context.Context, codigo string) (*entity.Catalogo, error) {
	if codigo == "" {
		return nil, domain.ErrDatosInvalidos
	}
	return u.catalogoRepo.ObtenerPorCodigo(ctx, codigo)
}

func (u *catalogoUseCase) ListarCatalogos(ctx context.Context, input port.ObtenerCatalogosInput) (*port.ObtenerCatalogosOutput, error) {
	if input.Limite <= 0 {
		input.Limite = 10
	}
	if input.Offset < 0 {
		input.Offset = 0
	}

	catalogos, total, err := u.catalogoRepo.Listar(ctx, input.SoloActivos, input.Limite, input.Offset)
	if err != nil {
		return nil, err
	}

	return &port.ObtenerCatalogosOutput{
		Catalogos: catalogos,
		Total:     total,
		Limite:    input.Limite,
		Offset:    input.Offset,
	}, nil
}

func (u *catalogoUseCase) CrearCatalogo(ctx context.Context, catalogo *entity.Catalogo) error {
	if catalogo.Codigo == "" || catalogo.Nombre == "" {
		return domain.ErrDatosInvalidos
	}

	existente, _ := u.catalogoRepo.ObtenerPorCodigo(ctx, catalogo.Codigo)
	if existente != nil {
		return domain.ErrCodigoDuplicado
	}

	catalogo.CreatedAt = time.Now()
	catalogo.Estado = "ACTIVO"
	return u.catalogoRepo.Guardar(ctx, catalogo)
}

func (u *catalogoUseCase) InactivarCatalogo(ctx context.Context, id string, usuarioModificador string) error {
	if id == "" || usuarioModificador == "" {
		return domain.ErrDatosInvalidos
	}

	catalogo, err := u.catalogoRepo.ObtenerPorID(ctx, id)
	if err != nil {
		return err
	}

	catalogo.Inactivar()
	return u.catalogoRepo.ActualizarEstado(ctx, id, false, usuarioModificador)
}