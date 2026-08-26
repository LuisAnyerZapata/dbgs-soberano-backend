package usecase

import (
    "context"
    "encoding/json"
    "fmt"
    "time"
    "database/sql"

    "DBGS_SOBERANO_BACKEND/internal/application/port"
    "DBGS_SOBERANO_BACKEND/internal/domain"
    "DBGS_SOBERANO_BACKEND/internal/domain/entity"
    "DBGS_SOBERANO_BACKEND/internal/domain/repository"

    "github.com/google/uuid"
)

const (
    permisoColeccionActualizar = "colecciones:actualizar"
    permisoColeccionEliminar   = "colecciones:eliminar"
)

type coleccionUseCase struct {
    coleccionRepo repository.ColeccionDinamicaRepository
    seguridad     repository.SeguridadRepository
}

func NewColeccionUseCase(repo repository.ColeccionDinamicaRepository, seguridad repository.SeguridadRepository) port.ColeccionPort {
    return &coleccionUseCase{coleccionRepo: repo, seguridad: seguridad}
}

// CrearColeccion orquesta todo el proceso de creación dinámica
func (uc *coleccionUseCase) CrearColeccion(ctx context.Context, input port.CrearColeccionInput) (*entity.ColeccionRegistro, error) {
    // 1. Validaciones de negocio iniciales
    if input.Nombre == "" || len(input.Campos) == 0 {
        return nil, domain.ErrDatosInvalidos
    }

    // 2. Generar el DDL de la tabla de forma segura
    definicion := &entity.ColeccionDefinicion{
        Nombre:       input.Nombre,
        Descripcion: input.Descripcion,
        Campos:       input.Campos,
        InstitucionID: input.InstitucionID,
    }

    sqlCreacionTabla, err := GenerarSQLCreacionTabla(definicion)
    if err != nil {
        return nil, err
    }

    // 3. Generar el SQL para vincular el trigger de auditoría forense (Dominio 7)
    nombreFisico := ObtenerNombreTablaCompleto(input.Nombre)
    sqlTrigger := GenerarSQLTriggerAuditoria(nombreFisico)

    // 4. Ejecutar DDLs en la base de datos
    if err := uc.coleccionRepo.EjecutarDDL(ctx, sqlCreacionTabla); err != nil {
        return nil, err
    }
    // Toda tabla dinámica nace auditada: INSERT/UPDATE/DELETE quedan registrados
    // en la bitácora inmutable sin intervención posterior.
    if err := uc.coleccionRepo.EjecutarDDL(ctx, sqlTrigger); err != nil {
        return nil, err
    }

    // 5. Serializar la estructura a JSON
    estructuraJSON, err := json.Marshal(input.Campos)
    if err != nil {
        return nil, domain.ErrErrorInterno
    }

    // 6. Extraer usuario que ejecuta la acción
    createdBy := "system"
    if usuarioCtx, ok := ctx.Value("user").(*entity.Usuario); ok {
        createdBy = usuarioCtx.Username
    }

    // 7. Manejo seguro de IDs nulos para PostgreSQL usando sql.NullString
    institucionIDSeguro := sql.NullString{
        String: input.InstitucionID,
        Valid:  input.InstitucionID != "", // Si es vacío, Valid=false y PQ enviará NULL a la BD
    }

    // 8. Guardar metadatos en el diccionario
    registro := &entity.ColeccionRegistro{
        ID:             uuid.New().String(),
        NombreLogico:   input.Nombre,
        NombreFisico:   nombreFisico,
        Descripcion:    input.Descripcion,
        InstitucionID:  institucionIDSeguro,
        EstructuraJSON: estructuraJSON,
        EstaActiva:     true,
        CreatedAt:      time.Now(),
        CreatedBy:      createdBy,
    }

    // Sobreescribimos solo el campo problemático antes de enviarlo a la BD
    registro.InstitucionID = institucionIDSeguro

    if err := uc.coleccionRepo.GuardarMetadatos(ctx, registro); err != nil {
        return nil, err
    }

    return registro, nil
}

// ListarColecciones obtiene el inventario de tablas dinámicas
func (uc *coleccionUseCase) ListarColecciones(ctx context.Context, limite, offset int) (*port.ListarColeccionesOutput, error) {
    if limite <= 0 {
        limite = 20
    }
    colecciones, total, err := uc.coleccionRepo.ListarMetadatos(ctx, limite, offset)
    if err != nil {
        return nil, err
    }
    return &port.ListarColeccionesOutput{Colecciones: colecciones, Total: total}, nil
}

