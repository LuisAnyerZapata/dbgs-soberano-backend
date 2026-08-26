package usecase

import (
    "context"
    "fmt"
    "time"

    "DBGS_SOBERANO_BACKEND/internal/application/port"
    "DBGS_SOBERANO_BACKEND/internal/domain"
    "DBGS_SOBERANO_BACKEND/internal/domain/entity"
    "DBGS_SOBERANO_BACKEND/internal/domain/repository"

    "github.com/golang-jwt/jwt/v5"
    "golang.org/x/crypto/bcrypt"
)

// Coste de procesamiento para los hashes bcrypt (estándar de seguridad actual)
const bcryptCost = 12

// seguridadUseCase implementa la lógica de negocio para autenticación, autorización y configuración inicial
type seguridadUseCase struct {
    repo      repository.SeguridadRepository
    jwtSecret string
    tokenTTL  time.Duration
}

// NewSeguridadUseCase inicializa el caso de uso inyectando el repositorio y la configuración de seguridad
func NewSeguridadUseCase(repo repository.SeguridadRepository, jwtSecret string, tokenTTLMinutes int) port.SeguridadUseCasePort {
    return &seguridadUseCase{
        repo:      repo,
        jwtSecret: jwtSecret,
        tokenTTL:  time.Duration(tokenTTLMinutes) * time.Minute,
    }
}

// =========================================================================================================
// FLUJO DE BOOTSTRAPPING (FIRST-TIME SETUP)
// =========================================================================================================

// VerificarEstadoSetup consulta si ya existe algún usuario con privilegios de Super Administrador
func (uc *seguridadUseCase) VerificarEstadoSetup(ctx context.Context) (*port.SetupStatusResult, error) {
    count, err := uc.repo.ContarSuperAdmins(ctx)
    if err != nil {
        return nil, err
    }
    return &port.SetupStatusResult{Inicializado: count > 0}, nil
}

// EjecutarSetup crea el usuario raíz si y solo si el sistema no ha sido inicializado previamente.
// Esto previene la creación de múltiples superadministradores o la sobreescritura de credenciales.
func (uc *seguridadUseCase) EjecutarSetup(ctx context.Context, input port.EjecutarSetupInput) (*port.LoginResult, error) {
    // 1. Doble verificación de seguridad
    status, err := uc.VerificarEstadoSetup(ctx)
    if err != nil {
        return nil, err
    }
    if status.Inicializado {
        return nil, domain.ErrAccesoNoAutorizado // El sistema ya fue configurado, rechazar petición
    }

    // 2. Validaciones de negocio básicas
    if input.Username == "" || input.Password == "" {
        return nil, domain.ErrDatosInvalidos
    }

    // 3. Cifrado seguro de la contraseña
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcryptCost)
    if err != nil {
        return nil, domain.ErrErrorInterno
    }

    // 4. Auto-provisionamiento del rol principal.
    // El sistema asegura su propia existencia sin depender de seeds manuales previos.
    rolID, err := uc.repo.AsegurarRolSuperAdmin(ctx)
    if err != nil {
        return nil, fmt.Errorf("error crítico al inicializar el rol de administrador principal: %w", err)
    }

    // 5. Persistencia del nuevo usuario asignándole el rol soberano principal
    nuevoUsuario, err := uc.repo.CrearUsuarioAdmin(ctx, input.Username, input.Email, string(hashedPassword), rolID)
    if err != nil {
        return nil, err
    }

    // 6. Generación de tokens de sesión para loguearse automáticamente
    return uc.generarTokens(nuevoUsuario)
}

// =========================================================================================================
// FLUJO DE AUTENTICACIÓN CONTINUA
// =========================================================================================================

// Login valida credenciales y emite un par de tokens JWT
func (uc *seguridadUseCase) Login(ctx context.Context, username, password string) (*port.LoginResult, error) {
    if username == "" || password == "" {
        return nil, domain.ErrDatosInvalidos
    }

    usuario, err := uc.repo.AutenticarUsuario(ctx, username, password)
    if err != nil {
        return nil, err
    }

    return uc.generarTokens(usuario)
}

// ValidarToken decodifica un JWT y verifica su firma y tiempo de expiración
func (uc *seguridadUseCase) ValidarToken(ctx context.Context, tokenStr string) (*port.TokenValidationResult, error) {
    token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, fmt.Errorf("método de firma inesperado: %v", token.Header["alg"])
        }
        return []byte(uc.jwtSecret), nil
    })

    if err != nil {
        return nil, domain.ErrAccesoNoAutorizado
    }

    if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
        return &port.TokenValidationResult{
            Valid:     true,
            UserID:    claims["sub"].(string),
            Username:  claims["username"].(string),
            Rol:       claims["rol"].(string),
            ExpiresAt: int64(claims["exp"].(float64)),
        }, nil
    }

    return nil, domain.ErrAccesoNoAutorizado
}

