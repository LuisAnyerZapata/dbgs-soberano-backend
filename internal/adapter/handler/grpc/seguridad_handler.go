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

	result, err := h.seguridadUseCase.Login(ctx, req.Username, req.Password)
	if err != nil {
		return nil, mapDomainErrorToGRPC(err)
	}

	return &pb.LoginResponse{
		AccessToken: result.AccessToken,
		TokenType:   result.TokenType,
		ExpiresIn:   result.ExpiresIn,
	}, nil
}

func (h *SeguridadHandler) ValidarToken(ctx context.Context, req *pb.ValidarTokenRequest) (*pb.ValidarTokenResponse, error) {
	if req.Token == "" {
		return nil, status.Error(codes.InvalidArgument, "el token es obligatorio")
	}

	result, err := h.seguridadUseCase.ValidarToken(ctx, req.Token)
	if err != nil {
		return nil, mapDomainErrorToGRPC(err)
	}

	return &pb.ValidarTokenResponse{EsValido: result.Valid, Username: result.Username, UsuarioId: result.UserID, Rol: result.Rol}, nil
}

func (h *SeguridadHandler) VerificarPermiso(ctx context.Context, req *pb.VerificarPermisoRequest) (*pb.VerificarPermisoResponse, error) {
	if req.UsuarioId == "" || req.Recurso == "" || req.Accion == "" {
		return nil, status.Error(codes.InvalidArgument, "usuario, recurso y acción son obligatorios")
	}

	permitido, err := h.seguridadUseCase.ValidarAcceso(ctx, port.ValidarAccesoInput{Username: req.UsuarioId, Permiso: req.Recurso + ":" + req.Accion})
	if err != nil {
		return nil, mapDomainErrorToGRPC(err)
	}

	return &pb.VerificarPermisoResponse{Permitido: permitido}, nil
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