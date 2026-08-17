package usecase

import (
    "context"
    "crypto/rand"
    "crypto/sha256"
    "encoding/hex"
    "errors"
    "fmt"
    "time"

    "DBGS_SOBERANO_BACKEND/internal/application/port"
    "DBGS_SOBERANO_BACKEND/internal/domain"
    "DBGS_SOBERANO_BACKEND/internal/domain/entity"
    "DBGS_SOBERANO_BACKEND/internal/domain/repository"
)

type integracionUseCase struct {
    integracionRepo repository.IntegracionRepository
}

func NewIntegracionUseCase(repo repository.IntegracionRepository) port.IntegracionPort {
    return &integracionUseCase{integracionRepo: repo}
}

// RegistrarCliente crea un nuevo sistema cliente y le asigna una API Key segura.
// Retorna la entidad con el TokenPlano para que la UI lo muestre AL USUARIO UNA ÚNICA VEZ.
func (u *integracionUseCase) RegistrarCliente(ctx context.Context, input port.RegistrarClienteInput) (*entity.ClienteIntegracion, error) {
    if input.Nombre == "" || input.Tipo == "" {
        return nil, domain.ErrDatosInvalidos
    }
    if input.VersionContrato == "" {
        input.VersionContrato = "v1"
    }

    // 1. Generar una API Key criptográficamente segura (texto plano)
    tokenPlano, err := generarTokenSeguro()
    if err != nil {
        return nil, errors.New("error crítico al generar token seguro")
    }

    // 2. Hashear la API Key para almacenarla en la base de datos
    hash := sha256.Sum256([]byte(tokenPlano))
    tokenHash := hex.EncodeToString(hash[:])

    cliente := &entity.ClienteIntegracion{
        Nombre:          input.Nombre,
        Tipo:            input.Tipo,
        VersionContrato: input.VersionContrato,
        Scopes:          input.Scopes,
        TokenHash:       tokenHash, // Se guarda el hash
        TokenPlano:      tokenPlano, // Se guarda temporalmente para retornarlo a la UI
        Estado:          true,
        CreatedAt:       time.Now(),
        UpdatedAt:       time.Now(),
    }

    if err := u.integracionRepo.GuardarCliente(ctx, cliente); err != nil {
        return nil, err
    }
    
    return cliente, nil
}

// ValidarAcceso es el método que consume el interceptor gRPC.
// Recibe el token en texto plano, lo hashea, y busca el hash en la BD en un solo paso seguro.
func (u *integracionUseCase) ValidarAcceso(ctx context.Context, input port.ValidarAccesoIntegracionInput) (*entity.ClienteIntegracion, error) {
    if input.ClienteID == "" || input.Token == "" {
        return nil, domain.ErrDatosInvalidos
    }

    // BUENA PRÁCTICA: Hasheamos el token entrante y buscamos directamente por hash.
    // NUNCA buscamos por ID y luego comparamos textos planos.
    hash := sha256.Sum256([]byte(input.Token))
    hashStr := hex.EncodeToString(hash[:])

    cliente, err := u.integracionRepo.ValidarCredenciales(ctx, hashStr)
    if err != nil {
        return nil, domain.ErrAccesoNoAutorizado
    }

    // Validación opcional de versión de contrato
    if input.VersionContrato != "" && cliente.VersionContrato != input.VersionContrato {
        return nil, domain.ErrAccesoNoAutorizado
    }
    
    return cliente, nil
}

// RegistrarSolicitud guarda un log de traza de la petición externa
func (u *integracionUseCase) RegistrarSolicitud(ctx context.Context, input port.RegistrarSolicitudInput) error {
    if input.ClienteID == "" || input.Metodo == "" || input.Recurso == "" {
        return domain.ErrDatosInvalidos
    }
    solicitud := &entity.SolicitudIntegracion{
        ClienteID: input.ClienteID,
        Metodo:    input.Metodo,
        Recurso:   input.Recurso,
        Estado:    input.Estado,
        Fecha:     time.Now(),
        Detalles:  input.Detalles,
    }
    return u.integracionRepo.GuardarSolicitud(ctx, solicitud)
}

// generarTokenSeguro crea una string aleatoria criptográficamente segura para usar como API Key
func generarTokenSeguro() (string, error) {
    b := make([]byte, 32) // 256 bits de entropía
    _, err := rand.Read(b)
    if err != nil {
        return "", err
    }
    return fmt.Sprintf("dbgs_%s", hex.EncodeToString(b)), nil
}