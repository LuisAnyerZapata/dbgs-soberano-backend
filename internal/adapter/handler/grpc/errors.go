package grpc

import (
	"errors"

	"DBGS_SOBERANO_BACKEND/internal/domain"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// mapDomainErrorToGRPC convierte un error de dominio a un error de estado gRPC sanitizado
func mapDomainErrorToGRPC(err error) error {
	if err == nil {
		return nil
	}

	switch {
	case errors.Is(err, domain.ErrEntidadNoEncontrada):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, domain.ErrDatosInvalidos):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, domain.ErrAccesoNoAutorizado):
		return status.Error(codes.PermissionDenied, err.Error())
	case errors.Is(err, domain.ErrCodigoDuplicado):
		return status.Error(codes.AlreadyExists, err.Error())
	case errors.Is(err, domain.ErrRegistroInactivo):
		return status.Error(codes.FailedPrecondition, err.Error())
	default:
		return status.Error(codes.Internal, "ha ocurrido un error interno al procesar la solicitud")
	}
}