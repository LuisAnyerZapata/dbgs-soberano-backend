package grpc

import (
	"context"
	"fmt"

	pb "DBGS_SOBERANO_BACKEND/api/proto/v1"
	"DBGS_SOBERANO_BACKEND/internal/application/port"
	"DBGS_SOBERANO_BACKEND/internal/domain"
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

func (h *SeguridadHandler) GetSetupStatus(ctx context.Context, req *pb.GetSetupStatusRequest) (*pb.GetSetupStatusResponse, error) {
	statusResult, err := h.seguridadUseCase.VerificarEstadoSetup(ctx)
	if err != nil {
		return nil, mapDomainErrorToGRPC(err)
	}
	return &pb.GetSetupStatusResponse{Initialized: statusResult.Inicializado}, nil
}

func (h *SeguridadHandler) CreateSetup(ctx context.Context, req *pb.CreateSetupRequest) (*pb.LoginResponse, error) {
	errs := domain.NewValidator()
	errs.Add(domain.ValidateRequired("username", req.GetUsername()))
	errs.Add(domain.ValidateRequired("password", req.GetPassword()))
	errs.Add(domain.ValidateMinLength("password", req.GetPassword(), 8))
	if err := errs.Validate(); err != nil {
		return nil, mapDomainErrorToGRPC(err)
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

func (h *SeguridadHandler) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	errs := domain.NewValidator()
	errs.Add(domain.ValidateRequired("email", req.GetEmail()))
	errs.Add(domain.ValidateRequired("password", req.GetPassword()))
	if err := errs.Validate(); err != nil {
		return nil, mapDomainErrorToGRPC(err)
	}

	result, err := h.seguridadUseCase.Login(ctx, req.Email, req.Password)
	if err != nil {
		return nil, mapDomainErrorToGRPC(err)
	}

	return &pb.LoginResponse{
		AccessToken: result.AccessToken,
		TokenType:   result.TokenType,
		ExpiresIn:   result.ExpiresIn,
	}, nil
}

func (h *SeguridadHandler) ValidarToken(ctx context.Context, req *pb.ValidarTokenRequest) (*pb.ValidarTokenResponse, error) {
	if err := domain.ValidateRequired("token", req.GetToken()); err != nil {
		return nil, mapDomainErrorToGRPC(err)
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

func (h *SeguridadHandler) VerificarPermiso(ctx context.Context, req *pb.VerificarPermisoRequest) (*pb.VerificarPermisoResponse, error) {
	errs := domain.NewValidator()
	errs.Add(domain.ValidateRequired("usuario_id", req.GetUsuarioId()))
	errs.Add(domain.ValidateRequired("recurso", req.GetRecurso()))
	errs.Add(domain.ValidateRequired("accion", req.GetAccion()))
	if err := errs.Validate(); err != nil {
		return nil, mapDomainErrorToGRPC(err)
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
	errs := domain.NewValidator()
	errs.Add(domain.ValidateName("nombre", req.GetNombre()))
	if err := errs.Validate(); err != nil {
		return nil, mapDomainErrorToGRPC(err)
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
	if err := domain.ValidateUUID("id", req.GetId()); err != nil {
		return nil, mapDomainErrorToGRPC(err)
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
	errs := domain.NewValidator()
	errs.Add(domain.ValidateUUID("id", req.GetId()))
	errs.Add(domain.ValidateName("nombre", req.GetNombre()))
	if err := errs.Validate(); err != nil {
		return nil, mapDomainErrorToGRPC(err)
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
	if err := domain.ValidateUUID("id", req.GetId()); err != nil {
		return nil, mapDomainErrorToGRPC(err)
	}

	if err := h.seguridadUseCase.EliminarRol(ctx, req.GetId()); err != nil {
		return nil, mapDomainErrorToGRPC(err)
	}

	return &pb.EliminarRolResponse{Mensaje: "Rol eliminado exitosamente"}, nil
}

func (h *SeguridadHandler) VincularPermisoRol(ctx context.Context, req *pb.VincularPermisoRolRequest) (*pb.VincularPermisoRolResponse, error) {
	errs := domain.NewValidator()
	errs.Add(domain.ValidateUUID("rol_id", req.GetRolId()))
	if len(req.GetPermisos()) == 0 {
		errs.Add(domain.RequiredError("permisos"))
	}
	if err := errs.Validate(); err != nil {
		return nil, mapDomainErrorToGRPC(err)
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
	errs := domain.NewValidator()
	errs.Add(domain.ValidateUUID("rol_id", req.GetRolId()))
	errs.Add(domain.ValidateRequired("permiso_codigo", req.GetPermisoCodigo()))
	if err := errs.Validate(); err != nil {
		return nil, mapDomainErrorToGRPC(err)
	}

	if err := h.seguridadUseCase.DesvincularPermiso(ctx, req.GetRolId(), req.GetPermisoCodigo()); err != nil {
		return nil, mapDomainErrorToGRPC(err)
	}

	return &pb.DesvincularPermisoRolResponse{Mensaje: "Permiso desvinculado exitosamente"}, nil
}

func (h *SeguridadHandler) ListarPermisosRol(ctx context.Context, req *pb.ListarPermisosRolRequest) (*pb.ListarPermisosRolResponse, error) {
	if err := domain.ValidateUUID("rol_id", req.GetRolId()); err != nil {
		return nil, mapDomainErrorToGRPC(err)
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
	errs := domain.NewValidator()
	errs.Add(domain.ValidateNoSpaces("username", req.GetUsername()))
	errs.Add(domain.ValidateMinLength("username", req.GetUsername(), 3))
	errs.Add(domain.ValidateMaxLength("username", req.GetUsername(), 50))
	errs.Add(domain.ValidatePassword("password", req.GetPassword()))
	errs.Add(domain.ValidateUUID("rol_id", req.GetRolId()))
	if req.GetEmail() != "" {
		errs.Add(domain.ValidateEmail("email", req.GetEmail()))
	}
	if err := errs.Validate(); err != nil {
		return nil, mapDomainErrorToGRPC(err)
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
	if err := domain.ValidateUUID("id", req.GetId()); err != nil {
		return nil, mapDomainErrorToGRPC(err)
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
	errs := domain.NewValidator()
	errs.Add(domain.ValidateUUID("id", req.GetId()))
	errs.Add(domain.ValidateUUID("rol_id", req.GetRolId()))
	if req.GetEmail() != "" {
		errs.Add(domain.ValidateEmail("email", req.GetEmail()))
	}
	if err := errs.Validate(); err != nil {
		return nil, mapDomainErrorToGRPC(err)
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
	if err := domain.ValidateUUID("id", req.GetId()); err != nil {
		return nil, mapDomainErrorToGRPC(err)
	}

	if err := h.seguridadUseCase.EliminarUsuario(ctx, req.GetId()); err != nil {
		return nil, mapDomainErrorToGRPC(err)
	}

	return &pb.EliminarUsuarioResponse{Mensaje: "Usuario eliminado exitosamente"}, nil
}
