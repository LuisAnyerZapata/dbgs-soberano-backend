package usecase

import (
	"context"
	"time"

	"DBGS_SOBERANO_BACKEND/internal/application/port"
	"DBGS_SOBERANO_BACKEND/internal/domain"
	"DBGS_SOBERANO_BACKEND/internal/domain/entity"
	"DBGS_SOBERANO_BACKEND/internal/domain/repository"
)

type datasetUseCase struct {
	datasetRepo repository.DatasetRepository
}

func NewDatasetUseCase(repo repository.DatasetRepository) port.DatasetPort {
	return &datasetUseCase{
		datasetRepo: repo,
	}
}

func (u *datasetUseCase) ObtenerFuentePorID(ctx context.Context, id string) (*entity.FuenteDato, error) {
	if id == "" {
		return nil, domain.ErrDatosInvalidos
	}
	return u.datasetRepo.ObtenerFuentePorID(ctx, id)
}

func (u *datasetUseCase) ListarFuentes(ctx context.Context) ([]entity.FuenteDato, error) {
	return u.datasetRepo.ListarFuentes(ctx)
}

func (u *datasetUseCase) ObtenerDatasetPorID(ctx context.Context, id string) (*entity.ConjuntoDato, error) {
	if id == "" {
		return nil, domain.ErrDatosInvalidos
	}
	return u.datasetRepo.ObtenerDatasetPorID(ctx, id)
}

func (u *datasetUseCase) ListarDatasets(ctx context.Context, input port.ObtenerDatasetsInput) (*port.ObtenerDatasetsOutput, error) {
	if input.Limite <= 0 {
		input.Limite = 10
	}
	if input.Offset < 0 {
		input.Offset = 0
	}

	datasets, total, err := u.datasetRepo.ListarDatasets(ctx, input.Clasificacion, input.Propietario, input.Limite, input.Offset)
	if err != nil {
		return nil, err
	}

	return &port.ObtenerDatasetsOutput{
		Datasets: datasets,
		Total:    total,
		Limite:   input.Limite,
		Offset:   input.Offset,
	}, nil
}

func (u *datasetUseCase) CrearDataset(ctx context.Context, dataset *entity.ConjuntoDato) (*entity.ConjuntoDato, error) {
	if err := dataset.EsValido(); err != nil {
		return nil, err
	}

	if dataset.FuenteDatoID != "" {
		if _, err := u.datasetRepo.ObtenerFuentePorID(ctx, dataset.FuenteDatoID); err != nil {
			return nil, err
		}
	}

	dataset.CreatedAt = time.Now()
	dataset.UpdatedAt = dataset.CreatedAt
	dataset.Estado = true
	if dataset.CreatedBy == "" {
		dataset.CreatedBy = "system"
	}
	if dataset.UpdatedBy == "" {
		dataset.UpdatedBy = dataset.CreatedBy
	}

	if err := u.datasetRepo.GuardarDataset(ctx, dataset); err != nil {
		return nil, err
	}

	return dataset, nil
}

func (u *datasetUseCase) ActualizarDataset(ctx context.Context, dataset *entity.ConjuntoDato) (*entity.ConjuntoDato, error) {
	if dataset.ID == "" || dataset.Nombre == "" {
		return nil, domain.ErrDatosInvalidos
	}

	existente, err := u.datasetRepo.ObtenerDatasetPorID(ctx, dataset.ID)
	if err != nil {
		return nil, err
	}

	if dataset.FuenteDatoID == "" {
		dataset.FuenteDatoID = existente.FuenteDatoID
	}
	if dataset.PropietarioDato == "" {
		dataset.PropietarioDato = existente.PropietarioDato
	}
	if dataset.Clasificacion == "" {
		dataset.Clasificacion = existente.Clasificacion
	}

	dataset.UpdatedAt = time.Now()
	if dataset.UpdatedBy == "" {
		dataset.UpdatedBy = existente.UpdatedBy
	}

	return u.datasetRepo.ActualizarDataset(ctx, dataset)
}
