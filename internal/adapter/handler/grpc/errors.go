package grpc

import (
	"errors"

	"DBGS_SOBERANO_BACKEND/internal/domain"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// mapDomainErrorToGRPC convierte un error de dominio a un error de estado gRPC sanitizado.
// Usa errors.As() para detectar errores estructurados (AppError) y extraer información segura.
// NUNCA expone detalles internos al cliente.
func mapDomainErrorToGRPC(err error) error {
	if err == nil {
		return nil
	}

	// 1. Intentar extraer un AppError estructurado
	var appErr *domain.AppError
	if errors.As(err, &appErr) {
		return status.Error(appErr.Code.GRPCCode(), appErr.Safe)
	}

	// 2. Intentar extraer un ValidationError (múltiples errores)
	var valErr *domain.ValidationError
	if errors.As(err, &valErr) {
		if app := valErr.ToAppError(); app != nil {
			return status.Error(codes.InvalidArgument, app.Safe)
		}
	}

	// 3. Fallback: errores legacy (sentinel errors)
	switch {
	case errors.Is(err, domain.ErrEntidadNoEncontrada):
		return status.Error(codes.NotFound, "Recurso no encontrado")
	case errors.Is(err, domain.ErrDatosInvalidos):
		return status.Error(codes.InvalidArgument, "Datos inválidos")
	case errors.Is(err, domain.ErrAccesoNoAutorizado):
		return status.Error(codes.PermissionDenied, "Acceso denegado")
	case errors.Is(err, domain.ErrCodigoDuplicado):
		return status.Error(codes.AlreadyExists, "El registro ya existe")
	case errors.Is(err, domain.ErrRegistroInactivo):
		return status.Error(codes.FailedPrecondition, "Registro inactivo")
	case errors.Is(err, domain.ErrSintaxisInvalida):
		return status.Error(codes.InvalidArgument, "Formato de datos inválido")
	default:
		// NUNCA enviar detalles del error interno al cliente
		return status.Error(codes.Internal, "Error interno del servidor")
	}
}
