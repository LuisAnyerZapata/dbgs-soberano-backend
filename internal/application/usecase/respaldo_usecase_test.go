package usecase

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"DBGS_SOBERANO_BACKEND/internal/application/port"
	"DBGS_SOBERANO_BACKEND/internal/domain"
	"DBGS_SOBERANO_BACKEND/internal/domain/entity"
	"DBGS_SOBERANO_BACKEND/internal/domain/repository"
)

// =========================================================================================================
// STUBS
// =========================================================================================================

type stubRespaldoRepository struct {
	respaldos      []entity.RespaldoOperacion
	restauraciones []entity.Restauracion
	logs           []entity.LogOperativo
	metricas       []entity.MetricaSistema
}

func (s *stubRespaldoRepository) GuardarRespaldo(ctx context.Context, respaldo *entity.RespaldoOperacion) error {
	s.respaldos = append(s.respaldos, *respaldo)
	return nil
}

func (s *stubRespaldoRepository) ObtenerRespaldo(ctx context.Context, id string) (*entity.RespaldoOperacion, error) {
	for i := range s.respaldos {
		if s.respaldos[i].ID == id {
			copia := s.respaldos[i]
			return &copia, nil
		}
	}
	return nil, entity.ErrEntidadNoEncontrada
}

func (s *stubRespaldoRepository) ListarRespaldos(ctx context.Context, limite int) ([]entity.RespaldoOperacion, error) {
	if limite > len(s.respaldos) {
		limite = len(s.respaldos)
	}
	return append([]entity.RespaldoOperacion{}, s.respaldos[:limite]...), nil
}

func (s *stubRespaldoRepository) ActualizarRespaldo(ctx context.Context, respaldo *entity.RespaldoOperacion) error {
	for i := range s.respaldos {
		if s.respaldos[i].ID == respaldo.ID {
			s.respaldos[i] = *respaldo
			return nil
		}
	}
	return entity.ErrEntidadNoEncontrada
}

func (s *stubRespaldoRepository) EliminarRespaldo(ctx context.Context, id string) error {
	for i := range s.respaldos {
		if s.respaldos[i].ID == id {
			s.respaldos = append(s.respaldos[:i], s.respaldos[i+1:]...)
			return nil
		}
	}
	return entity.ErrEntidadNoEncontrada
}

func (s *stubRespaldoRepository) GuardarRestauracion(ctx context.Context, restauracion *entity.Restauracion) error {
	s.restauraciones = append(s.restauraciones, *restauracion)
	return nil
}

func (s *stubRespaldoRepository) ActualizarRestauracion(ctx context.Context, restauracion *entity.Restauracion) error {
	for i := range s.restauraciones {
		if s.restauraciones[i].ID == restauracion.ID {
			s.restauraciones[i].Estado = restauracion.Estado
			s.restauraciones[i].Validado = restauracion.Validado
			s.restauraciones[i].Observaciones = restauracion.Observaciones
			return nil
		}
	}
	return entity.ErrEntidadNoEncontrada
}

func (s *stubRespaldoRepository) GuardarLog(ctx context.Context, logOp *entity.LogOperativo) error {
	s.logs = append(s.logs, *logOp)
	return nil
}

func (s *stubRespaldoRepository) GuardarMetrica(ctx context.Context, metrica *entity.MetricaSistema) error {
	s.metricas = append(s.metricas, *metrica)
	return nil
}

var _ repository.RespaldoRepository = (*stubRespaldoRepository)(nil)

type stubSeguridadPermisos struct {
	permisos map[string]bool // clave "rolID|permiso"
}

func (s *stubSeguridadPermisos) ValidarPermiso(ctx context.Context, rolID, permiso string) (bool, error) {
	return s.permisos[rolID+"|"+permiso], nil
}

