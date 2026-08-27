package postgres

import (
	"database/sql"
	"errors"
	"log"
	"strings"

	"DBGS_SOBERANO_BACKEND/internal/domain"

	"github.com/lib/pq"
)

// =========================================================================================================
// DETECCIÓN DE ERRORES POSTGRESQL
// =========================================================================================================

// PostgresErrorClassifier clasifica errores de PostgreSQL a errores de dominio
type PostgresErrorClassifier struct {
	component string
}

// NewPostgresErrorClassifier crea un clasificador con contexto del componente
func NewPostgresErrorClassifier(component string) *PostgresErrorClassifier {
	return &PostgresErrorClassifier{component: component}
}

// Classify convierte un error de PostgreSQL a un error de dominio apropiado
func (c *PostgresErrorClassifier) Classify(err error) error {
	if err == nil {
		return nil
	}

	// 1. sql.ErrNoRows → NotFound
	if errors.Is(err, sql.ErrNoRows) {
		return domain.NotFoundError(c.component)
	}

	// 2. Errores PostgreSQL específicos por código
	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		switch pqErr.Code {
		case "23505": // unique_violation
			log.Printf("WARN BD (%s): duplicado - %s", c.component, pqErr.Detail)
			return domain.DuplicateError(c.component)

		case "23503": // foreign_key_violation
			log.Printf("WARN BD (%s): FK violation - %s", c.component, pqErr.Detail)
			return domain.InvalidArgument.WithMessage(
				"la operación viola una restricción de integridad referencial",
			)

		case "23502": // not_null_violation
			log.Printf("WARN BD (%s): campo nulo - %s", c.component, pqErr.Detail)
			return domain.RequiredError(pqErr.Column)

		case "22P02": // invalid_text_representation
			log.Printf("WARN BD (%s): sintaxis inválida - %s", c.component, pqErr.Detail)
			return domain.SyntaxError

		case "42P01": // undefined_table
			log.Printf("ERROR BD (%s): tabla no existe - %s", c.component, pqErr.Detail)
			return domain.NotFoundError("tabla")

		case "42703": // undefined_column
			log.Printf("ERROR BD (%s): columna no existe - %s", c.component, pqErr.Detail)
			return domain.InvalidArgument.WithMessage("la estructura de datos no coincide con el esquema")

		case "08000", "08003", "08006", "57P01", "57P02", "57P03":
			// Connection errors — reintentables
			log.Printf("ERROR BD (%s): error de conexión - %v", c.component, err)
			return domain.InternalError.WithMessage("servicio de base de datos temporalmente no disponible")

		case "40001": // serialization_failure
			log.Printf("WARN BD (%s): conflicto de concurrencia - %v", c.component, err)
			return domain.InternalError.ConFLICT().WithMessage("conflicto de concurrencia, intente nuevamente")

		default:
			// Otros errores PostgreSQL no clasificados
			log.Printf("ERROR BD (%s): código=%s detalle=%s", c.component, pqErr.Code, pqErr.Detail)
			return domain.InternalError
		}
	}

	// 3. Errores de conexión genéricos
	if strings.Contains(err.Error(), "connection refused") ||
		strings.Contains(err.Error(), "dial tcp") ||
		strings.Contains(err.Error(), "timeout") {
		log.Printf("ERROR BD (%s): error de conexión - %v", c.component, err)
		return domain.InternalError.WithMessage("servicio de base de datos temporalmente no disponible")
	}

	// 4. Error genérico
	log.Printf("ERROR BD (%s): %v", c.component, err)
	return domain.InternalError
}

// MustBeNotFound retorna true si el error es de tipo NotFound
func MustBeNotFound(err error) bool {
	var appErr *domain.AppError
	if errors.As(err, &appErr) {
		return appErr.Code == domain.CodeNotFound
	}
	return errors.Is(err, domain.ErrEntidadNoEncontrada)
}

// MustBeDuplicate retorna true si el error es de tipo AlreadyExists
func MustBeDuplicate(err error) bool {
	var appErr *domain.AppError
	if errors.As(err, &appErr) {
		return appErr.Code == domain.CodeAlreadyExists
	}
	return errors.Is(err, domain.ErrCodigoDuplicado)
}

// MissingRepoMethod marca un método del repositorio que aún no está implementado
func MissingRepoMethod(name string) error {
	log.Printf("WARNING: método %s no implementado en repositorio", name)
	return domain.InternalError.WithMessage("funcionalidad no disponible")
}

// isUniqueViolation retorna true si el error es una violación de constraint único
// de PostgreSQL (código 23505).
func isUniqueViolation(err error) bool {
	var pqErr *pq.Error
	return errors.As(err, &pqErr) && pqErr.Code == "23505"
}
