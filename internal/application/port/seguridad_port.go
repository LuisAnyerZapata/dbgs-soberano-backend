package port

import (
	"context"

	"DBGS_SOBERANO_BACKEND/internal/domain/entity"
)

type ValidarAccesoInput struct {
	Username string
	Permiso  string
}

type LoginResult struct {
	AccessToken string
	TokenType   string
	ExpiresIn   int64
	Username    string
}

type TokenValidationResult struct {
	Valid     bool
	Username  string
	UserID    string
	Rol       string
	ExpiresAt int64
}

type SetupStatusResult struct {
    Inicializado bool
}

type EjecutarSetupInput struct {
    Username string
    Password string
    Email    string
}


// SeguridadUseCasePort define las operaciones de seguridad disponibles para los adaptadores de entrada
type SeguridadUseCasePort interface {
    // --- Setup Inicial ---
    VerificarEstadoSetup(ctx context.Context) (*SetupStatusResult, error)
    EjecutarSetup(ctx context.Context, input EjecutarSetupInput) (*LoginResult, error)

    ObtenerPerfilUsuario(ctx context.Context, username string) (*entity.Usuario, error)
    ValidarAcceso(ctx context.Context, input ValidarAccesoInput) (bool, error)
    Login(ctx context.Context, username, password string) (*LoginResult, error)
    ValidarToken(ctx context.Context, token string) (*TokenValidationResult, error)

    // CRUD de Roles
    CrearRol(ctx context.Context, nombre, descripcion string) (*entity.Rol, error)
    ListarRoles(ctx context.Context) ([]entity.Rol, error)
    ObtenerRol(ctx context.Context, id string) (*entity.Rol, error)
    ActualizarRol(ctx context.Context, id, nombre, descripcion string) error
    EliminarRol(ctx context.Context, id string) error

    // Permisos de Roles
    VincularPermisos(ctx context.Context, rolID string, codigos []string) (int64, error)
    DesvincularPermiso(ctx context.Context, rolID, permisoID string) error
    ListarPermisosRol(ctx context.Context, rolID string) ([]string, error)
}