func (s *stubSeguridadPermisos) ObtenerUsuarioPorUsername(ctx context.Context, username string) (*entity.Usuario, error) {
	return &entity.Usuario{Username: username}, nil
}
func (s *stubSeguridadPermisos) AutenticarUsuario(ctx context.Context, username, password string) (*entity.Usuario, error) {
	return nil, domain.ErrAccesoNoAutorizado
}
func (s *stubSeguridadPermisos) ObtenerRolPorID(ctx context.Context, rolID string) (*entity.Rol, error) {
	return &entity.Rol{ID: rolID}, nil
}
func (s *stubSeguridadPermisos) ContarSuperAdmins(ctx context.Context) (int64, error) { return 1, nil }
func (s *stubSeguridadPermisos) CrearUsuarioAdmin(ctx context.Context, username, email, passwordHash, rolID string) (*entity.Usuario, error) {
	return &entity.Usuario{Username: username}, nil
}
func (s *stubSeguridadPermisos) AsegurarRolSuperAdmin(ctx context.Context) (string, error) {
	return "rol-admin", nil
}

func (s *stubSeguridadPermisos) CrearRol(ctx context.Context, nombre, descripcion string) (*entity.Rol, error) {
	return &entity.Rol{ID: "stub-rol", Nombre: nombre, Descripcion: descripcion}, nil
}
func (s *stubSeguridadPermisos) ListarRoles(ctx context.Context) ([]entity.Rol, error) {
	return nil, nil
}
func (s *stubSeguridadPermisos) ObtenerRolPorNombre(ctx context.Context, nombre string) (*entity.Rol, error) {
	return &entity.Rol{ID: "stub-rol", Nombre: nombre}, nil
}
func (s *stubSeguridadPermisos) ActualizarRol(ctx context.Context, id, nombre, descripcion string) error {
	return nil
}
func (s *stubSeguridadPermisos) EliminarRol(ctx context.Context, id string) error {
	return nil
}
func (s *stubSeguridadPermisos) VincularPermisos(ctx context.Context, rolID string, codigos []string) (int64, error) {
	return int64(len(codigos)), nil
}
func (s *stubSeguridadPermisos) DesvincularPermiso(ctx context.Context, rolID, permisoCodigo string) error {
	return nil
}
func (s *stubSeguridadPermisos) ListarPermisosRol(ctx context.Context, rolID string) ([]string, error) {
	return nil, nil
}

func (s *stubSeguridadPermisos) CrearUsuario(ctx context.Context, username, email, passwordHash, rolID string, esTecnico bool) (*entity.Usuario, error) {
	return &entity.Usuario{ID: "stub-user", Username: username, Email: email, RolID: rolID, EsTecnico: esTecnico, Estado: true}, nil
}
func (s *stubSeguridadPermisos) ListarUsuarios(ctx context.Context) ([]entity.Usuario, error) {
	return nil, nil
}
func (s *stubSeguridadPermisos) ObtenerUsuarioPorID(ctx context.Context, id string) (*entity.Usuario, error) {
	return &entity.Usuario{ID: id, Username: "stub"}, nil
}
func (s *stubSeguridadPermisos) ActualizarUsuario(ctx context.Context, id, email, rolID string, esTecnico, estado bool) error {
	return nil
}
func (s *stubSeguridadPermisos) EliminarUsuario(ctx context.Context, id string) error {
	return nil
}

var _ repository.SeguridadRepository = (*stubSeguridadPermisos)(nil)

// stubMotor simula pg_dump/pg_restore creando/borrando archivos reales en un directorio temporal.
type stubMotor struct {
	dumpsDir          string
	errCreacion       error
	errRestaurar      error
	llamadasRestaurar []string
}

func (m *stubMotor) EjecutarCreacion(ctx context.Context, destinoDir string) (string, error) {
	if m.errCreacion != nil {
		return "", m.errCreacion
	}
	ruta := filepath.Join(destinoDir, fmt.Sprintf("dbgs_backup_%d.dump", time.Now().UnixNano()))
	if err := os.WriteFile(ruta, []byte("CONTENIDO_DUMP_FALSO"), 0600); err != nil {
		return "", err
	}
	return ruta, nil
}

func (m *stubMotor) EjecutarRestauracion(ctx context.Context, rutaArchivo string) error {
	m.llamadasRestaurar = append(m.llamadasRestaurar, rutaArchivo)
	return m.errRestaurar
}

