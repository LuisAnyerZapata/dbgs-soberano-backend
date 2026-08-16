# ==============================================================================
# DBGS_SOBERANO_BACKEND - Automation Makefile
# ==============================================================================

APP_NAME := dbgs_soberano_backend
MAIN_PACKAGE := ./cmd/server/main.go
BUILD_DIR := ./bin
PROTO_DIR := ./api/proto/v1
THIRD_PARTY_DIR := ./api/proto/third_party
GOOGLE_APIS_DIR := $(THIRD_PARTY_DIR)/google/api
MIGRATIONS_DIR := ./db/migrations

GOPATH := $(shell go env GOPATH)
export PATH := $(GOPATH)/bin:$(PATH)

CONFIG_JSON := ./config/config.json
DB_USER ?= $(or $(shell python3 -c 'import json; data=json.load(open("$(CONFIG_JSON)")); print(data.get("database", {}).get("user", "postgres"))' 2>/dev/null), postgres)
DB_PASSWORD ?= $(or $(shell python3 -c 'import json; data=json.load(open("$(CONFIG_JSON)")); print(data.get("database", {}).get("password", "postgres"))' 2>/dev/null), postgres)
DB_HOST ?= $(or $(shell python3 -c 'import json; data=json.load(open("$(CONFIG_JSON)")); print(data.get("database", {}).get("host", "localhost"))' 2>/dev/null), localhost)
DB_PORT ?= $(or $(shell python3 -c 'import json; data=json.load(open("$(CONFIG_JSON)")); print(data.get("database", {}).get("port", "5432"))' 2>/dev/null), 5432)
DB_NAME ?= $(or $(shell python3 -c 'import json; data=json.load(open("$(CONFIG_JSON)")); print(data.get("database", {}).get("name", "dbgs_soberano"))' 2>/dev/null), dbgs_soberano)
DB_SSLMODE ?= $(or $(shell python3 -c 'import json; data=json.load(open("$(CONFIG_JSON)")); print(data.get("database", {}).get("ssl_mode", "disable"))' 2>/dev/null), disable)
DATABASE_URL := "postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSLMODE)"

.PHONY: help setup proto-deps proto proto-clean clean build run dev test test-coverage fmt lint deps install-migrate migrate-up migrate-down seed backup restore

.DEFAULT_GOAL := help

## help: Muestra los comandos disponibles con sus descripciones
help:
	@echo "Uso: make [target]"
	@echo ""
	@echo "Comandos disponibles:"
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' | sed -e 's/^/ /'

# ==============================================================================
# CONFIGURACIÓN INICIAL DEL ENTORNO
# ==============================================================================

## setup: Instala todas las herramientas necesarias de Go para compilar .proto (incluyendo gRPC-Gateway)
setup: proto-deps
	@echo "==> Instalando plugins de Go para Protobuf..."
	@go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	@go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	@go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest
	@echo "==> Herramientas instaladas correctamente en $(GOPATH)/bin"

## proto-deps: Descarga las dependencias de Google API necesarias para las anotaciones REST
proto-deps:
	@mkdir -p $(GOOGLE_APIS_DIR)
	@if [ ! -f $(GOOGLE_APIS_DIR)/annotations.proto ]; then \
		echo "==> Descargando dependencias de Google API para anotaciones HTTP..."; \
		curl -sSL https://raw.githubusercontent.com/googleapis/googleapis/master/google/api/annotations.proto -o $(GOOGLE_APIS_DIR)/annotations.proto; \
		curl -sSL https://raw.githubusercontent.com/googleapis/googleapis/master/google/api/http.proto -o $(GOOGLE_APIS_DIR)/http.proto; \
		echo "==> Dependencias descargadas en $(THIRD_PARTY_DIR)"; \
	else \
		echo "==> Dependencias de Google API ya existen, omitiendo descarga."; \
	fi

# ==============================================================================
# PROTOBUF / gRPC / REST-GATEWAY
# ==============================================================================

