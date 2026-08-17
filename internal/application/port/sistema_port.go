package port

import (
    "context"
    "DBGS_SOBERANO_BACKEND/internal/domain/entity"
)

type SistemaPort interface {
    ObtenerHealth(ctx context.Context) (*entity.SistemaHealth, error)
    ObtenerVersion(ctx context.Context) (*entity.SistemaVersion, error)
}