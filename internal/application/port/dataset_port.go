package port

import (
	"context"

	"DBGS_SOBERANO_BACKEND/internal/domain/entity"
)

type ObtenerDatasetsInput struct {
	Clasificacion string
	Propietario   string
	Limite        int
	Offset        int
}

type ObtenerDatasetsOutput struct {
	Datasets []entity.ConjuntoDato `json:"datasets"`
	Total    int64                 `json:"total"`
	Limite   int                   `json:"limite"`
	Offset   int                   `json:"offset"`
}

type DatasetPort interface {
	ObtenerFuentePorID(ctx context.Context, id string) (*entity.FuenteDato, error)
	ListarFuentes(ctx context.Context) ([]entity.FuenteDato, error)
	ObtenerDatasetPorID(ctx context.Context, id string) (*entity.ConjuntoDato, error)
	ListarDatasets(ctx context.Context, input ObtenerDatasetsInput) (*ObtenerDatasetsOutput, error)
	CrearDataset(ctx context.Context, dataset *entity.ConjuntoDato) (*entity.ConjuntoDato, error)
	ActualizarDataset(ctx context.Context, dataset *entity.ConjuntoDato) (*entity.ConjuntoDato, error)
}
