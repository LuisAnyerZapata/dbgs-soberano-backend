package domain

import (
	"net/mail"
	"regexp"
	"strings"
)

// =========================================================================================================
// VALIDADORES — Funciones reutilizables de validación de datos
// =========================================================================================================

var (
	uuidRegex   = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
	codigoRegex = regexp.MustCompile(`^[a-z][a-z0-9_:.]{0,99}$`)
	nameRegex   = regexp.MustCompile(`^[a-zA-ZáéíóúÁÉÍÓÚñÑüÜ][a-zA-Z0-9áéíóúÁÉÍÓÚñÑüÜ _.-]{0,99}$`)
)

// ValidateRequired valida que un campo string no esté vacío
func ValidateRequired(field, value string) *AppError {
	if strings.TrimSpace(value) == "" {
		return RequiredError(field)
	}
	return nil
}

// ValidateEmail valida que un campo tenga formato de email válido
func ValidateEmail(field, value string) *AppError {
	if value == "" {
		return nil // email es opcional
	}
	if _, err := mail.ParseAddress(value); err != nil {
		return InvalidFormatError(field, "email válido (ej: usuario@dominio.com)")
	}
	return nil
}

// ValidateUUID valida que un campo tenga formato UUID válido
func ValidateUUID(field, value string) *AppError {
	if value == "" {
		return RequiredError(field)
	}
	if !uuidRegex.MatchString(value) {
		return InvalidFormatError(field, "UUID válido (ej: 550e8400-e29b-41d4-a716-446655440000)")
	}
	return nil
}

// ValidateMinLength valida la longitud mínima de un campo
func ValidateMinLength(field, value string, min int) *AppError {
	if len(strings.TrimSpace(value)) < min {
		return MinLengthError(field, min)
	}
	return nil
}

// ValidateMaxLength valida la longitud máxima de un campo
func ValidateMaxLength(field, value string, max int) *AppError {
	if len(value) > max {
		return MaxLengthError(field, max)
	}
	return nil
}

// ValidateCodigo valida un código funcional (alfanumérico + guiones bajos)
func ValidateCodigo(field, value string) *AppError {
	if value == "" {
		return RequiredError(field)
	}
	if !codigoRegex.MatchString(value) {
		return InvalidFormatError(field, "código alfanumérico (minúsculas, guiones bajos, puntos)")
	}
	return nil
}

// ValidateName valida un nombre legible
func ValidateName(field, value string) *AppError {
	if value == "" {
		return RequiredError(field)
	}
	if !nameRegex.MatchString(value) {
		return InvalidFormatError(field, "nombre alfanumérico (letras, números, espacios, guiones)")
	}
	return nil
}

// ValidatePassword valida la fortaleza de una contraseña
func ValidatePassword(field, value string) *AppError {
	if value == "" {
		return RequiredError(field)
	}
	if len(value) < 8 {
		return &AppError{
			Code:    CodeInvalidArgument,
			Field:   field,
			Message: "la contraseña debe tener al menos 8 caracteres",
			Safe:    "Contraseña demasiado corta",
		}
	}
	if len(value) > 128 {
		return MaxLengthError(field, 128)
	}
	return nil
}

// ValidateEnum valida que un campo esté dentro de valores permitidos
func ValidateEnum(field, value string, allowed []string) *AppError {
	if value == "" {
		return nil
	}
	for _, a := range allowed {
		if strings.EqualFold(a, value) {
			return nil
		}
	}
	return InvalidValueError(field, allowed)
}

// ValidateRange valida que un número esté dentro de un rango
func ValidateRange(field string, value, min, max int) *AppError {
	if value < min || value > max {
		return RangeError(field, min, max)
	}
	return nil
}

// ValidateNoSpaces valida que un campo no contenga espacios
func ValidateNoSpaces(field, value string) *AppError {
	if value == "" {
		return RequiredError(field)
	}
	if strings.Contains(value, " ") {
		return InvalidFormatError(field, "sin espacios")
	}
	return nil
}

// =========================================================================================================
// BATCH VALIDATOR — Validación en cadena
// =========================================================================================================

// ValidationError acumula múltiples errores de validación
type ValidationError struct {
	Errors []*AppError `json:"errors"`
}

// Error implementa la interfaz error
func (v *ValidationError) Error() string {
	msgs := make([]string, len(v.Errors))
	for i, e := range v.Errors {
		msgs[i] = e.Error()
	}
	return strings.Join(msgs, "; ")
}

// HasErrors retorna true si hay errores acumulados
func (v *ValidationError) HasErrors() bool {
	return len(v.Errors) > 0
}

// Add agrega un error de validación (ignora nil)
func (v *ValidationError) Add(err *AppError) {
	if err != nil {
		v.Errors = append(v.Errors, err)
	}
}

// ToAppError convierte el ValidationError a un AppError único
func (v *ValidationError) ToAppError() *AppError {
	if !v.HasErrors() {
		return nil
	}
	if len(v.Errors) == 1 {
		return v.Errors[0]
	}
	// Múltiples errores: concatenar mensajes seguros
	msgs := make([]string, len(v.Errors))
	for i, e := range v.Errors {
		msgs[i] = e.Safe
	}
	return &AppError{
		Code:    CodeInvalidArgument,
		Message: v.Error(),
		Safe:    strings.Join(msgs, "; "),
	}
}

// NewValidator crea un validador en cadena
func NewValidator() *ValidationError {
	return &ValidationError{}
}

// Validate ejecuta todas las validaciones y retorna el primer error (o nil)
func (v *ValidationError) Validate() error {
	if !v.HasErrors() {
		return nil
	}
	return v.ToAppError()
}

// ValidateAll ejecuta todas las validaciones y retorna todos los errores
func (v *ValidationError) ValidateAll() *ValidationError {
	return v
}
