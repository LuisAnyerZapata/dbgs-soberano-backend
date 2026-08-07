package usecase

import (
	"context"
	"testing"
	"time"

	"DBGS_SOBERANO_BACKEND/internal/application/port"
	"DBGS_SOBERANO_BACKEND/internal/domain"
	"DBGS_SOBERANO_BACKEND/internal/domain/entity"
	"DBGS_SOBERANO_BACKEND/internal/domain/repository"
)

type stubAuditoriaRepository struct {
	eventos []entity.AuditoriaEvento
}

func (s *stubAuditoriaRepository) RegistrarEvento(ctx context.Context, evento *entity.AuditoriaEvento) error {
	if evento.ID == "" {
		evento.ID = "evt-1"
	}
	s.eventos = append(s.eventos, *evento)
	return nil
}

func (s *stubAuditoriaRepository) ListarEventos(ctx context.Context, usuarioID, operacion, resultado string, fechaInicio, fechaFin *time.Time, limite, offset int) ([]entity.AuditoriaEvento, int64, error) {
	var result []entity.AuditoriaEvento
	for _, evt := range s.eventos {
		if usuarioID != "" && evt.UsuarioID != usuarioID {
			continue
		}
		if operacion != "" && evt.Operacion != operacion {
			continue
		}
		if resultado != "" && evt.Resultado != resultado {
			continue
		}
		result = append(result, evt)
	}
	return result, int64(len(result)), nil
}

var _ repository.AuditoriaRepository = (*stubAuditoriaRepository)(nil)

func TestRegistrarEventoValido(t *testing.T) {
	repo := &stubAuditoriaRepository{}
	uc := NewAuditoriaUseCase(repo)

	created, err := uc.RegistrarEvento(context.Background(), port.RegistrarEventoInput{
		UsuarioID: "u1",
		Username:  "admin",
		Operacion: "CREAR_CATALOGO",
		Recurso:   "catalogo",
		Resultado: "EXITO",
		IPOrigen:  "127.0.0.1",
	})
	if err != nil {
		t.Fatalf("RegistrarEvento() error = %v", err)
	}
	if created.ID == "" {
		t.Fatal("RegistrarEvento() debe devolver un ID")
	}
	if len(repo.eventos) != 1 {
		t.Fatalf("RegistrarEvento() count = %d, want 1", len(repo.eventos))
	}
}

func TestRegistrarEventoRechazaOperacionVacia(t *testing.T) {
	repo := &stubAuditoriaRepository{}
	uc := NewAuditoriaUseCase(repo)

	_, err := uc.RegistrarEvento(context.Background(), port.RegistrarEventoInput{Resultado: "EXITO"})
	if err != domain.ErrDatosInvalidos {
		t.Fatalf("RegistrarEvento() error = %v, want %v", err, domain.ErrDatosInvalidos)
	}
}

func TestConsultarBitacoraFiltraPorUsuarioYOperacion(t *testing.T) {
	repo := &stubAuditoriaRepository{}
	repo.eventos = []entity.AuditoriaEvento{
		{ID: "evt-1", UsuarioID: "u1", Operacion: "CREAR_CATALOGO", Resultado: "EXITO"},
		{ID: "evt-2", UsuarioID: "u2", Operacion: "CONSULTAR_CATALOGO", Resultado: "EXITO"},
	}
	uc := NewAuditoriaUseCase(repo)

	output, err := uc.ConsultarBitacora(context.Background(), port.ConsultarAuditoriaInput{UsuarioID: "u1", Operacion: "CREAR_CATALOGO", Limite: 10, Offset: 0})
	if err != nil {
		t.Fatalf("ConsultarBitacora() error = %v", err)
	}
	if len(output.Eventos) != 1 {
		t.Fatalf("ConsultarBitacora() count = %d, want 1", len(output.Eventos))
	}
	if output.Eventos[0].ID != "evt-1" {
		t.Fatalf("ConsultarBitacora() id = %q, want evt-1", output.Eventos[0].ID)
	}
}