// ObtenerPerfilUsuario busca un usuario por su nombre de usuario sin validar contraseñas
func (uc *seguridadUseCase) ObtenerPerfilUsuario(ctx context.Context, username string) (*entity.Usuario, error) {
    return uc.repo.ObtenerUsuarioPorUsername(ctx, username)
}

// ValidarAcceso comprueba si un usuario posee un permiso específico basado en su rol
func (uc *seguridadUseCase) ValidarAcceso(ctx context.Context, input port.ValidarAccesoInput) (bool, error) {
    usuario, err := uc.repo.ObtenerUsuarioPorUsername(ctx, input.Username)
    if err != nil {
        return false, err
    }
    return uc.repo.ValidarPermiso(ctx, usuario.RolID, input.Permiso)
}

// =========================================================================================================
// MÉTODOS AUXILIARES PRIVADOS
// =========================================================================================================

// generarTokens es una función privada que encapsula la lógica de creación del JWT
func (uc *seguridadUseCase) generarTokens(usuario *entity.Usuario) (*port.LoginResult, error) {
    // Obtener detalles del rol para incluirlos en el token
    rol, err := uc.repo.ObtenerRolPorID(context.Background(), usuario.RolID)
    if err != nil {
        return nil, domain.ErrErrorInterno
    }

    // Crear claims (payload del token)
    claims := jwt.MapClaims{
        "sub":      usuario.ID,
        "username": usuario.Username,
        "rol":      rol.Nombre,
        "exp":      time.Now().Add(uc.tokenTTL).Unix(),
        "iat":      time.Now().Unix(),
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    tokenString, err := token.SignedString([]byte(uc.jwtSecret))
    if err != nil {
        return nil, domain.ErrErrorInterno
    }

    return &port.LoginResult{
        AccessToken: tokenString,
        TokenType:   "Bearer",
        ExpiresIn:   int64(uc.tokenTTL.Seconds()),
        Username:    usuario.Username,
    }, nil
}

// =========================================================================================================
// CRUD DE ROLES
// =========================================================================================================

// CrearRol crea un nuevo rol en el sistema
func (uc *seguridadUseCase) CrearRol(ctx context.Context, nombre, descripcion string) (*entity.Rol, error) {
    if nombre == "" {
        return nil, domain.ErrDatosInvalidos
    }
    return uc.repo.CrearRol(ctx, nombre, descripcion)
}

// ListarRoles obtiene todos los roles registrados
func (uc *seguridadUseCase) ListarRoles(ctx context.Context) ([]entity.Rol, error) {
    return uc.repo.ListarRoles(ctx)
}

// ObtenerRol busca un rol por ID
func (uc *seguridadUseCase) ObtenerRol(ctx context.Context, id string) (*entity.Rol, error) {
    rol, err := uc.repo.ObtenerRolPorID(ctx, id)
    if err != nil {
        return nil, err
    }
    return rol, nil
}

// ActualizarRol modifica nombre y descripción de un rol
func (uc *seguridadUseCase) ActualizarRol(ctx context.Context, id, nombre, descripcion string) error {
    if id == "" || nombre == "" {
        return domain.ErrDatosInvalidos
    }
    return uc.repo.ActualizarRol(ctx, id, nombre, descripcion)
}

// EliminarRol elimina un rol del sistema
func (uc *seguridadUseCase) EliminarRol(ctx context.Context, id string) error {
    if id == "" {
        return domain.ErrDatosInvalidos
    }
    return uc.repo.EliminarRol(ctx, id)
}

// VincularPermisos vincula múltiples permisos a un rol
func (uc *seguridadUseCase) VincularPermisos(ctx context.Context, rolID string, codigos []string) (int64, error) {
    if rolID == "" || len(codigos) == 0 {
        return 0, domain.ErrDatosInvalidos
    }
    return uc.repo.VincularPermisos(ctx, rolID, codigos)
}

// DesvincularPermiso elimina un permiso de un rol
func (uc *seguridadUseCase) DesvincularPermiso(ctx context.Context, rolID, permisoID string) error {
    if rolID == "" || permisoID == "" {
        return domain.ErrDatosInvalidos
    }
    return uc.repo.DesvincularPermiso(ctx, rolID, permisoID)
}

// ListarPermisosRol obtiene los códigos de permisos de un rol
func (uc *seguridadUseCase) ListarPermisosRol(ctx context.Context, rolID string) ([]string, error) {
    if rolID == "" {
        return nil, domain.ErrDatosInvalidos
    }
    return uc.repo.ListarPermisosRol(ctx, rolID)
}