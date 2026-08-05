package grpc

import (
	"context"

	dbgsv1 "DBGS_SOBERANO_BACKEND/api/proto/v1"
	"DBGS_SOBERANO_BACKEND/internal/application/port"
)

type DatasetsHandler struct {
	dbgsv1.UnimplementedDatasetsServiceServer
	useCase port.DatasetPort
}

func NewDatasetsHandler(uc port.DatasetPort) *DatasetsHandler {
	return &DatasetsHandler{
		useCase: uc,
	}
}

func (h *DatasetsHandler) ListarDatasets(ctx context.Context, req *dbgsv1.ListarDatasetsRequest) (*dbgsv1.ListarDatasetsResponse, error) {
	input := port.ObtenerDatasetsInput{
		Clasificacion: req.GetFiltroClasificacion().String(),
		Limite:        int(req.GetLimit()),
		Offset:        int(req.GetOffset()),
	}

	output, err := h.useCase.ListarDatasets(ctx, input)
	if err != nil {
		return nil, mapDomainErrorToGRPC(err)
	}

	var pbDatasets []*dbgsv1.Dataset
	for _, ds := range output.Datasets {
		pbDatasets = append(pbDatasets, &dbgsv1.Dataset{
			Id:           ds.ID,
			FuenteDatoId: ds.FuenteDatoID,
			Nombre:       ds.Nombre,
			Proposito:    ds.Proposito,
			Estado:       ds.Estado,
		})
	}

	return &dbgsv1.ListarDatasetsResponse{
		Datasets:     pbDatasets,
		TotalRecords: int32(output.Total),
	}, nil
}