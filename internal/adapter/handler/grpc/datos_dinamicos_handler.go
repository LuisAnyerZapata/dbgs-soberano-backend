package grpc

import (
    "context"

    dbgsv1 "DBGS_SOBERANO_BACKEND/api/proto/v1"
    "DBGS_SOBERANO_BACKEND/internal/application/port"

    "google.golang.org/protobuf/types/known/structpb"
)

type DatosDinamicosHandler struct {
    dbgsv1.UnimplementedDatosDinamicosServiceServer
    useCase port.DatosDinamicosPort
}

func NewDatosDinamicosHandler(uc port.DatosDinamicosPort) *DatosDinamicosHandler {
    return &DatosDinamicosHandler{useCase: uc}
}

// mapToStruct convierte un mapa de Go genérico a un Struct de Protobuf de forma segura
func mapToStruct(m map[string]interface{}) (*structpb.Struct, error) {
    if m == nil {
        return structpb.NewStruct(nil) // Se le pasa un mapa nulo para cumplir la firma
    }
    return structpb.NewStruct(m)
}

// mapsToStructs convierte un arreglo de mapas a un arreglo de Structs de Protobuf
func mapsToStructs(maps []map[string]interface{}) ([]*structpb.Struct, error) {
    structs := make([]*structpb.Struct, len(maps))
    for i, m := range maps {
        s, err := mapToStruct(m)
        if err != nil {
            return nil, err
        }
        structs[i] = s
    }
    return structs, nil
}

func (h *DatosDinamicosHandler) ListarRegistros(ctx context.Context, req *dbgsv1.ListarRegistrosRequest) (*dbgsv1.ListarRegistrosResponse, error) {
    input := port.ListarRegistrosInput{
        NombreTabla: req.GetNombreTabla(),
        Limite:      int(req.GetLimit()),
        Offset:      int(req.GetOffset()),
    }

    output, err := h.useCase.ListarRegistros(ctx, input)
    if err != nil {
        return nil, mapDomainErrorToGRPC(err)
    }

    structs, err := mapsToStructs(output.Registros)
    if err != nil {
        return nil, mapDomainErrorToGRPC(err)
    }

    return &dbgsv1.ListarRegistrosResponse{
        Registros: structs,
        Total:     int32(output.Total),
    }, nil
}

func (h *DatosDinamicosHandler) ObtenerRegistro(ctx context.Context, req *dbgsv1.ObtenerRegistroRequest) (*dbgsv1.ObtenerRegistroResponse, error) {
    input := port.ObtenerRegistroInput{
        NombreTabla: req.GetNombreTabla(),
        ID:          req.GetId(),
    }

    registro, err := h.useCase.ObtenerRegistro(ctx, input)
    if err != nil {
        return nil, mapDomainErrorToGRPC(err)
    }

    s, err := mapToStruct(registro)
    if err != nil {
        return nil, mapDomainErrorToGRPC(err)
    }

    return &dbgsv1.ObtenerRegistroResponse{Registro: s}, nil
}

func (h *DatosDinamicosHandler) CrearRegistro(ctx context.Context, req *dbgsv1.CrearRegistroRequest) (*dbgsv1.CrearRegistroResponse, error) {
    // Convertimos el Struct del Proto a un mapa Go usando el método nativo
    datosMap := req.GetDatos().AsMap()
    
    input := port.CrearRegistroInput{
        NombreTabla: req.GetNombreTabla(),
        Datos:       datosMap,
    }

    nuevoID, err := h.useCase.CrearRegistro(ctx, input)
    if err != nil {
        return nil, mapDomainErrorToGRPC(err)
    }

    return &dbgsv1.CrearRegistroResponse{Id: nuevoID}, nil
}

func (h *DatosDinamicosHandler) ActualizarRegistro(ctx context.Context, req *dbgsv1.ActualizarRegistroRequest) (*dbgsv1.ActualizarRegistroResponse, error) {
    datosMap := req.GetDatos().AsMap()

    input := port.ActualizarRegistroInput{
        NombreTabla: req.GetNombreTabla(),
        ID:          req.GetId(),
        Datos:       datosMap,
    }

    registro, err := h.useCase.ActualizarRegistro(ctx, input)
    if err != nil {
        return nil, mapDomainErrorToGRPC(err)
    }

    s, err := mapToStruct(registro)
    if err != nil {
        return nil, mapDomainErrorToGRPC(err)
    }

    return &dbgsv1.ActualizarRegistroResponse{Registro: s}, nil
}

func (h *DatosDinamicosHandler) EliminarRegistro(ctx context.Context, req *dbgsv1.EliminarRegistroRequest) (*dbgsv1.EliminarRegistroResponse, error) {
    input := port.EliminarRegistroInput{
        NombreTabla: req.GetNombreTabla(),
        ID:          req.GetId(),
    }

    err := h.useCase.EliminarRegistro(ctx, input)
    if err != nil {
        return nil, mapDomainErrorToGRPC(err)
    }

    return &dbgsv1.EliminarRegistroResponse{
        Exitoso: true,
        Mensaje: "Registro eliminado correctamente",
    }, nil
}