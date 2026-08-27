package entity

import "DBGS_SOBERANO_BACKEND/internal/domain"

// Errores re-exportados del dominio para uso en la capa de repositorio
var (
	ErrEntidadNoEncontrada = domain.ErrEntidadNoEncontrada
	ErrCodigoDuplicado     = domain.ErrCodigoDuplicado
	ErrAccesoNoAutorizado  = domain.ErrAccesoNoAutorizado
	ErrDatosInvalidos      = domain.ErrDatosInvalidos
	ErrRegistroInactivo    = domain.ErrRegistroInactivo
	ErrErrorInterno        = domain.ErrErrorInterno
	ErrSintaxisInvalida    = domain.ErrSintaxisInvalida
)

// AppError re-exportado para uso en repositorios
type AppError = domain.AppError

// Códigos de error re-exportados
const (
	CodeNotFound         = domain.CodeNotFound
	CodeAlreadyExists    = domain.CodeAlreadyExists
	CodeInvalidArgument  = domain.CodeInvalidArgument
	CodePermissionDenied = domain.CodePermissionDenied
	CodeInternal         = domain.CodeInternal
	CodeInactive         = domain.CodeInactive
	CodeSyntaxError      = domain.CodeSyntaxError
	CodeConflict         = domain.CodeConflict
)
