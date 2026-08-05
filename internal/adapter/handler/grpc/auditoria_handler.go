package grpc

import (
	"context"

	dbgsv1 "DBGS_SOBERANO_BACKEND/api/proto/v1"
	"DBGS_SOBERANO_BACKEND/internal/application/port"
)

type AuditoriaHandler struct {
	dbgsv1.UnimplementedAuditoriaServiceServer
	useCase port.AuditoriaPort
}

func NewAuditoriaHandler(uc port.AuditoriaPort) *AuditoriaHandler {
	return &AuditoriaHandler{
		useCase: uc,
	}
}

func (h *AuditoriaHandler) RegistrarEvento(ctx context.Context, req *dbgsv1.RegistrarEventoRequest) (*dbgsv1.RegistrarEventoResponse, error) {
	input := port.RegistrarEventoInput{
		UsuarioID: req.GetUsuarioId(),
		Username:  req.GetUsername(),
		Operacion: req.GetOperacion(),
		Recurso:   req.GetRecurso(),
		Detalles:  req.GetDetalles(),
		Resultado: req.GetResultado(),
		IPOrigen:  req.GetIpOrigen(),
	}

	evento, err := h.useCase.RegistrarEvento(ctx, input)
	if err != nil {
		return nil, mapDomainErrorToGRPC(err)
	}

	return &dbgsv1.RegistrarEventoResponse{
		EventoId:   evento.ID,
		Registrado: true,
	}, nil
}

func (h *AuditoriaHandler) ConsultarEventos(ctx context.Context, req *dbgsv1.ConsultarEventosRequest) (*dbgsv1.ConsultarEventosResponse, error) {
	input := port.ConsultarAuditoriaInput{
		UsuarioID: req.GetUsuarioId(),
		Resultado: "",
		Limite:    int(req.GetLimit()),
		Offset:    int(req.GetOffset()),
	}

	output, err := h.useCase.ConsultarBitacora(ctx, input)
	if err != nil {
		return nil, mapDomainErrorToGRPC(err)
	}

	var pbEventos []*dbgsv1.EventoAuditoria
	for _, e := range output.Eventos {
		pbEventos = append(pbEventos, &dbgsv1.EventoAuditoria{
			Id:        e.ID,
			UsuarioId: e.UsuarioID,
			Username:  e.Username,
			Operacion: e.Operacion,
			Recurso:   e.Recurso,
			Detalles:  e.Detalles,
			Resultado: e.Resultado,
			IpOrigen:  e.IPOrigen,
		})
	}

	return &dbgsv1.ConsultarEventosResponse{
		Eventos: pbEventos,
		Total:   int32(output.Total),
	}, nil
}