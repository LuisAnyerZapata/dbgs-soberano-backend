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

// AuthInterceptor intercepta las llamadas gRPC para verificar JWT (Humanos) o API Keys (Máquinas)
type AuthInterceptor struct {
    seguridadUseCase   port.SeguridadUseCasePort
    integracionUseCase port.IntegracionPort // Nuevo: Inyectamos el caso de uso de integración aquí
}

// NewAuthInterceptor crea el interceptor unificado inyectando ambas dependencias
func NewAuthInterceptor(seguridadUC port.SeguridadUseCasePort, integracionUC port.IntegracionPort) *AuthInterceptor {
    return &AuthInterceptor{
        seguridadUseCase:   seguridadUC,
        integracionUseCase: integracionUC,
    }
}

// Unary returns a gRPC unary server interceptor unificado
func (i *AuthInterceptor) Unary() grpc.UnaryServerInterceptor {
    return func(
        ctx context.Context,
        req interface{},
        info *grpc.UnaryServerInfo,
        handler grpc.UnaryHandler,
    ) (interface{}, error) {
        // 1. Rutas públicas (Siempre excluidas de cualquier autenticación)
        if info.FullMethod == "/dbgs.v1.SeguridadService/Login" ||
            info.FullMethod == "/dbgs.v1.SeguridadService/ValidarToken" ||
            info.FullMethod == "/dbgs.v1.SeguridadService/GetSetupStatus" ||
            info.FullMethod == "/dbgs.v1.SeguridadService/CreateSetup" ||
            info.FullMethod == "/dbgs.v1.SistemaService/GetHealth" ||
            info.FullMethod == "/dbgs.v1.SistemaService/GetVersion" {
            return handler(ctx, req)
        }

        md, ok := metadata.FromIncomingContext(ctx)
        if !ok {
            return nil, status.Error(codes.Unauthenticated, "metadatos de autenticación no encontrados")
        }

        // 2. Estrategia A: Autenticación por API Key (Integración de Máquinas)
        // Buscamos las cabeceras x-api-token y x-client-id
        if apiTokens := md.Get("x-api-token"); len(apiTokens) > 0 {
            if clientIDs := md.Get("x-client-id"); len(clientIDs) > 0 {
                cliente, err := i.integracionUseCase.ValidarAcceso(ctx, port.ValidarAccesoIntegracionInput{
                    ClienteID: clientIDs[0],
                    Token:     apiTokens[0],
                })
                if err != nil {
                    return nil, status.Error(codes.PermissionDenied, "acceso de integración no autorizado (API Key inválida o inactiva)")
                }
                // Inyectamos el cliente de integración en el contexto para que los Use Cases lo sepan
                newCtx := context.WithValue(ctx, KeyClienteIntegracion, cliente)
                return handler(newCtx, req)
            }
        }

        // 3. Estrategia B: Autenticación por JWT (Usuarios Humanos)
        // Si no había API Key, exigimos el token JWT tradicional
        authHeader := md["authorization"]
        if len(authHeader) == 0 {
            return nil, status.Error(codes.Unauthenticated, "token de autorización (JWT o API Key) requerido")
        }

        tokenStr := strings.TrimPrefix(authHeader[0], "Bearer ")
        if tokenStr == "" {
            return nil, status.Error(codes.Unauthenticated, "formato de token inválido")
        }

        validated, err := i.seguridadUseCase.ValidarToken(ctx, tokenStr)
        if err != nil || !validated.Valid {
            return nil, status.Error(codes.Unauthenticated, "token JWT inválido o expirado")
        }

        usuario, err := i.seguridadUseCase.ObtenerPerfilUsuario(ctx, validated.Username)
        if err != nil {
            return nil, status.Error(codes.PermissionDenied, "acceso denegado: cuenta inactiva o no registrada")
        }

        newCtx := context.WithValue(ctx, KeyUsuario, usuario)
        return handler(newCtx, req)
    }
}