-include deploy.env

## Set binary commands
GO := go
DOCKER := docker
NPM := npm
NODEMON := nodemon

## Set the flags
GOFLAGS := #-mod=vendor
LDFLAGS := -ldflags="-s -w"
GCFLAGS := -gcflags="-m -l"
TESTFLAGS := -v
# DOCKERBUILDFLAGS := --no-cache
NPMFLAGS := --no-audit --no-fund

## Define the source files
SERVER_SOURCES := ./cmd/serve.go

## Define the target binary name
SERVER_TARGET := g-fe-server
SERVER_TARGET_FE := ./web/build
SERVER_DOCKERFILE := ./tools/docker/Dockerfile.server
SERVER_DEPLOY_TAG ?= g-fe-service:0.0.1
SERVER_TAG = $(word 1,$(subst :, ,$(SERVER_DEPLOY_TAG)))
SERVER_VERSION = $(word 2,$(subst :, ,$(SERVER_DEPLOY_TAG)))

## Define the runtime args
SERVE_ARGS := -ctx=/fe -static=$(SERVER_TARGET_FE) -host=localhost -port=3000
OTEL_ARGS := -otel-enabled=true --otlp-url=http://localhost:4317
OIDC_ARGS := -oidc-issuer=http://localhost:8080/realms/gfes -oidc-client-id=ps -oidc-client-secret=tefnJ7pbekZuTV7vPVpI3VHPNto7LlOy -oidc-scopes=openid,profile,email
MONGO_ARGS := -db-mongo-password=fe_password -db-mongo-user=fe_user -db-mongo-url=mongodb://localhost:27017/fe_db

build-fe:
	@$(NPM) $(NPMFLAGS) --prefix ./web/ui i
	@$(NPM) --prefix ./web/ui test
	@$(NPM) --prefix ./web/ui run build

build-server:
	# @$(GO) test $(TESTFLAGS) ./...
	$(GO) build $(GOFLAGS) $(LDFLAGS) $(GCFLAGS) -o $(SERVER_TARGET) $(SERVER_SOURCES)

watch-fe:
	@$(NPM) --prefix ./web/ui i
	@$(NPM) --prefix ./web/ui run watch

watch-server:
	@$(NODEMON) --watch './**/*.go' --signal SIGTERM --exec $(GO) run $(GOFLAGS) $(LDFLAGS) $(SERVER_SOURCES) $(SERVE_ARGS) $(OTEL_ARGS) $(NO_OIDC_ARGS) $(OIDC_ARGS) $(MONGO_ARGS)

# run-server: build-fe
run-server:
	$(GO) run $(GOFLAGS) $(LDFLAGS) $(GCFLAGS) $(SERVER_SOURCES) $(SERVE_ARGS) $(OTEL_ARGS) $(OIDC_ARGS) $(MONGO_ARGS)

clean:
	-@rm -f $(SERVER_TARGET)
	-@rm -rf $(SERVER_TARGET_FE)

deploy: clean
	@$(DOCKER) run -d --network host --rm -v /var/run/docker.sock:/var/run/docker.sock --name socat alpine/socat tcp-listen:12345,fork,reuseaddr,ignoreeof unix-connect:/var/run/docker.sock
	-$(DOCKER) build --network host \
    --platform linux/amd64 --output type=docker \
    --build-arg TAG_NAME=$(SERVER_TAG) \
    --build-arg TAG_VERSION=$(SERVER_VERSION) \
    -t $(SERVER_DEPLOY_TAG) -f $(SERVER_DOCKERFILE) $(DOCKERBUILDFLAGS) .
	@$(DOCKER) stop socat

run-docker:
	$(DOCKER) run --rm --network host --platform linux/amd64 --name gfe \
    -e CONTEXT_ROOT=/fe -e STATIC_PATH=$(SERVER_TARGET_FE) -e SERVE_HOST=localhost -e SERVE_PORT=3000 \
    -e OTEL_ENABLED=true -e OTLP_URL=http://localhost:4317 \
    -e OIDC_ISSUER=http://localhost:8080/realms/gfes -e OIDC_CLIENT_ID=ps -e OIDC_CLIENT_SECRET=tefnJ7pbekZuTV7vPVpI3VHPNto7LlOy -e OIDC_SCOPES=openid,profile,email \
    -e DB_MONGO_PASSWORD=fe_password -e DB_MONGO_USER=fe_user -e DB_MONGO_URL=mongodb://localhost:27017/fe_db \
    $(SERVER_DEPLOY_TAG)
