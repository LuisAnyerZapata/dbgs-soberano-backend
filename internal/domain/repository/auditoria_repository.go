package repository

import (
	"context"
	"time"

	"DBGS_SOBERANO_BACKEND/internal/domain/entity"
)

// AuditoriaRepository define las operaciones de escritura y lectura de la bitácora de eventos
type AuditoriaRepository interface {
	// RegistrarEvento inserta una traza de auditoría de forma inmutable.
	// Debe devolver el ID generado populando el campo evento.ID cuando corresponda.
	RegistrarEvento(ctx context.Context, evento *entity.AuditoriaEvento) error

	// ListarEventos consulta las trazas con filtros por usuario, operación, resultado y rango de fechas.
	ListarEventos(ctx context.Context, usuarioID, operacion, resultado string, fechaInicio, fechaFin *time.Time, limite, offset int) ([]entity.AuditoriaEvento, int64, error)
}
