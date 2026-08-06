package repository

import (
	"context"

	"DBGS_SOBERANO_BACKEND/internal/domain/entity"
)

// SeguridadRepository define las operaciones para validar usuarios, roles y permisos (RBAC)
type SeguridadRepository interface {
	// ObtenerUsuarioPorUsername busca la identidad del usuario o cuenta técnica
	ObtenerUsuarioPorUsername(ctx context.Context, username string) (*entity.Usuario, error)

	// AutenticarUsuario valida credenciales de acceso de forma inicial.
	AutenticarUsuario(ctx context.Context, username, password string) (*entity.Usuario, error)

	// ObtenerRolPorID obtiene las definiciones del rol y sus permisos asociados
	ObtenerRolPorID(ctx context.Context, rolID string) (*entity.Rol, error)

	// ValidarPermiso verifica si un rol específico posee el permiso requerido
	ValidarPermiso(ctx context.Context, rolID, permiso string) (bool, error)
}
