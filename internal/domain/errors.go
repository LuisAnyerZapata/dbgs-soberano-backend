package domain

import (
	"errors"
	"fmt"
	"strings"

	"google.golang.org/grpc/codes"
)

// =========================================================================================================
// CÓDIGOS DE ERROR ESTRUCTURADOS
// =========================================================================================================

// ErrorCode clasifica los errores del sistema para respuestas controladas
type ErrorCode string

const (
	CodeNotFound         ErrorCode = "NOT_FOUND"
	CodeAlreadyExists    ErrorCode = "ALREADY_EXISTS"
	CodeInvalidArgument  ErrorCode = "INVALID_ARGUMENT"
	CodePermissionDenied ErrorCode = "PERMISSION_DENIED"
	CodeInternal         ErrorCode = "INTERNAL_ERROR"
	CodeInactive         ErrorCode = "INACTIVE"
	CodeSyntaxError      ErrorCode = "SYNTAX_ERROR"
	CodeConflict         ErrorCode = "CONFLICT"
)

// HTTPStatus retorna el código HTTP asociado a cada código de error
func (c ErrorCode) HTTPStatus() int {
	switch c {
	case CodeNotFound:
		return 404
	case CodeAlreadyExists:
		return 409
	case CodeInvalidArgument:
		return 400
	case CodePermissionDenied:
		return 403
	case CodeInactive:
		return 412
	case CodeSyntaxError:
		return 400
	case CodeConflict:
		return 409
	default:
		return 500
	}
}

// GRPCCode retorna el código gRPC asociado a cada código de error
func (c ErrorCode) GRPCCode() codes.Code {
	switch c {
	case CodeNotFound:
		return codes.NotFound
	case CodeAlreadyExists:
		return codes.AlreadyExists
	case CodeInvalidArgument:
		return codes.InvalidArgument
	case CodePermissionDenied:
		return codes.PermissionDenied
	case CodeInactive:
		return codes.FailedPrecondition
	case CodeSyntaxError:
		return codes.InvalidArgument
	case CodeConflict:
		return codes.AlreadyExists
	default:
		return codes.Internal
	}
}

// =========================================================================================================
// TIPO AppError — Error estructurado del dominio
// =========================================================================================================

// AppError es el tipo de error principal del sistema.
// Lleva información estructurada: código, campo, mensaje interno y mensaje seguro para el cliente.
type AppError struct {
	Code    ErrorCode `json:"code"`
	Field   string    `json:"field,omitempty"`
	Message string    `json:"message"`
	Safe    string    `json:"safe"`
	Err     error     `json:"-"`
}

// Error implementa la interfaz error
func (e *AppError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("[%s] campo '%s': %s", e.Code, e.Field, e.Message)
	}
	return fmt.Sprintf("[%s]: %s", e.Code, e.Message)
}

// Unwrap permite usar errors.Is() y errors.As() con errores envueltos
func (e *AppError) Unwrap() error {
	return e.Err
}

// ConField retorna una copia del error con el campo especificado
func (e *AppError) ConField(field string) *AppError {
	return &AppError{
		Code:    e.Code,
		Field:   field,
		Message: e.Message,
		Safe:    e.Safe,
		Err:     e.Err,
	}
}

// ConWrap retorna una copia del error envolviendo otro error
func (e *AppError) ConWrap(err error) *AppError {
	return &AppError{
		Code:    e.Code,
		Field:   e.Field,
		Message: e.Message,
		Safe:    e.Safe,
		Err:     err,
	}
}

// WithMessage retorna una copia del error con un mensaje adicional
func (e *AppError) WithMessage(msg string) *AppError {
	return &AppError{
		Code:    e.Code,
		Field:   e.Field,
		Message: msg,
		Safe:    e.Safe,
		Err:     e.Err,
	}
}

// ConFLICT retorna una copia del error con código de conflicto
func (e *AppError) ConFLICT() *AppError {
	return &AppError{
		Code:    CodeConflict,
		Field:   e.Field,
		Message: e.Message,
		Safe:    e.Safe,
		Err:     e.Err,
	}
}

// =========================================================================================================
// ERRORES SENTINEL ESTRUCTURADOS
// =================================================================================================_

// NotFound — recurso no encontrado
var NotFound = &AppError{
	Code:    CodeNotFound,
	Message: "el recurso solicitado no existe",
	Safe:    "Recurso no encontrado",
}

// AlreadyExists — registro duplicado
var AlreadyExists = &AppError{
	Code:    CodeAlreadyExists,
	Message: "el registro ya se encuentra registrado",
	Safe:    "El registro ya existe",
}

// InvalidArgument — datos inválidos
var InvalidArgument = &AppError{
	Code:    CodeInvalidArgument,
	Message: "los datos proporcionados no cumplen con las reglas de validación",
	Safe:    "Datos inválidos",
}

// PermissionDenied — sin permisos
var PermissionDenied = &AppError{
	Code:    CodePermissionDenied,
	Message: "no tiene permisos suficientes para realizar esta operación",
	Safe:    "Acceso denegado",
}

