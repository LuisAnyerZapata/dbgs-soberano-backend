package grpc

import (
    "context"

    pb "DBGS_SOBERANO_BACKEND/api/proto/v1"
    "DBGS_SOBERANO_BACKEND/internal/application/port"
)

type SistemaHandler struct {
    pb.UnimplementedSistemaServiceServer
    useCase port.SistemaPort
}

func NewSistemaHandler(uc port.SistemaPort) *SistemaHandler {
    return &SistemaHandler{useCase: uc}
}

func (h *SistemaHandler) GetHealth(ctx context.Context, req *pb.GetHealthRequest) (*pb.GetHealthResponse, error) {
    health, err := h.useCase.ObtenerHealth(ctx)
    if err != nil {
        // Si falla la verificación, asumimos que el sistema está caído
        return &pb.GetHealthResponse{Status: "DOWN"}, nil
    }

    return &pb.GetHealthResponse{
        Status:       health.Estado,
        UptimeSeconds: health.UptimeSegundos,
        Database: &pb.HealthCheck{
            Status:   health.BaseDeDatos.Estado,
            Component: health.BaseDeDatos.Componente,
            LatencyMs: health.BaseDeDatos.LatenciaMs,
        },
    }, nil
}

func (h *SistemaHandler) GetVersion(ctx context.Context, req *pb.GetVersionRequest) (*pb.GetVersionResponse, error) {
    version, err := h.useCase.ObtenerVersion(ctx)
    if err != nil {
        return nil, mapDomainErrorToGRPC(err)
    }

    return &pb.GetVersionResponse{
        Version:     version.Version,
        GitCommit:   version.GitCommit,
        BuildDate:   version.BuildDate,
        GoVersion:   version.GoVersion,
        Environment: version.Environment,
        Engine:      version.Engine,
    }, nil
}