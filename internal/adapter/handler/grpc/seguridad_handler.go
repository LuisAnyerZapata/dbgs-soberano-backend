package grpc

import (
    "context"

    pb "DBGS_SOBERANO_BACKEND/api/proto/v1"
    "DBGS_SOBERANO_BACKEND/internal/application/port"

    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"
)

// SeguridadHandler implementa los métodos gRPC para el servicio de seguridad
type SeguridadHandler struct {
    pb.UnimplementedSeguridadServiceServer
    seguridadUseCase port.SeguridadUseCasePort
}

// NewSeguridadHandler inyecta el caso de uso de seguridad en el adaptador de entrada
func NewSeguridadHandler(seguridadUC port.SeguridadUseCasePort) *SeguridadHandler {
    return &SeguridadHandler{seguridadUseCase: seguridadUC}
}

// =========================================================================================================
// ENDPOINTS DE BOOTSTRAPPING (SETUP)
// =========================================================================================================

// GetSetupStatus permite a la interfaz saber si debe renderizar la pantalla de bienvenida
func (h *SeguridadHandler) GetSetupStatus(ctx context.Context, req *pb.GetSetupStatusRequest) (*pb.GetSetupStatusResponse, error) {
    statusResult, err := h.seguridadUseCase.VerificarEstadoSetup(ctx)
    if err != nil {
        return nil, mapDomainErrorToGRPC(err)
    }

    return &pb.GetSetupStatusResponse{Initialized: statusResult.Inicializado}, nil
}

// CreateSetup orquesta la creación del primer superadministrador de forma segura
func (h *SeguridadHandler) CreateSetup(ctx context.Context, req *pb.CreateSetupRequest) (*pb.LoginResponse, error) {
    // Sanitización básica a nivel de API
    if req.GetUsername() == "" || req.GetPassword() == "" {
        return nil, status.Error(codes.InvalidArgument, "username y password son obligatorios para el setup")
    }

    input := port.EjecutarSetupInput{
        Username: req.GetUsername(),
        Password: req.GetPassword(),
        Email:    req.GetEmail(),
    }

    result, err := h.seguridadUseCase.EjecutarSetup(ctx, input)
    if err != nil {
        return nil, mapDomainErrorToGRPC(err)
    }

    return &pb.LoginResponse{
        AccessToken: result.AccessToken,
        TokenType:   result.TokenType,
        ExpiresIn:   result.ExpiresIn,
    }, nil
}

// =========================================================================================================
// ENDPOINTS DE AUTENTICACIÓN CONTINUA
// =========================================================================================================

// Login autentica a un usuario y devuelve un JWT
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

// ValidarToken verifica la vigencia de un token JWT
func (h *SeguridadHandler) ValidarToken(ctx context.Context, req *pb.ValidarTokenRequest) (*pb.ValidarTokenResponse, error) {
    if req.Token == "" {
        return nil, status.Error(codes.InvalidArgument, "el token es obligatorio")
    }

    result, err := h.seguridadUseCase.ValidarToken(ctx, req.Token)
    if err != nil {
        return nil, mapDomainErrorToGRPC(err)
    }

    return &pb.ValidarTokenResponse{
        EsValido:  result.Valid,
        Username:  result.Username,
        UsuarioId: result.UserID,
        Rol:       result.Rol,
    }, nil
}

// VerificarPermiso comprueba si un usuario tiene acceso a un recurso específico
func (h *SeguridadHandler) VerificarPermiso(ctx context.Context, req *pb.VerificarPermisoRequest) (*pb.VerificarPermisoResponse, error) {
    if req.UsuarioId == "" || req.Recurso == "" || req.Accion == "" {
        return nil, status.Error(codes.InvalidArgument, "usuario, recurso y acción son obligatorios")
    }

    permitido, err := h.seguridadUseCase.ValidarAcceso(ctx, port.ValidarAccesoInput{
        Username: req.UsuarioId,
        Permiso:  req.Recurso + ":" + req.Accion,
    })
    if err != nil {
        return nil, mapDomainErrorToGRPC(err)
    }

    return &pb.VerificarPermisoResponse{Permitido: permitido}, nil
}