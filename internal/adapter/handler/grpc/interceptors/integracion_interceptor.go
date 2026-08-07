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

type IntegracionInterceptor struct {
	integracionUseCase port.IntegracionPort
}

func NewIntegracionInterceptor(integracionUC port.IntegracionPort) *IntegracionInterceptor {
	return &IntegracionInterceptor{integracionUseCase: integracionUC}
}

func (i *IntegracionInterceptor) Unary() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if strings.Contains(info.FullMethod, "/Integration") || strings.Contains(info.FullMethod, "/Integracion") {
			md, ok := metadata.FromIncomingContext(ctx)
			if !ok {
				return nil, status.Error(codes.Unauthenticated, "metadata de integración no encontrado")
			}

			var clienteID string
			var token string
			var version string
			if vals := md.Get("x-client-id"); len(vals) > 0 {
				clienteID = vals[0]
			}
			if vals := md.Get("x-api-token"); len(vals) > 0 {
				token = vals[0]
			}
			if vals := md.Get("x-api-version"); len(vals) > 0 {
				version = vals[0]
			}

			if clienteID == "" || token == "" {
				return nil, status.Error(codes.Unauthenticated, "cliente y token de integración requeridos")
			}

			cliente, err := i.integracionUseCase.ValidarAcceso(ctx, port.ValidarAccesoIntegracionInput{
				ClienteID:       clienteID,
				Token:           token,
				VersionContrato: version,
			})
			if err != nil {
				return nil, status.Error(codes.PermissionDenied, "acceso de integración no autorizado")
			}
			ctx = context.WithValue(ctx, "integracion_cliente", cliente)
			return handler(ctx, req)
		}
		return handler(ctx, req)
	}
}
