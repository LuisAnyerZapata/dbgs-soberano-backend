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

type catalogoUseCase struct {
    catalogoRepo repository.CatalogoRepository
}

// NewCatalogoUseCase instancia el caso de uso de catálogos inyectando el repositorio
func NewCatalogoUseCase(repo repository.CatalogoRepository) port.CatalogoUseCasePort {
    return &catalogoUseCase{
        catalogoRepo: repo,
    }
}

// ObtenerPorID busca un catálogo específico por su UUID
func (u *catalogoUseCase) ObtenerPorID(ctx context.Context, id string) (*entity.Catalogo, error) {
    if id == "" {
        return nil, domain.ErrDatosInvalidos
    }
    return u.catalogoRepo.ObtenerPorID(ctx, id)
}

// ObtenerPorCodigo busca un catálogo específico por su código de negocio
func (u *catalogoUseCase) ObtenerPorCodigo(ctx context.Context, codigo string) (*entity.Catalogo, error) {
    if codigo == "" {
        return nil, domain.ErrDatosInvalidos
    }
    return u.catalogoRepo.ObtenerPorCodigo(ctx, codigo)
}

// ListarCatalogos obtiene una lista paginada de catálogos, aplicando filtros de activos si es necesario
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

// CrearCatalogo valida e inserta un nuevo catálogo en el sistema
func (u *catalogoUseCase) CrearCatalogo(ctx context.Context, catalogo *entity.Catalogo) (*entity.Catalogo, error) {
    if catalogo.Codigo == "" || catalogo.Nombre == "" {
        return nil, domain.ErrDatosInvalidos
    }

    // Prevención de duplicados a nivel de aplicación
    existente, _ := u.catalogoRepo.ObtenerPorCodigo(ctx, catalogo.Codigo)
    if existente != nil {
        return nil, domain.ErrCodigoDuplicado
    }

    // --- NUEVO: Generación soberana del ID en el dominio ---
    // Si no se proporcionó un ID desde fuera, la aplicación lo genera.
    // Esto evita depender de DEFAULTS de la base de datos y problemas de tipos NULL.
    if catalogo.ID == "" {
        catalogo.ID = uuid.New().String()
    }

    // Metadatos de auditoría
    catalogo.CreatedAt = time.Now()
    catalogo.UpdatedAt = catalogo.CreatedAt
    catalogo.Estado = true
    
    if catalogo.CreatedBy == "" {
        if usuarioCtx, ok := ctx.Value("user").(*entity.Usuario); ok {
            catalogo.CreatedBy = usuarioCtx.Username
        } else {
            catalogo.CreatedBy = "system"
        }
    }
    catalogo.UpdatedBy = catalogo.CreatedBy

    if err := u.catalogoRepo.Guardar(ctx, catalogo); err != nil {
        return nil, err
    }

    return catalogo, nil
}

// ActualizarCatalogo modifica los campos permitidos de un catálogo existente
func (u *catalogoUseCase) ActualizarCatalogo(ctx context.Context, catalogo *entity.Catalogo) (*entity.Catalogo, error) {
    if catalogo.ID == "" || catalogo.Nombre == "" {
        return nil, domain.ErrDatosInvalidos
    }

    existente, err := u.catalogoRepo.ObtenerPorID(ctx, catalogo.ID)
    if err != nil {
        return nil, err
    }

    // Validación de cambio de código para evitar duplicados
    if catalogo.Codigo == "" {
        catalogo.Codigo = existente.Codigo
    } else if existente.Codigo != catalogo.Codigo {
        otro, _ := u.catalogoRepo.ObtenerPorCodigo(ctx, catalogo.Codigo)
        if otro != nil {
            return nil, domain.ErrCodigoDuplicado
        }
    }

    // Metadatos de auditoría
    catalogo.UpdatedAt = time.Now()
    if catalogo.UpdatedBy == "" {
        // Extracción segura del usuario autenticado desde el contexto
        if usuarioCtx, ok := ctx.Value("user").(*entity.Usuario); ok {
            catalogo.UpdatedBy = usuarioCtx.Username
        } else {
            catalogo.UpdatedBy = existente.UpdatedBy
        }
    }

    return u.catalogoRepo.Actualizar(ctx, catalogo)
}

// InactivarCatalogo realiza una baja lógica del catálogo (cambia estado a false)
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

// EliminarCatalogo realiza una baja lógica del catálogo (seguridad soberana: se prefiere inactivar)
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