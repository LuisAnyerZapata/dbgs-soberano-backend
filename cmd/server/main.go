package main

import (
    "context"
    "fmt"
    "log"
    "net"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"

    // gRPC nativo y utilidades
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"
    "google.golang.org/grpc/reflection"

    // Adaptadores de entrada (Handlers gRPC) e Interceptores
    grpcHandler "DBGS_SOBERANO_BACKEND/internal/adapter/handler/grpc"
    grpcInterceptors "DBGS_SOBERANO_BACKEND/internal/adapter/handler/grpc/interceptors"

    // Adaptadores de salida (Repositorios Postgres)
    "DBGS_SOBERANO_BACKEND/internal/adapter/repository/postgres"

    // Casos de Uso (Aplicación)
    "DBGS_SOBERANO_BACKEND/internal/application/usecase"

    // Configuración global
    "DBGS_SOBERANO_BACKEND/config"

    // Paquetes generados por Protobuf
    pb "DBGS_SOBERANO_BACKEND/api/proto/v1"

    // grpc-gateway: Adaptador que traduce HTTP/JSON a gRPC binario
    "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
)

func main() {
    // ====================================================================
    // 1. INICIALIZACIÓN DE INFRAESTRUCTURA Y CONFIGURACIÓN
    // ====================================================================
    cfg, err := config.LoadConfig(".")
    if err != nil {
        log.Fatalf("Error crítico al cargar la configuración: %v", err)
    }

    db, err := postgres.NewPostgresConnection(cfg.Database)
    if err != nil {
        log.Fatalf("Error crítico al conectar con PostgreSQL: %v", err)
    }
    defer db.Close()

    // ====================================================================
    // 2. INYECCIÓN DE DEPENDENCIAS (Arquitectura Hexagonal)
    // ====================================================================
    // Se instancian los puertos de salida (Repositorios) y se inyectan
    // en los puertos de entrada (Casos de Uso). El dominio permanece aislado.
    
    catalogoRepo := postgres.NewCatalogoPostgresRepository(db)
    auditoriaRepo := postgres.NewAuditoriaPostgresRepository(db)
    datasetRepo := postgres.NewDatasetPostgresRepository(db)
    seguridadRepo := postgres.NewSeguridadPostgresRepository(db)

    catalogoUseCase := usecase.NewCatalogoUseCase(catalogoRepo)
    auditoriaUseCase := usecase.NewAuditoriaUseCase(auditoriaRepo)
    datasetUseCase := usecase.NewDatasetUseCase(datasetRepo)
    seguridadUseCase := usecase.NewSeguridadUseCase(seguridadRepo)
    integracionUseCase := usecase.NewIntegracionUseCase(nil)

    // Controladores que implementan las interfaces gRPC generadas
    catalogoHandler := grpcHandler.NewCatalogoHandler(catalogoUseCase)
    auditoriaHandler := grpcHandler.NewAuditoriaHandler(auditoriaUseCase)
    datasetHandler := grpcHandler.NewDatasetsHandler(datasetUseCase)
    seguridadHandler := grpcHandler.NewSeguridadHandler(seguridadUseCase)
    
    // Interceptores para autorización y validación de integraciones
    integracionInterceptor := grpcInterceptors.NewIntegracionInterceptor(integracionUseCase)
    authInterceptor := grpcInterceptors.NewAuthInterceptor(seguridadUseCase)

    // ====================================================================
    // 3. SERVIDOR gRPC NATIVO (Comunicación interna / Móvil / Escritorio)
    // ====================================================================
    grpcServer := grpc.NewServer(
        grpc.ChainUnaryInterceptor(
            integracionInterceptor.Unary(),
            authInterceptor.Unary(),
        ),
    )

    // Registro de servicios en el servidor gRPC puro
    pb.RegisterCatalogosServiceServer(grpcServer, catalogoHandler)
    pb.RegisterAuditoriaServiceServer(grpcServer, auditoriaHandler)
    pb.RegisterDatasetsServiceServer(grpcServer, datasetHandler)
    pb.RegisterSeguridadServiceServer(grpcServer, seguridadHandler)

    // Habilitar reflexión para herramientas de depuración como grpcurl
    reflection.Register(grpcServer)

    // ====================================================================
    // 4. SERVIDOR HTTP / REST (grpc-gateway para Aplicaciones Web)
    // ====================================================================
    // El Mux de runtime actúa como un proxy inverso en proceso.
    // Traduce peticiones HTTP/JSON entrantes a peticiones gRPC salientes.
    ctx := context.Background()
    gatewayMux := runtime.NewServeMux()

    // Configuración para que el Gateway se conecte al servidor gRPC local.
    // Se usa insecure porque la comunicación es en loopback (localhost).
    endpoint := fmt.Sprintf("localhost:%d", cfg.Server.Port)
    dialOpts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}

    // Registro de los handlers generados por protoc-gen-grpc-gateway.
    // Estos handlers exponen las rutas REST definidas en los archivos .proto
    if err := pb.RegisterSeguridadServiceHandlerFromEndpoint(ctx, gatewayMux, endpoint, dialOpts); err != nil {
        log.Fatalf("Fallo al registrar SeguridadService en el Gateway: %v", err)
    }
    if err := pb.RegisterCatalogosServiceHandlerFromEndpoint(ctx, gatewayMux, endpoint, dialOpts); err != nil {
        log.Fatalf("Fallo al registrar CatalogosService en el Gateway: %v", err)
    }
    if err := pb.RegisterDatasetsServiceHandlerFromEndpoint(ctx, gatewayMux, endpoint, dialOpts); err != nil {
        log.Fatalf("Fallo al registrar DatasetsService en el Gateway: %v", err)
    }
    if err := pb.RegisterAuditoriaServiceHandlerFromEndpoint(ctx, gatewayMux, endpoint, dialOpts); err != nil {
        log.Fatalf("Fallo al registrar AuditoriaService en el Gateway: %v", err)
    }

    // Se envuelve el Mux en un middleware CORS. Esto es estrictamente necesario
    // para que los navegadores web permitan el consumo de esta API desde otros dominios/puertos.
    httpHandler := corsMiddleware(gatewayMux)
    httpServer := &http.Server{
        Addr:    ":8080", // Puerto expuesto para el frontend web
        Handler: httpHandler,
    }

    // ====================================================================
    // 5. EJECUCIÓN CONCURRENTE Y GRACEFUL SHUTDOWN
    // ====================================================================
    // Canal para escuchar señales de interrupción del SO (Ctrl+C, docker stop, etc.)
    stop := make(chan os.Signal, 1)
    signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

    // Iniciar gRPC en una goroutine independiente
    go func() {
        lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Server.Port))
        if err != nil {
            log.Fatalf("Error al abrir el puerto TCP para gRPC: %v", err)
        }
        log.Printf("🚀 Servidor gRPC nativo escuchando en puerto :%d", cfg.Server.Port)
        if err := grpcServer.Serve(lis); err != nil {
            log.Fatalf("Error irrecuperable en el servidor gRPC: %v", err)
        }
    }()

    // Iniciar HTTP REST en otra goroutine independiente
    go func() {
        log.Printf("🌐 Servidor REST (JSON) para Web escuchando en puerto :8080")
        if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatalf("Error irrecuperable en el servidor HTTP: %v", err)
        }
    }()

    // El programa principal se bloquea aquí hasta recibir una señal de parada
    <-stop
    log.Println("🛑 Señal de apagado recibida. Iniciando cierre gradual...")

    // Contexto con timeout para forzar el cierre si los servidores tardan mucho
    shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    // Apagar el servidor HTTP REST primero (para dejar de recibir tráfico web)
    if err := httpServer.Shutdown(shutdownCtx); err != nil {
        log.Printf("⚠️ Error durante el apagado del servidor HTTP: %v", err)
    }

    // Apagar el servidor gRPC (espera que terminen las peticiones en curso)
    grpcServer.GracefulStop()

    log.Println("✅ Todos los servidores han sido detenidos de forma segura.")
}

// corsMiddleware inyecta las cabeceras necesarias para el Intercambio de Recursos de Origen Cruzado (CORS).
// Sin esto, los navegadores bloquearían las peticiones fetch/axios desde un frontend en otro puerto.
func corsMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        origin := r.Header.Get("Origin")
        
        // Permitir cualquier origen en desarrollo. 
        // NOTA DE SEGURIDAD: En producción, reemplazar "*" por el dominio específico del frontend.
        if origin != "" {
            w.Header().Set("Access-Control-Allow-Origin", origin)
            w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
            // 'Authorization' es vital aquí: es por donde viaja el JWT desde el navegador
            w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
            w.Header().Set("Access-Control-Expose-Headers", "Content-Length")
        }

        // Las peticiones "Preflight" envían método OPTIONS. Debemos responder 204 y cortar la cadena.
        if r.Method == http.MethodOptions {
            w.WriteHeader(http.StatusNoContent)
            return
        }

        // Si no es preflight, continuar con el handler normal (grpc-gateway)
        next.ServeHTTP(w, r)
    })
}