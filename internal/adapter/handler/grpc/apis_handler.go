package grpc

import (
	"context"

	dbgsv1 "DBGS_SOBERANO_BACKEND/api/proto/v1"
	"DBGS_SOBERANO_BACKEND/internal/application/port"
	"DBGS_SOBERANO_BACKEND/internal/domain"
	"DBGS_SOBERANO_BACKEND/internal/domain/entity"

	"google.golang.org/protobuf/types/known/timestamppb"
)

type ApisHandler struct {
	dbgsv1.UnimplementedApisServiceServer
	useCase port.ApiPublicadaPort
}

func NewApisHandler(uc port.ApiPublicadaPort) *ApisHandler {
	return &ApisHandler{useCase: uc}
}

func toApiProto(a entity.ApiPublicada) *dbgsv1.ApiPublicada {
	return &dbgsv1.ApiPublicada{
		Id:             a.ID,
		Name:           a.Name,
		Description:    a.Description,
		Slug:           a.Slug,
		ConnectionId:   a.ConnectionID,
		ConnectionName: a.ConnectionName,
		Schema:         a.Schema,
		Table:          a.Table,
		MaxRows:        int32(a.MaxRows),
		Active:         a.Active,
		ApiKey:         a.APIKey,
		Endpoint:       a.Endpoint,
		CreatedBy:      a.CreatedBy,
		CreatedAt:      timestamppb.New(a.CreatedAt),
	}
}

func (h *ApisHandler) CrearApi(ctx context.Context, req *dbgsv1.CrearApiRequest) (*dbgsv1.ApiResponse, error) {
	if err := domain.ValidateRequired("name", req.GetName()); err != nil {
		return nil, mapDomainErrorToGRPC(err)
	}
	a := &entity.ApiPublicada{
		Name:         req.GetName(),
		Description:  req.GetDescription(),
		Slug:         req.GetSlug(),
		ConnectionID: req.GetConnectionId(),
		Schema:       req.GetSchema(),
		Table:        req.GetTable(),
		MaxRows:      int(req.GetMaxRows()),
	}
	created, err := h.useCase.CrearApi(ctx, a)
	if err != nil {
		return nil, mapDomainErrorToGRPC(err)
	}
	return &dbgsv1.ApiResponse{Api: toApiProto(*created)}, nil
}

func (h *ApisHandler) ListarApis(ctx context.Context, req *dbgsv1.ListarApisRequest) (*dbgsv1.ListarApisResponse, error) {
	list, total, err := h.useCase.ListarApis(ctx, int(req.GetLimit()), int(req.GetOffset()))
	if err != nil {
		return nil, mapDomainErrorToGRPC(err)
	}
	var protoApis []*dbgsv1.ApiPublicada
	for _, a := range list {
		protoApis = append(protoApis, toApiProto(a))
	}
	return &dbgsv1.ListarApisResponse{
		Apis:         protoApis,
		TotalRecords: total,
	}, nil
}

func (h *ApisHandler) ObtenerApi(ctx context.Context, req *dbgsv1.ObtenerApiRequest) (*dbgsv1.ApiResponse, error) {
	a, err := h.useCase.ObtenerApi(ctx, req.GetId())
	if err != nil {
		return nil, mapDomainErrorToGRPC(err)
	}
	return &dbgsv1.ApiResponse{Api: toApiProto(*a)}, nil
}

func (h *ApisHandler) CambiarEstadoApi(ctx context.Context, req *dbgsv1.CambiarEstadoApiRequest) (*dbgsv1.ApiResponse, error) {
	a, err := h.useCase.CambiarEstadoApi(ctx, req.GetId(), req.GetActive())
	if err != nil {
		return nil, mapDomainErrorToGRPC(err)
	}
	return &dbgsv1.ApiResponse{Api: toApiProto(*a)}, nil
}

func (h *ApisHandler) EliminarApi(ctx context.Context, req *dbgsv1.EliminarApiRequest) (*dbgsv1.EliminarApiResponse, error) {
	if err := h.useCase.EliminarApi(ctx, req.GetId()); err != nil {
		return nil, mapDomainErrorToGRPC(err)
	}
	return &dbgsv1.EliminarApiResponse{Id: req.GetId(), Accion: "eliminada"}, nil
}