## proto: Compila todos los archivos .proto a código Go (Stubs gRPC + Traductores REST)
proto: proto-deps
	@echo "==> Verificando plugins de Protobuf..."
	@command -v protoc-gen-go >/dev/null 2>&1 || go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	@command -v protoc-gen-go-grpc >/dev/null 2>&1 || go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	@command -v protoc-gen-grpc-gateway >/dev/null 2>&1 || go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest
	@echo "==> Compilando archivos Protobuf (.proto)..."
	@protoc \
		-I$(PROTO_DIR) \
		-I$(THIRD_PARTY_DIR) \
		--go_out=$(PROTO_DIR) --go_opt=paths=source_relative \
		--go-grpc_out=$(PROTO_DIR) --go-grpc_opt=paths=source_relative \
		--grpc-gateway_out=$(PROTO_DIR) --grpc-gateway_opt=paths=source_relative \
		$(PROTO_DIR)/*.proto
	@echo "==> Archivos .pb.go generados exitosamente dentro de $(PROTO_DIR)"

## proto-clean: Elimina los archivos autogenerados de Protobuf y las dependencias de terceros
proto-clean:
	@echo "==> Limpiando archivos autogenerados de Protobuf..."
	@rm -f $(PROTO_DIR)/*.pb.go
	@rm -rf $(THIRD_PARTY_DIR)

# ==============================================================================
# CONSTRUCCIÓN Y EJECUCIÓN
# ==============================================================================

## build: Compila el binario ejecutable en ./bin/
build: proto
	@echo "==> Compilando binario Go..."
	@mkdir -p $(BUILD_DIR)
	@go build -ldflags="-w -s" -o $(BUILD_DIR)/$(APP_NAME) $(MAIN_PACKAGE)
	@echo "==> Binario compilado en $(BUILD_DIR)/$(APP_NAME)"

## run: Compila y ejecuta el servidor de producción
run: build
	@echo "==> Iniciando servidor..."
	@$(BUILD_DIR)/$(APP_NAME)

## dev: Ejecuta el servidor en modo desarrollo
dev:
	@if command -v air > /dev/null; then \
		air; \
	else \
		echo "==> 'air' no está instalado. Ejecutando con 'go run'..."; \
		go run $(MAIN_PACKAGE); \
	fi

## clean: Limpia binarios y artefactos temporales
clean: proto-clean
	@echo "==> Limpiando directorio de construcción..."
	@rm -rf $(BUILD_DIR)
	@go clean

# ==============================================================================
# CALIDAD DE CÓDIGO Y PRUEBAS
# ==============================================================================

## test: Ejecuta todas las pruebas unitarias e integración
test:
	@echo "==> Ejecutando pruebas..."
	@go test -v -race ./...

## test-coverage: Ejecuta pruebas y genera reporte de cobertura HTML
test-coverage:
	@echo "==> Calculando cobertura de pruebas..."
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "==> Reporte generado en coverage.html"

## fmt: Formatea todo el código Go según los estándares oficiales
fmt:
	@echo "==> Formateando código con gofmt..."
	@go fmt ./...

## lint: Ejecuta el linter de Go
lint:
	@echo "==> Verificando calidad de código con golangci-lint..."
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run ./...; \
	else \
		echo "ERROR: golangci-lint no está instalado."; \
	fi

## deps: Sincroniza y verifica las dependencias del go.mod
deps:
	@echo "==> Descargando y limpiando dependencias..."
	@go mod download
	@go mod tidy

## install-migrate: Instala golang-migrate en $(GOPATH)/bin
install-migrate:
	@echo "==> Instalando golang-migrate..."
	@go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	@echo "==> golang-migrate instalado correctamente en $(GOPATH)/bin"

# ==============================================================================
# BASE DE DATOS Y MIGRACIONES
# ==============================================================================

## migrate-up: Aplica todas las migraciones SQL pendientes hacia adelante
migrate-up: install-migrate
	@echo "==> Ejecutando migraciones SQL (UP)..."
	@migrate -path $(MIGRATIONS_DIR) -database $(DATABASE_URL) up

## migrate-down: Revierte la última migración SQL aplicada
migrate-down: install-migrate
	@echo "==> Revirtiendo última migración SQL (DOWN)..."
	@migrate -path $(MIGRATIONS_DIR) -database $(DATABASE_URL) down 1

## seed: Carga los datos de prueba y catálogos de referencia iniciales
seed:
	@echo "==> Insertando datos iniciales y semillas (Seeds)..."
	@PGPASSWORD=$(DB_PASSWORD) psql -v ON_ERROR_STOP=1 -h $(DB_HOST) -p $(DB_PORT) -U $(DB_USER) -d $(DB_NAME) -f db/seeds/initial_data.sql
	@PGPASSWORD=$(DB_PASSWORD) psql -v ON_ERROR_STOP=1 -h $(DB_HOST) -p $(DB_PORT) -U $(DB_USER) -d $(DB_NAME) -f db/seeds/01_catalogos_referencia.sql
	@PGPASSWORD=$(DB_PASSWORD) psql -v ON_ERROR_STOP=1 -h $(DB_HOST) -p $(DB_PORT) -U $(DB_USER) -d $(DB_NAME) -f db/seeds/02_datos_prueba_sinteticos.sql
	@PGPASSWORD=$(DB_PASSWORD) psql -v ON_ERROR_STOP=1 -h $(DB_HOST) -p $(DB_PORT) -U $(DB_USER) -d $(DB_NAME) -f db/seeds/seed_users.sql
	@echo "==> Semillas cargadas correctamente."

## backup: Genera un respaldo .dump de la base de datos
backup:
	@echo "==> Creando respaldo de la base de datos..."
	@bash ./db/backup/backup_dbgs.sh

## restore: Restaura una copia de seguridad .dump. Uso: make restore BACKUP_FILE=path/to/file.dump
restore:
	@if [ -z "$(BACKUP_FILE)" ]; then \
		echo "ERROR: Debes indicar el archivo de respaldo. Ejemplo: make restore BACKUP_FILE=db/backup/dumps/backup.dump"; \
	exit 1; \
	fi
	@echo "==> Restaurando respaldo $(BACKUP_FILE)..."
	@AUTO_APPROVE=1 bash ./db/backup/restore_dbgs.sh "$(BACKUP_FILE)"