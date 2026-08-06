package usecase

import (
	"context"
	"fmt"
	"os"
	"time"

	"DBGS_SOBERANO_BACKEND/internal/application/port"
	"DBGS_SOBERANO_BACKEND/internal/domain"
	"DBGS_SOBERANO_BACKEND/internal/domain/entity"
	"DBGS_SOBERANO_BACKEND/internal/domain/repository"
	jwt "github.com/golang-jwt/jwt/v5"
)

type seguridadUseCase struct {
	seguridadRepo repository.SeguridadRepository
}

func NewSeguridadUseCase(repo repository.SeguridadRepository) port.SeguridadUseCasePort {
	return &seguridadUseCase{
		seguridadRepo: repo,
	}
}

func (u *seguridadUseCase) ObtenerPerfilUsuario(ctx context.Context, username string) (*entity.Usuario, error) {
	if username == "" {
		return nil, domain.ErrDatosInvalidos
	}

	usuario, err := u.seguridadRepo.ObtenerUsuarioPorUsername(ctx, username)
	if err != nil {
		return nil, err
	}

	if !usuario.EstaActivo() {
		return nil, domain.ErrRegistroInactivo
	}

	return usuario, nil
}

func (u *seguridadUseCase) Login(ctx context.Context, username, password string) (*port.LoginResult, error) {
	if username == "" || password == "" {
		return nil, domain.ErrDatosInvalidos
	}

	usuario, err := u.seguridadRepo.AutenticarUsuario(ctx, username, password)
	if err != nil {
		return nil, domain.ErrAccesoNoAutorizado
	}

	if !usuario.EstaActivo() {
		return nil, domain.ErrRegistroInactivo
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": usuario.ID,
		"username": usuario.Username,
		"role": usuario.RolID,
		"exp": time.Now().Add(24 * time.Hour).Unix(),
		"iat": time.Now().Unix(),
	})

	secret := os.Getenv("DBGS_JWT_SECRET")
	if secret == "" {
		secret = "dbgs-dev-secret"
	}

	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return nil, domain.ErrErrorInterno
	}

	return &port.LoginResult{
		AccessToken: tokenString,
		TokenType:   "Bearer",
		ExpiresIn:   86400,
		Username:    usuario.Username,
	}, nil
}

func (u *seguridadUseCase) ValidarToken(ctx context.Context, token string) (*port.TokenValidationResult, error) {
	if token == "" {
		return nil, domain.ErrDatosInvalidos
	}

	secret := os.Getenv("DBGS_JWT_SECRET")
	if secret == "" {
		secret = "dbgs-dev-secret"
	}

	parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil || !parsedToken.Valid {
		return &port.TokenValidationResult{Valid: false}, nil
	}

	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok {
		return &port.TokenValidationResult{Valid: false}, nil
	}

	return &port.TokenValidationResult{
		Valid:     true,
		Username:  fmt.Sprint(claims["username"]),
		UserID:    fmt.Sprint(claims["sub"]),
		Rol:       fmt.Sprint(claims["role"]),
		ExpiresAt: int64(claims["exp"].(float64)),
	}, nil
}

func (u *seguridadUseCase) ValidarAcceso(ctx context.Context, input port.ValidarAccesoInput) (bool, error) {
	if input.Username == "" || input.Permiso == "" {
		return false, domain.ErrDatosInvalidos
	}

	usuario, err := u.seguridadRepo.ObtenerUsuarioPorUsername(ctx, input.Username)
	if err != nil {
		return false, err
	}

	if !usuario.EstaActivo() {
		return false, domain.ErrRegistroInactivo
	}

	return u.seguridadRepo.ValidarPermiso(ctx, usuario.RolID, input.Permiso)
}