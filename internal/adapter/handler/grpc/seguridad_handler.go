package grpc

import (
	"context"

	pb "DBGS_SOBERANO_BACKEND/api/proto/v1"
	"DBGS_SOBERANO_BACKEND/internal/application/port"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type SeguridadHandler struct {
	pb.UnimplementedSeguridadServiceServer
	seguridadUseCase port.SeguridadUseCasePort
}

func NewSeguridadHandler(seguridadUC port.SeguridadUseCasePort) *SeguridadHandler {
	return &SeguridadHandler{seguridadUseCase: seguridadUC}
}

func (h *SeguridadHandler) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	if req.Username == "" || req.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "el usuario y la contraseña son obligatorios")
	}

	// En un flujo real, aquí se autentica y se genera el JWT mediante el caso de uso
	// Por ahora resolvemos la llamada a la capa de aplicación:
	usuario, err := h.seguridadUseCase.ObtenerPerfilUsuario(ctx, req.Username)
	if err != nil {
		return nil, mapDomainErrorToGRPC(err)
	}

	if !usuario.Estado {
		return nil, status.Error(codes.PermissionDenied, "el usuario se encuentra inactivo")
	}

	return &pb.LoginResponse{
		AccessToken: "jwt_token_simulado_dbgs",
		TokenType:   "Bearer",
		ExpiresIn:   86400, // 24 horas
	}, nil
}

func (h *SeguridadHandler) ValidarToken(ctx context.Context, req *pb.ValidarTokenRequest) (*pb.ValidarTokenResponse, error) {
	return &pb.ValidarTokenResponse{EsValido: true, Username: "demo"}, nil
}

func (h *SeguridadHandler) VerificarPermiso(ctx context.Context, req *pb.VerificarPermisoRequest) (*pb.VerificarPermisoResponse, error) {
	return &pb.VerificarPermisoResponse{Permitido: true}, nil
}

func (h *SeguridadHandler) ObtenerPerfil(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	if req.Username == "" {
		return nil, status.Error(codes.InvalidArgument, "el nombre de usuario es requerido")
	}

	usuario, err := h.seguridadUseCase.ObtenerPerfilUsuario(ctx, req.Username)
	if err != nil {
		return nil, mapDomainErrorToGRPC(err)
	}

	return &pb.LoginResponse{
		AccessToken: usuario.Username,
		TokenType:   "Bearer",
		ExpiresIn:   86400,
	}, nil
}