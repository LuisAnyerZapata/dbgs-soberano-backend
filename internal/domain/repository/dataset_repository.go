package repository

import (
	"context"

	"DBGS_SOBERANO_BACKEND/internal/domain/entity"
)

// DatasetRepository define las operaciones de persistencia para Fuentes y Conjuntos de Datos
type DatasetRepository interface {
	// ObtenerFuentePorID busca una fuente de datos por su ID
	ObtenerFuentePorID(ctx context.Context, id string) (*entity.FuenteDato, error)

	// ListarFuentes retorna todas las fuentes de datos registradas
	ListarFuentes(ctx context.Context) ([]entity.FuenteDato, error)

	// ObtenerDatasetPorID busca un conjunto de datos por su ID
	ObtenerDatasetPorID(ctx context.Context, id string) (*entity.ConjuntoDato, error)

	// ListarDatasets retorna conjuntos de datos con filtros por clasificación, propietario y paginación
	ListarDatasets(ctx context.Context, clasificacion, propietario string, limite, offset int) ([]entity.ConjuntoDato, int64, error)

	// GuardarDataset inserta o actualiza un conjunto de datos
	GuardarDataset(ctx context.Context, dataset *entity.ConjuntoDato) error

	// ActualizarDataset actualiza un conjunto de datos existente
	ActualizarDataset(ctx context.Context, dataset *entity.ConjuntoDato) (*entity.ConjuntoDato, error)
}
