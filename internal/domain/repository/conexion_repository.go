package repository

import (
	"context"

	"DBGS_SOBERANO_BACKEND/internal/domain/entity"
)

// ConexionRepository define las operaciones de persistencia de las conexiones
// registradas (metadatos) a bases de datos externas.
type ConexionRepository interface {
	Guardar(ctx context.Context, c *entity.Conexion) error
	ObtenerPorID(ctx context.Context, id string) (*entity.Conexion, error)
	Listar(ctx context.Context, limite, offset int) ([]entity.Conexion, int64, error)
	Actualizar(ctx context.Context, c *entity.Conexion) (*entity.Conexion, error)
	Eliminar(ctx context.Context, id string) error
}

// ApiPublicadaRepository define las operaciones de persistencia de las APIs
// públicas registradas sobre tablas externas.
type ApiPublicadaRepository interface {
	Guardar(ctx context.Context, a *entity.ApiPublicada) error
	ObtenerPorID(ctx context.Context, id string) (*entity.ApiPublicada, error)
	Listar(ctx context.Context, limite, offset int) ([]entity.ApiPublicada, int64, error)
	Actualizar(ctx context.Context, a *entity.ApiPublicada) (*entity.ApiPublicada, error)
	Eliminar(ctx context.Context, id string) error
	SlugEnUso(ctx context.Context, slug, excluirID string) (bool, error)
}
