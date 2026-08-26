package grpc

import (
    "context"
    "fmt"

    pb "DBGS_SOBERANO_BACKEND/api/proto/v1"
    "DBGS_SOBERANO_BACKEND/internal/application/port"

    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"
)

// SeguridadHandler implementa los métodos gRPC para el servicio de seguridad
type SeguridadHandler struct {
    pb.UnimplementedSeguridadServiceServer
    seguridadUseCase port.SeguridadUseCasePort
}

// NewSeguridadHandler inyecta el caso de uso de seguridad en el adaptador de entrada
func NewSeguridadHandler(seguridadUC port.SeguridadUseCasePort) *SeguridadHandler {
    return &SeguridadHandler{seguridadUseCase: seguridadUC}
}

// =========================================================================================================
// ENDPOINTS DE BOOTSTRAPPING (SETUP)
// =========================================================================================================

// GetSetupStatus permite a la interfaz saber si debe renderizar la pantalla de bienvenida
func (h *SeguridadHandler) GetSetupStatus(ctx context.Context, req *pb.GetSetupStatusRequest) (*pb.GetSetupStatusResponse, error) {
    statusResult, err := h.seguridadUseCase.VerificarEstadoSetup(ctx)
    if err != nil {
        return nil, mapDomainErrorToGRPC(err)
    }

    return &pb.GetSetupStatusResponse{Initialized: statusResult.Inicializado}, nil
}

// CreateSetup orquesta la creación del primer superadministrador de forma segura
func (h *SeguridadHandler) CreateSetup(ctx context.Context, req *pb.CreateSetupRequest) (*pb.LoginResponse, error) {
    // Sanitización básica a nivel de API
    if req.GetUsername() == "" || req.GetPassword() == "" {
        return nil, status.Error(codes.InvalidArgument, "username y password son obligatorios para el setup")
    }

    input := port.EjecutarSetupInput{
        Username: req.GetUsername(),
        Password: req.GetPassword(),
        Email:    req.GetEmail(),
    }

    result, err := h.seguridadUseCase.EjecutarSetup(ctx, input)
    if err != nil {
        return nil, mapDomainErrorToGRPC(err)
    }

    return &pb.LoginResponse{
        AccessToken: result.AccessToken,
        TokenType:   result.TokenType,
        ExpiresIn:   result.ExpiresIn,
    }, nil
}

// =========================================================================================================
// ENDPOINTS DE AUTENTICACIÓN CONTINUA
// =========================================================================================================

// Login autentica a un usuario y devuelve un JWT
func (h *SeguridadHandler) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
    if req.Username == "" || req.Password == "" {
        return nil, status.Error(codes.InvalidArgument, "el usuario y la contraseña son obligatorios")
    }

    result, err := h.seguridadUseCase.Login(ctx, req.Username, req.Password)
    if err != nil {
        return nil, mapDomainErrorToGRPC(err)
    }

    return &pb.LoginResponse{
        AccessToken: result.AccessToken,
        TokenType:   result.TokenType,
        ExpiresIn:   result.ExpiresIn,
    }, nil
}

// ValidarToken verifica la vigencia de un token JWT
func (h *SeguridadHandler) ValidarToken(ctx context.Context, req *pb.ValidarTokenRequest) (*pb.ValidarTokenResponse, error) {
    if req.Token == "" {
        return nil, status.Error(codes.InvalidArgument, "el token es obligatorio")
    }

    result, err := h.seguridadUseCase.ValidarToken(ctx, req.Token)
    if err != nil {
        return nil, mapDomainErrorToGRPC(err)
    }

    return &pb.ValidarTokenResponse{
        EsValido:  result.Valid,
        Username:  result.Username,
        UsuarioId: result.UserID,
        Rol:       result.Rol,
    }, nil
}

// VerificarPermiso comprueba si un usuario tiene acceso a un recurso específico
func (h *SeguridadHandler) VerificarPermiso(ctx context.Context, req *pb.VerificarPermisoRequest) (*pb.VerificarPermisoResponse, error) {
    if req.UsuarioId == "" || req.Recurso == "" || req.Accion == "" {
        return nil, status.Error(codes.InvalidArgument, "usuario, recurso y acción son obligatorios")
    }

    permitido, err := h.seguridadUseCase.ValidarAcceso(ctx, port.ValidarAccesoInput{
        Username: req.UsuarioId,
        Permiso:  req.Recurso + ":" + req.Accion,
    })
    if err != nil {
        return nil, mapDomainErrorToGRPC(err)
    }

    return &pb.VerificarPermisoResponse{Permitido: permitido}, nil
}

// =========================================================================================================
// CRUD DE ROLES
// =========================================================================================================

