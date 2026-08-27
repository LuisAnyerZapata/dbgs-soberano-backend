package grpc

import (
	"context"
	"log"

	dbgsv1 "DBGS_SOBERANO_BACKEND/api/proto/v1"
	"DBGS_SOBERANO_BACKEND/internal/application/port"
	"DBGS_SOBERANO_BACKEND/internal/domain/entity"

	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type ConexionesHandler struct {
	dbgsv1.UnimplementedConexionesServiceServer
	useCase port.ConexionPort
}

func NewConexionesHandler(uc port.ConexionPort) *ConexionesHandler {
	return &ConexionesHandler{useCase: uc}
}

func motorFromString(s string) dbgsv1.MotorBaseDatos {
	switch s {
	case "mysql":
		return dbgsv1.MotorBaseDatos_MYSQL
	case "postgres", "postgresql", "pgsql":
		return dbgsv1.MotorBaseDatos_POSTGRESQL
	default:
		return dbgsv1.MotorBaseDatos_POSTGRESQL
	}
}

func toConexionProto(c entity.Conexion) *dbgsv1.Conexion {
	return &dbgsv1.Conexion{
		Id:        c.ID,
		Name:      c.Name,
		Engine:    motorFromString(c.Engine),
		Host:      c.Host,
		Port:      int32(c.Port),
		User:      c.User,
		Database:  c.Database,
		SslMode:   c.SSLMode,
		ReadOnly:  c.ReadOnly,
		CreatedBy: c.CreatedBy,
		CreatedAt: timestamppb.New(c.CreatedAt),
	}
}

func (h *ConexionesHandler) ProbarConexion(ctx context.Context, req *dbgsv1.ProbarConexionRequest) (*dbgsv1.ProbarConexionResponse, error) {
	c := &entity.Conexion{
		Name:     req.GetName(),
		Engine:   req.GetEngine(),
		Host:     req.GetHost(),
		Port:     int(req.GetPort()),
		User:     req.GetUser(),
		Database: req.GetDatabase(),
		SSLMode:  req.GetSslMode(),
	}
	res, err := h.useCase.ProbarConexion(ctx, c, req.GetPassword())
	if err != nil {
		return nil, mapDomainErrorToGRPC(err)
	}
	return &dbgsv1.ProbarConexionResponse{
		Ok:        res.OK,
		Message:   res.Message,
		LatencyMs: res.LatencyMS,
	}, nil
}

func (h *ConexionesHandler) CrearConexion(ctx context.Context, req *dbgsv1.CrearConexionRequest) (*dbgsv1.ConexionResponse, error) {
	c := &entity.Conexion{
		Name:     req.GetName(),
		Engine:   req.GetEngine(),
		Host:     req.GetHost(),
		Port:     int(req.GetPort()),
		User:     req.GetUser(),
		Database: req.GetDatabase(),
		SSLMode:  req.GetSslMode(),
		ReadOnly: req.GetReadOnly(),
	}
	created, err := h.useCase.CrearConexion(ctx, c, req.GetPassword())
	if err != nil {
		return nil, mapDomainErrorToGRPC(err)
	}
	return &dbgsv1.ConexionResponse{Connection: toConexionProto(*created)}, nil
}

func (h *ConexionesHandler) ListarConexiones(ctx context.Context, req *dbgsv1.ListarConexionesRequest) (*dbgsv1.ListarConexionesResponse, error) {
	list, total, err := h.useCase.ListarConexiones(ctx, int(req.GetLimit()), int(req.GetOffset()))
	if err != nil {
		return nil, mapDomainErrorToGRPC(err)
	}
	var protoConns []*dbgsv1.Conexion
	for _, c := range list {
		protoConns = append(protoConns, toConexionProto(c))
	}
	return &dbgsv1.ListarConexionesResponse{
		Connections:  protoConns,
		TotalRecords: total,
	}, nil
}

func (h *ConexionesHandler) ObtenerConexion(ctx context.Context, req *dbgsv1.ObtenerConexionRequest) (*dbgsv1.ConexionResponse, error) {
	c, err := h.useCase.ObtenerConexion(ctx, req.GetId())
	if err != nil {
		return nil, mapDomainErrorToGRPC(err)
	}
	return &dbgsv1.ConexionResponse{Connection: toConexionProto(*c)}, nil
}

func (h *ConexionesHandler) ActualizarConexion(ctx context.Context, req *dbgsv1.ActualizarConexionRequest) (*dbgsv1.ConexionResponse, error) {
	c := &entity.Conexion{
		ID:       req.GetId(),
		Name:     req.GetName(),
		Engine:   req.GetEngine(),
		Host:     req.GetHost(),
		Port:     int(req.GetPort()),
		User:     req.GetUser(),
		Database: req.GetDatabase(),
		SSLMode:  req.GetSslMode(),
		ReadOnly: req.GetReadOnly(),
	}
	updated, err := h.useCase.ActualizarConexion(ctx, c, req.GetPassword())
	if err != nil {
		return nil, mapDomainErrorToGRPC(err)
	}
	return &dbgsv1.ConexionResponse{Connection: toConexionProto(*updated)}, nil
}

func (h *ConexionesHandler) EliminarConexion(ctx context.Context, req *dbgsv1.EliminarConexionRequest) (*dbgsv1.EliminarConexionResponse, error) {
	if err := h.useCase.EliminarConexion(ctx, req.GetId()); err != nil {
		return nil, mapDomainErrorToGRPC(err)
	}
	return &dbgsv1.EliminarConexionResponse{Id: req.GetId(), Accion: "eliminada"}, nil
}

func (h *ConexionesHandler) ListarEsquemas(ctx context.Context, req *dbgsv1.ListarEsquemasRequest) (*dbgsv1.ListarEsquemasResponse, error) {
	schemas, err := h.useCase.ListarEsquemas(ctx, req.GetId())
	if err != nil {
		return nil, mapDomainErrorToGRPC(err)
	}
	return &dbgsv1.ListarEsquemasResponse{Schemas: schemas}, nil
}

func (h *ConexionesHandler) ListarTablas(ctx context.Context, req *dbgsv1.ListarTablasRequest) (*dbgsv1.ListarTablasResponse, error) {
	tables, err := h.useCase.ListarTablas(ctx, req.GetId(), req.GetSchema())
	if err != nil {
		return nil, mapDomainErrorToGRPC(err)
	}
	return &dbgsv1.ListarTablasResponse{Tables: tables}, nil
}

func (h *ConexionesHandler) ExplorarDatos(ctx context.Context, req *dbgsv1.ExplorarDatosRequest) (*dbgsv1.ExplorarDatosResponse, error) {
	res, err := h.useCase.ExplorarDatos(ctx, req.GetId(), req.GetSchema(), req.GetTable(), int(req.GetLimit()), int(req.GetOffset()))
	if err != nil {
		log.Printf("ExplorarDatos error (id=%s schema=%q table=%q): %v", req.GetId(), req.GetSchema(), req.GetTable(), err)
		return nil, mapDomainErrorToGRPC(err)
	}
	cols := make([]*dbgsv1.ColumnaExterna, 0, len(res.Columns))
	for _, c := range res.Columns {
		cols = append(cols, &dbgsv1.ColumnaExterna{
			Name:       c.Name,
			DataType:   c.DataType,
			Nullable:   c.Nullable,
			PrimaryKey: c.PrimaryKey,
		})
	}
	rows := make([]*structpb.Struct, 0, len(res.Rows))
	for _, r := range res.Rows {
		st, err := structpb.NewStruct(r)
		if err == nil {
			rows = append(rows, st)
		}
	}
	return &dbgsv1.ExplorarDatosResponse{
		Columns:      cols,
		Rows:         rows,
		TotalRecords: res.Total,
	}, nil
}
