package usecase

import (
    "context"
    "encoding/json"
    "fmt"

    "DBGS_SOBERANO_BACKEND/internal/application/port"
    "DBGS_SOBERANO_BACKEND/internal/domain"
    "DBGS_SOBERANO_BACKEND/internal/domain/entity"
    "DBGS_SOBERANO_BACKEND/internal/domain/repository"
)

type datosDinamicosUseCase struct {
    coleccionRepo repository.ColeccionDinamicaRepository
    datosRepo     repository.DatosDinamicosRepository
}

func NewDatosDinamicosUseCase(coleccionRepo repository.ColeccionDinamicaRepository, datosRepo repository.DatosDinamicosRepository) port.DatosDinamicosPort {
    return &datosDinamicosUseCase{
        coleccionRepo: coleccionRepo,
        datosRepo:     datosRepo,
    }
}

// validarEstructuraYCampos es el método privado que garantiza que nadie pueda inyectar 
// campos falsos o maliciosos que no estén en el diccionario de datos.
func (uc *datosDinamicosUseCase) validarEstructuraYCampos(nombreLogico string, payload map[string]interface{}) (*entity.ColeccionRegistro, []entity.CampoDinamico, error) {
    // 1. Buscar en el diccionario de datos (Metadatos) por nombre lógico exacto
    registro, err := uc.coleccionRepo.ObtenerMetadatosPorNombre(context.Background(), nombreLogico)
    if err != nil {
        return nil, nil, err
    }

    if !registro.EstaActiva {
        return nil, nil, domain.ErrEntidadNoEncontrada
    }

    // 2. Deserializar el JSONB de la estructura permitida
    var camposPermitidos []entity.CampoDinamico
    if err := json.Unmarshal(registro.EstructuraJSON, &camposPermitidos); err != nil {
        return nil, nil, domain.ErrErrorInterno
    }

    // 3. Validar que los campos del payload estén permitidos
    camposSoberanos := map[string]bool{"id": true, "created_at": true, "updated_at": true, "created_by": true, "updated_by": true}
    for key := range payload {
        if camposSoberanos[key] {
            continue // Permitir que pase si la UI los envía, el repo los ignorará
        }
        permitido := false
        for _, campo := range camposPermitidos {
            if campo.Nombre == key {
                permitido = true
                break
            }
        }
        if !permitido {
            // Se envuelve el error de dominio para que el handler lo mapee como InvalidArgument (y no Internal)
            return nil, nil, fmt.Errorf("%w: el campo '%s' no está definido en la estructura de la tabla '%s'", domain.ErrDatosInvalidos, key, nombreLogico)
        }
    }

    return registro, camposPermitidos, nil
}

func (uc *datosDinamicosUseCase) ListarRegistros(ctx context.Context, input port.ListarRegistrosInput) (*port.ListarRegistrosOutput, error) {
    if input.NombreTabla == "" {
        return nil, domain.ErrDatosInvalidos
    }
    if input.Limite <= 0 || input.Limite > 100 {
        input.Limite = 20
    }

    // Validamos que la tabla exista (el payload es nil en un GET list)
    _, _, err := uc.validarEstructuraYCampos(input.NombreTabla, nil)
    if err != nil {
        return nil, err
    }

    nombreFisico := "dyn_" + input.NombreTabla
    registros, total, err := uc.datosRepo.Listar(ctx, nombreFisico, input.Limite, input.Offset)
    if err != nil {
        return nil, err
    }

    return &port.ListarRegistrosOutput{Registros: registros, Total: total}, nil
}

func (uc *datosDinamicosUseCase) ObtenerRegistro(ctx context.Context, input port.ObtenerRegistroInput) (map[string]interface{}, error) {
    if input.NombreTabla == "" || input.ID == "" {
        return nil, domain.ErrDatosInvalidos
    }

    _, _, err := uc.validarEstructuraYCampos(input.NombreTabla, nil)
    if err != nil {
        return nil, err
    }

    nombreFisico := "dyn_" + input.NombreTabla
    return uc.datosRepo.ObtenerPorID(ctx, nombreFisico, input.ID)
}

func (uc *datosDinamicosUseCase) CrearRegistro(ctx context.Context, input port.CrearRegistroInput) (string, error) {
    if input.NombreTabla == "" || len(input.Datos) == 0 {
        return "", domain.ErrDatosInvalidos // Cambiado de nil a ""
    }

    // Validar estructura Y campos del payload
    _, _, err := uc.validarEstructuraYCampos(input.NombreTabla, input.Datos)
    if err != nil {
        return "", err // Cambiado de nil a ""
    }

    nombreFisico := "dyn_" + input.NombreTabla
    createdBy := "system"
    if usuarioCtx, ok := ctx.Value(domain.CtxKeyUsuario).(*entity.Usuario); ok {
        createdBy = usuarioCtx.Username
    }

    return uc.datosRepo.Insertar(ctx, nombreFisico, input.Datos, createdBy)
}

func (uc *datosDinamicosUseCase) ActualizarRegistro(ctx context.Context, input port.ActualizarRegistroInput) (map[string]interface{}, error) {
    if input.NombreTabla == "" || input.ID == "" || len(input.Datos) == 0 {
        return nil, domain.ErrDatosInvalidos
    }

    _, _, err := uc.validarEstructuraYCampos(input.NombreTabla, input.Datos)
    if err != nil {
        return nil, err
    }

    nombreFisico := "dyn_" + input.NombreTabla
    updatedBy := "system"
    if usuarioCtx, ok := ctx.Value(domain.CtxKeyUsuario).(*entity.Usuario); ok {
        updatedBy = usuarioCtx.Username
    }

    return uc.datosRepo.Actualizar(ctx, nombreFisico, input.ID, input.Datos, updatedBy)
}

func (uc *datosDinamicosUseCase) EliminarRegistro(ctx context.Context, input port.EliminarRegistroInput) error {
    if input.NombreTabla == "" || input.ID == "" {
        return domain.ErrDatosInvalidos
    }

    _, _, err := uc.validarEstructuraYCampos(input.NombreTabla, nil)
    if err != nil {
        return err
    }

    nombreFisico := "dyn_" + input.NombreTabla
    return uc.datosRepo.Eliminar(ctx, nombreFisico, input.ID)
}