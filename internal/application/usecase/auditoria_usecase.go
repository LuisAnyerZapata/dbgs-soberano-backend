package usecase

import (
	"context"
	"time"

	"DBGS_SOBERANO_BACKEND/internal/application/port"
	"DBGS_SOBERANO_BACKEND/internal/domain"
	"DBGS_SOBERANO_BACKEND/internal/domain/entity"
	"DBGS_SOBERANO_BACKEND/internal/domain/repository"
)

type auditoriaUseCase struct {
	auditoriaRepo repository.AuditoriaRepository
}

func NewAuditoriaUseCase(repo repository.AuditoriaRepository) port.AuditoriaPort {
	return &auditoriaUseCase{
		auditoriaRepo: repo,
	}
}

func (u *auditoriaUseCase) RegistrarEvento(ctx context.Context, input port.RegistrarEventoInput) (*entity.AuditoriaEvento, error) {
	if input.Operacion == "" || input.Resultado == "" {
		return nil, domain.ErrDatosInvalidos
	}

	evento := &entity.AuditoriaEvento{
		UsuarioID:     input.UsuarioID,
		Username:      input.Username,
		Operacion:     input.Operacion,
		Recurso:       input.Recurso,
		Detalles:      input.Detalles,
		Resultado:     input.Resultado,
		IPOrigen:      input.IPOrigen,
		FechaCreacion: time.Now(),
	}

	if err := u.auditoriaRepo.RegistrarEvento(ctx, evento); err != nil {
		return nil, err
	}

	return evento, nil
}

func (u *auditoriaUseCase) ConsultarBitacora(ctx context.Context, input port.ConsultarAuditoriaInput) (*port.ConsultarAuditoriaOutput, error) {
	if input.Limite <= 0 {
		input.Limite = 20
	}
	if input.Offset < 0 {
		input.Offset = 0
	}

	eventos, total, err := u.auditoriaRepo.ListarEventos(ctx, input.UsuarioID, input.Resultado, input.Limite, input.Offset)
	if err != nil {
		return nil, err
	}

	return &port.ConsultarAuditoriaOutput{
		Eventos: eventos,
		Total:   total,
		Limite:  input.Limite,
		Offset:  input.Offset,
	}, nil
}