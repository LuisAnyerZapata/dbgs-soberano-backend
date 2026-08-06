package repository

import (
	"context"

	"DBGS_SOBERANO_BACKEND/internal/domain/entity"
)

// CatalogoRepository define las operaciones de persistencia para Catálogos
type CatalogoRepository interface {
	// ObtenerPorID busca un catálogo por su identificador único
	ObtenerPorID(ctx context.Context, id string) (*entity.Catalogo, error)

	// ObtenerPorCodigo busca un catálogo por su código funcional
	ObtenerPorCodigo(ctx context.Context, codigo string) (*entity.Catalogo, error)

	// Listar retorna los catálogos aplicando filtros de vigencia y paginación
	Listar(ctx context.Context, soloActivos bool, limite, offset int) ([]entity.Catalogo, int64, error)

	// Guardar inserta un nuevo catálogo en el almacén de datos
	Guardar(ctx context.Context, catalogo *entity.Catalogo) error

	// Actualizar actualiza los campos modificables de un catálogo existente
	Actualizar(ctx context.Context, catalogo *entity.Catalogo) (*entity.Catalogo, error)

	// ActualizarEstado modifica la vigencia lógica del catálogo
	ActualizarEstado(ctx context.Context, id string, estado bool, usuarioModificador string) error
}
