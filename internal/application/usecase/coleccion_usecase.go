package usecase

import (
    "context"
    "encoding/json"
    "time"
    "database/sql"

    "DBGS_SOBERANO_BACKEND/internal/application/port"
    "DBGS_SOBERANO_BACKEND/internal/domain"
    "DBGS_SOBERANO_BACKEND/internal/domain/entity"
    "DBGS_SOBERANO_BACKEND/internal/domain/repository"

    "github.com/google/uuid"
)

type coleccionUseCase struct {
    coleccionRepo repository.ColeccionDinamicaRepository
}

func NewColeccionUseCase(repo repository.ColeccionDinamicaRepository) port.ColeccionPort {
    return &coleccionUseCase{coleccionRepo: repo}
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

    // 3. Generar el SQL para vincular el trigger de auditoría
    nombreFisico := ObtenerNombreTablaCompleto(input.Nombre)
    sqlTrigger := GenerarSQLTriggerAuditoria(nombreFisico)

    // 4. Ejecutar DDLs en la base de datos
    if err := uc.coleccionRepo.EjecutarDDL(ctx, sqlCreacionTabla); err != nil {
        return nil, err
    }
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