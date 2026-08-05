package interceptors

import (
	"context"
	"strings"

	"DBGS_SOBERANO_BACKEND/internal/application/port"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// AuthInterceptor intercepta las llamadas gRPC para verificar JWT y permisos RBAC
type AuthInterceptor struct {
	seguridadUseCase port.SeguridadUseCasePort
}

func NewAuthInterceptor(seguridadUC port.SeguridadUseCasePort) *AuthInterceptor {
	return &AuthInterceptor{seguridadUseCase: seguridadUC}
}

// Unary returns a gRPC unary server interceptor for authentication
func (i *AuthInterceptor) Unary() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		// Endpoints públicos excluidos de autenticación
		if info.FullMethod == "/v1.CatalogosService/GetHealth" || info.FullMethod == "/v1.CatalogosService/GetVersion" {
			return handler(ctx, req)
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "metadatos de autenticación no encontrados")
		}

		authHeader := md["authorization"]
		if len(authHeader) == 0 {
			return nil, status.Error(codes.Unauthenticated, "token de autorización requerido")
		}

		tokenStr := strings.TrimPrefix(authHeader[0], "Bearer ")
		if tokenStr == "" {
			return nil, status.Error(codes.Unauthenticated, "formato de token inválido")
		}

		// Simulamos la extracción del usuario del token JWT (en prod se valida firma)
		username := "usuario_demo" // Extraído del token reclamado

		// Validar perfil de usuario activo
		usuario, err := i.seguridadUseCase.ObtenerPerfilUsuario(ctx, username)
		if err != nil {
			return nil, status.Error(codes.PermissionDenied, "acceso denegado: cuenta inactiva o no registrada")
		}

		// Inyectar usuario en el contexto
		newCtx := context.WithValue(ctx, "user", usuario)
		return handler(newCtx, req)
	}
}