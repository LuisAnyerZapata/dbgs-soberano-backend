package repository

import (
    "context"
    "DBGS_SOBERANO_BACKEND/internal/domain/entity"
)

// InstitucionRepository define el contrato para la persistencia de entidades gubernamentales
type InstitucionRepository interface {
    ObtenerPorID(ctx context.Context, id string) (*entity.Institucion, error)
    Listar(ctx context.Context, soloActivos bool) ([]entity.Institucion, error)
}