// ActualizarColeccion agrega nuevas columnas a una tabla dinámica existente (solo aditivo)
func (uc *coleccionUseCase) ActualizarColeccion(ctx context.Context, input port.ActualizarColeccionInput) (*port.ActualizarColeccionOutput, error) {
    if _, err := uc.autorizar(ctx, permisoColeccionActualizar); err != nil {
        return nil, err
    }

    if input.Nombre == "" || len(input.Campos) == 0 {
        return nil, domain.ErrDatosInvalidos
    }

    // Cargar metadatos existentes
    registro, err := uc.coleccionRepo.ObtenerMetadatosPorNombre(ctx, input.Nombre)
    if err != nil {
        return nil, err
    }

    if !registro.EstaActiva {
        return nil, domain.ErrRegistroInactivo
    }

    // Parsear estructura actual para identificar campos existentes
    var camposExistentes []entity.CampoDinamico
    if err := json.Unmarshal(registro.EstructuraJSON, &camposExistentes); err != nil {
        return nil, domain.ErrErrorInterno
    }

    existentes := make(map[string]bool)
    for _, c := range camposExistentes {
        existentes[c.Nombre] = true
    }

    // Filtrar solo campos nuevos
    var camposNuevos []entity.CampoDinamico
    for _, c := range input.Campos {
        if !existentes[c.Nombre] {
            camposNuevos = append(camposNuevos, c)
        }
    }

    if len(camposNuevos) == 0 {
        return &port.ActualizarColeccionOutput{
            ID: registro.ID, NombreLogico: registro.NombreLogico, CamposAgregados: 0,
        }, nil
    }

    // Generar y ejecutar DDL
    sqlAlter, err := GenerarSQLAgregarColumnas(registro.NombreFisico, camposNuevos)
    if err != nil {
        return nil, err
    }

    if err := uc.coleccionRepo.EjecutarDDL(ctx, sqlAlter); err != nil {
        return nil, err
    }

    // Merge de estructura JSON
    camposActualizados := append(camposExistentes, camposNuevos...)
    estructuraNueva, err := json.Marshal(camposActualizados)
    if err != nil {
        return nil, domain.ErrErrorInterno
    }

    if err := uc.coleccionRepo.ActualizarMetadatos(ctx, input.Nombre, estructuraNueva); err != nil {
        return nil, err
    }

    return &port.ActualizarColeccionOutput{
        ID:              registro.ID,
        NombreLogico:    registro.NombreLogico,
        CamposAgregados: len(camposNuevos),
    }, nil
}

// EliminarColeccion elimina o desactiva una colección dinámica
func (uc *coleccionUseCase) EliminarColeccion(ctx context.Context, input port.EliminarColeccionInput) (*port.EliminarColeccionOutput, error) {
    if _, err := uc.autorizar(ctx, permisoColeccionEliminar); err != nil {
        return nil, err
    }

    if input.Nombre == "" {
        return nil, domain.ErrDatosInvalidos
    }

    registro, err := uc.coleccionRepo.ObtenerMetadatosPorNombre(ctx, input.Nombre)
    if err != nil {
        return nil, err
    }

    if input.Confirmar {
        // Hard delete: DROP TABLE + DELETE metadata
        sqlDrop := fmt.Sprintf("DROP TABLE IF EXISTS %s", registro.NombreFisico)
        if err := uc.coleccionRepo.EjecutarDDL(ctx, sqlDrop); err != nil {
            return nil, err
        }
        if err := uc.coleccionRepo.EliminarMetadatos(ctx, input.Nombre); err != nil {
            return nil, err
        }
        return &port.EliminarColeccionOutput{
            NombreLogico:    input.Nombre,
            AccionRealizada: "eliminada",
        }, nil
    }

    // Soft delete: desactivar sin tocar la tabla física
    if err := uc.coleccionRepo.DesactivarMetadatos(ctx, input.Nombre); err != nil {
        return nil, err
    }
    return &port.EliminarColeccionOutput{
        NombreLogico:    input.Nombre,
        AccionRealizada: "desactivada",
    }, nil
}

// autorizar valida sesión y permiso granular.
func (uc *coleccionUseCase) autorizar(ctx context.Context, permiso string) (*entity.Usuario, error) {
    usuario, ok := ctx.Value("user").(*entity.Usuario)
    if !ok || usuario == nil {
        return nil, fmt.Errorf("%w: el dominio de colecciones exige un usuario autenticado", domain.ErrAccesoNoAutorizado)
    }
    permitido, err := uc.seguridad.ValidarPermiso(ctx, usuario.RolID, permiso)
    if err != nil {
        return nil, domain.ErrErrorInterno
    }
    if !permitido {
        return nil, fmt.Errorf("%w: su rol no posee el permiso '%s'", domain.ErrAccesoNoAutorizado, permiso)
    }
    return usuario, nil
}
