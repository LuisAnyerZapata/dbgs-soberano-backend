package repository

import (
	"context"

	"DBGS_SOBERANO_BACKEND/internal/domain/entity"
)

type RespaldoRepository interface {
	GuardarRespaldo(ctx context.Context, respaldo *entity.RespaldoOperacion) error
	ObtenerRespaldo(ctx context.Context, id string) (*entity.RespaldoOperacion, error)
	ListarRespaldos(ctx context.Context, limite int) ([]entity.RespaldoOperacion, error)
	ActualizarRespaldo(ctx context.Context, respaldo *entity.RespaldoOperacion) error
	EliminarRespaldo(ctx context.Context, id string) error

	GuardarRestauracion(ctx context.Context, restauracion *entity.Restauracion) error
	ActualizarRestauracion(ctx context.Context, restauracion *entity.Restauracion) error
	GuardarLog(ctx context.Context, log *entity.LogOperativo) error
	GuardarMetrica(ctx context.Context, metrica *entity.MetricaSistema) error
}
