package port

import (
	"context"
	"time"

	"DBGS_SOBERANO_BACKEND/internal/domain/entity"
)

type RegistrarEventoInput struct {
	UsuarioID string
	Username  string
	Operacion string
	Recurso   string
	Detalles  string
	Resultado string
	IPOrigen  string
}

type ConsultarAuditoriaInput struct {
	UsuarioID   string
	Resultado   string
	FechaInicio *time.Time
	FechaFin    *time.Time
	Limite      int
	Offset      int
}

type ConsultarAuditoriaOutput struct {
	Eventos []entity.AuditoriaEvento `json:"eventos"`
	Total   int64                    `json:"total"`
	Limite  int                      `json:"limite"`
	Offset  int                      `json:"offset"`
}

type AuditoriaPort interface {
	RegistrarEvento(ctx context.Context, input RegistrarEventoInput) (*entity.AuditoriaEvento, error)
	ConsultarBitacora(ctx context.Context, input ConsultarAuditoriaInput) (*ConsultarAuditoriaOutput, error)
}