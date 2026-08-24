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
    "google.golang.org/grpc/metadata"

    // Adaptadores de entrada (Handlers gRPC) e Interceptores
    grpcHandler "DBGS_SOBERANO_BACKEND/internal/adapter/handler/grpc"
    grpcInterceptors "DBGS_SOBERANO_BACKEND/internal/adapter/handler/grpc/interceptors"

    // Adaptadores de salida (Repositorios Postgres)
    "DBGS_SOBERANO_BACKEND/internal/adapter/repository/postgres"

    // Casos de Uso (Aplicación)
    "DBGS_SOBERANO_BACKEND/internal/application/usecase"

    // Adaptador de ejecución de scripts (pg_dump/pg_restore)
    "DBGS_SOBERANO_BACKEND/internal/adapter/executor"

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
    integracionRepo := postgres.NewIntegracionPostgresRepository(db)
    coleccionRepo := postgres.NewColeccionDinamicaPostgresRepository(db)
    datosDinamicosRepo := postgres.NewDatosDinamicosPostgresRepository(db)

    catalogoUseCase := usecase.NewCatalogoUseCase(catalogoRepo)
    auditoriaUseCase := usecase.NewAuditoriaUseCase(auditoriaRepo)
    datasetUseCase := usecase.NewDatasetUseCase(datasetRepo)
    datosDinamicosUseCase := usecase.NewDatosDinamicosUseCase(coleccionRepo, datosDinamicosRepo)
    
    // Se inyectan los parámetros de seguridad desde el config.json o variables de entorno
    seguridadUseCase := usecase.NewSeguridadUseCase(seguridadRepo, cfg.Security.JWTSecret, cfg.Security.TokenTTLMinutes)

    integracionUseCase := usecase.NewIntegracionUseCase(integracionRepo)
    coleccionUseCase := usecase.NewColeccionUseCase(coleccionRepo)

    // Dominio de Respaldos: el motor ejecuta los scripts bash de db/backup con las
    // credenciales centralizadas; el caso de uso orquesta pg_dump/pg_restore asíncrono.
    respaldoRepo := postgres.NewRespaldoPostgresRepository(db)
    motorRespaldos := executor.NewMotorScripts(
        cfg.Backup.ScriptsDir,
        cfg.Database.Host, cfg.Database.Port,
        cfg.Database.User, cfg.Database.Password, cfg.Database.Name,
    )
    respaldoUseCase := usecase.NewRespaldoUseCase(respaldoRepo, seguridadRepo, motorRespaldos, usecase.RespaldoConfig{
        DumpsDir:         cfg.Backup.DumpsDir,
        TimeoutEjecucion: time.Duration(cfg.Backup.TimeoutMinutos) * time.Minute,
    })

    // Nota: En producción, estos valores se inyectan con -ldflags en el Makefile
    sistemaUseCase := usecase.NewSistemaUseCase(db, usecase.BuildInfo{
        Version:    "v1.0.0-dev",
        GitCommit:  "local-dev",
        BuildDate:  time.Now().Format(time.RFC3339),
        GoVersion:  "1.26",
        Environment: "development",
    })

    // Controladores que implementan las interfaces gRPC generadas
    catalogoHandler := grpcHandler.NewCatalogoHandler(catalogoUseCase)
    auditoriaHandler := grpcHandler.NewAuditoriaHandler(auditoriaUseCase)
    datasetHandler := grpcHandler.NewDatasetsHandler(datasetUseCase)
    seguridadHandler := grpcHandler.NewSeguridadHandler(seguridadUseCase)
    sistemaHandler := grpcHandler.NewSistemaHandler(sistemaUseCase)
    coleccionHandler := grpcHandler.NewColeccionesHandler(coleccionUseCase)
    datosDinamicosHandler := grpcHandler.NewDatosDinamicosHandler(datosDinamicosUseCase)
    respaldoHandler := grpcHandler.NewRespaldoHandler(respaldoUseCase)
    
    // Interceptores para autorización y validación de integraciones
    unifiedAuthInterceptor := grpcInterceptors.NewAuthInterceptor(seguridadUseCase, integracionUseCase)

    // ====================================================================
    // 3. SERVIDOR gRPC NATIVO (Comunicación interna / Móvil / Escritorio)
    // ====================================================================
    grpcServer := grpc.NewServer(
        grpc.ChainUnaryInterceptor(
            unifiedAuthInterceptor.Unary(), // Único interceptor en la cadena
        ),
    )

    // Registro de servicios en el servidor gRPC puro
    pb.RegisterCatalogosServiceServer(grpcServer, catalogoHandler)
    pb.RegisterAuditoriaServiceServer(grpcServer, auditoriaHandler)
    pb.RegisterDatasetsServiceServer(grpcServer, datasetHandler)
    pb.RegisterSeguridadServiceServer(grpcServer, seguridadHandler)
    pb.RegisterSistemaServiceServer(grpcServer, sistemaHandler)
    pb.RegisterColeccionesServiceServer(grpcServer, coleccionHandler)
    pb.RegisterDatosDinamicosServiceServer(grpcServer, datosDinamicosHandler)
    pb.RegisterRespaldoServiceServer(grpcServer, respaldoHandler)

    // Habilitar reflexión para herramientas de depuración como grpcurl
    reflection.Register(grpcServer)

    // ====================================================================
    // 4. SERVIDOR HTTP / REST (grpc-gateway para Aplicaciones Web)
    // ====================================================================
    // El Mux de runtime actúa como un proxy inverso en proceso.
    // Traduce peticiones HTTP/JSON entrantes a peticiones gRPC salientes.
    ctx := context.Background()
    
    // Configuración crítica para gRPC-Gateway: 
    // Extrae cabeceras HTTP personalizadas y las inyecta en los metadatos gRPC
    // para que los interceptores puedan leerlas.
    gatewayMux := runtime.NewServeMux(
        runtime.WithMetadata(func(ctx context.Context, req *http.Request) metadata.MD {
            md := metadata.New(map[string]string{})
            
            // Extraer API Keys de integración (Máquina a Máquina)
            if val := req.Header.Get("x-api-token"); val != "" {
                md.Append("x-api-token", val)
            }
            if val := req.Header.Get("x-client-id"); val != "" {
                md.Append("x-client-id", val)
            }
            
            return md
        }),
    )

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
    if err := pb.RegisterSistemaServiceHandlerFromEndpoint(ctx, gatewayMux, endpoint, dialOpts); err != nil {
        log.Fatalf("Fallo al registrar SistemaService en el Gateway: %v", err)
    }
    if err := pb.RegisterColeccionesServiceHandlerFromEndpoint(ctx, gatewayMux, endpoint, dialOpts); err != nil {
        log.Fatalf("Fallo al registrar ColeccionesService en el Gateway: %v", err)
    }
    if err := pb.RegisterDatosDinamicosServiceHandlerFromEndpoint(ctx, gatewayMux, endpoint, dialOpts); err != nil {
        log.Fatalf("Fallo al registrar DatosDinamicosService en el Gateway: %v", err)
    }
    if err := pb.RegisterRespaldoServiceHandlerFromEndpoint(ctx, gatewayMux, endpoint, dialOpts); err != nil {
        log.Fatalf("Fallo al registrar RespaldoService en el Gateway: %v", err)
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