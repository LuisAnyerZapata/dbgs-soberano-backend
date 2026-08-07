package usecase

import (
	"context"
	"testing"

	"DBGS_SOBERANO_BACKEND/internal/application/port"
	"DBGS_SOBERANO_BACKEND/internal/domain"
	"DBGS_SOBERANO_BACKEND/internal/domain/entity"
	"DBGS_SOBERANO_BACKEND/internal/domain/repository"
)

type stubIntegracionRepository struct {
	clientes    map[string]*entity.ClienteIntegracion
	solicitudes []entity.SolicitudIntegracion
}

func newStubIntegracionRepository() *stubIntegracionRepository {
	return &stubIntegracionRepository{clientes: make(map[string]*entity.ClienteIntegracion)}
}

func (s *stubIntegracionRepository) GuardarCliente(ctx context.Context, cliente *entity.ClienteIntegracion) error {
	if cliente.ID == "" {
		cliente.ID = "cli-1"
	}
	s.clientes[cliente.ID] = cliente
	return nil
}

func (s *stubIntegracionRepository) ObtenerClientePorID(ctx context.Context, id string) (*entity.ClienteIntegracion, error) {
	cliente, ok := s.clientes[id]
	if !ok {
		return nil, entity.ErrEntidadNoEncontrada
	}
	return cliente, nil
}

func (s *stubIntegracionRepository) GuardarSolicitud(ctx context.Context, solicitud *entity.SolicitudIntegracion) error {
	s.solicitudes = append(s.solicitudes, *solicitud)
	return nil
}

var _ repository.IntegracionRepository = (*stubIntegracionRepository)(nil)

func TestRegistrarClienteGeneraTokenYVersion(t *testing.T) {
	repo := newStubIntegracionRepository()
	uc := NewIntegracionUseCase(repo)

	cliente, err := uc.RegistrarCliente(context.Background(), port.RegistrarClienteInput{
		Nombre:          "CRM externo",
		Tipo:            "api",
		VersionContrato: "v1",
		Scopes:          []string{"catalogos:leer"},
	})
	if err != nil {
		t.Fatalf("RegistrarCliente() error = %v", err)
	}
	if cliente.Token == "" {
		t.Fatal("RegistrarCliente() debe generar un token")
	}
	if cliente.VersionContrato != "v1" {
		t.Fatalf("RegistrarCliente() version = %q, want v1", cliente.VersionContrato)
	}
}

func TestValidarAccesoRechazaTokenInvalido(t *testing.T) {
	repo := newStubIntegracionRepository()
	repo.clientes["cli-1"] = &entity.ClienteIntegracion{ID: "cli-1", Nombre: "CRM", Token: "abc", Estado: true, VersionContrato: "v1"}
	uc := NewIntegracionUseCase(repo)

	_, err := uc.ValidarAcceso(context.Background(), port.ValidarAccesoIntegracionInput{ClienteID: "cli-1", Token: "otro", VersionContrato: "v1"})
	if err != domain.ErrAccesoNoAutorizado {
		t.Fatalf("ValidarAcceso() error = %v, want %v", err, domain.ErrAccesoNoAutorizado)
	}
}

func TestRegistrarSolicitudGuardaLog(t *testing.T) {
	repo := newStubIntegracionRepository()
	uc := NewIntegracionUseCase(repo)

	err := uc.RegistrarSolicitud(context.Background(), port.RegistrarSolicitudInput{ClienteID: "cli-1", Metodo: "ListarCatalogos", Recurso: "/CatalogosService/ListarCatalogos", Estado: "OK"})
	if err != nil {
		t.Fatalf("RegistrarSolicitud() error = %v", err)
	}
	if len(repo.solicitudes) != 1 {
		t.Fatalf("RegistrarSolicitud() count = %d, want 1", len(repo.solicitudes))
	}
}