func (h *SeguridadHandler) CrearRol(ctx context.Context, req *pb.CrearRolRequest) (*pb.CrearRolResponse, error) {
    if req.GetNombre() == "" {
        return nil, status.Error(codes.InvalidArgument, "el nombre del rol es obligatorio")
    }

    rol, err := h.seguridadUseCase.CrearRol(ctx, req.GetNombre(), req.GetDescription())
    if err != nil {
        return nil, mapDomainErrorToGRPC(err)
    }

    return &pb.CrearRolResponse{
        Id:          rol.ID,
        Nombre:      rol.Nombre,
        Description: rol.Descripcion,
        Mensaje:     "Rol creado exitosamente",
    }, nil
}

func (h *SeguridadHandler) ListarRoles(ctx context.Context, req *pb.ListarRolesRequest) (*pb.ListarRolesResponse, error) {
    roles, err := h.seguridadUseCase.ListarRoles(ctx)
    if err != nil {
        return nil, mapDomainErrorToGRPC(err)
    }

    var pbRoles []*pb.RolProto
    for _, r := range roles {
        pbRoles = append(pbRoles, &pb.RolProto{
            Id:          r.ID,
            Nombre:      r.Nombre,
            Description: r.Descripcion,
        })
    }

    return &pb.ListarRolesResponse{Roles: pbRoles}, nil
}

func (h *SeguridadHandler) ObtenerRol(ctx context.Context, req *pb.ObtenerRolRequest) (*pb.ObtenerRolResponse, error) {
    if req.GetId() == "" {
        return nil, status.Error(codes.InvalidArgument, "el ID del rol es obligatorio")
    }

    rol, err := h.seguridadUseCase.ObtenerRol(ctx, req.GetId())
    if err != nil {
        return nil, mapDomainErrorToGRPC(err)
    }

    return &pb.ObtenerRolResponse{
        Id:          rol.ID,
        Nombre:      rol.Nombre,
        Description: rol.Descripcion,
    }, nil
}

func (h *SeguridadHandler) ActualizarRol(ctx context.Context, req *pb.ActualizarRolRequest) (*pb.ActualizarRolResponse, error) {
    if req.GetId() == "" || req.GetNombre() == "" {
        return nil, status.Error(codes.InvalidArgument, "el ID y nombre del rol son obligatorios")
    }

    if err := h.seguridadUseCase.ActualizarRol(ctx, req.GetId(), req.GetNombre(), req.GetDescription()); err != nil {
        return nil, mapDomainErrorToGRPC(err)
    }

    return &pb.ActualizarRolResponse{
        Id:          req.GetId(),
        Nombre:      req.GetNombre(),
        Description: req.GetDescription(),
        Mensaje:     "Rol actualizado exitosamente",
    }, nil
}

func (h *SeguridadHandler) EliminarRol(ctx context.Context, req *pb.EliminarRolRequest) (*pb.EliminarRolResponse, error) {
    if req.GetId() == "" {
        return nil, status.Error(codes.InvalidArgument, "el ID del rol es obligatorio")
    }

    if err := h.seguridadUseCase.EliminarRol(ctx, req.GetId()); err != nil {
        return nil, mapDomainErrorToGRPC(err)
    }

    return &pb.EliminarRolResponse{Mensaje: "Rol eliminado exitosamente"}, nil
}

func (h *SeguridadHandler) VincularPermisoRol(ctx context.Context, req *pb.VincularPermisoRolRequest) (*pb.VincularPermisoRolResponse, error) {
    if req.GetRolId() == "" || len(req.GetPermisos()) == 0 {
        return nil, status.Error(codes.InvalidArgument, "el ID del rol y al menos un permiso son obligatorios")
    }

    vinculados, err := h.seguridadUseCase.VincularPermisos(ctx, req.GetRolId(), req.GetPermisos())
    if err != nil {
        return nil, mapDomainErrorToGRPC(err)
    }

    return &pb.VincularPermisoRolResponse{
        Vinculados: int32(vinculados),
        Mensaje:    fmt.Sprintf("%d permiso(s) vinculado(s) exitosamente", vinculados),
    }, nil
}

func (h *SeguridadHandler) DesvincularPermisoRol(ctx context.Context, req *pb.DesvincularPermisoRolRequest) (*pb.DesvincularPermisoRolResponse, error) {
    if req.GetRolId() == "" || req.GetPermisoCodigo() == "" {
        return nil, status.Error(codes.InvalidArgument, "el ID del rol y el código del permiso son obligatorios")
    }

    if err := h.seguridadUseCase.DesvincularPermiso(ctx, req.GetRolId(), req.GetPermisoCodigo()); err != nil {
        return nil, mapDomainErrorToGRPC(err)
    }

    return &pb.DesvincularPermisoRolResponse{Mensaje: "Permiso desvinculado exitosamente"}, nil
}

