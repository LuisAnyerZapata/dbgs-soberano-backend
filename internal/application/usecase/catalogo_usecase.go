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

func (u *catalogoUseCase) CrearCatalogo(ctx context.Context, catalogo *entity.Catalogo) (*entity.Catalogo, error) {
	if catalogo.Codigo == "" || catalogo.Nombre == "" {
		return nil, domain.ErrDatosInvalidos
	}

	existente, _ := u.catalogoRepo.ObtenerPorCodigo(ctx, catalogo.Codigo)
	if existente != nil {
		return nil, domain.ErrCodigoDuplicado
	}

	catalogo.CreatedAt = time.Now()
	catalogo.UpdatedAt = catalogo.CreatedAt
	catalogo.Estado = true
	if catalogo.CreatedBy == "" {
		catalogo.CreatedBy = "system"
	}
	catalogo.UpdatedBy = catalogo.CreatedBy

	if err := u.catalogoRepo.Guardar(ctx, catalogo); err != nil {
		return nil, err
	}

	return catalogo, nil
}

func (u *catalogoUseCase) ActualizarCatalogo(ctx context.Context, catalogo *entity.Catalogo) (*entity.Catalogo, error) {
	if catalogo.ID == "" || catalogo.Nombre == "" {
		return nil, domain.ErrDatosInvalidos
	}

	existente, err := u.catalogoRepo.ObtenerPorID(ctx, catalogo.ID)
	if err != nil {
		return nil, err
	}

	if catalogo.Codigo == "" {
		catalogo.Codigo = existente.Codigo
	} else if existente.Codigo != catalogo.Codigo {
		otro, _ := u.catalogoRepo.ObtenerPorCodigo(ctx, catalogo.Codigo)
		if otro != nil {
			return nil, domain.ErrCodigoDuplicado
		}
	}

	catalogo.UpdatedAt = time.Now()
	if catalogo.UpdatedBy == "" {
		catalogo.UpdatedBy = existente.UpdatedBy
	}

	return u.catalogoRepo.Actualizar(ctx, catalogo)
}

func (u *catalogoUseCase) InactivarCatalogo(ctx context.Context, id string, usuarioModificador string) error {
	if id == "" || usuarioModificador == "" {
		return domain.ErrDatosInvalidos
	}

	_, err := u.catalogoRepo.ObtenerPorID(ctx, id)
	if err != nil {
		return err
	}

	return u.catalogoRepo.ActualizarEstado(ctx, id, false, usuarioModificador)
}

func (u *catalogoUseCase) EliminarCatalogo(ctx context.Context, id string, usuarioModificador string) error {
	if id == "" || usuarioModificador == "" {
		return domain.ErrDatosInvalidos
	}

	_, err := u.catalogoRepo.ObtenerPorID(ctx, id)
	if err != nil {
		return err
	}

	return u.catalogoRepo.ActualizarEstado(ctx, id, false, usuarioModificador)
}
