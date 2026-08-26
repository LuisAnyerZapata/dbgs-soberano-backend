package usecase

import (
    "context"
    "encoding/json"
    "fmt"
    "strings"
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

// ActualizarColeccion edita una tabla dinámica: agregar, renombrar, cambiar tipo y eliminar columnas; renombrar tabla.
func (uc *coleccionUseCase) ActualizarColeccion(ctx context.Context, input port.ActualizarColeccionInput) (*port.ActualizarColeccionOutput, error) {
    if _, err := uc.autorizar(ctx, permisoColeccionActualizar); err != nil {
        return nil, err
    }

    if input.Nombre == "" {
        return nil, domain.ErrDatosInvalidos
    }

    tieneOperaciones := len(input.CamposAgregar) > 0 || len(input.CamposRenombrar) > 0 ||
        len(input.CamposTipo) > 0 || len(input.CamposEliminar) > 0 || input.NuevoNombre != ""
    if !tieneOperaciones && input.Descripcion == "" {
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

    // Validar que operaciones destructivas tengan confirmación
    if len(input.CamposEliminar) > 0 && !input.Confirmar {
        return nil, fmt.Errorf("%w: eliminación de columnas requiere confirmar=true", domain.ErrDatosInvalidos)
    }
    if input.NuevoNombre != "" && input.NuevoNombre != input.Nombre && !input.Confirmar {
        return nil, fmt.Errorf("%w: renombrar tabla requiere confirmar=true", domain.ErrDatosInvalidos)
    }

    // Parsear estructura actual
    var camposExistentes []entity.CampoDinamico
    if err := json.Unmarshal(registro.EstructuraJSON, &camposExistentes); err != nil {
        return nil, domain.ErrErrorInterno
    }

    existentes := make(map[string]bool)
    for _, c := range camposExistentes {
        existentes[c.Nombre] = true
    }

    output := &port.ActualizarColeccionOutput{
        ID:           registro.ID,
        NombreLogico: registro.NombreLogico,
    }

    // 1. Renombrar tabla (DDL + metadata)
    nombreFisicoActual := registro.NombreFisico
    if input.NuevoNombre != "" && input.NuevoNombre != input.Nombre {
        sqlRename, err := GenerarSQLRenombrarTabla(nombreFisicoActual, input.NuevoNombre)
        if err != nil {
            return nil, err
        }
        if err := uc.coleccionRepo.EjecutarDDL(ctx, sqlRename); err != nil {
            return nil, err
        }
        nuevoFisico := DBGS_SCHEMA + ".dyn_" + strings.ToLower(input.NuevoNombre)
        if err := uc.coleccionRepo.RenombrarMetadatos(ctx, input.Nombre, input.NuevoNombre, nuevoFisico); err != nil {
            return nil, err
        }
        output.NombreTablaAnterior = nombreFisicoActual
        output.NombreLogico = input.NuevoNombre
        nombreFisicoActual = nuevoFisico
    }

    // 2. Agregar columnas
    var camposNuevos []entity.CampoDinamico
    for _, c := range input.CamposAgregar {
        if !existentes[c.Nombre] {
            camposNuevos = append(camposNuevos, c)
        }
    }
    if len(camposNuevos) > 0 {
        sqlAdd, err := GenerarSQLAgregarColumnas(nombreFisicoActual, camposNuevos)
        if err != nil {
            return nil, err
        }
        if err := uc.coleccionRepo.EjecutarDDL(ctx, sqlAdd); err != nil {
            return nil, err
        }
        output.CamposAgregados = len(camposNuevos)
    }

    // 3. Renombrar columnas
    for _, r := range input.CamposRenombrar {
        sqlRename, err := GenerarSQLRenombrarColumna(nombreFisicoActual, r.NombreActual, r.NombreNuevo)
        if err != nil {
            return nil, err
        }
        if sqlRename == "" {
            continue
        }
        if err := uc.coleccionRepo.EjecutarDDL(ctx, sqlRename); err != nil {
            return nil, err
        }
        output.CamposRenombrados++
    }

    // 4. Cambiar tipo de columnas
    for _, t := range input.CamposTipo {
        sqlType, err := GenerarSQLCambiarTipoColumna(nombreFisicoActual, t.Nombre, entity.FieldType(t.NuevoTipo))
        if err != nil {
            return nil, err
        }
        if err := uc.coleccionRepo.EjecutarDDL(ctx, sqlType); err != nil {
            return nil, err
        }
        output.CamposTipoCambiado++
    }

    // 5. Eliminar columnas (solo con confirmar)
    for _, nombreCol := range input.CamposEliminar {
        sqlDrop, err := GenerarSQUEliminarColumna(nombreFisicoActual, nombreCol)
        if err != nil {
            return nil, err
        }
        if err := uc.coleccionRepo.EjecutarDDL(ctx, sqlDrop); err != nil {
            return nil, err
        }
        output.CamposEliminados++
    }

    // Reconstruir estructura JSON desde la BD real
    // Recargar metadatos si se renombró la tabla
    nombreActualRef := input.Nombre
    if input.NuevoNombre != "" && output.NombreTablaAnterior != "" {
        nombreActualRef = input.NuevoNombre
    }
    registroFinal, err := uc.coleccionRepo.ObtenerMetadatosPorNombre(ctx, nombreActualRef)
    if err != nil {
        return nil, err
    }

    // Reconstruir: partir de existentes, aplicar cambios al array
    var camposActualizados []entity.CampoDinamico
    for _, c := range camposExistentes {
        // Saltar columnas eliminadas
        eliminada := false
        for _, del := range input.CamposEliminar {
            if c.Nombre == del {
                eliminada = true
                break
            }
        }
        if eliminada {
            continue
        }
        // Aplicar renombrados
        for _, rn := range input.CamposRenombrar {
            if c.Nombre == rn.NombreActual {
                c.Nombre = rn.NombreNuevo
                break
            }
        }
        // Aplicar cambio de tipo
        for _, ct := range input.CamposTipo {
            if c.Nombre == ct.Nombre {
                c.Tipo = entity.FieldType(ct.NuevoTipo)
                break
            }
        }
        camposActualizados = append(camposActualizados, c)
    }
    // Agregar los nuevos
    camposActualizados = append(camposActualizados, camposNuevos...)

    estructuraNueva, err := json.Marshal(camposActualizados)
    if err != nil {
        return nil, domain.ErrErrorInterno
    }

    // Actualizar descripción si cambió
    if input.Descripcion != "" {
        registroFinal.Descripcion = input.Descripcion
    }

    if err := uc.coleccionRepo.ActualizarMetadatos(ctx, nombreActualRef, estructuraNueva); err != nil {
        return nil, err
    }

    return output, nil
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