func (h *SeguridadHandler) ListarPermisosRol(ctx context.Context, req *pb.ListarPermisosRolRequest) (*pb.ListarPermisosRolResponse, error) {
    if req.GetRolId() == "" {
        return nil, status.Error(codes.InvalidArgument, "el ID del rol es obligatorio")
    }

    permisos, err := h.seguridadUseCase.ListarPermisosRol(ctx, req.GetRolId())
    if err != nil {
        return nil, mapDomainErrorToGRPC(err)
    }

    return &pb.ListarPermisosRolResponse{Permisos: permisos}, nil
}

// =========================================================================================================
// CRUD DE USUARIOS
// =========================================================================================================

func (h *SeguridadHandler) CrearUsuario(ctx context.Context, req *pb.CrearUsuarioRequest) (*pb.CrearUsuarioResponse, error) {
    if req.GetUsername() == "" || req.GetPassword() == "" || req.GetRolId() == "" {
        return nil, status.Error(codes.InvalidArgument, "username, password y rol_id son obligatorios")
    }

    usuario, err := h.seguridadUseCase.CrearUsuario(ctx, req.GetUsername(), req.GetEmail(), req.GetPassword(), req.GetRolId(), req.GetEsTecnico())
    if err != nil {
        return nil, mapDomainErrorToGRPC(err)
    }

    return &pb.CrearUsuarioResponse{
        Id:        usuario.ID,
        Username:  usuario.Username,
        Email:     usuario.Email,
        RolId:     usuario.RolID,
        EsTecnico: usuario.EsTecnico,
        Estado:    usuario.Estado,
        Mensaje:   "Usuario creado exitosamente",
    }, nil
}

func (h *SeguridadHandler) ListarUsuarios(ctx context.Context, req *pb.ListarUsuariosRequest) (*pb.ListarUsuariosResponse, error) {
    usuarios, err := h.seguridadUseCase.ListarUsuarios(ctx)
    if err != nil {
        return nil, mapDomainErrorToGRPC(err)
    }

    var pbUsuarios []*pb.UsuarioProto
    for _, u := range usuarios {
        pbUsuarios = append(pbUsuarios, &pb.UsuarioProto{
            Id:        u.ID,
            Username:  u.Username,
            Email:     u.Email,
            RolId:     u.RolID,
            RolNombre: u.Rol.Nombre,
            EsTecnico: u.EsTecnico,
            Estado:    u.Estado,
        })
    }

    return &pb.ListarUsuariosResponse{Usuarios: pbUsuarios}, nil
}

func (h *SeguridadHandler) ObtenerUsuario(ctx context.Context, req *pb.ObtenerUsuarioRequest) (*pb.ObtenerUsuarioResponse, error) {
    if req.GetId() == "" {
        return nil, status.Error(codes.InvalidArgument, "el ID del usuario es obligatorio")
    }

    usuario, err := h.seguridadUseCase.ObtenerUsuario(ctx, req.GetId())
    if err != nil {
        return nil, mapDomainErrorToGRPC(err)
    }

    return &pb.ObtenerUsuarioResponse{
        Id:        usuario.ID,
        Username:  usuario.Username,
        Email:     usuario.Email,
        RolId:     usuario.RolID,
        RolNombre: usuario.Rol.Nombre,
        EsTecnico: usuario.EsTecnico,
        Estado:    usuario.Estado,
    }, nil
}

func (h *SeguridadHandler) ActualizarUsuario(ctx context.Context, req *pb.ActualizarUsuarioRequest) (*pb.ActualizarUsuarioResponse, error) {
    if req.GetId() == "" || req.GetRolId() == "" {
        return nil, status.Error(codes.InvalidArgument, "el ID del usuario y rol_id son obligatorios")
    }

    if err := h.seguridadUseCase.ActualizarUsuario(ctx, req.GetId(), req.GetEmail(), req.GetRolId(), req.GetEsTecnico(), req.GetEstado()); err != nil {
        return nil, mapDomainErrorToGRPC(err)
    }

    return &pb.ActualizarUsuarioResponse{
        Id:        req.GetId(),
        Email:     req.GetEmail(),
        RolId:     req.GetRolId(),
        EsTecnico: req.GetEsTecnico(),
        Estado:    req.GetEstado(),
        Mensaje:   "Usuario actualizado exitosamente",
    }, nil
}

func (h *SeguridadHandler) EliminarUsuario(ctx context.Context, req *pb.EliminarUsuarioRequest) (*pb.EliminarUsuarioResponse, error) {
    if req.GetId() == "" {
        return nil, status.Error(codes.InvalidArgument, "el ID del usuario es obligatorio")
    }

    if err := h.seguridadUseCase.EliminarUsuario(ctx, req.GetId()); err != nil {
        return nil, mapDomainErrorToGRPC(err)
    }

    return &pb.EliminarUsuarioResponse{Mensaje: "Usuario eliminado exitosamente"}, nil
}