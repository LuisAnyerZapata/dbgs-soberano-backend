package grpc

import (
	"context"

	dbgsv1 "DBGS_SOBERANO_BACKEND/api/proto/v1"
	"DBGS_SOBERANO_BACKEND/internal/application/port"
	"DBGS_SOBERANO_BACKEND/internal/domain"

	"google.golang.org/protobuf/types/known/structpb"
)

type DatosDinamicosHandler struct {
	dbgsv1.UnimplementedDatosDinamicosServiceServer
	useCase port.DatosDinamicosPort
}

func NewDatosDinamicosHandler(uc port.DatosDinamicosPort) *DatosDinamicosHandler {
	return &DatosDinamicosHandler{useCase: uc}
}

func mapToStruct(m map[string]interface{}) (*structpb.Struct, error) {
	if m == nil {
		return structpb.NewStruct(nil)
	}
	return structpb.NewStruct(m)
}

func mapsToStructs(maps []map[string]interface{}) ([]*structpb.Struct, error) {
	structs := make([]*structpb.Struct, len(maps))
	for i, m := range maps {
		s, err := mapToStruct(m)
		if err != nil {
			return nil, err
		}
		structs[i] = s
	}
	return structs, nil
}

func (h *DatosDinamicosHandler) ListarRegistros(ctx context.Context, req *dbgsv1.ListarRegistrosRequest) (*dbgsv1.ListarRegistrosResponse, error) {
	if err := domain.ValidateRequired("nombre_tabla", req.GetNombreTabla()); err != nil {
		return nil, mapDomainErrorToGRPC(err)
	}

	input := port.ListarRegistrosInput{
		NombreTabla: req.GetNombreTabla(),
		Limite:      int(req.GetLimit()),
		Offset:      int(req.GetOffset()),
	}

	output, err := h.useCase.ListarRegistros(ctx, input)
	if err != nil {
		return nil, mapDomainErrorToGRPC(err)
	}

	structs, err := mapsToStructs(output.Registros)
	if err != nil {
		return nil, mapDomainErrorToGRPC(err)
	}

	return &dbgsv1.ListarRegistrosResponse{
		Registros: structs,
		Total:     int32(output.Total),
	}, nil
}

func (h *DatosDinamicosHandler) ObtenerRegistro(ctx context.Context, req *dbgsv1.ObtenerRegistroRequest) (*dbgsv1.ObtenerRegistroResponse, error) {
	errs := domain.NewValidator()
	errs.Add(domain.ValidateRequired("nombre_tabla", req.GetNombreTabla()))
	errs.Add(domain.ValidateRequired("id", req.GetId()))
	if err := errs.Validate(); err != nil {
		return nil, mapDomainErrorToGRPC(err)
	}

	input := port.ObtenerRegistroInput{
		NombreTabla: req.GetNombreTabla(),
		ID:          req.GetId(),
	}

	registro, err := h.useCase.ObtenerRegistro(ctx, input)
	if err != nil {
		return nil, mapDomainErrorToGRPC(err)
	}

	s, err := mapToStruct(registro)
	if err != nil {
		return nil, mapDomainErrorToGRPC(err)
	}

	return &dbgsv1.ObtenerRegistroResponse{Registro: s}, nil
}

func (h *DatosDinamicosHandler) CrearRegistro(ctx context.Context, req *dbgsv1.CrearRegistroRequest) (*dbgsv1.CrearRegistroResponse, error) {
	errs := domain.NewValidator()
	errs.Add(domain.ValidateRequired("nombre_tabla", req.GetNombreTabla()))
	if req.GetDatos() == nil || len(req.GetDatos().AsMap()) == 0 {
		errs.Add(domain.RequiredError("datos"))
	}
	if err := errs.Validate(); err != nil {
		return nil, mapDomainErrorToGRPC(err)
	}

	datosMap := req.GetDatos().AsMap()

	input := port.CrearRegistroInput{
		NombreTabla: req.GetNombreTabla(),
		Datos:       datosMap,
	}

	nuevoID, err := h.useCase.CrearRegistro(ctx, input)
	if err != nil {
		return nil, mapDomainErrorToGRPC(err)
	}

	return &dbgsv1.CrearRegistroResponse{Id: nuevoID}, nil
}

func (h *DatosDinamicosHandler) ActualizarRegistro(ctx context.Context, req *dbgsv1.ActualizarRegistroRequest) (*dbgsv1.ActualizarRegistroResponse, error) {
	errs := domain.NewValidator()
	errs.Add(domain.ValidateRequired("nombre_tabla", req.GetNombreTabla()))
	errs.Add(domain.ValidateRequired("id", req.GetId()))
	if req.GetDatos() == nil || len(req.GetDatos().AsMap()) == 0 {
		errs.Add(domain.RequiredError("datos"))
	}
	if err := errs.Validate(); err != nil {
		return nil, mapDomainErrorToGRPC(err)
	}

	datosMap := req.GetDatos().AsMap()

	input := port.ActualizarRegistroInput{
		NombreTabla: req.GetNombreTabla(),
		ID:          req.GetId(),
		Datos:       datosMap,
	}

	registro, err := h.useCase.ActualizarRegistro(ctx, input)
	if err != nil {
		return nil, mapDomainErrorToGRPC(err)
	}

	s, err := mapToStruct(registro)
	if err != nil {
		return nil, mapDomainErrorToGRPC(err)
	}

	return &dbgsv1.ActualizarRegistroResponse{Registro: s}, nil
}

func (h *DatosDinamicosHandler) EliminarRegistro(ctx context.Context, req *dbgsv1.EliminarRegistroRequest) (*dbgsv1.EliminarRegistroResponse, error) {
	errs := domain.NewValidator()
	errs.Add(domain.ValidateRequired("nombre_tabla", req.GetNombreTabla()))
	errs.Add(domain.ValidateRequired("id", req.GetId()))
	if err := errs.Validate(); err != nil {
		return nil, mapDomainErrorToGRPC(err)
	}

	input := port.EliminarRegistroInput{
		NombreTabla: req.GetNombreTabla(),
		ID:          req.GetId(),
	}

	err := h.useCase.EliminarRegistro(ctx, input)
	if err != nil {
		return nil, mapDomainErrorToGRPC(err)
	}

	return &dbgsv1.EliminarRegistroResponse{
		Exitoso: true,
		Mensaje: "Registro eliminado correctamente",
	}, nil
}
