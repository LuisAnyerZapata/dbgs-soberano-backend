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

type stubCatalogoRepository struct {
	catalogos map[string]*entity.Catalogo
}

func newStubCatalogoRepository() *stubCatalogoRepository {
	return &stubCatalogoRepository{catalogos: make(map[string]*entity.Catalogo)}
}

func (s *stubCatalogoRepository) ObtenerPorID(ctx context.Context, id string) (*entity.Catalogo, error) {
	if c, ok := s.catalogos[id]; ok {
		return c, nil
	}
	return nil, entity.ErrEntidadNoEncontrada
}

func (s *stubCatalogoRepository) ObtenerPorCodigo(ctx context.Context, codigo string) (*entity.Catalogo, error) {
	for _, c := range s.catalogos {
		if c.Codigo == codigo {
			return c, nil
		}
	}
	return nil, entity.ErrEntidadNoEncontrada
}

func (s *stubCatalogoRepository) Listar(ctx context.Context, soloActivos bool, limite, offset int) ([]entity.Catalogo, int64, error) {
	var list []entity.Catalogo
	for _, c := range s.catalogos {
		if !soloActivos || c.Estado {
			list = append(list, *c)
		}
	}
	return list, int64(len(list)), nil
}

func (s *stubCatalogoRepository) Guardar(ctx context.Context, catalogo *entity.Catalogo) error {
	if catalogo.ID == "" {
		catalogo.ID = "id-" + catalogo.Codigo
	}
	s.catalogos[catalogo.ID] = catalogo
	return nil
}

func (s *stubCatalogoRepository) Actualizar(ctx context.Context, catalogo *entity.Catalogo) (*entity.Catalogo, error) {
	if _, ok := s.catalogos[catalogo.ID]; !ok {
		return nil, entity.ErrEntidadNoEncontrada
	}
	s.catalogos[catalogo.ID] = catalogo
	return catalogo, nil
}

func (s *stubCatalogoRepository) ActualizarEstado(ctx context.Context, id string, estado bool, usuarioModificador string) error {
	c, ok := s.catalogos[id]
	if !ok {
		return entity.ErrEntidadNoEncontrada
	}
	c.Estado = estado
	c.UpdatedBy = usuarioModificador
	c.UpdatedAt = time.Now()
	return nil
}

var _ repository.CatalogoRepository = (*stubCatalogoRepository)(nil)

func TestCrearCatalogoValido(t *testing.T) {
	repo := newStubCatalogoRepository()
	uc := NewCatalogoUseCase(repo)

	catalogo := &entity.Catalogo{Codigo: "CAT001", Nombre: "Catálogo prueba"}
	created, err := uc.CrearCatalogo(context.Background(), catalogo)
	if err != nil {
		t.Fatalf("CrearCatalogo() error = %v", err)
	}
	if created.ID == "" {
		t.Fatal("CrearCatalogo() debe generar un ID")
	}
	if !created.Estado {
		t.Fatal("CrearCatalogo() debe establecer estado true")
	}
	if created.CreatedBy != "system" {
		t.Fatalf("CrearCatalogo() createdBy = %q, want system", created.CreatedBy)
	}
}

func TestCrearCatalogoCodigoDuplicado(t *testing.T) {
	repo := newStubCatalogoRepository()
	repo.catalogos["id-1"] = &entity.Catalogo{ID: "id-1", Codigo: "CAT001", Nombre: "Existente", Estado: true}
	uc := NewCatalogoUseCase(repo)

	_, err := uc.CrearCatalogo(context.Background(), &entity.Catalogo{Codigo: "CAT001", Nombre: "Duplicado"})
	if err != domain.ErrCodigoDuplicado {
		t.Fatalf("CrearCatalogo() error = %v, want %v", err, domain.ErrCodigoDuplicado)
	}
}

func TestActualizarCatalogoPreservaCodigoSiNoSeProporciona(t *testing.T) {
	repo := newStubCatalogoRepository()
	repo.catalogos["id-1"] = &entity.Catalogo{ID: "id-1", Codigo: "CAT001", Nombre: "Original", Estado: true, UpdatedBy: "creator"}
	uc := NewCatalogoUseCase(repo)

	updated, err := uc.ActualizarCatalogo(context.Background(), &entity.Catalogo{ID: "id-1", Nombre: "Actualizado"})
	if err != nil {
		t.Fatalf("ActualizarCatalogo() error = %v", err)
	}
	if updated.Codigo != "CAT001" {
		t.Fatalf("ActualizarCatalogo() Codigo = %q, want CAT001", updated.Codigo)
	}
	if updated.Nombre != "Actualizado" {
		t.Fatalf("ActualizarCatalogo() Nombre = %q, want Actualizado", updated.Nombre)
	}
}

func TestEliminarCatalogoDesactivaRegistro(t *testing.T) {
	repo := newStubCatalogoRepository()
	repo.catalogos["id-1"] = &entity.Catalogo{ID: "id-1", Codigo: "CAT001", Nombre: "Original", Estado: true}
	uc := NewCatalogoUseCase(repo)

	err := uc.EliminarCatalogo(context.Background(), "id-1", "system")
	if err != nil {
		t.Fatalf("EliminarCatalogo() error = %v", err)
	}
	if repo.catalogos["id-1"].Estado {
		t.Fatal("EliminarCatalogo() debe dejar el catálogo inactivo")
	}
}

func TestListarCatalogosDevuelveSoloActivos(t *testing.T) {
	repo := newStubCatalogoRepository()
	repo.catalogos["id-1"] = &entity.Catalogo{ID: "id-1", Codigo: "CAT001", Nombre: "Activo", Estado: true}
	repo.catalogos["id-2"] = &entity.Catalogo{ID: "id-2", Codigo: "CAT002", Nombre: "Inactivo", Estado: false}
	uc := NewCatalogoUseCase(repo)

	output, err := uc.ListarCatalogos(context.Background(), port.ObtenerCatalogosInput{SoloActivos: true, Limite: 10, Offset: 0})
	if err != nil {
		t.Fatalf("ListarCatalogos() error = %v", err)
	}
	if len(output.Catalogos) != 1 {
		t.Fatalf("ListarCatalogos() count = %d, want 1", len(output.Catalogos))
	}
}
