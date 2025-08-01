# G-FE-Server Makefile

-include deploy.env

# Default target
.DEFAULT_GOAL := help

# Declare phony targets
.PHONY: help clean test-server vet build-server watch-server run-server run-ssl-server build-fe watch-fe deploy run-docker start-deps stop-deps

## Binary commands
GO := go
DOCKER := docker
NPM := npm
NODEMON := nodemon

## Build flags
GOFLAGS := #-mod=vendor
LDFLAGS := -ldflags="-s -w"
GCFLAGS := -gcflags="-m -l"
TESTFLAGS := -v -count=1 -timeout=2s
# DOCKERBUILDFLAGS := --no-cache
NPMFLAGS := --no-audit --no-fund

## Project configuration
SERVER_SOURCES := ./cmd/serve.go
SERVER_TARGET := g-fe-server
SERVER_TARGET_FE := ./web/ui/dist
SERVER_DOCKERFILE := ./tools/docker/Dockerfile.server
SERVER_DEPLOY_TAG ?= g-fe-service:0.0.1
SERVER_TAG = $(word 1,$(subst :, ,$(SERVER_DEPLOY_TAG)))
SERVER_VERSION = $(word 2,$(subst :, ,$(SERVER_DEPLOY_TAG)))

## Runtime arguments
SERVE_ARGS := -ctx=/fe -static=$(SERVER_TARGET_FE) -host=localhost -port=3000 -session-key="my secure session key" -session-secure=true -session-same-site=Strict
OTEL_ARGS := -otel-enabled=true --otlp-url=http://localhost:4317
OIDC_ARGS := -oidc-issuer=http://localhost:8080/realms/gfes -oidc-client-id=ps -oidc-client-secret=DTntfqG8c9Hn77zBEQkIfuPQbdOvsWcw -oidc-scopes=openid,profile,email
MONGO_ARGS := -db-mongo-password=fe_password -db-mongo-user=fe_user -db-mongo-url=mongodb://localhost:27017/fe_db?w=1
UNLEASH_ARGS := -unleash-enabled=true -unleash-url=http://localhost:3063/api -unleash-app-name=fe-server -unleash-token=default:development.f9e56e74a070c76b577840b2adb2ca195d394a2c3bd8915a93e6d617 -unleash-environment=production
AIW_ARGS := -aiw-fqdn=http://localhost:3000/fe

# ==============================================================================
# HELP TARGET
# ==============================================================================

help: ## Display this help message
	@echo "G-FE-Server Build System"
	@echo "========================"
	@echo ""
	@echo "Available targets:"
	@echo ""
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)
	@echo ""
	@echo "Examples:"
	@echo "  make build-server    # Build the Go server binary"
	@echo "  make run-server      # Run the server with all dependencies"
	@echo "  make start-deps      # Start Docker dependencies"
	@echo ""

# ==============================================================================
# DEVELOPMENT TARGETS
# ==============================================================================

clean: ## Remove build artifacts and clean workspace
	-@rm -f $(SERVER_TARGET)
	-@rm -rf $(SERVER_TARGET_FE)

test-server: ## Run Go tests for server components
	@$(GO) test $(TESTFLAGS) $(shell $(GO) list ./... | grep -vE '/tools/|/web/')

vet: ## Run Go vet for static analysis
	@$(GO) vet $(shell $(GO) list ./... | grep -vE '/tools/|/web/')

# ==============================================================================
# BUILD TARGETS
# ==============================================================================

build-server: ## Build the Go server binary
	$(GO) build $(GOFLAGS) $(LDFLAGS) $(GCFLAGS) -o $(SERVER_TARGET) $(SERVER_SOURCES)

build-fe: ## Build the frontend application
	@$(NPM) $(NPMFLAGS) --prefix ./web/ui i
	@$(NPM) --prefix ./web/ui test
	@$(NPM) --prefix ./web/ui run build

# ==============================================================================
# DEVELOPMENT TARGETS
# ==============================================================================