// =========================================================================================================
// HELPERS
// =========================================================================================================

const (
	rolConPermisos = "rol-admin"
	rolSinPermisos = "rol-auditor"
)

func nuevoUseCase(repo *stubRespaldoRepository, motor port.MotorRespaldo, dumpsDir string) port.RespaldoPort {
	seguridad := &stubSeguridadPermisos{
		permisos: map[string]bool{
			rolConPermisos + "|respaldo:ejecutar":     true,
			rolConPermisos + "|restauracion:ejecutar": true,
		},
	}
	return NewRespaldoUseCase(repo, seguridad, motor, RespaldoConfig{DumpsDir: dumpsDir})
}

func ctxAdmin() context.Context {
	return context.WithValue(context.Background(), domain.CtxKeyUsuario, &entity.Usuario{
		ID: "u-1", Username: "admin", RolID: rolConPermisos,
	})
}

func ctxAuditor() context.Context {
	return context.WithValue(context.Background(), domain.CtxKeyUsuario, &entity.Usuario{
		ID: "u-2", Username: "auditor", RolID: rolSinPermisos,
	})
}

// esperarEstado hace polling del registro hasta que alcance el estado esperado.
func esperarEstado(t *testing.T, repo *stubRespaldoRepository, id, estado string) entity.RespaldoOperacion {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		for i := range repo.respaldos {
			if repo.respaldos[i].ID == id && repo.respaldos[i].Estado == estado {
				return repo.respaldos[i]
			}
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("el respaldo %s nunca alcanzó el estado %s", id, estado)
	return entity.RespaldoOperacion{}
}

// =========================================================================================================
// AUTORIZACIÓN
// =========================================================================================================

func TestCrearRespaldoSinSesionRechazado(t *testing.T) {
	repo := &stubRespaldoRepository{}
	uc := nuevoUseCase(repo, &stubMotor{dumpsDir: t.TempDir()}, t.TempDir())

	_, err := uc.CrearRespaldo(context.Background(), port.CrearRespaldoInput{})
	if err == nil || !errors.Is(err, domain.ErrAccesoNoAutorizado) {
		t.Fatalf("esperaba ErrAccesoNoAutorizado sin sesión, obtuve %v", err)
	}
}

func TestCrearRespaldoSinPermisoRechazado(t *testing.T) {
	repo := &stubRespaldoRepository{}
	uc := nuevoUseCase(repo, &stubMotor{dumpsDir: t.TempDir()}, t.TempDir())

	_, err := uc.CrearRespaldo(ctxAuditor(), port.CrearRespaldoInput{})
	if err == nil || !errors.Is(err, domain.ErrAccesoNoAutorizado) {
		t.Fatalf("esperaba rechazo por permiso faltante, obtuve %v", err)
	}
	if len(repo.respaldos) != 0 {
		t.Fatal("no debe registrarse nada si falla la autorización")
	}
}

// =========================================================================================================
// CREACIÓN ASÍNCRONA
// =========================================================================================================

func TestCrearRespaldoFlujoAsincronoCompleto(t *testing.T) {
	repo := &stubRespaldoRepository{}
	motor := &stubMotor{}
	uc := nuevoUseCase(repo, motor, t.TempDir())

	output, err := uc.CrearRespaldo(ctxAdmin(), port.CrearRespaldoInput{Tipo: "FULL"})
	if err != nil {
		t.Fatalf("CrearRespaldo() error = %v", err)
	}
	if output.Respaldo.Estado != entity.EstadoEnProgreso {
		t.Fatalf("estado inicial = %q, want EN_PROGRESO", output.Respaldo.Estado)
	}
	if output.Respaldo.UsuarioCreador != "admin" {
		t.Fatalf("usuario_creador = %q, want admin (tomado del contexto)", output.Respaldo.UsuarioCreador)
	}

	final := esperarEstado(t, repo, output.Respaldo.ID, entity.EstadoCompletado)
	if final.RutaArchivo == "" {
		t.Fatal("la ruta del archivo debe registrarse al completar")
	}
	if final.TamanoBytes <= 0 {
		t.Fatal("el tamaño debe capturarse del archivo físico")
	}
	if final.FechaFinalizacion.IsZero() {
		t.Fatal("debe registrarse la fecha de finalización")
	}
}

func TestCrearRespaldoFalloMotorMarcaFallido(t *testing.T) {
	repo := &stubRespaldoRepository{}
	motor := &stubMotor{errCreacion: fmt.Errorf("pg_dump no existe")}
	uc := nuevoUseCase(repo, motor, t.TempDir())

	output, _ := uc.CrearRespaldo(ctxAdmin(), port.CrearRespaldoInput{})
	final := esperarEstado(t, repo, output.Respaldo.ID, entity.EstadoFallido)
	if final.RutaArchivo != "" {
		t.Fatal("un respaldo fallido no debe declarar ruta")
	}
}

// =========================================================================================================
// DESCARGA Y RESTAURACIÓN
// =========================================================================================================

func TestDescargarRespaldoDevuelveContenido(t *testing.T) {
	repo := &stubRespaldoRepository{}
	dumpsDir := t.TempDir()
	ruta := filepath.Join(dumpsDir, "dbgs_backup_ok.dump")
	if err := os.WriteFile(ruta, []byte("BYTES_DE_PRUEBA"), 0600); err != nil {
		t.Fatal(err)
	}
	repo.respaldos = append(repo.respaldos, entity.RespaldoOperacion{
		ID: "b-ok", Estado: entity.EstadoCompletado, RutaArchivo: ruta,
	})
	uc := nuevoUseCase(repo, &stubMotor{}, dumpsDir)

	output, err := uc.DescargarRespaldo(ctxAdmin(), port.DescargarRespaldoInput{ID: "b-ok"})
	if err != nil {
		t.Fatalf("DescargarRespaldo() error = %v", err)
	}
	if string(output.Contenido) != "BYTES_DE_PRUEBA" || output.NombreArchivo != "dbgs_backup_ok.dump" {
		t.Fatalf("descarga incorrecta: %+v", output)
	}
}

func TestDescargarRespaldoRechazaRutaFueraDelRepositorio(t *testing.T) {
	repo := &stubRespaldoRepository{}
	archivoAjeno := filepath.Join(t.TempDir(), "robo.dump")
	if err := os.WriteFile(archivoAjeno, []byte("x"), 0600); err != nil {
		t.Fatal(err)
	}
	repo.respaldos = append(repo.respaldos, entity.RespaldoOperacion{
		ID: "b-mal", Estado: entity.EstadoCompletado, RutaArchivo: archivoAjeno,
	})
	uc := nuevoUseCase(repo, &stubMotor{}, t.TempDir()) // dumpsDir distinto al del archivo

	_, err := uc.DescargarRespaldo(ctxAdmin(), port.DescargarRespaldoInput{ID: "b-mal"})
	if err == nil || !errors.Is(err, domain.ErrAccesoNoAutorizado) {
		t.Fatalf("path traversal debe rechazarse con ErrAccesoNoAutorizado, obtuve %v", err)
	}
}

func TestRestaurarRechazaSinConfirmacion(t *testing.T) {
	repo := &stubRespaldoRepository{}
	uc := nuevoUseCase(repo, &stubMotor{}, t.TempDir())

	_, err := uc.RestaurarRespaldo(ctxAdmin(), port.RestaurarBackupInput{BackupID: "b-1", Confirmar: false})
	if err == nil || !errors.Is(err, domain.ErrDatosInvalidos) {
		t.Fatalf("sin confirmación explícita debe rechazarse, obtuve %v", err)
	}
}

func TestRestaurarSoloAceptaBackupCompletado(t *testing.T) {
	repo := &stubRespaldoRepository{}
	repo.respaldos = append(repo.respaldos, entity.RespaldoOperacion{ID: "b-pend", Estado: entity.EstadoEnProgreso})
	uc := nuevoUseCase(repo, &stubMotor{}, t.TempDir())

	_, err := uc.RestaurarRespaldo(ctxAdmin(), port.RestaurarBackupInput{BackupID: "b-pend", Confirmar: true})
	if err == nil || !errors.Is(err, domain.ErrDatosInvalidos) {
		t.Fatalf("un respaldo EN_PROGRESO no debe restaurarse, obtuve %v", err)
	}
}

func TestRestaurarFlujoAsincronoCompleto(t *testing.T) {
	repo := &stubRespaldoRepository{}
	dumpsDir := t.TempDir()
	ruta := filepath.Join(dumpsDir, "dbgs_backup_restaurable.dump")
	if err := os.WriteFile(ruta, []byte("dump"), 0600); err != nil {
		t.Fatal(err)
	}
	repo.respaldos = append(repo.respaldos, entity.RespaldoOperacion{
		ID: "b-full", Estado: entity.EstadoCompletado, RutaArchivo: ruta, FechaCreacion: time.Now(),
	})

	motor := &stubMotor{}
	uc := nuevoUseCase(repo, motor, dumpsDir)

	output, err := uc.RestaurarRespaldo(ctxAdmin(), port.RestaurarBackupInput{BackupID: "b-full", Confirmar: true, Usuario: "dba"})
	if err != nil {
		t.Fatalf("RestaurarRespaldo() error = %v", err)
	}

	deadline := time.Now().Add(2 * time.Second)
	var restauracion entity.Restauracion
	for time.Now().Before(deadline) {
		if repo.restauraciones[0].Estado == entity.EstadoCompletado {
			restauracion = repo.restauraciones[0]
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if restauracion.Estado != entity.EstadoCompletado {
		t.Fatalf("la restauración nunca completó: %+v", repo.restauraciones)
	}
	if len(motor.llamadasRestaurar) != 1 || motor.llamadasRestaurar[0] != ruta {
		t.Fatalf("pg_restore debe invocarse exactamente una vez sobre %s", ruta)
	}
	_ = output
}

// =========================================================================================================
// RETENCIÓN
// =========================================================================================================

func TestAplicarRetencionEliminaVencidosYRespetaEnCurso(t *testing.T) {
	repo := &stubRespaldoRepository{}
	dumpsDir := t.TempDir()

	crear := func(id string, edadDias int, estado string) entity.RespaldoOperacion {
		ruta := filepath.Join(dumpsDir, id+".dump")
		if err := os.WriteFile(ruta, []byte("x"), 0600); err != nil {
			t.Fatal(err)
		}
		b := entity.RespaldoOperacion{
			ID: id, Estado: estado, RutaArchivo: ruta,
			FechaCreacion: time.Now().AddDate(0, 0, -edadDias),
		}
		repo.respaldos = append(repo.respaldos, b)
		return b
	}
	viejo := crear("viejo", 90, entity.EstadoCompletado)
	enCursoVencido := crear("curso-vencido", 90, entity.EstadoEnProgreso)
	reciente := crear("reciente", 1, entity.EstadoCompletado)

	uc := nuevoUseCase(repo, &stubMotor{}, dumpsDir)
	output, err := uc.AplicarRetencion(ctxAdmin(), port.RetencionInput{DiasRetencion: 30, MaximoBackups: 10})
	if err != nil {
		t.Fatalf("AplicarRetencion() error = %v", err)
	}

	if len(output.IDsEliminados) != 1 || output.IDsEliminados[0] != viejo.ID {
		t.Fatalf("solo el respaldo vencido terminal debe eliminarse: %v", output.IDsEliminados)
	}
	if _, err := os.Stat(viejo.RutaArchivo); !os.IsNotExist(err) {
		t.Fatal("el archivo del respaldo eliminado debe borrarse físicamente")
	}
	if _, err := os.Stat(enCursoVencido.RutaArchivo); err != nil {
		t.Fatal("una operación EN_PROGRESO jamás debe tocarse aunque esté vencida")
	}
	encontradoReciente := false
	for _, b := range repo.respaldos {
		if b.ID == reciente.ID {
			encontradoReciente = true
		}
	}
	if !encontradoReciente {
		t.Fatal("el respaldo reciente debe sobrevivir a la retención")
	}
}
