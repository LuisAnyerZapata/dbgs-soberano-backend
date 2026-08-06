package port

import (
	"context"

	"DBGS_SOBERANO_BACKEND/internal/domain/entity"
)

type ObtenerCatalogosInput struct {
	SoloActivos bool
	Limite      int
	Offset      int
}

type ObtenerCatalogosOutput struct {
	Catalogos []entity.Catalogo `json:"catalogos"`
	Total     int64             `json:"total"`
	Limite    int               `json:"limite"`
	Offset    int               `json:"offset"`
}

type CatalogoUseCasePort interface {
	ObtenerPorID(ctx context.Context, id string) (*entity.Catalogo, error)
	ObtenerPorCodigo(ctx context.Context, codigo string) (*entity.Catalogo, error)
	ListarCatalogos(ctx context.Context, input ObtenerCatalogosInput) (*ObtenerCatalogosOutput, error)
	CrearCatalogo(ctx context.Context, catalogo *entity.Catalogo) (*entity.Catalogo, error)
	ActualizarCatalogo(ctx context.Context, catalogo *entity.Catalogo) (*entity.Catalogo, error)
	InactivarCatalogo(ctx context.Context, id string, usuarioModificador string) error
	EliminarCatalogo(ctx context.Context, id string, usuarioModificador string) error
}
