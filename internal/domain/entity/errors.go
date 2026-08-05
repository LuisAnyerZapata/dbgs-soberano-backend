package entity

import "DBGS_SOBERANO_BACKEND/internal/domain"

var (
	ErrEntidadNoEncontrada = domain.ErrEntidadNoEncontrada
	ErrCodigoDuplicado     = domain.ErrCodigoDuplicado
	ErrAccesoNoAutorizado  = domain.ErrAccesoNoAutorizado
	ErrDatosInvalidos      = domain.ErrDatosInvalidos
	ErrRegistroInactivo    = domain.ErrRegistroInactivo
	ErrErrorInterno        = domain.ErrErrorInterno
)