watch-server: ## Run server in watch mode with auto-reload
	@$(NODEMON) --watch './**/*.go' --signal SIGTERM --exec $(GO) run $(GOFLAGS) $(LDFLAGS) $(SERVER_SOURCES) $(SERVE_ARGS) $(OTEL_ARGS) $(NO_OIDC_ARGS) $(OIDC_ARGS) $(MONGO_ARGS) $(UNLEASH_ARGS) $(AIW_ARGS)

watch-fe: ## Run frontend in watch mode with auto-reload
	@$(NPM) --prefix ./web/ui i
	@$(NPM) --prefix ./web/ui run watch

run-server: ## Run the server with all configured services
	$(GO) run $(GOFLAGS) $(LDFLAGS) $(GCFLAGS) $(SERVER_SOURCES) $(SERVE_ARGS) $(OTEL_ARGS) $(OIDC_ARGS) $(MONGO_ARGS) $(UNLEASH_ARGS) $(AIW_ARGS)

run-ssl-server: ## Run the server with SSL/TLS enabled
	@TMPDIR=$$(mktemp -d) && \
	openssl req -x509 -nodes -days 1 -newkey rsa:2048 \
    -keyout $$TMPDIR/server.key -out $$TMPDIR/server.crt \
    -subj "/CN=localhost" && \
	$(GO) run $(GOFLAGS) $(LDFLAGS) $(GCFLAGS) $(SERVER_SOURCES) \
    $(SERVE_ARGS) $(OTEL_ARGS) $(OIDC_ARGS) $(MONGO_ARGS) $(UNLEASH_ARGS) $(AIW_ARGS) \
    -protocol=https -tls-cert=$$TMPDIR/server.crt -tls-key=$$TMPDIR/server.key

# ==============================================================================
# DOCKER TARGETS
# ==============================================================================

deploy: clean ## Build and deploy Docker image
	@$(DOCKER) run -d --network host --rm -v /var/run/docker.sock:/var/run/docker.sock --name socat alpine/socat tcp-listen:12345,fork,reuseaddr,ignoreeof unix-connect:/var/run/docker.sock
	-$(DOCKER) build --network host \
    --platform linux/amd64 --output type=docker \
    --build-arg TAG_NAME=$(SERVER_TAG) \
    --build-arg TAG_VERSION=$(SERVER_VERSION) \
    -t $(SERVER_DEPLOY_TAG) -f $(SERVER_DOCKERFILE) $(DOCKERBUILDFLAGS) .
	@$(DOCKER) stop socat

run-docker: ## Run the application in Docker container
	$(DOCKER) run --rm --network host --platform linux/amd64 --name gfe \
    -e CONTEXT_ROOT=/fe -e STATIC_PATH=$(SERVER_TARGET_FE) -e SERVE_HOST=localhost -e SERVE_PORT=3000 \
    -e SESSION_KEY="my secure session key" -e SESSION_SECURE=false \
    -e OTEL_ENABLED=true -e OTLP_URL=http://localhost:4317 \
    -e OIDC_ISSUER=http://localhost:8080/realms/gfes -e OIDC_CLIENT_ID=ps -e OIDC_CLIENT_SECRET=tefnJ7pbekZuTV7vPVpI3VHPNto7LlOy -e OIDC_SCOPES=openid,profile,email \
    -e DB_MONGO_PASSWORD=fe_password -e DB_MONGO_USER=fe_user -e DB_MONGO_URL=mongodb://localhost:27017/fe_db?w=1 \
    -e UNLEASH_ENABLED=true -e UNLEASH_URL=http://localhost:4242/api -e UNLEASH_APP_NAME=fe-server -e UNLEASH_TOKEN=default:development.f9e56e74a070c76b577840b2adb2ca195d394a2c3bd8915a93e6d617 \
     -e AIW_FQDN=http://localhost:3000/fe \
    $(SERVER_DEPLOY_TAG)

# ==============================================================================
# DEPENDENCY MANAGEMENT
# ==============================================================================

start-deps: ## Start all external dependencies via Docker Compose
	@echo "Starting dependencies using Docker Compose..."
	@$(DOCKER) compose -p gfe -f ./tools/compose/docker-compose.yml up -d

stop-deps: ## Stop all external dependencies
	@echo "Stopping dependencies..."
	@$(DOCKER) compose -p gfe -f ./tools/compose/docker-compose.yml down
