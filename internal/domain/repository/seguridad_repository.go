package repository

import (
    "context"
    "DBGS_SOBERANO_BACKEND/internal/domain/entity"
)

// SeguridadRepository define el contrato para la persistencia de seguridad y usuarios
type SeguridadRepository interface {
    ObtenerUsuarioPorUsername(ctx context.Context, username string) (*entity.Usuario, error)
    AutenticarUsuario(ctx context.Context, username, password string) (*entity.Usuario, error)
    ObtenerRolPorID(ctx context.Context, rolID string) (*entity.Rol, error)
    ValidarPermiso(ctx context.Context, rolID, permiso string) (bool, error)
    
    // ContarSuperAdmins verifica cuántos usuarios activos con rol de administrador principal existen.
    // Fundamental para el flujo de Bootstrapping (First-Time Setup).
    ContarSuperAdmins(ctx context.Context) (int64, error)

    // CrearUsuarioAdmin inserta el primer usuario del sistema durante el setup.
    // Retorna el usuario creado (sin el hash en la respuesta) o un error.
    CrearUsuarioAdmin(ctx context.Context, username, email, passwordHash, rolID string) (*entity.Usuario, error)

    // AsegurarRolSuperAdmin garantiza que el rol de máximo privilegio exista en la base de datos.
    // Si no existe, lo crea. Si existe, devuelve su ID. Vital para el Bootstrapping sin dependencia de Seeds.
    AsegurarRolSuperAdmin(ctx context.Context) (string, error)

    // CRUD de Roles
    CrearRol(ctx context.Context, nombre, descripcion string) (*entity.Rol, error)
    ListarRoles(ctx context.Context) ([]entity.Rol, error)
    ObtenerRolPorNombre(ctx context.Context, nombre string) (*entity.Rol, error)
    ActualizarRol(ctx context.Context, id, nombre, descripcion string) error
    EliminarRol(ctx context.Context, id string) error

    // Permisos de Roles
    VincularPermisos(ctx context.Context, rolID string, codigos []string) (int64, error)
    DesvincularPermiso(ctx context.Context, rolID, permisoID string) error
    ListarPermisosRol(ctx context.Context, rolID string) ([]string, error)

    // CRUD de Usuarios
    CrearUsuario(ctx context.Context, username, email, passwordHash, rolID string, esTecnico bool) (*entity.Usuario, error)
    ListarUsuarios(ctx context.Context) ([]entity.Usuario, error)
    ObtenerUsuarioPorID(ctx context.Context, id string) (*entity.Usuario, error)
    ActualizarUsuario(ctx context.Context, id, email, rolID string, esTecnico, estado bool) error
    EliminarUsuario(ctx context.Context, id string) error
}