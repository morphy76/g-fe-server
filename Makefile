-include deploy.env

# Set binary commands
GO := go
DOCKER := docker
NPM := npm
NODEMON := nodemon

# Set the flags
#GOFLAGS := -mod=vendor
LDFLAGS := -ldflags="-s -w"
GCFLAGS := -gcflags="-m -l"
TESTFLAGS := -v
# DOCKERBUILDFLAGS := --no-cache
NPMFLAGS := --no-audit --no-fund

# Cross-cutting runtime args
OTEL_ARGS := -otel-enabled=false
OIDC_ARGS := -oidc-issuer=http://localhost:28080/realms/gfes -oidc-client-id=ps -oidc-client-secret=BA4eYsij3vDerLdQTRp6khSKWSDQWdLr -oidc-scopes=openid,profile,email,offline_access

# Server

## Define the source files
SERVER_SOURCES := ./cmd/server/serve.go

## Define the target binary name
SERVER_TARGET := g-fe-server
SERVER_TARGET_FE := ./web/build
SERVER_DOCKERFILE := ./tools/docker/Dockerfile.server
SERVER_DEPLOY_TAG ?= g-fe-service:0.0.1

## Runtime args
SERVER_SERVE_ARGS := -ctx=/fe -static=$(TARGET_FE) -host=localhost

# Service (example)
## Define the target binary name
SERVICE_TARGET := g-be-service
SERVICE_DOCKERFILE := ./tools/docker/Dockerfile.service
SERVICE_DEPLOY_TAG ?= g-be-service:0.0.1

## Define the source files
SERVICE_SOURCES := ./cmd/example/example.go

## Runtime args
SERVICE_SERVE_ARGS := -ctx=/be -host=localhost
SERVICE_MONGO_ARGS := -db=1 -db-mongo-url=mongodb://127.0.0.1:27017/go_db -db-mongo-user=go -db-mongo-password=go

build-fe:
	@$(NPM) $(NPMFLAGS) --prefix ./web/ui i
	@$(NPM) --prefix ./web/ui test
	@$(NPM) --prefix ./web/ui run build

build-server:
	@$(GO) test $(TESTFLAGS) ./...
	@$(GO) build $(GOFLAGS) $(LDFLAGS) $(GCFLAGS) -o $(SERVER_TARGET) $(SERVER_SOURCES)

build-service:
	@$(GO) test $(TESTFLAGS) ./...
	@$(GO) build $(GOFLAGS) $(LDFLAGS) $(GCFLAGS) -o $(SERVICE_TARGET) $(SERVICE_SOURCES)

watch-fe:
	@$(NPM) --prefix ./web/ui i
	@$(NPM) --prefix ./web/ui run watch

watch-server:
	@$(NODEMON) --watch './**/*.go' --signal SIGTERM --exec $(GO) run $(GOFLAGS) $(LDFLAGS) $(SERVER_SOURCES) $(SERVER_SERVE_ARGS) -trace $(OTEL_ARGS)

watch-service:
	@$(NODEMON) --watch './**/*.go' --signal SIGTERM --exec $(GO) run $(GOFLAGS) $(LDFLAGS) $(SERVICE_SOURCES) $(SERVICE_SERVE_ARGS) -trace $(OTEL_ARGS)

watch-service-mongo:
	@$(NODEMON) --watch './**/*.go' --signal SIGTERM --exec $(GO) run $(GOFLAGS) $(LDFLAGS) $(SERVICE_SOURCES) $(SERVICE_SERVE_ARGS) -trace $(OTEL_ARGS) $(SERVICE_MONGO_ARGS)

run-server: build-fe
	$(GO) run $(GOFLAGS) $(LDFLAGS) $(GCFLAGS) $(SERVER_SOURCES) $(SERVER_SERVE_ARGS) $(OTEL_ARGS) $(OIDC_ARGS)

run-service:
	$(GO) run $(GOFLAGS) $(LDFLAGS) $(GCFLAGS) $(SERVICE_SOURCES) $(SERVICE_SERVE_ARGS) $(OTEL_ARGS) $(OIDC_ARGS)

run-service-mongo:
	$(GO) run $(GOFLAGS) $(LDFLAGS) $(GCFLAGS) $(SERVICE_SOURCES) $(SERVICE_SERVE_ARGS) $(SERVICE_MONGO_ARGS) $(OTEL_ARGS) $(OIDC_ARGS)

# Define the clean target
clean:
	-@rm -f $(SERVER_TARGET)
	-@rm -f $(SERVICE_TARGET)
	-@rm -rf $(SERVER_TARGET_FE)

deploy: clean
  SERVER_TAG = $(word 1,$(subst :, ,$(SERVER_DEPLOY_TAG)))
  SERVER_VERSION = $(word 2,$(subst :, ,$(SERVER_DEPLOY_TAG)))
  SERVICE_TAG = $(word 1,$(subst :, ,$(SERVICE_DEPLOY_TAG)))
  SERVICE_VERSION = $(word 2,$(subst :, ,$(SERVICE_DEPLOY_TAG)))
	@$(DOCKER) run -d --network host --rm -v /var/run/docker.sock:/var/run/docker.sock --name socat alpine/socat tcp-listen:12345,fork,reuseaddr,ignoreeof unix-connect:/var/run/docker.sock
	-$(DOCKER) build --network host \
    --build-arg TAG_NAME=$(SERVER_TAG) \
    --build-arg TAG_VERSION=$(SERVER_VERSION) \
    -t $(SERVER_DEPLOY_TAG) -f $(SERVER_DOCKERFILE) $(DOCKERBUILDFLAGS) .
	-$(DOCKER) build --network host \
    --build-arg TAG_NAME=$(SERVICE_TAG) \
    --build-arg TAG_VERSION=$(SERVICE_VERSION) \
    -t $(SERVICE_DEPLOY_TAG) -f $(SERVICE_DOCKERFILE) $(DOCKERBUILDFLAGS) .
	@$(DOCKER) stop socat
