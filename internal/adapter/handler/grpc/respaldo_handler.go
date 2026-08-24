package grpc

import (
	"context"
	"time"

	pb "DBGS_SOBERANO_BACKEND/api/proto/v1"
	"DBGS_SOBERANO_BACKEND/internal/application/port"
	"DBGS_SOBERANO_BACKEND/internal/domain/entity"
)

type RespaldoHandler struct {
	pb.UnimplementedRespaldoServiceServer
	useCase port.RespaldoPort
}

func NewRespaldoHandler(uc port.RespaldoPort) *RespaldoHandler {
	return &RespaldoHandler{useCase: uc}
}

func (h *RespaldoHandler) CrearRespaldo(ctx context.Context, req *pb.CrearRespaldoRequest) (*pb.CrearRespaldoResponse, error) {
	output, err := h.useCase.CrearRespaldo(ctx, port.CrearRespaldoInput{
		Tipo:          req.GetTipo(),
		RetencionDias: int(req.GetRetencionDias()),
		Detalles:      req.GetDetalles(),
	})
	if err != nil {
		return nil, mapDomainErrorToGRPC(err)
	}
	return &pb.CrearRespaldoResponse{Respaldo: mapearRespaldo(output.Respaldo)}, nil
}

func (h *RespaldoHandler) ObtenerRespaldo(ctx context.Context, req *pb.ObtenerRespaldoRequest) (*pb.ObtenerRespaldoResponse, error) {
	output, err := h.useCase.ObtenerRespaldo(ctx, req.GetId())
	if err != nil {
		return nil, mapDomainErrorToGRPC(err)
	}
	return &pb.ObtenerRespaldoResponse{Respaldo: mapearRespaldo(output.Respaldo)}, nil
}

func (h *RespaldoHandler) ListarRespaldos(ctx context.Context, req *pb.ListarRespaldosRequest) (*pb.ListarRespaldosResponse, error) {
	output, err := h.useCase.ListarRespaldos(ctx, int(req.GetLimite()))
	if err != nil {
		return nil, mapDomainErrorToGRPC(err)
	}

	respaldos := make([]*pb.Respaldo, 0, len(output.Respaldos))
	for i := range output.Respaldos {
		respaldos = append(respaldos, mapearRespaldo(&output.Respaldos[i]))
	}
	return &pb.ListarRespaldosResponse{Respaldos: respaldos, Total: output.Total}, nil
}

func (h *RespaldoHandler) DescargarRespaldo(ctx context.Context, req *pb.DescargarRespaldoRequest) (*pb.DescargarRespaldoResponse, error) {
	output, err := h.useCase.DescargarRespaldo(ctx, port.DescargarRespaldoInput{ID: req.GetId()})
	if err != nil {
		return nil, mapDomainErrorToGRPC(err)
	}
	return &pb.DescargarRespaldoResponse{
		Id:            output.ID,
		NombreArchivo: output.NombreArchivo,
		Contenido:     output.Contenido,
	}, nil
}

func (h *RespaldoHandler) RestaurarRespaldo(ctx context.Context, req *pb.RestaurarRespaldoRequest) (*pb.RestaurarRespaldoResponse, error) {
	output, err := h.useCase.RestaurarRespaldo(ctx, port.RestaurarBackupInput{
		BackupID:  req.GetBackupId(),
		Confirmar: req.GetConfirmar(),
	})
	if err != nil {
		return nil, mapDomainErrorToGRPC(err)
	}
	return &pb.RestaurarRespaldoResponse{Restauracion: mapearRestauracion(output.Restauracion)}, nil
}

func (h *RespaldoHandler) AplicarRetencion(ctx context.Context, req *pb.AplicarRetencionRequest) (*pb.AplicarRetencionResponse, error) {
	output, err := h.useCase.AplicarRetencion(ctx, port.RetencionInput{
		DiasRetencion: int(req.GetDiasRetencion()),
		MaximoBackups: int(req.GetMaximoBackups()),
	})
	if err != nil {
		return nil, mapDomainErrorToGRPC(err)
	}
	return &pb.AplicarRetencionResponse{IdsEliminados: output.IDsEliminados}, nil
}

// =========================================================================================================
// TRADUCTORES proto <-> dominio
// =========================================================================================================

func mapearRespaldo(r *entity.RespaldoOperacion) *pb.Respaldo {
	if r == nil {
		return nil
	}
	return &pb.Respaldo{
		Id:                r.ID,
		Tipo:              r.Tipo,
		Estado:            r.Estado,
		RutaArchivo:       r.RutaArchivo,
		TamanoBytes:       r.TamanoBytes,
		Detalles:          r.Detalles,
		RetencionDias:     int32(r.RetencionDias),
		UsuarioCreador:    r.UsuarioCreador,
		FechaCreacion:     formatearFecha(r.FechaCreacion),
		FechaFinalizacion: formatearFecha(r.FechaFinalizacion),
	}
}

func mapearRestauracion(r *entity.Restauracion) *pb.Restauracion {
	if r == nil {
		return nil
	}
	return &pb.Restauracion{
		Id:            r.ID,
		BackupId:      r.BackupID,
		Usuario:       r.Usuario,
		Estado:        r.Estado,
		Validado:      r.Validado,
		Observaciones: r.Observaciones,
		FechaCreacion: formatearFecha(r.FechaCreacion),
	}
}

func formatearFecha(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format(time.RFC3339)
}
