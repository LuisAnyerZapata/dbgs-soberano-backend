package domain

import "errors"

// Errores estandarizados del dominio DBGS
var (
	ErrEntidadNoEncontrada = errors.New("el recurso solicitado no existe")
	ErrCodigoDuplicado     = errors.New("el código funcional ya se encuentra registrado")
	ErrAccesoNoAutorizado  = errors.New("no tiene permisos suficientes para realizar esta operación")
	ErrDatosInvalidos      = errors.New("los datos proporcionados no cumplen con las reglas de validación")
	ErrRegistroInactivo    = errors.New("el registro se encuentra inactivo y no puede ser operado")
	ErrErrorInterno        = errors.New("ha ocurrido un error interno en el procesamiento de la solicitud")
)