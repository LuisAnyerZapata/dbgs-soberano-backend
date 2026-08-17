package usecase

import (
    "context"
    "time"
    "github.com/google/uuid"

    "DBGS_SOBERANO_BACKEND/internal/application/port"
    "DBGS_SOBERANO_BACKEND/internal/domain"
    "DBGS_SOBERANO_BACKEND/internal/domain/entity"
    "DBGS_SOBERANO_BACKEND/internal/domain/repository"
)

type datasetUseCase struct {
    datasetRepo repository.DatasetRepository
}

// NewDatasetUseCase instancia el caso de uso de datasets inyectando el repositorio
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

// ListarDatasets obtiene conjuntos de datos aplicando filtros opcionales de clasificación
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

// CrearDataset valida la existencia de la fuente de datos padre y crea el nuevo conjunto
func (u *datasetUseCase) CrearDataset(ctx context.Context, dataset *entity.ConjuntoDato) (*entity.ConjuntoDato, error) {
    // Ejecuta el método de validación propio de la entidad (que asumo existe en tu código)
    if err := dataset.EsValido(); err != nil {
        return nil, err
    }

    // Generación soberana del ID en el dominio (igual que en catálogos)
    if dataset.ID == "" {
        dataset.ID = uuid.New().String()
    }

    // Regla de integridad: Verificar que la Fuente de Datos realmente existe
    if dataset.FuenteDatoID != "" {
        if _, err := u.datasetRepo.ObtenerFuentePorID(ctx, dataset.FuenteDatoID); err != nil {
            return nil, err // Retorna ErrEntidadNoEncontrada si la fuente no existe
        }
    }

    // Metadatos de auditoría
    dataset.CreatedAt = time.Now()
    dataset.UpdatedAt = dataset.CreatedAt
    dataset.Estado = true
    
    // Extracción segura del usuario autenticado desde el contexto inyectado por el AuthInterceptor
    if dataset.CreatedBy == "" {
        if usuarioCtx, ok := ctx.Value("user").(*entity.Usuario); ok {
            dataset.CreatedBy = usuarioCtx.Username
        } else {
            dataset.CreatedBy = "system"
        }
    }
    if dataset.UpdatedBy == "" {
        dataset.UpdatedBy = dataset.CreatedBy
    }

    if err := u.datasetRepo.GuardarDataset(ctx, dataset); err != nil {
        return nil, err
    }

    return dataset, nil
}

// ActualizarDataset modifica los campos permitidos de un conjunto de datos
func (u *datasetUseCase) ActualizarDataset(ctx context.Context, dataset *entity.ConjuntoDato) (*entity.ConjuntoDato, error) {
    if dataset.ID == "" || dataset.Nombre == "" {
        return nil, domain.ErrDatosInvalidos
    }

    existente, err := u.datasetRepo.ObtenerDatasetPorID(ctx, dataset.ID)
    if err != nil {
        return nil, err
    }

    // Merge de campos: si no se envían, se mantienen los existentes
    if dataset.FuenteDatoID == "" {
        dataset.FuenteDatoID = existente.FuenteDatoID
    }
    if dataset.PropietarioDato == "" {
        dataset.PropietarioDato = existente.PropietarioDato
    }
    if dataset.Clasificacion == "" {
        dataset.Clasificacion = existente.Clasificacion
    }

    // Metadatos de auditoría
    dataset.UpdatedAt = time.Now()
    if dataset.UpdatedBy == "" {
        // Extracción segura del usuario autenticado desde el contexto
        if usuarioCtx, ok := ctx.Value("user").(*entity.Usuario); ok {
            dataset.UpdatedBy = usuarioCtx.Username
        } else {
            dataset.UpdatedBy = existente.UpdatedBy
        }
    }

    return u.datasetRepo.ActualizarDataset(ctx, dataset)
}