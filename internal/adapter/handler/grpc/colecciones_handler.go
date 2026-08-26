package grpc

import (
    "context"
    "time"

    pb "DBGS_SOBERANO_BACKEND/api/proto/v1"
    "DBGS_SOBERANO_BACKEND/internal/application/port"
    "DBGS_SOBERANO_BACKEND/internal/domain/entity"


)

type ColeccionesHandler struct {
    pb.UnimplementedColeccionesServiceServer
    useCase port.ColeccionPort
}

func NewColeccionesHandler(uc port.ColeccionPort) *ColeccionesHandler {
    return &ColeccionesHandler{useCase: uc}
}

func (h *ColeccionesHandler) CrearColeccion(ctx context.Context, req *pb.CrearColeccionRequest) (*pb.CrearColeccionResponse, error) {
    // 1. Mapear del Proto a la Entidad de Dominio
    var camposDominio []entity.CampoDinamico
    for _, c := range req.Campos {
        camposDominio = append(camposDominio, entity.CampoDinamico{
            Nombre:     c.GetNombre(),
            Tipo:       entity.FieldType(c.GetTipo()),
            Nulo:       c.GetNulo(),
            Unico:      c.GetUnico(),
            Descripcion: c.GetDescripcion(),
        })
    }

    input := port.CrearColeccionInput{
        Nombre:       req.GetNombre(),
        Descripcion: req.GetDescripcion(),
        Campos:       camposDominio,
        InstitucionID: req.GetInstitucionId(),
    }

    // 2. Ejecutar la lógica de negocio (Fase 2)
    registro, err := h.useCase.CrearColeccion(ctx, input)
    if err != nil {
        return nil, mapDomainErrorToGRPC(err)
    }

    return &pb.CrearColeccionResponse{
        Id:           registro.ID,
        NombreLogico: registro.NombreLogico,
        NombreFisico: registro.NombreFisico,
        Mensaje:      "Colección creada y tabla dinámica generada exitosamente",
    }, nil
}

func (h *ColeccionesHandler) ListarColecciones(ctx context.Context, req *pb.ListarColeccionesRequest) (*pb.ListarColeccionesResponse, error) {
    output, err := h.useCase.ListarColecciones(ctx, int(req.GetLimit()), int(req.GetOffset()))
    if err != nil {
        return nil, mapDomainErrorToGRPC(err)
    }

    var pbColecciones []*pb.ColeccionRegistroProto
    for _, c := range output.Colecciones {
        pbColecciones = append(pbColecciones, &pb.ColeccionRegistroProto{
            Id:             c.ID,
            NombreLogico:   c.NombreLogico,
            NombreFisico:   c.NombreFisico,
            Descripcion:    c.Descripcion,
            EstructuraJson: string(c.EstructuraJSON),
            EstaActiva:     c.EstaActiva,
            CreatedAt:      c.CreatedAt.Format(time.RFC3339),
        })
    }

    return &pb.ListarColeccionesResponse{
        Colecciones: pbColecciones,
        Total:       int32(output.Total),
    }, nil
}

func (h *ColeccionesHandler) ActualizarColeccion(ctx context.Context, req *pb.ActualizarColeccionRequest) (*pb.ActualizarColeccionResponse, error) {
    var camposDominio []entity.CampoDinamico
    for _, c := range req.Campos {
        camposDominio = append(camposDominio, entity.CampoDinamico{
            Nombre:      c.GetNombre(),
            Tipo:        entity.FieldType(c.GetTipo()),
            Nulo:        c.GetNulo(),
            Unico:       c.GetUnico(),
            Descripcion: c.GetDescripcion(),
        })
    }

    input := port.ActualizarColeccionInput{
        Nombre: req.GetNombre(),
        Campos: camposDominio,
    }

    output, err := h.useCase.ActualizarColeccion(ctx, input)
    if err != nil {
        return nil, mapDomainErrorToGRPC(err)
    }

    return &pb.ActualizarColeccionResponse{
        Id:             output.ID,
        NombreLogico:   output.NombreLogico,
        CamposAgregados: int32(output.CamposAgregados),
        Mensaje:        "Colección actualizada exitosamente",
    }, nil
}

func (h *ColeccionesHandler) EliminarColeccion(ctx context.Context, req *pb.EliminarColeccionRequest) (*pb.EliminarColeccionResponse, error) {
    input := port.EliminarColeccionInput{
        Nombre:   req.GetNombre(),
        Confirmar: req.GetConfirmar(),
    }

    output, err := h.useCase.EliminarColeccion(ctx, input)
    if err != nil {
        return nil, mapDomainErrorToGRPC(err)
    }

    var mensaje string
    if output.AccionRealizada == "eliminada" {
        mensaje = "Colección eliminada permanentemente"
    } else {
        mensaje = "Colección desactivada (tabla física preservada)"
    }

    return &pb.EliminarColeccionResponse{
        NombreLogico:    output.NombreLogico,
        AccionRealizada: output.AccionRealizada,
        Mensaje:         mensaje,
    }, nil
}