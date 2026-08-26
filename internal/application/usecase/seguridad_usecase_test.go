package usecase

import (
	"context"
	"testing"

	"DBGS_SOBERANO_BACKEND/internal/application/port"
	"DBGS_SOBERANO_BACKEND/internal/domain"
	"DBGS_SOBERANO_BACKEND/internal/domain/entity"
	"DBGS_SOBERANO_BACKEND/internal/domain/repository"
)

type stubSeguridadRepository struct {
	usuario *entity.Usuario
	rol     *entity.Rol
}

func (s *stubSeguridadRepository) ObtenerUsuarioPorUsername(ctx context.Context, username string) (*entity.Usuario, error) {
	if s.usuario == nil || s.usuario.Username != username {
		return nil, entity.ErrEntidadNoEncontrada
	}
	return s.usuario, nil
}

func (s *stubSeguridadRepository) AutenticarUsuario(ctx context.Context, username, password string) (*entity.Usuario, error) {
	if username == "admin" && password == "dbgs-admin" {
		return s.usuario, nil
	}
	return nil, entity.ErrAccesoNoAutorizado
}

func (s *stubSeguridadRepository) ObtenerRolPorID(ctx context.Context, rolID string) (*entity.Rol, error) {
	if s.rol == nil || s.rol.ID != rolID {
		return nil, entity.ErrEntidadNoEncontrada
	}
	return s.rol, nil
}

func (s *stubSeguridadRepository) ValidarPermiso(ctx context.Context, rolID, permiso string) (bool, error) {
	for _, p := range s.rol.Permisos {
		if p == permiso {
			return true, nil
		}
	}
	return false, nil
}

func (s *stubSeguridadRepository) ContarSuperAdmins(ctx context.Context) (int64, error) {
	if s.usuario == nil || !s.usuario.Estado {
		return 0, nil
	}
	return 1, nil
}

func (s *stubSeguridadRepository) CrearUsuarioAdmin(ctx context.Context, username, email, passwordHash, rolID string) (*entity.Usuario, error) {
	s.usuario = &entity.Usuario{ID: "u-nuevo", Username: username, Email: email, RolID: rolID, Estado: true}
	return s.usuario, nil
}

func (s *stubSeguridadRepository) AsegurarRolSuperAdmin(ctx context.Context) (string, error) {
	if s.rol == nil {
		s.rol = &entity.Rol{ID: "r-admin", Nombre: "ADMIN_PLATFORM", Estado: true}
	}
	return s.rol.ID, nil
}

func (s *stubSeguridadRepository) CrearRol(ctx context.Context, nombre, descripcion string) (*entity.Rol, error) {
	return &entity.Rol{ID: "stub-rol", Nombre: nombre, Descripcion: descripcion}, nil
}
func (s *stubSeguridadRepository) ListarRoles(ctx context.Context) ([]entity.Rol, error) {
	return nil, nil
}
func (s *stubSeguridadRepository) ObtenerRolPorNombre(ctx context.Context, nombre string) (*entity.Rol, error) {
	return &entity.Rol{ID: "stub-rol", Nombre: nombre}, nil
}
func (s *stubSeguridadRepository) ActualizarRol(ctx context.Context, id, nombre, descripcion string) error {
	return nil
}
func (s *stubSeguridadRepository) EliminarRol(ctx context.Context, id string) error {
	return nil
}
func (s *stubSeguridadRepository) VincularPermisos(ctx context.Context, rolID string, codigos []string) (int64, error) {
	return int64(len(codigos)), nil
}
func (s *stubSeguridadRepository) DesvincularPermiso(ctx context.Context, rolID, permisoCodigo string) error {
	return nil
}
func (s *stubSeguridadRepository) ListarPermisosRol(ctx context.Context, rolID string) ([]string, error) {
	return nil, nil
}

func (s *stubSeguridadRepository) CrearUsuario(ctx context.Context, username, email, passwordHash, rolID string, esTecnico bool) (*entity.Usuario, error) {
	return &entity.Usuario{ID: "stub-user", Username: username, Email: email, RolID: rolID, EsTecnico: esTecnico, Estado: true}, nil
}
func (s *stubSeguridadRepository) ListarUsuarios(ctx context.Context) ([]entity.Usuario, error) {
	return nil, nil
}
func (s *stubSeguridadRepository) ObtenerUsuarioPorID(ctx context.Context, id string) (*entity.Usuario, error) {
	return &entity.Usuario{ID: id, Username: "stub"}, nil
}
func (s *stubSeguridadRepository) ActualizarUsuario(ctx context.Context, id, email, rolID string, esTecnico, estado bool) error {
	return nil
}
func (s *stubSeguridadRepository) EliminarUsuario(ctx context.Context, id string) error {
	return nil
}

var _ repository.SeguridadRepository = (*stubSeguridadRepository)(nil)

func TestLoginAutenticaUsuarioYGeneraToken(t *testing.T) {
	repo := &stubSeguridadRepository{
		usuario: &entity.Usuario{ID: "u1", Username: "admin", Email: "admin@example.com", RolID: "r1", Estado: true},
		rol:     &entity.Rol{ID: "r1", Nombre: "ADMIN", Permisos: []string{"usuarios:leer"}, Estado: true},
	}

	uc := NewSeguridadUseCase(repo, "secreto-de-prueba", 60)
	login, err := uc.Login(context.Background(), "admin", "dbgs-admin")
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	if login.AccessToken == "" {
		t.Fatal("Login() debe devolver un token")
	}
}

func TestValidarAccesoConPermisoValido(t *testing.T) {
	repo := &stubSeguridadRepository{
		usuario: &entity.Usuario{ID: "u1", Username: "admin", Email: "admin@example.com", RolID: "r1", Estado: true},
		rol:     &entity.Rol{ID: "r1", Nombre: "ADMIN", Permisos: []string{"usuarios:leer"}, Estado: true},
	}

	uc := NewSeguridadUseCase(repo, "secreto-de-prueba", 60)
	permitido, err := uc.ValidarAcceso(context.Background(), port.ValidarAccesoInput{Username: "admin", Permiso: "usuarios:leer"})
	if err != nil {
		t.Fatalf("ValidarAcceso() error = %v", err)
	}
	if !permitido {
		t.Fatal("ValidarAcceso() esperaba true")
	}
}

func TestLoginRechazaCredencialesInvalidas(t *testing.T) {
	repo := &stubSeguridadRepository{
		usuario: &entity.Usuario{ID: "u1", Username: "admin", Email: "admin@example.com", RolID: "r1", Estado: true},
		rol:     &entity.Rol{ID: "r1", Nombre: "ADMIN", Permisos: []string{"usuarios:leer"}, Estado: true},
	}

	uc := NewSeguridadUseCase(repo, "secreto-de-prueba", 60)
	_, err := uc.Login(context.Background(), "admin", "password-malo")
	if err == nil {
		t.Fatal("Login() esperaba error para credenciales inválidas")
	}
	if err != domain.ErrAccesoNoAutorizado {
		t.Fatalf("Login() error = %v, want %v", err, domain.ErrAccesoNoAutorizado)
	}
}
