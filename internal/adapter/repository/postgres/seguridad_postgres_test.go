package postgres

import (
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestVerificarCredencialConHashValido(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("dbgs-admin"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("GenerateFromPassword() error = %v", err)
	}

	if !verificarCredencial(string(hash), "dbgs-admin") {
		t.Fatal("verificarCredencial() esperaba true para un hash válido")
	}
}

func TestVerificarCredencialConHashInvalido(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("dbgs-admin"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("GenerateFromPassword() error = %v", err)
	}

	if verificarCredencial(string(hash), "password-malo") {
		t.Fatal("verificarCredencial() esperaba false para una contraseña incorrecta")
	}
}