// InternalError — error interno
var InternalError = &AppError{
	Code:    CodeInternal,
	Message: "ha ocurrido un error interno en el procesamiento de la solicitud",
	Safe:    "Error interno del servidor",
}

// Inactive — registro inactivo
var Inactive = &AppError{
	Code:    CodeInactive,
	Message: "el registro se encuentra inactivo y no puede ser operado",
	Safe:    "Registro inactivo",
}

// SyntaxError — error de sintaxis SQL
var SyntaxError = &AppError{
	Code:    CodeSyntaxError,
	Message: "error de sintaxis en la base de datos al procesar el payload dinámico",
	Safe:    "Formato de datos inválido",
}

// =========================================================================================================
// ERRORES DE VALIDACIÓN — Campoes específicos
// =========================================================================================================

// RequiredError retorna un error de campo requerido
func RequiredError(field string) *AppError {
	return &AppError{
		Code:    CodeInvalidArgument,
		Field:   field,
		Message: fmt.Sprintf("el campo '%s' es obligatorio", field),
		Safe:    fmt.Sprintf("Campo '%s' requerido", field),
	}
}

// InvalidFormatError retorna un error de formato inválido
func InvalidFormatError(field, expected string) *AppError {
	return &AppError{
		Code:    CodeInvalidArgument,
		Field:   field,
		Message: fmt.Sprintf("el campo '%s' tiene un formato inválido, se esperaba: %s", field, expected),
		Safe:    fmt.Sprintf("Campo '%s' con formato inválido", field),
	}
}

// MinLengthError retorna un error de longitud mínima
func MinLengthError(field string, min int) *AppError {
	return &AppError{
		Code:    CodeInvalidArgument,
		Field:   field,
		Message: fmt.Sprintf("el campo '%s' debe tener al menos %d caracteres", field, min),
		Safe:    fmt.Sprintf("Campo '%s' demasiado corto", field),
	}
}

// MaxLengthError retorna un error de longitud máxima
func MaxLengthError(field string, max int) *AppError {
	return &AppError{
		Code:    CodeInvalidArgument,
		Field:   field,
		Message: fmt.Sprintf("el campo '%s' no debe exceder %d caracteres", field, max),
		Safe:    fmt.Sprintf("Campo '%s' demasiado largo", field),
	}
}

// InvalidValueError retorna un error de valor no permitido
func InvalidValueError(field string, allowed []string) *AppError {
	return &AppError{
		Code:    CodeInvalidArgument,
		Field:   field,
		Message: fmt.Sprintf("el campo '%s' debe ser uno de: %s", field, strings.Join(allowed, ", ")),
		Safe:    fmt.Sprintf("Campo '%s' con valor no permitido", field),
	}
}

// RangeError retorna un error de rango numérico
func RangeError(field string, min, max int) *AppError {
	return &AppError{
		Code:    CodeInvalidArgument,
		Field:   field,
		Message: fmt.Sprintf("el campo '%s' debe estar entre %d y %d", field, min, max),
		Safe:    fmt.Sprintf("Campo '%s' fuera de rango", field),
	}
}

// NotFoundError retorna un error de recurso no encontrado con contexto
func NotFoundError(entity string) *AppError {
	return &AppError{
		Code:    CodeNotFound,
		Message: fmt.Sprintf("el recurso '%s' no fue encontrado", entity),
		Safe:    "Recurso no encontrado",
	}
}

// DuplicateError retorna un error de registro duplicado con contexto
func DuplicateError(entity string) *AppError {
	return &AppError{
		Code:    CodeAlreadyExists,
		Message: fmt.Sprintf("el registro de '%s' ya existe", entity),
		Safe:    "El registro ya existe",
	}
}

// =========================================================================================================
// FUNCIONES AUXILIARES
// =========================================================================================================

// NewAppError crea un AppError a partir de un error existente
func NewAppError(code ErrorCode, msg string) *AppError {
	return &AppError{
		Code:    code,
		Message: msg,
		Safe:    msg,
	}
}

// WrapNotFoundError convierte sql.ErrNoRows a AppError de tipo NotFound
func WrapNotFoundError(err error, entity string) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, errors.New("sql: no rows in result set")) || strings.Contains(err.Error(), "no rows") {
		return NotFoundError(entity)
	}
	return InternalError.ConWrap(err)
}

// =========================================================================================================
// ERRORES LEGACY (compatibilidad)
// =========================================================================================================

// Estos errores se mantienen para compatibilidad con código existente.
// Se recomienda usar los tipos AppError en nuevo código.
var (
	ErrEntidadNoEncontrada = NotFound
	ErrCodigoDuplicado     = AlreadyExists
	ErrAccesoNoAutorizado  = PermissionDenied
	ErrDatosInvalidos      = InvalidArgument
	ErrRegistroInactivo    = Inactive
	ErrErrorInterno        = InternalError
	ErrSintaxisInvalida    = SyntaxError
)
