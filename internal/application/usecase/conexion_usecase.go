package usecase

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"time"

	"DBGS_SOBERANO_BACKEND/internal/application/port"
	"DBGS_SOBERANO_BACKEND/internal/domain"
	"DBGS_SOBERANO_BACKEND/internal/domain/entity"
	"DBGS_SOBERANO_BACKEND/internal/domain/repository"

	"github.com/google/uuid"
)

// conexionCifrador cifra/descifra las contraseñas de las conexiones externas
// usando AES-GCM con una clave derivada del secreto del servidor.
type conexionCifrador struct {
	key [32]byte
}

func newConexionCifrador(secret string) *conexionCifrador {
	return &conexionCifrador{key: sha256.Sum256([]byte(secret))}
}

func (c *conexionCifrador) cifrar(plaintext string) (string, error) {
	if plaintext == "" {
		return "", nil
	}
	block, err := aes.NewCipher(c.key[:])
	if err != nil {
		return "", domain.InternalError.ConWrap(err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", domain.InternalError.ConWrap(err)
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", domain.InternalError.ConWrap(err)
	}
	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func (c *conexionCifrador) descifrar(encoded string) (string, error) {
	if encoded == "" {
		return "", nil
	}
	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", domain.InternalError.ConWrap(err)
	}
	block, err := aes.NewCipher(c.key[:])
	if err != nil {
		return "", domain.InternalError.ConWrap(err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", domain.InternalError.ConWrap(err)
	}
	if len(data) < gcm.NonceSize() {
		return "", domain.InternalError
	}
	plaintext, err := gcm.Open(nil, data[:gcm.NonceSize()], data[gcm.NonceSize():], nil)
	if err != nil {
		return "", domain.InternalError.ConWrap(err)
	}
	return string(plaintext), nil
}

type conexionUseCase struct {
	conexionRepo repository.ConexionRepository
	externa      port.ConexionExternaPort
	apiRepo      repository.ApiPublicadaRepository
	cifrador     *conexionCifrador
}

// NewConexionUseCase instancia el caso de uso de conexiones a bases externas.
func NewConexionUseCase(
	repo repository.ConexionRepository,
	externa port.ConexionExternaPort,
	apiRepo repository.ApiPublicadaRepository,
	secret string,
) port.ConexionPort {
	return &conexionUseCase{
		conexionRepo: repo,
		externa:      externa,
		apiRepo:      apiRepo,
		cifrador:     newConexionCifrador(secret),
	}
}

// NewApiPublicadaUseCase instancia el caso de uso de APIs públicas reutilizando
// el repositorio de conexiones (para resolver el nombre de la conexión padre).
func NewApiPublicadaUseCase(
	repo repository.ApiPublicadaRepository,
	conexionRepo repository.ConexionRepository,
	secret string,
) port.ApiPublicadaPort {
	return &conexionUseCase{
		apiRepo:      repo,
		conexionRepo: conexionRepo,
		cifrador:     newConexionCifrador(secret),
	}
}

func (u *conexionUseCase) credenciales(c *entity.Conexion, password string) (port.ConexionCredenciales, error) {
	plain, err := u.cifrador.descifrar(c.PasswordHash)
	if err != nil {
		return port.ConexionCredenciales{}, err
	}
	if password != "" {
		plain = password
	}
	return port.ConexionCredenciales{
		Engine:   c.Engine,
		Host:     c.Host,
		Port:     c.Port,
		User:     c.User,
		Password: plain,
		Database: c.Database,
		SSLMode:  c.SSLMode,
	}, nil
}

func (u *conexionUseCase) ProbarConexion(ctx context.Context, c *entity.Conexion, password string) (*port.PruebaConexionResult, error) {
	if c != nil && c.EsValido() == nil {
		creds, err := u.credenciales(c, password)
		if err != nil {
			return nil, err
		}
		return u.externa.Probar(ctx, creds)
	}
	// Prueba sin persistir: usar directamente los datos recibidos
	creds := port.ConexionCredenciales{
		Engine:   c.Engine,
		Host:     c.Host,
		Port:     c.Port,
		User:     c.User,
		Password: password,
		Database: c.Database,
		SSLMode:  c.SSLMode,
	}
	return u.externa.Probar(ctx, creds)
}

func (u *conexionUseCase) CrearConexion(ctx context.Context, c *entity.Conexion, password string) (*entity.Conexion, error) {
	if err := c.EsValido(); err != nil {
		return nil, err
	}
	cifrado, err := u.cifrador.cifrar(password)
	if err != nil {
		return nil, err
	}
	c.ID = uuid.New().String()
	c.PasswordHash = cifrado
	c.CreatedAt = time.Now()
	if usuarioCtx, ok := ctx.Value(domain.CtxKeyUsuario).(*entity.Usuario); ok {
		c.CreatedBy = usuarioCtx.Username
	} else {
		c.CreatedBy = "system"
	}
	if err := u.conexionRepo.Guardar(ctx, c); err != nil {
		return nil, err
	}
	return c, nil
}

func (u *conexionUseCase) ListarConexiones(ctx context.Context, limite, offset int) ([]entity.Conexion, int64, error) {
	if limite <= 0 {
		limite = 50
	}
	if offset < 0 {
		offset = 0
	}
	return u.conexionRepo.Listar(ctx, limite, offset)
}

func (u *conexionUseCase) ObtenerConexion(ctx context.Context, id string) (*entity.Conexion, error) {
	if id == "" {
		return nil, domain.InvalidArgument
	}
	return u.conexionRepo.ObtenerPorID(ctx, id)
}

func (u *conexionUseCase) ActualizarConexion(ctx context.Context, c *entity.Conexion, password string) (*entity.Conexion, error) {
	if c.ID == "" {
		return nil, domain.InvalidArgument
	}
	existente, err := u.conexionRepo.ObtenerPorID(ctx, c.ID)
	if err != nil {
		return nil, err
	}
	if c.Name == "" {
		c.Name = existente.Name
	}
	if c.Engine == "" {
		c.Engine = existente.Engine
	}
	if c.Host == "" {
		c.Host = existente.Host
	}
	if c.Port <= 0 {
		c.Port = existente.Port
	}
	if c.User == "" {
		c.User = existente.User
	}
	if c.Database == "" {
		c.Database = existente.Database
	}
	if c.SSLMode == "" {
		c.SSLMode = existente.SSLMode
	}
	// Si se envía una nueva contraseña, se recifra; si no, se conserva el hash
	cifrado := existente.PasswordHash
	if password != "" {
		cifrado, err = u.cifrador.cifrar(password)
		if err != nil {
			return nil, err
		}
	}
	c.PasswordHash = cifrado
	c.CreatedAt = existente.CreatedAt
	c.CreatedBy = existente.CreatedBy
	return u.conexionRepo.Actualizar(ctx, c)
}

func (u *conexionUseCase) EliminarConexion(ctx context.Context, id string) error {
	if id == "" {
		return domain.InvalidArgument
	}
	return u.conexionRepo.Eliminar(ctx, id)
}

func (u *conexionUseCase) ListarEsquemas(ctx context.Context, id string) ([]string, error) {
	c, err := u.conexionRepo.ObtenerPorID(ctx, id)
	if err != nil {
		return nil, err
	}
	creds, err := u.credenciales(c, "")
	if err != nil {
		return nil, err
	}
	return u.externa.ListarEsquemas(ctx, creds)
}

func (u *conexionUseCase) ListarTablas(ctx context.Context, id, schema string) ([]string, error) {
	c, err := u.conexionRepo.ObtenerPorID(ctx, id)
	if err != nil {
		return nil, err
	}
	creds, err := u.credenciales(c, "")
	if err != nil {
		return nil, err
	}
	return u.externa.ListarTablas(ctx, creds, schema)
}

func (u *conexionUseCase) ExplorarDatos(ctx context.Context, id, schema, tabla string, limite, offset int) (*port.TablaExterna, error) {
	if tabla == "" {
		return nil, domain.RequiredError("table")
	}
	c, err := u.conexionRepo.ObtenerPorID(ctx, id)
	if err != nil {
		return nil, err
	}
	if c.ReadOnly && false { // la exploración siempre es de solo lectura
	}
	creds, err := u.credenciales(c, "")
	if err != nil {
		return nil, err
	}
	return u.externa.ExplorarDatos(ctx, creds, schema, tabla, limite, offset)
}

// ApiPublicada

func (u *conexionUseCase) CrearApi(ctx context.Context, a *entity.ApiPublicada) (*entity.ApiPublicada, error) {
	if err := a.EsValido(); err != nil {
		return nil, err
	}
	if enUso, err := u.apiRepo.SlugEnUso(ctx, a.Slug, ""); err != nil {
		return nil, err
	} else if enUso {
		return nil, domain.AlreadyExists.ConField("slug")
	}
	conn, err := u.conexionRepo.ObtenerPorID(ctx, a.ConnectionID)
	if err != nil {
		return nil, err
	}
	a.ConnectionName = conn.Name
	a.ID = uuid.New().String()
	if a.MaxRows <= 0 {
		a.MaxRows = 500
	}
	a.APIKey = "dg_live_" + uuid.NewString()[:24]
	a.Endpoint = fmt.Sprintf("/api/v1/public/%s", a.Slug)
	a.Active = true
	a.CreatedAt = time.Now()
	if usuarioCtx, ok := ctx.Value(domain.CtxKeyUsuario).(*entity.Usuario); ok {
		a.CreatedBy = usuarioCtx.Username
	} else {
		a.CreatedBy = "system"
	}
	if err := u.apiRepo.Guardar(ctx, a); err != nil {
		return nil, err
	}
	return a, nil
}

func (u *conexionUseCase) ListarApis(ctx context.Context, limite, offset int) ([]entity.ApiPublicada, int64, error) {
	if limite <= 0 {
		limite = 50
	}
	if offset < 0 {
		offset = 0
	}
	return u.apiRepo.Listar(ctx, limite, offset)
}

func (u *conexionUseCase) ObtenerApi(ctx context.Context, id string) (*entity.ApiPublicada, error) {
	if id == "" {
		return nil, domain.InvalidArgument
	}
	return u.apiRepo.ObtenerPorID(ctx, id)
}

func (u *conexionUseCase) CambiarEstadoApi(ctx context.Context, id string, active bool) (*entity.ApiPublicada, error) {
	a, err := u.apiRepo.ObtenerPorID(ctx, id)
	if err != nil {
		return nil, err
	}
	a.Active = active
	return u.apiRepo.Actualizar(ctx, a)
}

func (u *conexionUseCase) EliminarApi(ctx context.Context, id string) error {
	if id == "" {
		return domain.InvalidArgument
	}
	return u.apiRepo.Eliminar(ctx, id)
}
