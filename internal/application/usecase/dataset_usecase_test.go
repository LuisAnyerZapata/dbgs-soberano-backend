package usecase

import (
	"context"
	"testing"
	"time"

	"DBGS_SOBERANO_BACKEND/internal/application/port"
	"DBGS_SOBERANO_BACKEND/internal/domain/entity"
	"DBGS_SOBERANO_BACKEND/internal/domain/repository"
)

type stubDatasetRepository struct {
	datasets map[string]*entity.ConjuntoDato
	fuentes  map[string]*entity.FuenteDato
}

func newStubDatasetRepository() *stubDatasetRepository {
	return &stubDatasetRepository{
		datasets: make(map[string]*entity.ConjuntoDato),
		fuentes:  make(map[string]*entity.FuenteDato),
	}
}

func (s *stubDatasetRepository) ObtenerFuentePorID(ctx context.Context, id string) (*entity.FuenteDato, error) {
	if f, ok := s.fuentes[id]; ok {
		return f, nil
	}
	return nil, entity.ErrEntidadNoEncontrada
}

func (s *stubDatasetRepository) ListarFuentes(ctx context.Context) ([]entity.FuenteDato, error) {
	return nil, nil
}

func (s *stubDatasetRepository) ObtenerDatasetPorID(ctx context.Context, id string) (*entity.ConjuntoDato, error) {
	if d, ok := s.datasets[id]; ok {
		return d, nil
	}
	return nil, entity.ErrEntidadNoEncontrada
}

func (s *stubDatasetRepository) ListarDatasets(ctx context.Context, clasificacion, propietario string, limite, offset int) ([]entity.ConjuntoDato, int64, error) {
	var list []entity.ConjuntoDato
	for _, d := range s.datasets {
		if (clasificacion == "" || d.Clasificacion == clasificacion) && (propietario == "" || d.PropietarioDato == propietario) {
			list = append(list, *d)
		}
	}
	return list, int64(len(list)), nil
}

func (s *stubDatasetRepository) GuardarDataset(ctx context.Context, dataset *entity.ConjuntoDato) error {
	if dataset.ID == "" {
		dataset.ID = "id-" + dataset.Nombre
	}
	s.datasets[dataset.ID] = dataset
	return nil
}

func (s *stubDatasetRepository) ActualizarDataset(ctx context.Context, dataset *entity.ConjuntoDato) (*entity.ConjuntoDato, error) {
	if _, ok := s.datasets[dataset.ID]; !ok {
		return nil, entity.ErrEntidadNoEncontrada
	}
	s.datasets[dataset.ID] = dataset
	return dataset, nil
}

var _ repository.DatasetRepository = (*stubDatasetRepository)(nil)

func TestCrearDatasetValido(t *testing.T) {
	repo := newStubDatasetRepository()
	repo.fuentes["f1"] = &entity.FuenteDato{ID: "f1", Nombre: "Fuente prueba", Estado: true}
	uc := NewDatasetUseCase(repo)

	dataset := &entity.ConjuntoDato{FuenteDatoID: "f1", Nombre: "Dataset prueba"}
	created, err := uc.CrearDataset(context.Background(), dataset)
	if err != nil {
		t.Fatalf("CrearDataset() error = %v", err)
	}
	if created.ID == "" {
		t.Fatal("CrearDataset() debe generar un ID")
	}
	if !created.Estado {
		t.Fatal("CrearDataset() debe establecer estado true")
	}
	if created.CreatedBy != "system" {
		t.Fatalf("CrearDataset() createdBy = %q, want system", created.CreatedBy)
	}
}

func TestCrearDatasetRechazaFuenteInexistente(t *testing.T) {
	repo := newStubDatasetRepository()
	uc := NewDatasetUseCase(repo)

	_, err := uc.CrearDataset(context.Background(), &entity.ConjuntoDato{FuenteDatoID: "fuente-no-existe", Nombre: "Dataset invalido"})
	if err == nil {
		t.Fatal("CrearDataset() esperaba error para fuente inexistente")
	}
	if err != entity.ErrEntidadNoEncontrada {
		t.Fatalf("CrearDataset() error = %v, want %v", err, entity.ErrEntidadNoEncontrada)
	}
}

func TestActualizarDatasetPreservaCamposNulos(t *testing.T) {
	repo := newStubDatasetRepository()
	repo.datasets["id-1"] = &entity.ConjuntoDato{ID: "id-1", FuenteDatoID: "f1", Nombre: "Original", PropietarioDato: "owner", Clasificacion: "PUBLICO", UpdatedBy: "creator", UpdatedAt: time.Now()}
	uc := NewDatasetUseCase(repo)

	updated, err := uc.ActualizarDataset(context.Background(), &entity.ConjuntoDato{ID: "id-1", Nombre: "Actualizado"})
	if err != nil {
		t.Fatalf("ActualizarDataset() error = %v", err)
	}
	if updated.FuenteDatoID != "f1" {
		t.Fatalf("ActualizarDataset() FuenteDatoID = %q, want f1", updated.FuenteDatoID)
	}
	if updated.PropietarioDato != "owner" {
		t.Fatalf("ActualizarDataset() PropietarioDato = %q, want owner", updated.PropietarioDato)
	}
	if updated.Clasificacion != "PUBLICO" {
		t.Fatalf("ActualizarDataset() Clasificacion = %q, want PUBLICO", updated.Clasificacion)
	}
}

func TestListarDatasetsFiltraPorClasificacion(t *testing.T) {
	repo := newStubDatasetRepository()
	repo.datasets["id-1"] = &entity.ConjuntoDato{ID: "id-1", Nombre: "D1", Clasificacion: "PUBLICO"}
	repo.datasets["id-2"] = &entity.ConjuntoDato{ID: "id-2", Nombre: "D2", Clasificacion: "CONFIDENCIAL"}
	uc := NewDatasetUseCase(repo)

	output, err := uc.ListarDatasets(context.Background(), port.ObtenerDatasetsInput{Clasificacion: "PUBLICO", Limite: 10, Offset: 0})
	if err != nil {
		t.Fatalf("ListarDatasets() error = %v", err)
	}
	if len(output.Datasets) != 1 {
		t.Fatalf("ListarDatasets() count = %d, want 1", len(output.Datasets))
	}
}
