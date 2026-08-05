package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	grpcHandler "DBGS_SOBERANO_BACKEND/internal/adapter/handler/grpc"
	"DBGS_SOBERANO_BACKEND/internal/adapter/repository/postgres"
	"DBGS_SOBERANO_BACKEND/internal/application/usecase"
	"DBGS_SOBERANO_BACKEND/config"
	pb "DBGS_SOBERANO_BACKEND/api/proto/v1"
)

func main() {
	// Carga la configuración del entorno
	cfg, err := config.LoadConfig(".")
	if err != nil {
		log.Fatalf("Error al cargar la configuración: %v", err)
	}

	// Conexión a base de datos PostgreSQL utilizando la configuración deserializada
	db, err := postgres.NewPostgresConnection(cfg.Database)
	if err != nil {
		log.Fatalf("Error al conectar con PostgreSQL: %v", err)
	}
	defer db.Close()

	// Repositorios de almacenamiento
	catalogoRepo := postgres.NewCatalogoPostgresRepository(db)
	auditoriaRepo := postgres.NewAuditoriaPostgresRepository(db)
	datasetRepo := postgres.NewDatasetPostgresRepository(db)
	seguridadRepo := postgres.NewSeguridadPostgresRepository(db)

	// Capa de aplicación (Casos de Uso)
	catalogoUseCase := usecase.NewCatalogoUseCase(catalogoRepo)
	auditoriaUseCase := usecase.NewAuditoriaUseCase(auditoriaRepo)
	datasetUseCase := usecase.NewDatasetUseCase(datasetRepo)
	seguridadUseCase := usecase.NewSeguridadUseCase(seguridadRepo)

	// Controladores de transporte gRPC
	catalogoHandler := grpcHandler.NewCatalogoHandler(catalogoUseCase)
	auditoriaHandler := grpcHandler.NewAuditoriaHandler(auditoriaUseCase)
	datasetHandler := grpcHandler.NewDatasetsHandler(datasetUseCase)
	seguridadHandler := grpcHandler.NewSeguridadHandler(seguridadUseCase)

	// Servidor gRPC
	grpcServer := grpc.NewServer()

	// Registro de los servicios gRPC
	pb.RegisterCatalogosServiceServer(grpcServer, catalogoHandler)
	pb.RegisterAuditoriaServiceServer(grpcServer, auditoriaHandler)
	pb.RegisterDatasetsServiceServer(grpcServer, datasetHandler)
	pb.RegisterSeguridadServiceServer(grpcServer, seguridadHandler)

	// Habilitar gRPC Reflection
	reflection.Register(grpcServer)

	// Inicio del listener TCP
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Server.Port))
	if err != nil {
		log.Fatalf("Error al escuchar en el puerto %d: %v", cfg.Server.Port, err)
	}

	// Graceful Shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Printf("Servidor gRPC iniciado y escuchando en el puerto %d", cfg.Server.Port)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("Error al servir gRPC: %v", err)
		}
	}()

	<-stop
	log.Println("Iniciando apagado gradual del servidor...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	stopped := make(chan struct{})
	go func() {
		grpcServer.GracefulStop()
		close(stopped)
	}()

	select {
	case <-stopped:
		log.Println("Servidor apagado correctamente.")
	case <-ctx.Done():
		log.Println("Tiempo de espera agotado. Forzando parada del servidor...")
		grpcServer.Stop()
	}
}