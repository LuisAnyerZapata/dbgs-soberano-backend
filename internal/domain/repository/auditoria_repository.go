package repository

import (
	"context"

	"DBGS_SOBERANO_BACKEND/internal/domain/entity"
)

// AuditoriaRepository define las operaciones de escritura y lectura de la bitácora de eventos
type AuditoriaRepository interface {
	// RegistrarEvento inserta una traza de auditoría de forma inmutable
	RegistrarEvento(ctx context.Context, evento *entity.AuditoriaEvento) error

	// ListarEventos consulta las trazas con filtros por rango de fechas, usuario y resultado (Exclusivo para Auditores)
	ListarEventos(ctx context.Context, usuarioID, resultado string, limite, offset int) ([]entity.AuditoriaEvento, int64, error)
}
