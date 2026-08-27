package grpc

import (
	"context"
	"time"

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

func (h *AuditoriaHandler) ConsultarEventos(ctx context.Context, req *dbgsv1.ConsultarEventosRequest) (*dbgsv1.ConsultarEventosResponse, error) {
	var inicio *time.Time
	var fin *time.Time
	if ts := req.GetFechaInicio(); ts != nil {
		t := ts.AsTime()
		inicio = &t
	}
	if ts := req.GetFechaFin(); ts != nil {
		t := ts.AsTime()
		fin = &t
	}

	input := port.ConsultarAuditoriaInput{
		UsuarioID:   req.GetUsuarioId(),
		Operacion:   req.GetOperacion(),
		Resultado:   "",
		FechaInicio: inicio,
		FechaFin:    fin,
		Limite:      int(req.GetLimit()),
		Offset:      int(req.GetOffset()),
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
