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

type stubRespaldoRepository struct {
	respaldos     []entity.RespaldoOperacion
	restauracion  []entity.Restauracion
	logs          []entity.LogOperativo
	metricas      []entity.MetricaSistema
}

func (s *stubRespaldoRepository) GuardarRespaldo(ctx context.Context, respaldo *entity.RespaldoOperacion) error {
	s.respaldos = append(s.respaldos, *respaldo)
	return nil
}

func (s *stubRespaldoRepository) ObtenerRespaldo(ctx context.Context, id string) (*entity.RespaldoOperacion, error) {
	for _, item := range s.respaldos {
		if item.ID == id {
			copy := item
			return &copy, nil
		}
	}
	return nil, entity.ErrEntidadNoEncontrada
}

func (s *stubRespaldoRepository) ListarRespaldos(ctx context.Context, limite int) ([]entity.RespaldoOperacion, error) {
	if limite <= 0 || limite > len(s.respaldos) {
		limite = len(s.respaldos)
	}
	result := make([]entity.RespaldoOperacion, 0, limite)
	for _, item := range s.respaldos {
		result = append(result, item)
		if len(result) == limite {
			break
		}
	}
	return result, nil
}

func (s *stubRespaldoRepository) ActualizarRespaldo(ctx context.Context, respaldo *entity.RespaldoOperacion) error {
	for i, item := range s.respaldos {
		if item.ID == respaldo.ID {
			s.respaldos[i] = *respaldo
			return nil
		}
	}
	return entity.ErrEntidadNoEncontrada
}

func (s *stubRespaldoRepository) EliminarRespaldo(ctx context.Context, id string) error {
	for i, item := range s.respaldos {
		if item.ID == id {
			s.respaldos = append(s.respaldos[:i], s.respaldos[i+1:]...)
			return nil
		}
	}
	return entity.ErrEntidadNoEncontrada
}

func (s *stubRespaldoRepository) GuardarRestauracion(ctx context.Context, restauracion *entity.Restauracion) error {
	s.restauracion = append(s.restauracion, *restauracion)
	return nil
}

func (s *stubRespaldoRepository) GuardarLog(ctx context.Context, log *entity.LogOperativo) error {
	s.logs = append(s.logs, *log)
	return nil
}

func (s *stubRespaldoRepository) GuardarMetrica(ctx context.Context, metrica *entity.MetricaSistema) error {
	s.metricas = append(s.metricas, *metrica)
	return nil
}

var _ repository.RespaldoRepository = (*stubRespaldoRepository)(nil)

func TestCrearRespaldoGeneraRegistroValido(t *testing.T) {
	repo := &stubRespaldoRepository{}
	uc := NewRespaldoUseCase(repo)

	resp, err := uc.CrearRespaldo(context.Background(), port.CrearRespaldoInput{Tipo: "FULL", Ruta: "/tmp/backup.sql", Detalles: "respaldo de prueba"})
	if err != nil {
		t.Fatalf("CrearRespaldo() error = %v", err)
	}
	if resp.ID == "" {
		t.Fatal("CrearRespaldo() debe generar un ID")
	}
	if resp.Estado != "COMPLETADO" {
		t.Fatalf("CrearRespaldo() estado = %q, want COMPLETADO", resp.Estado)
	}
}

func TestValidarRestauracionAceptaBackupCompleto(t *testing.T) {
	repo := &stubRespaldoRepository{}
	repo.respaldos = append(repo.respaldos, entity.RespaldoOperacion{ID: "backup-1", Tipo: "FULL", Estado: "COMPLETADO", RutaArchivo: "/tmp/backup.sql", FechaCreacion: time.Now(), RetencionDias: 7})
	uc := NewRespaldoUseCase(repo)

	res, err := uc.ValidarRestauracion(context.Background(), port.RestaurarBackupInput{BackupID: "backup-1", Validar: true, Usuario: "admin"})
	if err != nil {
		t.Fatalf("ValidarRestauracion() error = %v", err)
	}
	if !res.Validado {
		t.Fatal("ValidarRestauracion() debe marcar la restauración como validada")
	}
}

func TestValidarRestauracionRechazaBackupInvalido(t *testing.T) {
	repo := &stubRespaldoRepository{}
	repo.respaldos = append(repo.respaldos, entity.RespaldoOperacion{ID: "backup-2", Tipo: "FULL", Estado: "PENDIENTE", RutaArchivo: "/tmp/backup.sql", FechaCreacion: time.Now(), RetencionDias: 7})
	uc := NewRespaldoUseCase(repo)

	_, err := uc.ValidarRestauracion(context.Background(), port.RestaurarBackupInput{BackupID: "backup-2", Validar: true, Usuario: "admin"})
	if err == nil {
		t.Fatal("ValidarRestauracion() esperaba error para backup no válido")
	}
	if err != domain.ErrDatosInvalidos {
		t.Fatalf("ValidarRestauracion() error = %v, want %v", err, domain.ErrDatosInvalidos)
	}
}

func TestAplicarRetencionEliminaRespaldosViejos(t *testing.T) {
	repo := &stubRespaldoRepository{}
	oldBackup := entity.RespaldoOperacion{ID: "old", Tipo: "FULL", Estado: "COMPLETADO", RutaArchivo: "/tmp/old.sql", FechaCreacion: time.Now().AddDate(0, 0, -40), RetencionDias: 7}
	repo.respaldos = append(repo.respaldos, oldBackup)
	uc := NewRespaldoUseCase(repo)

	removed, err := uc.AplicarRetencion(context.Background(), port.RetencionInput{DiasRetencion: 30, MaximoBackups: 5})
	if err != nil {
		t.Fatalf("AplicarRetencion() error = %v", err)
	}
	if len(removed) != 1 {
		t.Fatalf("AplicarRetencion() removed = %d, want 1", len(removed))
	}
}

func TestRegistrarLogYMetricas(t *testing.T) {
	repo := &stubRespaldoRepository{}
	uc := NewRespaldoUseCase(repo)

	log, err := uc.RegistrarLog(context.Background(), port.RegistrarLogInput{Nivel: "INFO", Modulo: "api", Mensaje: "inicio"})
	if err != nil {
		t.Fatalf("RegistrarLog() error = %v", err)
	}
	if log.ID == "" {
		t.Fatal("RegistrarLog() debe generar un ID")
	}

	metric, err := uc.RegistrarMetrica(context.Background(), port.RegistrarMetricaInput{Nombre: "peticiones", Valor: 12.5, Unidad: "req/min"})
	if err != nil {
		t.Fatalf("RegistrarMetrica() error = %v", err)
	}
	if metric.ID == "" {
		t.Fatal("RegistrarMetrica() debe generar un ID")
	}
}

func TestHealthCheckMarcaEstadoBasico(t *testing.T) {
	repo := &stubRespaldoRepository{}
	uc := NewRespaldoUseCase(repo)

	hc, err := uc.EjecutarHealthCheck(context.Background(), port.HealthCheckInput{Componente: "database"})
	if err != nil {
		t.Fatalf("EjecutarHealthCheck() error = %v", err)
	}
	if hc.Estado == "" {
		t.Fatal("EjecutarHealthCheck() debe devolver un estado")
	}
}
